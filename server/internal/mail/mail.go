package mail

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"lath/xman/internal/db"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/smtp"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

type Attachment struct {
	Filename    string
	ContentType string
	Body        []byte
}

func SendMailNewMessageNotification(to string, agencyName string, message db.Message) error {
	var messageType string
	switch message.MessageType {
	case "0501":
		messageType = "Anbietung"
	case "0503":
		messageType = "Abgabe"
	case "0505":
		messageType = "Bewertungsbestätigung"
	default:
		panic("unhandled message type: " + message.MessageType)
	}
	body := "<p>Es ist eine neue " + messageType + " von \"" + agencyName + "\" eingegangen.</p>\n"
	origin := os.Getenv("ORIGIN")
	if message.MessageType == "0501" || message.MessageType == "0503" {
		url := origin + "/nachricht/" + message.MessageHead.ProcessID.String() + "/" + string(message.MessageType)
		body += fmt.Sprintf("<p>Der Inhalt steht unter folgendem Link zur Verfügung: <a href=\"%s\">%s</a></p>\n", url, url)
	}
	body += "<p>Diese E-Mail stellt eine reine Benachrichtigung dar und sollte nicht veraktet werden.</p>\n"
	body += "<p>Sie bekommen diese E-Mail, weil Sie der abgebenden Stelle als zuständige(r) Archivar(in) zugeordnet sind.<br>\n" +
		fmt.Sprintf("Sie können Ihre Einstellungen für E-Mail-Benachrichtigungen unter <a href=\"%s\">%s</a> ändern.</p>", origin, origin)
	return sendMail(to, "Neue "+messageType+" von "+agencyName, body, []Attachment{})
}

func SendMailNewMessagePostOffice(to string, agencyName string, message db.Message) error {
	var messageType string
	switch message.MessageType {
	case "0501":
		messageType = "Anbietung"
	case "0503":
		messageType = "Abgabe"
	case "0505":
		messageType = "Bewertungsbestätigung"
	default:
		panic("unhandled message type: " + message.MessageType)
	}
	body := "<p>Es ist eine " + messageType + " von \"" + agencyName + "\" eingegangen.</p>\n"
	body += "<p>Im Anhang befindet sich die XML-Datei der xdomea-Nachricht.</p>\n"
	body += "<p>Diese E-Mail ist an die Poststelle adressiert und zur Veraktung vorgesehen.</p>\n"
	messageXML, err := os.ReadFile(message.MessagePath)
	if err != nil {
		panic(err)
	}
	return sendMail(to, messageType+" von "+agencyName, body, []Attachment{{
		Filename:    path.Base(message.MessagePath),
		ContentType: "application/xml",
		Body:        messageXML,
	}})
}

func SendMailReport(to string, process db.SubmissionProcess, report Attachment) error {
	agencyName := process.Agency.Name
	body := "<p>Die Abgabe von " + agencyName + " wurde erfolgreich archiviert.</p>"
	origin := os.Getenv("ORIGIN")
	body += "<p>Sie bekommen diese E-Mail, weil Sie die Archivierung der Aussonderung abgeschlossen haben.<br>\n" +
		fmt.Sprintf("Sie können Ihre Einstellungen für E-Mail-Benachrichtigungen unter <a href=\"%s\">%s</a> ändern.</p>", origin, origin)
	return sendMail(to, "Übernahmebericht für Abgabe von "+agencyName, body, []Attachment{report})
}

func SendMailProcessingError(to string, e db.ProcessingError) error {
	origin := os.Getenv("ORIGIN")
	message := "<p>Ein Fehler wurde in der Steuerungsstelle eingetragen.</p>\n"
	message += fmt.Sprintf("<p><strong>%s</strong></p>\n", e.Title)
	if e.ProcessID != uuid.Nil {
		url := origin + "/nachricht/" + e.ProcessID.String()
		message += fmt.Sprintf("<p>Nachricht: <a href=\"%s\">%s</a></p>\n", url, url)
	}
	if e.Info != "" {
		message += fmt.Sprintf("<p>%s</p>", strings.ReplaceAll(e.Info, "\n", "\n<br>"))
	}
	message += fmt.Sprintf("<p>Sie bekommen diese E-Mail, weil Sie sich als Administrator für Benachrichtigungen für Fehler eingetragen haben.<br>\n"+
		"Sie können Ihre Einstellungen für E-Mail-Benachrichtigungen unter <a href=\"%s\">%s</a> ändern.</p>",
		origin, origin,
	)
	return sendMail(to, "Fehler in Steuerungsstelle: "+e.Title, message, []Attachment{})
}

func sendMail(to, subject, body string, attachments []Attachment) error {
	addr := os.Getenv("SMTP_SERVER")
	if addr == "" {
		return nil
	}
	log.Println("Sending e-mail to " + to)
	from := os.Getenv("SMTP_FROM_EMAIL")
	content := getContent(to, from, subject, body, attachments)
	tlsMode := os.Getenv("SMTP_TLS_MODE")
	username := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")
	err := sendMailInner(addr, from, to, content, tlsMode, username, password)
	if err != nil {
		return fmt.Errorf("failed to send mail to %s: %w", to, err)
	}
	return nil
}

func getContent(to, from, subject, body string, attachments []Attachment) []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf(
		"From: X-MAN <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n",
		from, to, mime.QEncoding.Encode("utf-8", subject)))
	buf.WriteString("MIME-Version: 1.0\r\n")
	writer := multipart.NewWriter(buf)
	boundary := writer.Boundary()
	if len(attachments) > 0 {
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", boundary))
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	}
	buf.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	if len(attachments) > 0 {
		buf.WriteString("Content-Disposition: inline\r\n")
	}
	buf.WriteString("\r\n")
	buf.WriteString(encodeCRLF(body))
	for _, a := range attachments {
		buf.WriteString(fmt.Sprintf("\r\n\r\n--%s\r\n", boundary))
		buf.WriteString(fmt.Sprintf("Content-Type: %s\r\n", a.ContentType))
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%s\r\n\r\n", mime.QEncoding.Encode("utf-8", a.Filename)))
		b := make([]byte, base64.StdEncoding.EncodedLen(len(a.Body)))
		base64.StdEncoding.Encode(b, a.Body)
		buf.Write(b)
	}
	if len(attachments) > 0 {
		buf.WriteString(fmt.Sprintf("\r\n\r\n--%s--", boundary))
	}
	return buf.Bytes()
}

func encodeCRLF(input string) string {
	re := regexp.MustCompile(`\r?\n`)
	return re.ReplaceAllString(input, "\r\n")
}

// sendMailInner sends an email with the given configuration.
//
// This is largely equivalent to smtp.SendMail, but we have an explicit option
// to force or disable StartTLS independently of the extensions announced by the
// server.
func sendMailInner(
	addr, from, to string,
	content []byte,
	tlsMode string,
	username, password string,
) error {
	c, err := connect(addr, tlsMode, username, password)
	if err != nil {
		return err
	}
	defer c.Close()

	// Set the sender and recipient first
	if err := c.Mail(from); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}

	// Send the email body.
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(content)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	// Send the QUIT command and close the connection.
	return c.Quit()
}

// TestConnection tries to connect with the configured SMTP server.
func TestConnection() error {
	addr := os.Getenv("SMTP_SERVER")
	tlsMode := os.Getenv("SMTP_TLS_MODE")
	if tlsMode == "" {
		return errors.New("SMTP_TLS_MODE not set")
	} else if tlsMode != "off" && tlsMode != "tls" && tlsMode != "starttls" {
		return fmt.Errorf("invalid value for SMTP_TLS_MODE: %s", tlsMode)
	}
	username := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")

	c, err := connect(addr, tlsMode, username, password)
	if err != nil {
		return err
	}
	defer c.Close()
	err = c.Quit()
	return err
}

func connect(
	addr string,
	tlsMode string,
	username, password string,
) (*smtp.Client, error) {
	var c *smtp.Client
	host, _, _ := net.SplitHostPort(addr)

	// Connect to the remote SMTP server.
	if tlsMode == "tls" {
		tlsConfig := &tls.Config{
			ServerName: host,
		}
		tlsConn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return nil, err
		}
		c, err = smtp.NewClient(tlsConn, host)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		c, err = smtp.Dial(addr)
		if err != nil {
			return nil, err
		}
	}

	if tlsMode == "starttls" {
		if ok, _ := c.Extension("STARTTLS"); ok {
			config := &tls.Config{ServerName: host}
			if err := c.StartTLS(config); err != nil {
				c.Close()
				return nil, err
			}
		} else {
			c.Close()
			return nil, errors.New("server doesn't support STARTTLS")
		}
	}

	if username != "" {
		auth := smtp.PlainAuth("", username, password, host)
		if ok, _ := c.Extension("AUTH"); !ok {
			c.Close()
			return nil, errors.New("server doesn't support AUTH")
		}
		if err := c.Auth(auth); err != nil {
			c.Close()
			return nil, err
		}
	}
	return c, nil
}

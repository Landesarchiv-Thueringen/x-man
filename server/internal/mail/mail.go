package mail

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"mime"
	"net"
	"net/smtp"
	"os"
	"regexp"
)

func SendMailNew0501(to string, agencyName string, messageID string) {
	url := os.Getenv("ORIGIN") + "/nachricht/" + messageID
	sendMail(to, "Neue Anbietung von "+agencyName, fmt.Sprintf(
		"<p>Es ist eine neue Anbietung von \"%s\" eingegangen.</p>"+
			"<p>Der Inhalt steht unter folgendem Link zur Verfügung: <a href=\"%s\">%s</a></p>",
		agencyName, url, url,
	))
}

func SendMailProcessingError(to string, agencyName string, messageID string) {
	url := os.Getenv("ORIGIN") + "/nachricht/" + messageID
	sendMail(to, "Neue Anbietung von "+agencyName, fmt.Sprintf(
		"<p>Es ist eine neue Anbietung von \"%s\" eingegangen.</p>"+
			"<p>Der Inhalt steht unter folgendem Link zur Verfügung: <a href=\"%s\">%s</a></p>",
		agencyName, url, url,
	))
}

func sendMail(to, subject, body string) {
	addr := os.Getenv("SMTP_SERVER")
	if addr == "" {
		log.Println("Not sending e-mail since SMTP_SERVER is not configured")
		return
	} else {
		log.Println("Sending e-mail to " + to)
	}
	from := os.Getenv("SMTP_FROM_EMAIL")
	content := fmt.Sprintf(
		"From: X-MAN <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: text/html; charset=utf-8\r\n"+
			"\r\n%s",
		from, to,
		mime.QEncoding.Encode("utf-8", subject),
		encodeCRLF(body),
	)
	useStartTLS := os.Getenv("SMTP_STARTTLS") != "false"
	username := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")
	err := sendMailInner(addr, from, to, content, useStartTLS, username, password)
	if err != nil {
		log.Printf("Error sending e-mail: %v\n", err)
	}
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
	addr, from, to, content string,
	useStartTLS bool,
	username, password string,
) error {
	// Connect to the remote SMTP server.
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}

	host, _, _ := net.SplitHostPort(addr)
	if useStartTLS {
		if ok, _ := c.Extension("STARTTLS"); ok {
			config := &tls.Config{ServerName: host}
			if err = c.StartTLS(config); err != nil {
				return err
			}
		} else {
			return errors.New("server doesn't support STARTTLS")
		}
	}

	if username != "" {
		auth := smtp.PlainAuth("", username, password, host)
		if ok, _ := c.Extension("AUTH"); !ok {
			return errors.New("server doesn't support AUTH")
		}
		if err = c.Auth(auth); err != nil {
			return err
		}
	}

	// Set the sender and recipient first
	if err := c.Mail(from); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(wc, content)
	if err != nil {
		return err
	}
	err = wc.Close()
	if err != nil {
		return err
	}

	// Send the QUIT command and close the connection.
	err = c.Quit()
	if err != nil {
		return err
	}
	return nil
}

package dimag

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"lath/xman/internal/archive"
	"lath/xman/internal/db"
	"lath/xman/internal/xdomea"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const PortSFTP uint = 22

var sshClient *ssh.Client
var sftpClient *sftp.Client

func InitConnection() error {
	urlString := os.Getenv("DIMAG_SFTP_SERVER_URL")
	if urlString == "" {
		errorMessage := "DIMAG SFTP server URL not set"
		log.Println(errorMessage)
		return errors.New(errorMessage)
	}
	url, err := url.Parse(urlString)
	if err != nil {
		errorMessage := "could't parse dimag SFTP server URl"
		log.Println(errorMessage)
		return errors.New(errorMessage)
	}
	sftpUser := os.Getenv("DIMAG_SFTP_USER")
	if sftpUser == "" {
		errorMessage := "DIMAG SFTP user not set"
		log.Println(errorMessage)
		return errors.New(errorMessage)
	}
	// empty password is possible
	sftpPassword := os.Getenv("DIMAG_SFTP_PASSWORD")
	var auths []ssh.AuthMethod
	if sftpPassword != "" {
		auths = append(auths, ssh.Password(sftpPassword))
	}
	config := ssh.ClientConfig{
		User: sftpUser,
		Auth: auths,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			hostKeyString := os.Getenv("DIMAG_SFTP_HOST_KEY")
			splitHostKey := strings.Split(hostKeyString, " ")
			if len(splitHostKey) == 2 &&
				splitHostKey[0] == key.Type() &&
				splitHostKey[1] == base64.StdEncoding.EncodeToString(key.Marshal()) {
				return nil
			}
			return fmt.Errorf("failed to verify host key.\n\n"+
				"If you have connected to %s in the past, this could mean that someone is messing with your connection and tries to steal secrets!\n\n"+
				"If you are trying to connect to %s for the first time or changed the server's SSH keys, add the following line to your .env file and run the action again:\n\n"+
				"DIMAG_SFTP_HOST_KEY=\"%s %s\"",
				hostname, hostname,
				key.Type(), base64.StdEncoding.EncodeToString(key.Marshal()))
		},
	}
	addr := fmt.Sprintf("%s:%d", url.Host, PortSFTP)
	sshClient, err = ssh.Dial("tcp", addr, &config)
	if err != nil {
		log.Println(err)
		return err
	}
	sftpClient, err = sftp.NewClient(sshClient)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func CloseConnection() {
	sshClient.Close()
	sftpClient.Close()
}

func uploadArchivePackage(
	sftpClient *sftp.Client,
	process db.SubmissionProcess,
	message db.Message,
	archivePackage db.ArchivePackage,
) (string, error) {
	uploadDir := os.Getenv("DIMAG_SFTP_UPLOAD_DIR")
	importDir := "xman_import_" + uuid.NewString()
	importPath := filepath.Join(uploadDir, importDir)
	err := sftpClient.Mkdir(importPath)
	if err != nil {
		log.Println("sftpClient.Mkdir", err)
		return importDir, err
	}
	err = uploadXdomeaMessageFile(sftpClient, message, importPath, archivePackage)
	if err != nil {
		return importDir, err
	}
	err = uploadProtocol(sftpClient, process, importPath)
	if err != nil {
		return importDir, err
	}
	for _, primaryDocument := range archivePackage.PrimaryDocuments {
		filePath := filepath.Join(message.StoreDir, primaryDocument.Filename)
		_, err := os.Stat(filePath)
		if err != nil {
			log.Println("os.Stat(filePath)", err, filePath)
			return importDir, err
		}
		remotePath := filepath.Join(importPath, primaryDocument.Filename)
		err = uploadFile(sftpClient, filePath, remotePath)
		if err != nil {
			return importDir, err
		}
	}
	return importDir, uploadControlFile(sftpClient, message, archivePackage, importPath, importDir)
}

func uploadXdomeaMessageFile(
	sftpClient *sftp.Client,
	message db.Message,
	importPath string,
	archivePackage db.ArchivePackage,
) error {
	remotePath := getRemoteXmlPath(message, importPath)
	prunedMessage, err := xdomea.PruneMessage(message, archivePackage)
	if err != nil {
		return err
	}
	return createRemoteTextFile(sftpClient, prunedMessage, remotePath)
}

func uploadProtocol(sftpClient *sftp.Client, process db.SubmissionProcess, importPath string) error {
	remotePath := filepath.Join(importPath, archive.ProtocolFilename)
	protocol := archive.GenerateProtocol(process)
	return createRemoteTextFile(sftpClient, protocol, remotePath)
}

func uploadControlFile(
	sftpClient *sftp.Client,
	message db.Message,
	archivePackageData db.ArchivePackage,
	importPath string,
	importDir string,
) error {
	remotePath := filepath.Join(importPath, ControlFileName)
	controlFileXml := GenerateControlFile(message, archivePackageData, importDir)
	err := createRemoteTextFile(sftpClient, controlFileXml, remotePath)
	return err
}

func uploadFile(sftpClient *sftp.Client, localPath string, remotePath string) error {
	srcFile, err := os.Open(localPath)
	if err != nil {
		log.Println("os.Open(localPath)", err, localPath)
		return err
	}
	defer srcFile.Close()
	// the remote path must already exist
	dstFile, err := sftpClient.OpenFile(remotePath, (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
	if err != nil {
		log.Println("sftpClient.OpenFile", err, remotePath)
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		log.Println("io.Copy", err, dstFile, srcFile)
		return err
	}
	return nil
}

func createRemoteTextFile(sftpClient *sftp.Client, fileContent string, remotePath string) error {
	stringReader := strings.NewReader(fileContent)
	// the remote path must already exist
	dstFile, err := sftpClient.OpenFile(remotePath, (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
	if err != nil {
		log.Println(err)
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, stringReader)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

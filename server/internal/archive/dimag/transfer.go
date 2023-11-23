package dimag

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const PortSFTP = 22

func TransferToArchive(messageID uuid.UUID) {
	urlString := os.Getenv("DIMAG_SFTP_SERVER_URL")
	if urlString == "" {
		log.Fatal("DIMAG SFTP server URL not set")
	}
	url, err := url.Parse(urlString)
	if err != nil {
		log.Fatal("could't parse dimag SFTP server URl")
	}
	sftpUser := os.Getenv("DIMAG_SFTP_USER")
	if sftpUser == "" {
		log.Fatal("DIMAG SFTP user not set")
	}
	// empty password is possible
	sftpPassword := os.Getenv("DIMAG_SFTP_PASSWORD")
	var auths []ssh.AuthMethod
	if sftpPassword != "" {
		auths = append(auths, ssh.Password(sftpPassword))
	}
	config := ssh.ClientConfig{
		User:            sftpUser,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	addr := fmt.Sprintf("%s:%d", url.Host, PortSFTP)
	connection, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()
	sftpClient, err := sftp.NewClient(connection)
	if err != nil {
		log.Fatal(err)
	}
	UploadMessageFiles(sftpClient, messageID)
	sftpClient.Close()
}

func UploadMessageFiles(sftpClient *sftp.Client, messageID uuid.UUID) {
	uploadDir := os.Getenv("DIMAG_SFTP_UPLOAD_DIR")
	uniqueImportDir := "xman_import_" + uuid.NewString()
	importDir := filepath.Join(uploadDir, uniqueImportDir)
	log.Println(importDir)
	err := sftpClient.Mkdir(importDir)
	if err != nil {
		log.Println(err)
	}
}

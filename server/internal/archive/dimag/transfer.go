package dimag

import (
	"fmt"
	"io"
	"lath/xman/internal/db"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const PortSFTP uint = 22

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
	message, err := db.GetMessageByID(messageID)
	if err != nil {
		log.Fatal(err)
	}
	fileRecordObjects, err := db.GetAllFileRecordObjects(messageID)
	if err != nil {
		log.Fatal(err)
	}
	for _, fileRecordObject := range fileRecordObjects {
		err = uploadFileRecordObjectFiles(sftpClient, message, fileRecordObject)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func uploadFileRecordObjectFiles(
	sftpClient *sftp.Client,
	message db.Message,
	fileRecordObject db.FileRecordObject,
) error {
	uploadDir := os.Getenv("DIMAG_SFTP_UPLOAD_DIR")
	uniqueImportDir := "xman_import_" + uuid.NewString()
	importDir := filepath.Join(uploadDir, uniqueImportDir)
	err := sftpClient.Mkdir(importDir)
	if err != nil {
		log.Println(err)
		return err
	}
	primaryDocuments := fileRecordObject.GetPrimaryDocuments()
	for _, primaryDocument := range primaryDocuments {
		filePath := path.Join(message.StoreDir, primaryDocument.FileName)
		_, err := os.Stat(filePath)
		if err != nil {
			log.Println(err)
			return err
		}
		remotePath := filepath.Join(importDir, primaryDocument.FileName)
		err = uploadFile(sftpClient, filePath, remotePath)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func uploadFile(sftpClient *sftp.Client, localPath string, remotePath string) error {
	srcFile, err := os.Open(localPath)
	if err != nil {
		log.Println(err)
		return err
	}
	defer srcFile.Close()
	dstFile, err := sftpClient.OpenFile(remotePath, (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
	if err != nil {
		log.Println(err)
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

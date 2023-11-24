package dimag

import (
	"fmt"
	"io"
	"lath/xman/internal/db"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const PortSFTP uint = 22

func TransferToArchive(message db.Message) []string {
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
	defer sftpClient.Close()
	return UploadMessageFiles(sftpClient, message)
}

func UploadMessageFiles(sftpClient *sftp.Client, message db.Message) []string {
	fileRecordObjects, err := db.GetAllFileRecordObjects(message.ID)
	if err != nil {
		log.Fatal(err)
	}
	importDirs := []string{}
	for _, fileRecordObject := range fileRecordObjects {
		importDir, err := uploadFileRecordObjectFiles(sftpClient, message, fileRecordObject)
		if err != nil {
			log.Fatal(err)
		}
		importDirs = append(importDirs, importDir)
	}
	return importDirs
}

func uploadFileRecordObjectFiles(
	sftpClient *sftp.Client,
	message db.Message,
	fileRecordObject db.FileRecordObject,
) (string, error) {
	uploadDir := os.Getenv("DIMAG_SFTP_UPLOAD_DIR")
	uniqueImportDir := "xman_import_" + uuid.NewString()
	importDir := filepath.Join(uploadDir, uniqueImportDir)
	err := sftpClient.Mkdir(importDir)
	if err != nil {
		log.Println(err)
		return importDir, err
	}
	err = uploadXdomeaMessageFile(sftpClient, message, fileRecordObject, importDir)
	if err != nil {
		return importDir, err
	}
	primaryDocuments := fileRecordObject.GetPrimaryDocuments()
	for _, primaryDocument := range primaryDocuments {
		filePath := filepath.Join(message.StoreDir, primaryDocument.FileName)
		_, err := os.Stat(filePath)
		if err != nil {
			log.Println(err)
			return importDir, err
		}
		remotePath := primaryDocument.GetRemotePath(importDir)
		err = uploadFile(sftpClient, filePath, remotePath)
		if err != nil {
			return importDir, err
		}
	}
	return importDir, uploadControlFile(sftpClient, message, fileRecordObject, importDir)
}

func uploadXdomeaMessageFile(
	sftpClient *sftp.Client,
	message db.Message,
	fileRecordObject db.FileRecordObject,
	importDir string,
) error {
	remotePath := message.GetRemoteXmlPath(importDir)
	return uploadFile(sftpClient, message.MessagePath, remotePath)
}

func uploadControlFile(
	sftpClient *sftp.Client,
	message db.Message,
	fileRecordObject db.FileRecordObject,
	importDir string,
) error {
	remotePath := filepath.Join(importDir, ControlFileName)
	controlFileXml := GenerateControlFile(message, fileRecordObject, importDir)
	err := createRemoteTextFile(sftpClient, controlFileXml, remotePath)
	return err
}

func uploadFile(sftpClient *sftp.Client, localPath string, remotePath string) error {
	srcFile, err := os.Open(localPath)
	if err != nil {
		log.Println(err)
		return err
	}
	defer srcFile.Close()
	// the remote path must already exist
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

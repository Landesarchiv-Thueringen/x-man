package dimag

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"lath/xman/internal/db"
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

func uploadFileRecordObjectFiles(
	sftpClient *sftp.Client,
	message db.Message,
	archivePackageData db.ArchivePackage,
) (string, error) {
	uploadDir := os.Getenv("DIMAG_SFTP_UPLOAD_DIR")
	importDir := "xman_import_" + uuid.NewString()
	importPath := filepath.Join(uploadDir, importDir)
	err := sftpClient.Mkdir(importPath)
	if err != nil {
		log.Println(err)
		return importDir, err
	}
	err = uploadXdomeaMessageFile(sftpClient, message, importPath)
	if err != nil {
		return importDir, err
	}
	for _, primaryDocument := range archivePackageData.PrimaryDocuments {
		filePath := filepath.Join(message.StoreDir, primaryDocument.FileName)
		_, err := os.Stat(filePath)
		if err != nil {
			log.Println(err)
			return importDir, err
		}
		remotePath := primaryDocument.GetRemotePath(importPath)
		err = uploadFile(sftpClient, filePath, remotePath)
		if err != nil {
			return importDir, err
		}
	}
	return importDir, uploadControlFile(sftpClient, message, archivePackageData, importPath, importDir)
}

func uploadXdomeaMessageFile(
	sftpClient *sftp.Client,
	message db.Message,
	importPath string,
) error {
	remotePath := message.GetRemoteXmlPath(importPath)
	return uploadFile(sftpClient, message.MessagePath, remotePath)
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

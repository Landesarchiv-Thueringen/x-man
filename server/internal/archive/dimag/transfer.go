package dimag

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"lath/xman/internal/archive/shared"
	"lath/xman/internal/db"
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

type Connection struct {
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

func InitConnection() Connection {
	urlString := os.Getenv("DIMAG_SFTP_SERVER_URL")
	if urlString == "" {
		panic("missing env variable DIMAG_SFTP_SERVER_URL")
	}
	url, err := url.Parse(urlString)
	if err != nil {
		panic("failed to parse DIMAG SFTP server URL")
	}
	sftpUser := os.Getenv("DIMAG_SFTP_USER")
	if sftpUser == "" {
		panic("missing env variable DIMAG_SFTP_USER")
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
	sshClient, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		panic(err)
	}
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		panic(err)
	}
	return Connection{
		sshClient:  sshClient,
		sftpClient: sftpClient,
	}
}

func CloseConnection(c Connection) {
	err := c.sftpClient.Close()
	if err != nil {
		panic(err)
	}
	err = c.sshClient.Close()
	if err != nil {
		panic(err)
	}
}

func uploadArchivePackage(
	ctx context.Context,
	c Connection,
	process db.SubmissionProcess,
	message db.Message,
	archivePackage db.ArchivePackage,
) string {
	uploadDir := os.Getenv("DIMAG_SFTP_UPLOAD_DIR")
	importDir := "xman_import_" + uuid.NewString()
	importPath := filepath.Join(uploadDir, importDir)
	err := c.sftpClient.Mkdir(importPath)
	if err != nil {
		panic(err)
	}
	uploadXdomeaMessageFile(ctx, c.sftpClient, message, importPath, archivePackage)
	uploadProtocol(ctx, c.sftpClient, process, importPath)
	for _, primaryDocument := range archivePackage.PrimaryDocuments {
		filePath := filepath.Join(message.StoreDir, primaryDocument.Filename)
		_, err := os.Stat(filePath)
		if err != nil {
			panic(err)
		}
		remotePath := filepath.Join(importPath, primaryDocument.Filename)
		uploadFile(ctx, c.sftpClient, filePath, remotePath)
	}
	uploadControlFile(ctx, c.sftpClient, message, archivePackage, importPath, importDir)
	return importDir
}

func uploadXdomeaMessageFile(
	ctx context.Context,
	sftpClient *sftp.Client,
	message db.Message,
	importPath string,
	archivePackage db.ArchivePackage,
) {
	remotePath := getRemoteXmlPath(message, importPath)
	prunedMessage, err := shared.PruneMessage(message, archivePackage)
	if err != nil {
		panic(err)
	}
	createRemoteTextFile(ctx, sftpClient, prunedMessage, remotePath)
}

func uploadProtocol(
	ctx context.Context,
	sftpClient *sftp.Client,
	process db.SubmissionProcess,
	importPath string,
) {
	remotePath := filepath.Join(importPath, shared.ProtocolFilename)
	protocol := shared.GenerateProtocol(process)
	createRemoteTextFile(ctx, sftpClient, protocol, remotePath)
}

func uploadControlFile(
	ctx context.Context,
	sftpClient *sftp.Client,
	message db.Message,
	archivePackageData db.ArchivePackage,
	importPath string,
	importDir string,
) {
	remotePath := filepath.Join(importPath, ControlFileName)
	controlFileXml := GenerateControlFile(message, archivePackageData, importDir)
	createRemoteTextFile(ctx, sftpClient, controlFileXml, remotePath)
}

func uploadFile(ctx context.Context, sftpClient *sftp.Client, localPath string, remotePath string) {
	srcFile, err := os.Open(localPath)
	if err != nil {
		panic(err)
	}
	defer srcFile.Close()
	// the remote path must already exist
	dstFile, err := sftpClient.OpenFile(remotePath, (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
	if err != nil {
		panic(fmt.Sprintf("sftp: open %s: %v", remotePath, err))
	}
	defer dstFile.Close()
	_, err = copy(ctx, dstFile, srcFile)
	if err != nil {
		panic(err)
	}
}

func createRemoteTextFile(ctx context.Context, sftpClient *sftp.Client, fileContent string, remotePath string) {
	stringReader := strings.NewReader(fileContent)
	// the remote path must already exist
	dstFile, err := sftpClient.OpenFile(remotePath, (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
	if err != nil {
		panic(err)
	}
	defer dstFile.Close()
	_, err = copy(ctx, dstFile, stringReader)
	if err != nil {
		panic(err)
	}
}

// Adapted from https://gist.github.com/dillonstreator/3e9162e6e0d0929a6543a64f4564b604
type readerFunc func(p []byte) (n int, err error)

func (rf readerFunc) Read(p []byte) (n int, err error) { return rf(p) }
func copy(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	n, err := io.Copy(dst, readerFunc(func(p []byte) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			return src.Read(p)
		}
	}))
	return n, err
}

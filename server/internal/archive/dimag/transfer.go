package dimag

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"lath/xman/internal/archive/shared"
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

type Connection struct {
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

func TestConnection() error {
	c, err := InitConnection()
	if err != nil {
		return err
	}
	defer CloseConnection(c)
	err = testUploadDir(c.sftpClient)
	if err != nil {
		return err
	}
	GetCollectionIDs()
	return nil
}

func InitConnection() (Connection, error) {
	urlString := os.Getenv("DIMAG_SFTP_SERVER_URL")
	if urlString == "" {
		return Connection{}, fmt.Errorf("missing env variable DIMAG_SFTP_SERVER_URL")
	}
	url, err := url.Parse(urlString)
	if err != nil {
		return Connection{}, fmt.Errorf("failed to parse DIMAG SFTP server URL")
	}
	sftpUser := os.Getenv("DIMAG_SFTP_USER")
	if sftpUser == "" {
		return Connection{}, fmt.Errorf("missing env variable DIMAG_SFTP_USER")
	}
	// empty password is possible
	sftpPassword := os.Getenv("DIMAG_SFTP_PASSWORD")
	var auths []ssh.AuthMethod
	if sftpPassword != "" {
		auths = append(auths, ssh.Password(sftpPassword))
	}
	var sftpHostKey string
	serverState, ok := db.FindServerStateDIMAG()
	if ok {
		sftpHostKey = serverState.SFTPHostKey
	}
	config := ssh.ClientConfig{
		User: sftpUser,
		Auth: auths,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// Save the host key to the database when connecting for the first time.
			if len(sftpHostKey) == 0 {
				log.Println("Saving SFTP host key for DIMAG")
				sftpHostKey = fmt.Sprintf("%s %s", key.Type(), base64.StdEncoding.EncodeToString(key.Marshal()))
				db.UpsertServerStateDimagSFTPHostKey(sftpHostKey)
			}
			splitHostKey := strings.Split(sftpHostKey, " ")
			if len(splitHostKey) == 2 &&
				splitHostKey[0] == key.Type() &&
				splitHostKey[1] == base64.StdEncoding.EncodeToString(key.Marshal()) {
				return nil
			}
			return fmt.Errorf("failed to verify host key.\n\n" +
				"This could mean that someone is messing with your connection and tries to steal secrets!\n\n" +
				"If the server's SSH keys were changed, manually reset the \"dimag\" entry in the xman database in collection server_state")
		},
	}
	addr := fmt.Sprintf("%s:%d", url.Host, PortSFTP)
	sshClient, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		return Connection{}, err
	}
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return Connection{}, err
	}
	err = testUploadDir(sftpClient)
	if err != nil {
		return Connection{}, err
	}
	return Connection{
		sshClient:  sshClient,
		sftpClient: sftpClient,
	}, nil
}

func testUploadDir(c *sftp.Client) error {
	uploadDir := os.Getenv("DIMAG_SFTP_UPLOAD_DIR")
	if uploadDir == "" {
		return fmt.Errorf("missing env variable DIMAG_SFTP_UPLOAD_DIR")
	}
	_, err := c.Stat(uploadDir)
	if err != nil {
		return fmt.Errorf("failed to access DIMAG_SFTP_UPLOAD_DIR: sftp: stat %s: %w", uploadDir, err)
	}
	return nil
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
) (importDir string, err error) {
	uploadDir := os.Getenv("DIMAG_SFTP_UPLOAD_DIR")
	importDir = "xman_import_" + uuid.NewString()
	importPath := filepath.Join(uploadDir, importDir)
	err = c.sftpClient.Mkdir(importPath)
	if err != nil {
		return "", fmt.Errorf("sftp: mkdir %s: %w", importPath, err)
	}
	err = uploadXdomeaMessageFile(ctx, c.sftpClient, message, importPath, archivePackage)
	if err != nil {
		return "", err
	}
	err = uploadProtocol(ctx, c.sftpClient, process, importPath)
	if err != nil {
		return "", err
	}
	for _, primaryDocument := range archivePackage.PrimaryDocuments {
		filePath := filepath.Join(message.StoreDir, primaryDocument.Filename)
		_, err = os.Stat(filePath)
		if err != nil {
			return "", err
		}
		remotePath := filepath.Join(importPath, primaryDocument.Filename)
		err = uploadFile(ctx, c.sftpClient, filePath, remotePath)
		if err != nil {
			return "", err
		}
	}
	err = uploadControlFile(ctx, c.sftpClient, message, archivePackage, importPath, importDir)
	return importDir, err
}

func uploadXdomeaMessageFile(
	ctx context.Context,
	sftpClient *sftp.Client,
	message db.Message,
	importPath string,
	archivePackage db.ArchivePackage,
) error {
	remotePath := getRemoteXmlPath(message, importPath)
	prunedMessage := shared.PruneMessage(message, archivePackage)
	return createRemoteTextFile(ctx, sftpClient, prunedMessage, remotePath)
}

func uploadProtocol(
	ctx context.Context,
	sftpClient *sftp.Client,
	process db.SubmissionProcess,
	importPath string,
) error {
	remotePath := filepath.Join(importPath, shared.ProtocolFilename)
	protocol := shared.GenerateProtocol(process)
	return createRemoteTextFile(ctx, sftpClient, protocol, remotePath)
}

func uploadControlFile(
	ctx context.Context,
	sftpClient *sftp.Client,
	message db.Message,
	archivePackageData db.ArchivePackage,
	importPath string,
	importDir string,
) error {
	remotePath := filepath.Join(importPath, ControlFileName)
	controlFileXml := GenerateControlFile(message, archivePackageData, importDir)
	return createRemoteTextFile(ctx, sftpClient, controlFileXml, remotePath)
}

func uploadFile(ctx context.Context, sftpClient *sftp.Client, localPath string, remotePath string) error {
	srcFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	// the remote path must already exist
	dstFile, err := sftpClient.OpenFile(remotePath, (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
	if err != nil {
		return fmt.Errorf("sftp: open %s: %w", remotePath, err)
	}
	defer dstFile.Close()
	_, err = copy(ctx, dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("sftp: write file %s: %w", remotePath, err)
	}
	return nil
}

func createRemoteTextFile(ctx context.Context, sftpClient *sftp.Client, fileContent string, remotePath string) error {
	stringReader := strings.NewReader(fileContent)
	// the remote path must already exist
	dstFile, err := sftpClient.OpenFile(remotePath, (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
	if err != nil {
		return fmt.Errorf("sftp: open %s: %w", remotePath, err)
	}
	defer dstFile.Close()
	_, err = copy(ctx, dstFile, stringReader)
	if err != nil {
		return fmt.Errorf("sftp: write file %s: %w", remotePath, err)
	}
	return err
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

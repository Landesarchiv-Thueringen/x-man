package dimag

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"lath/xman/internal/db"
	"log"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const sftpPort uint = 22

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
	_, err = GetCollectionIDs()
	return err
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
	addr := fmt.Sprintf("%s:%d", url.Host, sftpPort)
	sshClient, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		return Connection{}, err
	}
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return Connection{}, err
	}
	err = createUploadDir(sftpClient)
	if err != nil {
		return Connection{}, err
	}
	return Connection{
		sshClient:  sshClient,
		sftpClient: sftpClient,
	}, nil
}

// createUploadDir creates the upload directory on the SFTP remote if it doesn't
// exist.
func createUploadDir(c *sftp.Client) error {
	sftpDir := os.Getenv("DIMAG_SFTP_DIR")
	if sftpDir == "" {
		return fmt.Errorf("missing env variable DIMAG_SFTP_DIR")
	}
	_, err := c.Stat(sftpDir)
	if err != nil {
		return fmt.Errorf("failed to access DIMAG_SFTP_DIR: sftp: stat %s: %w", sftpDir, err)
	}
	uploadDir := path.Join(sftpDir, "Import")
	_, err = c.Stat(uploadDir)
	if err != nil {
		err = c.Mkdir(uploadDir)
		if err != nil {
			return fmt.Errorf("failed create upload dir: sftp: mkdir %s: %w", uploadDir, err)
		}
	}
	return nil
}

func testUploadDir(c *sftp.Client) error {
	sftpDir := os.Getenv("DIMAG_SFTP_DIR")
	uploadDir := path.Join(sftpDir, "Import")
	_, err := c.Stat(uploadDir)
	if err != nil {
		return fmt.Errorf("failed to access upload dir: sftp: stat %s: %w", sftpDir, err)
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

// getUploadDir returns the remote directory name as which the BagIt will be
// uploaded.
func getUploadDir(bagit bagitHandle) string {
	return "Import/xman_bagit_" + bagit.ID()
}

// uploadBagit creates a remote import directory on the DIMAG server
// and uploads the given BagIt package to it.
func uploadBagit(
	ctx context.Context,
	c Connection,
	bagit bagitHandle,
) (remotePath string, err error) {
	uploadDir := getUploadDir(bagit)
	remotePath = filepath.Join(os.Getenv("DIMAG_SFTP_DIR"), uploadDir)
	log.Printf("Uploading %s...\n", uploadDir)
	err = uploadDirRecursive(ctx, c, bagit.Path(), remotePath)
	log.Println("Upload done")
	return uploadDir, err
}

func uploadDirRecursive(
	ctx context.Context, c Connection,
	localPath, remotePath string,
) error {
	entries, err := os.ReadDir(localPath)
	if err != nil {
		return err
	}
	err = c.sftpClient.Mkdir(remotePath)
	if err != nil {
		return fmt.Errorf("sftp: mkdir %s: %w", remotePath, err)
	}
	for _, entry := range entries {
		localEntryPath := filepath.Join(localPath, entry.Name())
		remoteEntryPath := filepath.Join(remotePath, entry.Name())
		if entry.IsDir() {
			err = uploadDirRecursive(ctx, c, localEntryPath, remoteEntryPath)
		} else {
			err = uploadFile(ctx, c.sftpClient, localEntryPath, remoteEntryPath)
		}
		if err != nil {
			return err
		}
	}
	return nil
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

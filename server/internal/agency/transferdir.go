package agency

import (
	"lath/xman/internal/auth"
	"lath/xman/internal/clearing"
	"lath/xman/internal/db"
	"lath/xman/internal/mail"
	"lath/xman/internal/messagestore"
	"lath/xman/internal/xdomea"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/studio-b12/gowebdav"
)

type URLScheme string

const (
	Local     URLScheme = "file"
	WebDAV    URLScheme = "webdav"
	WebDAVSec URLScheme = "webdavs"
)

var ticker time.Ticker
var stop chan bool

func TestTransferDir(dir string) bool {
	_, err := os.ReadDir(dir)
	return err == nil
}

func MonitorTransferDirs() {
	ticker = *time.NewTicker(time.Second * 5)
	stop = make(chan bool)
	go watchLoop(ticker, stop)
}

func watchLoop(timer time.Ticker, stop chan bool) {
	for {
		select {
		case <-stop:
			timer.Stop()
			return
		case <-timer.C:
			readMessages()
		}
	}
}

func readMessages() {
	agencies := db.GetAgencies()
	for _, agency := range agencies {
		transferDirURL, err := url.Parse(agency.TransferDirURL)
		if err != nil {
			panic(err)
		}
		switch transferDirURL.Scheme {
		case string(Local):
			readMessagesFromLocalFilesystem(agency, transferDirURL)
		case string(WebDAV):
			readMessagesFromWebDAV(agency, transferDirURL, false)
		case string(WebDAVSec):
			readMessagesFromWebDAV(agency, transferDirURL, true)
		default:
			panic("unknown transfer directory scheme")
		}
	}
}

func readMessagesFromLocalFilesystem(agency db.Agency, transferDirURL *url.URL) {
	rootDir := filepath.Join(transferDirURL.Path)
	files, err := os.ReadDir(rootDir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if !file.IsDir() && xdomea.IsMessage(file.Name()) {
			fullPath := filepath.Join(rootDir, file.Name())
			if !db.IsMessageAlreadyProcessed(fullPath) {
				log.Println("Processing new message " + fullPath)
				go func() {
					defer func() {
						if r := recover(); r != nil {
							log.Println("Error: readMessages panicked:", r)
							debug.PrintStack()
						}
					}()
					message, err := messagestore.StoreMessage(agency, fullPath)
					clearing.HandleError(err)
					if err == nil {
						for _, user := range agency.Users {
							address := auth.GetMailAddress(user.ID)
							preferences := db.GetUserInformation(user.ID).Preferences
							if preferences.MessageEmailNotifications {
								mail.SendMailNewMessage(address, agency.Name, message)
							}
						}
					}
				}()
			}
		}
	}
}

func readMessagesFromWebDAV(agency db.Agency, transferDirURL *url.URL, secure bool) {
	root := "https://" + transferDirURL.Host
	if !secure {
		root = "http://" + transferDirURL.Host
	}
	user := transferDirURL.User.Username()
	password, set := transferDirURL.User.Password()
	if !set {
		password = ""
	}
	c := gowebdav.NewClient(root, user, password)
	err := c.Connect()
	if err != nil {
		panic(err)
	}
	path := transferDirURL.Path
	files, err := c.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		log.Println(file.Name)
	}
}

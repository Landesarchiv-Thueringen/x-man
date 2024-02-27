package agency

import (
	"io/ioutil"
	"lath/xman/internal/db"
	"lath/xman/internal/messagestore"
	"lath/xman/internal/xdomea"
	"log"
	"os"
	"path/filepath"
	"time"
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
		files, err := ioutil.ReadDir(agency.TransferDir)
		if err != nil {
			panic(err)
		}
		for _, file := range files {
			if !file.IsDir() && xdomea.IsMessage(file.Name()) {
				fullPath := filepath.Join(agency.TransferDir, file.Name())
				if !db.IsMessageAlreadyProcessed(fullPath) {
					log.Println("Processing new message " + fullPath)
					go messagestore.StoreMessage(agency, fullPath)
				}
			}
		}
	}
}

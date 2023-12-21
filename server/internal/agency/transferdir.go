package agency

import (
	"io/ioutil"
	"lath/xman/internal/db"
	"lath/xman/internal/messagestore"
	"lath/xman/internal/xdomea"
	"log"
	"path/filepath"
	"time"
)

var ticker time.Ticker
var stop chan bool

func watchTransferDirectories(agencies []db.Agency) {
	ticker = *time.NewTicker(time.Second * 5)
	stop = make(chan bool)
	go watchLoop(agencies, ticker, stop)
}

func watchLoop(agencies []db.Agency, timer time.Ticker, stop chan bool) {
	for {
		select {
		case <-stop:
			timer.Stop()
			return
		case <-timer.C:
			readMessages(agencies)
		}
	}
}

func readMessages(agencies []db.Agency) {
	for _, agency := range agencies {
		files, err := ioutil.ReadDir(agency.TransferDir)
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range files {
			if !file.IsDir() && xdomea.IsMessage(file.Name()) {
				fullPath := filepath.Join(agency.TransferDir, file.Name())
				if !db.IsMessageAlreadyProcessed(fullPath) {
					go messagestore.StoreMessage(agency, fullPath)
				}
			}
		}
	}
}

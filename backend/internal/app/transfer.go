package app

import (
	"context"
	"lath/xman/internal/db"
	"log"
	"time"
)

var ticker time.Ticker

func MonitorTransferDirectories() {
	ticker = *time.NewTicker(time.Second * 5)
	for {
		<-ticker.C
		senders, err := db.FindSenders(context.Background())
		if err != nil {
			// TODO: add error to global error info
			continue
		}
		for _, sender := range senders {
			switch sender.TransferDir.TransferMode {
			case db.Local:
				log.Println("hello")
			}
		}
	}
}

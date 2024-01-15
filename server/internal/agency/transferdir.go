package agency

import (
	"lath/xman/internal/db"
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

		}
	}
}

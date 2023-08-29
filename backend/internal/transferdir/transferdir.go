package transferdir

import (
	"lath/xdomea/internal/messagestore"
	"lath/xdomea/internal/xdomea"
	"log"
	"math"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// example: https://github.com/fsnotify/fsnotify/blob/main/cmd/fsnotify/dedup.go

// Depending on the system, a single "write" can generate many Write events; for
// example compiling a large Go program can generate hundreds of Write events on
// the binary.
//
// The general strategy to deal with this is to wait a short time for more write
// events, resetting the wait period for every new event.

func Watch(paths ...string) {
	if len(paths) < 1 {
		log.Fatal("no transfer directories given")
	}

	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Start listening for events.
	go watchLoop(watcher)

	// Add all paths from the commandline.
	for _, path := range paths {
		err = watcher.Add(path)
		if err != nil {
			log.Fatal(path, err)
		}
	}

	<-make(chan struct{}) // Block forever
}

// Checks if a file event represents the creation of a new processable xdomea message.
// The corresponding file must match the xdomea naming conventions.
// The event must be a create event. It could be necessarry to add write events.
func isProcessableMessage(event fsnotify.Event) bool {
	return xdomea.IsMessage(event.Name) && event.Has(fsnotify.Create)
}

func watchLoop(watcher *fsnotify.Watcher) {
	var (
		// Wait 100ms for new events; each new event resets the timer.
		waitFor = 100 * time.Millisecond

		// Keep track of the timers, as path → timer.
		mutex  sync.Mutex
		timers = make(map[string]*time.Timer)

		// Callback we run.
		processEvent = func(event fsnotify.Event) {
			if isProcessableMessage(event) {
				go messagestore.StoreMessage(event.Name)
			}

			// Don't need to remove the timer if you don't have a lot of files.
			mutex.Lock()
			delete(timers, event.Name)
			mutex.Unlock()
		}
	)

	for {
		select {
		// Read from Errors.
		case err, ok := <-watcher.Errors:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			log.Fatal(err)
		// Read from Events.
		case event, ok := <-watcher.Events:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}

			// We just want to watch for file creation, so ignore everything
			// outside of Create and Write.
			if !event.Has(fsnotify.Create) && !event.Has(fsnotify.Write) {
				continue
			}

			// Get timer.
			mutex.Lock()
			timer, ok := timers[event.Name]
			mutex.Unlock()

			// No timer yet, so create one.
			if !ok {
				timer = time.AfterFunc(math.MaxInt64, func() { processEvent(event) })
				timer.Stop()

				mutex.Lock()
				timers[event.Name] = timer
				mutex.Unlock()
			}

			// Reset the timer for this path, so it will start from 100ms again.
			timer.Reset(waitFor)
		}
	}
}

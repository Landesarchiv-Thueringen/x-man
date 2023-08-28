package transferdir

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

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

func watchLoop(watcher *fsnotify.Watcher) {
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
			// Just print the event
			log.Print(event)
		}
	}
}

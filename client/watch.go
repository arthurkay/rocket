package client

import (
	"fmt"
	"os"
	"path/filepath"
	"rocket/log"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watch one or more files, but instead of watching the file directly it watches
// the parent directory. This solves various issues where files are frequently
// renamed, such as editors saving them.
func watch(c *Controller, opts *Options, files ...string) {
	if len(files) < 1 {
		log.Error("must specify at least one file to watch")
		os.Exit(1)
	}

	// Create a new watcher.
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("creating a new watcher: %s", err)
	}
	defer w.Close()

	// read configuration file
	config, err := LoadConfiguration(opts)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	go c.Run(config)

	// Start listening for events.
	go fileLoop(w, c, opts, files)

	// Add all files from the commandline.
	for _, p := range files {
		st, err := os.Lstat(p)
		if err != nil {
			log.Error("%s", err)
		}

		if st.IsDir() {
			log.Error("%q is a directory, not a file", p)
		}

		// Watch the directory, not the file itself.
		err = w.Add(filepath.Dir(p))
		if err != nil {
			log.Error("%q: %s", p, err)
		}
	}

	log.Info("ready; press ^C to exit")
	<-make(chan struct{}) // Block forever
}

func fileLoop(w *fsnotify.Watcher, c *Controller, opts *Options, files []string) {
	i := 0
	for {
		select {
		// Read from Errors.
		case err, ok := <-w.Errors:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			fmt.Printf("ERROR: %s", err)
		// Read from Events.
		case e, ok := <-w.Events:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}

			// Ignore files we're not interested in. Can use a
			// map[string]struct{} if you have a lot of files, but for just a
			// few files simply looping over a slice is faster.
			var found bool
			for _, f := range files {
				if f == e.Name {
					found = true
				}
			}
			if !found {
				continue
			}

			// Just print the event nicely aligned, and keep track how many
			// events we've seen.
			i++
			log.Error("%3d %s", i, e)
			// read configuration file
			time.Sleep(1 * time.Second)
			// Update the confifurations and prepare traffic migration to new controller
			newController := NewController()
			log.Info("Migrating traffic to another controller instance")
			for _, view := range c.views {
				newController.AddView(view)
			}
			newController.model = c.model
			//c.Shutdown("Killing old connection")
			c.model = nil
			c = newController
			config, err := LoadConfiguration(opts)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			newController.Shutdown("Good Bye")

			c.Run(config)
		}
	}
}

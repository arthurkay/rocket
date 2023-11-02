package client

import (
	"os"
	"path/filepath"
	"rocket/log"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watch one or more files, but instead of watching the file directly it watches
// the parent directory. This solves various issues where files are frequently
// renamed, such as editors saving them.
func watch(reload chan *bool, opts *Options, initController *Controller) {
	var files []string = []string{opts.config}
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

	// Start listening for events.
	go fileLoop(w, opts, reload, initController, files)

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

func fileLoop(w *fsnotify.Watcher, opts *Options, reload chan *bool, initController *Controller, files []string) {
	i := 0
	for {
		select {
		// Handle the reload action here
		case <-reload:
			log.Debug("Reloading....")
		// Read from Errors.
		case err, ok := <-w.Errors:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			log.Error("Oops: %s", err)
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
			log.Debug("%3d %s", i, e)
			time.Sleep(1 * time.Second)

			if e.Has(fsnotify.Write) {
				log.Debug("File changed, starting communication with the server")
				status := true
				select {
				case <-reload:
					log.Debug("Read from the reload channel")
				case reload <- &status:
					log.Debug("Sent to the reload channel")
					go newInstance(opts, reload, initController)
				default:
					log.Debug("No reload channel")
					status = true
				}

			}
		}
	}
}

func newInstance(opts *Options, reload chan *bool, c *Controller) {
	ready := <-reload
	if *ready {
		log.Debug("Loading new config file...")
		config, err := LoadConfiguration(opts)
		if err != nil {
			log.Error("%v", err)
			os.Exit(1)
		}
		config.InspectAddr = "disabled"
		for _, tunnel := range c.GetModel().tunnels {
			log.Debug("Tunnel %s", tunnel.PublicUrl)
		}
		//clientID := c.GetModel().id
		c.model = c.GetModel()
		c.doShutdown()

		c = NewController()

		log.Info("Reloaded config.")
		c.Run(config)
	}
}

package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/kupospelov/feeds-to-instapaper/config"
	"github.com/kupospelov/feeds-to-instapaper/instapaper"
	"github.com/kupospelov/feeds-to-instapaper/processor"
	"github.com/kupospelov/feeds-to-instapaper/state"
	"github.com/mmcdole/gofeed"
)

func scheduleCleanup(state *state.State) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	save := func() {
		err := state.Save()
		if err != nil {
			log.Fatalf("Failed to save state: %v", err)
		}
	}

	var saveOnce sync.Once
	defer saveOnce.Do(save)
	go func() {
		<-c
		saveOnce.Do(save)
		os.Exit(0)
	}()
}

func main() {
	log.SetFlags(0)

	config, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	state, err := state.Load()
	if err != nil {
		log.Fatalf("Failed to load state: %v", err)
	}

	scheduleCleanup(state)

	parser := gofeed.NewParser()
	instapaper := instapaper.New(config.Instapaper.Username, config.Instapaper.Password)
	proc := processor.New(parser, instapaper, state)

	err = proc.ProcessFeeds(config.Feeds.URLs)
	if err != nil {
		log.Fatalf("Error processing feeds: %v", err)
	}
}

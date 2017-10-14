package config

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/redforks/hal"
	"github.com/redforks/life"

	"github.com/redforks/testing/reset"
)

func start() {
	if reset.TestMode() {
		// In test mode, apply default options
		if err := initAllOptions(nil); err != nil {
			panic(err)
		}
		return
	}

	if dumpDefaultOptions {
		s, err := DumpDefaultOptions()
		if err != nil {
			panic(err)
		}
		fmt.Print(s)
		hal.Exit(0)
	}

	// only start signal monitor on non-test mode
	go monitorSignal()
	if err := Load(""); err != nil {
		log.Panic(err)
	}
}

// monitor USR1 signal to reload config and apply
func monitorSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)
	for range c {
		Reload()
	}
}

func init() {
	reset.Register(nil, func() {
		life.RegisterHook("config", 10, life.BeforeStarting, start)
	})
}

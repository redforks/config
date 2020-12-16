package config

import (
	"fmt"
	"log"

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

func init() {
	reset.Register(nil, func() {
		life.RegisterHook("config", 10, life.BeforeStarting, start)
	})
}

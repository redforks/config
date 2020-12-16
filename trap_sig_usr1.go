// +build linux

package config

import (
	"os"
	"os/signal"
	"syscall"
)

// monitor USR1 signal to reload config and apply
func monitorSignal() {
	c := make(chan os.Signal, 1)

	signal.Notify(c, syscall.SIGUSR1)

	for range c {
		Reload()
	}
}

// Package config package helps manage application configuration:
//
//  1. Each package do not need store and parse configuration
//  2. Application only need one line code to get configuration support, no
//     matter how many packages pulled.
//  3. Currently support configuration file only, later can store configuration
//     in a dedicate config server, suitable for micro service structure.
//  4. Monitor SIGUSR1, on SIGUSR1 reload configuration file, and apply to each
//     package.
//  5. Works with life package in mind. Only allow first init in life.Initing
//     phase, and not reload configuration in Shutingdown phase.
//
// Note: no order and dependency sort, although in current implementation,
// Init() called by Register order, which probabbly is go package dependency
// order, they maybe change in future implementation. The point is each package
// must work independently, matches the concept of life package: no order in
// life.Initing phase, just get the configuration, delay the work to Starting phase.
//
// If .Load() not called, all packages inited() with its default option, and no
// reload() support.
package config

import "github.com/redforks/testing/reset"

func init() {
	reset.Register(func() {
		filename = ""
	}, nil)
}

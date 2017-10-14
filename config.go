package config

import (
	"bytes"
	"fmt"
	"log"
	"regexp"

	"github.com/redforks/life"

	"github.com/redforks/testing/reset"

	"github.com/BurntSushi/toml"
	"github.com/urfave/cli"
)

const (
	tag = "config"
)

type optionRec struct {
	name    string
	op      Option
	creator OptionCreator
}

// The OptionCreator is a factory function to create option interface
type OptionCreator func() Option

var (
	// all access to options are in one goroutie, so no lock on it. During init
	// phase, it was called in main gorutine, during later, it was called in
	// monitor signal routine, which also has one gorutine.
	options []*optionRec

	// configuration file
	filename string
	loaded   bool

	// Flag is config file command line flag, add Flag to your cli.App.Flags.
	Flag cli.Flag = &cli.StringFlag{
		Name:        "config, c",
		Usage:       "specify config file",
		Destination: &filename,
	}

	dumpDefaultOptions bool

	// FlagDumpDefaultOptions dump default options to stdout if enabled, add to your cli.App.Flags.
	FlagDumpDefaultOptions cli.Flag = &cli.BoolFlag{
		Name:        "dumpDefaultOptions",
		Usage:       "Dump default options to stdout, save it as config file",
		Destination: &dumpDefaultOptions,
	}
)

// ResetInternal reset config package internal state, all registered options lost.
// It is mainly used in unit tests of config package itself.
func ResetInternal() {
	options = nil
	filename = ""
}

// Option is the interface config package manages. Must implement as a struct,
// and can be marshaled as toml, currently github.com/BurntSushi/toml.
type Option interface {
	// Called on configuration first load, and ensure it will during life.Initing
	// phase. If any error returned, abort following Init() and return this error.
	Init() error

	// Called on reload configuration, do not allow return error. If the new
	// config can not apply, log to log.
	Apply()
}

// Register an option object. Panic on name conflict.
// fnCreateDefault is a func create default value, must not return nil.
// fnCreateDefault may called multiple times, each time should create a new
// instance.
// Registered options not reset by spork/testing/reset package.
func Register(name string, fnCreateDefault OptionCreator) {
	// https://github.com/toml-lang/toml#user-content-spec:
	//
	// Keys may be either bare or quoted. Bare keys may only contain letters,
	// numbers, underscores, and dashes (A-Za-z0-9_-). Note that bare keys are
	// allowed to be composed of only digits, e.g. 1234.
	//
	// Although special character such as '.', '/' can be used by quote, but it
	// is complex and confusing, totally unnecessary.
	if ok, _ := regexp.Match("^[a-zA-Z0-9_-]+$", []byte(name)); !ok {
		log.Panicf("[%s] bad option name '%s'", tag, name)
	}

	life.EnsureStatef(life.Initing, "[%s] must register '%s' Option at life Initing phase", tag, name)

	if fnCreateDefault == nil {
		log.Panicf("[%s] can not register nil Option '%s'", tag, name)
	}

	for _, rec := range options {
		if rec.name == name {
			log.Panicf("[%s] option '%s' already registered", tag, name)
		}
	}
	options = append(options, &optionRec{name, nil, fnCreateDefault})
}

func optionChanged(op1, op2 Option) bool {
	buf1, buf2 := &bytes.Buffer{}, &bytes.Buffer{}
	if err := toml.NewEncoder(buf1).Encode(op1); err != nil {
		log.Panicf(err.Error())
	}
	if err := toml.NewEncoder(buf2).Encode(op2); err != nil {
		log.Panicf(err.Error())
	}

	return buf1.String() != buf2.String()
}

func getDefaultOptions() ([]Option, error) {
	opts := make([]Option, len(options))
	for i, rec := range options {
		if o := overrideDefOptions[rec.name]; o != nil {
			opts[i] = o
			delete(overrideDefOptions, rec.name)
		} else {
			opts[i] = rec.creator()
		}
	}
	if name, exist := getAnyKey(overrideDefOptions); exist {
		return nil, fmt.Errorf("[%s] overridden option \"%s\" not used, wrong option name?", tag, name)
	}
	return opts, nil
}

func getAnyKey(v map[string]Option) (string, bool) {
	for k := range v {
		return k, true
	}
	return "", false
}

func initAllOptions(opts []Option) error {
	var err error
	if opts == nil {
		if opts, err = getDefaultOptions(); err != nil {
			return err
		}
	}

	for i, rec := range options {
		log.Printf("[%s] initing option '%s'", tag, rec.name)
		if err := opts[i].Init(); err != nil {
			return err
		}
	}
	log.Printf("[%s] inited all options", tag)

	storeOptions(opts)
	return nil
}

func storeOptions(opts []Option) {
	for i, rec := range options {
		rec.op = opts[i]
	}
}

func init() {
	reset.Register(func() {
		loaded = false
	}, nil)
}

package config

import (
	"log"
	"os"
	"sync"

	"github.com/redforks/appinfo"
	"github.com/redforks/life"
	"github.com/redforks/xdgdirs"

	"github.com/BurntSushi/toml"
	"github.com/redforks/errors"
	"github.com/redforks/hal"
)

// Load config file manually. Normally loading config file is triggered by life
// package before life starting. For non-daemon applications, do not start life
// cycle, calling Load() manually in this case.
// Load can only be called once, it will fail if calling config.Load() and
// life.Start().
//
// Note: Use alternative config filename in server mode is not supported.
//
// If configFile is empty, use default config file resolve process, and this is
// recommended.
func Load(configFile string) (err error) {
	if loaded {
		return errors.Bugf("[%s] can not call Load() twice", tag)
	}

	loaded = true

	if configFile != "" {
		filename = configFile
	}

	if filename == "" {
		var err error
		filename, err = xdgdirs.ResolveConfigFile(appinfo.CodeName() + ".conf")

		if err != nil {
			return initAllOptions(nil)
		}
	}

	log.Printf("[%s] loading %s", tag, filename)
	var opts []Option
	if opts, err = loadConfigFile(); err != nil {
		return
	}

	return initAllOptions(opts)
}

var reloadLock sync.Mutex

// Reload config file and apply changed configurations.
func Reload() {
	reloadLock.Lock()
	defer reloadLock.Unlock()

	defer func() {
		e := recover()
		if e != nil {
			errors.Handle(nil, e)
			hal.Exit(20)
		}
	}()

	if life.State() == life.Shutingdown {
		log.Printf("[%s] abort reload on %s phase", tag, life.State())
		return
	}
	log.Printf("[%s] SIGUSR1 detected, reloading configfile '%s'", tag, filename)

	if filename == "" {
		// ignore reload() request if no config file attached
		return
	}

	var (
		opts    []Option
		err     error
		changed bool
	)
	if opts, err = loadConfigFile(); err != nil {
		log.Printf("[%s] error when reloading configfile: %s", tag, err.Error())
		return
	}

	for i, rec := range options {
		if optionChanged(opts[i], rec.op) {
			log.Printf("[%s] '%s' option chanegd, applying", tag, rec.name)
			opts[i].Apply()
			changed = true
		}
	}
	if changed {
		log.Printf("[%s] applied all changed options", tag)
	} else {
		log.Printf("[%s] no options changed", tag)
	}

	storeOptions(opts)
}

func loadConfigFile() (opts []Option, err error) {
	var (
		dict map[string]toml.Primitive
		md   toml.MetaData
	)

	if opts, err = getDefaultOptions(); err != nil {
		return
	}

	if md, err = toml.DecodeFile(filename, &dict); err != nil {
		if !os.IsNotExist(err) {
			return
		}

		// file not exist, log and continue
		log.Printf("[%s] %s", tag, err.Error())
		err = nil
	}

	for i, rec := range options {
		if p, ok := dict[rec.name]; ok {
			if err = md.PrimitiveDecode(p, opts[i]); err != nil {
				return
			}
		}
	}

	keys := md.Undecoded()
	if len(keys) != 0 {
		log.Printf("[%s] Warning fowlling keys in config file '%s' are unknow and ignored, possibly wrong spelling. %v", tag, filename, keys)
	}
	return
}

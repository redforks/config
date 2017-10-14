package config

import (
	"log"

	"github.com/redforks/testing/reset"
)

var (
	overrideDefOptions = make(map[string]Option)
)

// SetDefaultOptionForTest override default option object.
//
// A default option is created by default option creator function,
// by calling SetDefaultOption(), override default option for unit tests.
func SetDefaultOptionForTest(pkg string, op Option) {
	if !reset.TestMode() {
		panic("SetDefaultOptionForTest can only be used in unit tests")
	}

	if _, exist := overrideDefOptions[pkg]; exist {
		log.Panicf("[%s] SetDefaultOptionForTest: \"%s\" already set", tag, pkg)
	}

	overrideDefOptions[pkg] = op
}

func init() {
	reset.Register(func() {
		overrideDefOptions = make(map[string]Option)
	}, nil)
}

package config_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/redforks/config"

	"github.com/redforks/life"

	"github.com/redforks/testing/iotest"
	"github.com/redforks/testing/matcher"
	"github.com/redforks/testing/reset"

	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redforks/errors"
	"github.com/redforks/hal"
)

var _ = Describe("config", func() {

	var (
		testDir  iotest.TempTestDir
		filename string
	)

	var (
		writeConfigFile = func(content string) {
			Ω(ioutil.WriteFile(filename, []byte(content), os.ModePerm)).Should(Succeed())
		}

		loadNoError = func() {
			Ω(Load(filename)).Should(Succeed())
		}
	)

	BeforeEach(func() {
		reset.Enable()
		testDir = iotest.NewTempTestDir()
		filename = filepath.Join(testDir.Dir(), "app.conf")
		ops = nil
		initHits = nil
		applyHits = nil
		initErrors = nil
	})

	AfterEach(func() {
		ResetInternal()
		reset.Disable()
	})

	It("Empty - config file not exist", func() {
		loadNoError()
	})

	It("Empty - config file exist", func() {
		writeConfigFile("")
		loadNoError()
	})

	It("Open config file failed", func() {
		writeConfigFile("[ab")
		Ω(Load(filename)).Should(HaveOccurred())
	})

	It("Dup name", func() {
		Register("foo", newFakeOption(0))
		Ω(func() {
			Register("foo", newFakeOption(0))
		}).Should(matcher.Panics("[config] option 'foo' already registered"))
	})

	It("Bad name", func() {
		var badName = func(name string) {
			Ω(func() {
				Register(name, newFakeOption(0))
			}).Should(matcher.Panics(fmt.Sprintf("[config] bad option name '%s'", name)))
		}
		badName("")
		badName("a/b")
		badName("a.b")
	})

	It("option can not be nil", func() {
		Ω(func() {
			Register("foo", nil)
		}).Should(matcher.Panics("[config] can not register nil Option 'foo'"))
	})

	It("Only allow Register at life.Initing", func() {
		life.Start()
		Ω(func() {
			Register("foo", newFakeOption(0))
		}).Should(matcher.Panics("[config] must register 'foo' Option at life Initing phase"))
	})

	Context("One option", func() {

		BeforeEach(func() {
			Register("foo", newFakeOption(0))
		})

		It("Empty config file", func() {
			writeConfigFile("")
			loadNoError()
			Ω(ops[0].Name).Should(Equal("bar"))
			Ω(ops[0].Count).Should(Equal(3))
		})

		It("Config file with partial value", func() {
			writeConfigFile(`[foo]
Name = "foobar"
			`)
			loadNoError()
			Ω(ops[0].Name).Should(Equal("foobar"))
			Ω(ops[0].Count).Should(Equal(3))

			Ω(initHits[0]).Should(Equal(1))
			Ω(applyHits[0]).Should(Equal(0))
		})

		It("Abort if Init() return error", func() {
			writeConfigFile(`[foo]
Name = "foobar"
			`)
			initErrors = []error{errors.New("error")}
			Ω(Load(filename)).Should(Equal(initErrors[0]))
		})

		It("Init() called even not mentioned in config file", func() {
			writeConfigFile("")
			initErrors = []error{errors.New("error")}
			Ω(Load(filename)).Should(Equal(initErrors[0]))
		})

		It("Init() called when no config file", func() {
			initErrors = []error{errors.New("error")}
			Ω(Load(filename)).Should(Equal(initErrors[0]))
		})

		It("Reload on SIGUSR1", func() {
			loadNoError()
			Ω(initHits[0]).Should(Equal(1))
			Ω(applyHits[0]).Should(Equal(0))

			writeConfigFile(`[foo]
Name = "bar1"
			`)
			Reload()
			Ω(initHits[0]).Should(Equal(1))
			Ω(applyHits[0]).Should(Equal(1))
		})

		It("Abort reload in shutingdown", func() {
			life.Start()
			life.Shutdown()
			Ω(initHits[0]).Should(Equal(1))
			Ω(applyHits[0]).Should(Equal(0))
		})

		It("Init all options even if not load", func() {
			life.Start()
			Ω(initHits[0]).Should(Equal(1))
		})

		It("Abort reload if .Load() not called", func() {
			Reload()
			Ω(len(ops)).Should(Equal(0))
		})

		Context("panic in apply", func() {
			var (
				exitCodes []int
				errLog    []string
			)

			BeforeEach(func() {
				exitCodes = nil
				hal.Exit = func(n int) {
					exitCodes = append(exitCodes, n)
				}

				errLog = nil
				errors.SetHandler(func(_ context.Context, err interface{}) {
					errLog = append(errLog, fmt.Sprintf("%s", err))
				})

				loadNoError()
				writeConfigFile(`[foo]
Panic = true
			`)
			})

			AfterEach(func() {
				errors.SetHandler(nil)
			})

			It("Exit on panic", func() {
				Reload()
				Ω(exitCodes).Should(Equal([]int{20}))
			})

			It("report error", func() {
				Reload()
				Ω(errLog).Should(Equal([]string{"error"}))
			})

		})

	})

	Context("Multiple options", func() {

		It("Abort remains Inits", func() {
			initErrors = []error{nil, errors.New("error")}
			Register("op1", newFakeOption(0))
			Register("op2", newFakeOption(1))
			Register("op3", newFakeOption(2))
			Ω(Load(filename)).Should(Equal(initErrors[1]))
			Ω(initHits[0]).Should(Equal(1))
			Ω(initHits[1]).Should(Equal(1))
			Ω(ops).Should(HaveLen(2))
		})

		It("Only apply changed options", func() {
			Register("op1", newFakeOption(0))
			Register("op2", newFakeOption(1))
			Register("op3", newFakeOption(2))
			Register("op4", newFakeOption(3))
			writeConfigFile(`[op1]
Name = "foo"

[op2]
Name = "bar"

[op4]
Name = "op4"
`)
			loadNoError()

			writeConfigFile(`[op1]
Name = "foo"

[op2]
Name = "foobar"

[op3]
Name = "wow"
`)
			Reload()

			Ω(initHits[0]).Should(Equal(1))
			Ω(initHits[1]).Should(Equal(1))
			Ω(initHits[2]).Should(Equal(1))
			Ω(initHits[3]).Should(Equal(1))
			Ω(applyHits[0]).Should(Equal(0))
			Ω(applyHits[1]).Should(Equal(1))
			Ω(applyHits[2]).Should(Equal(1))
			Ω(applyHits[3]).Should(Equal(1))

			Ω(ops[0].Name).Should(Equal("foo"))
			Ω(ops[1].Name).Should(Equal("foobar"))
			Ω(ops[2].Name).Should(Equal("wow"))
			Ω(ops[3].Name).Should(Equal("bar"))

			writeConfigFile(`[op1]
Name = "foo"

[op2]
Name = "foobar1"
`)
			Reload()
			Ω(applyHits[0]).Should(Equal(0))
			Ω(applyHits[1]).Should(Equal(2))
			Ω(applyHits[2]).Should(Equal(2))
			Ω(applyHits[3]).Should(Equal(1))

			Ω(ops[0].Name).Should(Equal("foo"))
			Ω(ops[1].Name).Should(Equal("foobar1"))
			Ω(ops[2].Name).Should(Equal("bar"))
			Ω(ops[3].Name).Should(Equal("bar"))
		})

	})

	Context("Manual load", func() {

		It("Load twice", func() {
			loadNoError()
			Ω(Load(filename)).Should(MatchError("[config] can not call Load() twice"))
		})

	})

	Context("SetDefaultOptionForTest", func() {

		It("Test", func() {
			Register("op1", newFakeOption(0))
			Register("op2", newFakeOption(1))
			Register("op3", newFakeOption(2))
			SetDefaultOptionForTest("op2", &FakeOption{
				idx:   5,
				Name:  "override",
				Count: 4,
			})
			Ω(Load("")).Should(Succeed())
			Ω(initHits[0]).Should(Equal(1))
			Ω(initHits[5]).Should(Equal(1))
			Ω(initHits[2]).Should(Equal(1))
			Ω(initHits[1]).Should(Equal(0))
		})

		It("Detect wrong config name", func() {
			Register("op1", newFakeOption(0))
			SetDefaultOptionForTest("wrongName", &FakeOption{})
			Ω(Load("")).Should(MatchError("[config] overridden option \"wrongName\" not used, wrong option name?"))
		})

	})

})

var (
	ops                 []*FakeOption
	initErrors          []error
	initHits, applyHits []int
)

type FakeOption struct {
	idx   int
	Name  string
	Count int
	Panic bool
}

func newFakeOption(idx int) OptionCreator {
	return func() Option {
		return &FakeOption{idx: idx, Name: "bar", Count: 3}
	}
}

func (o *FakeOption) Init() error {
	o.initArrays()

	initHits[o.idx]++
	Ω(ops[o.idx]).Should(BeNil())
	ops[o.idx] = o

	if len(initErrors) > o.idx {
		return initErrors[o.idx]
	}
	return nil
}

func (o *FakeOption) Apply() {
	o.initArrays()

	ops[o.idx] = o
	applyHits[o.idx]++

	if o.Panic {
		panic("error")
	}
}

func (o *FakeOption) initArrays() {
	if len(ops) <= o.idx {
		newOps := make([]*FakeOption, o.idx+1)
		copy(newOps, ops)
		ops = newOps

		newInits := make([]int, o.idx+1)
		copy(newInits, initHits)
		initHits = newInits

		newApplys := make([]int, o.idx+1)
		copy(newApplys, applyHits)
		applyHits = newApplys
	}
}

type testService struct {
	configDir string
}

func (ts *testService) ResolveConfigFile(file string) string {
	return filepath.Join(ts.configDir, file)
}

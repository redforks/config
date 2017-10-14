package config_test

import (
	. "github.com/redforks/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redforks/testing/reset"
)

type SimpleOption struct {
	Name string
}

func (o SimpleOption) Init() error {
	return nil
}

func (o SimpleOption) Apply() {
}

func newSimpleOption(name string) OptionCreator {
	return func() Option {
		return SimpleOption{name}
	}
}

var _ = Describe("Dump", func() {

	BeforeEach(func() {
		reset.Enable()
	})

	AfterEach(func() {
		ResetInternal()
		reset.Disable()
	})

	It("No options", func() {
		Ω(DumpDefaultOptions()).Should(Equal("# default options for test\n\n"))
	})

	It("One Option", func() {
		Register("foo", newSimpleOption("bar"))
		Ω(DumpDefaultOptions()).Should(Equal(`# default options for test

# [foo]
# Name = "bar"
`))
	})

	It("Multiple options", func() {
		Register("foo", newSimpleOption("bar"))
		Register("bar", newSimpleOption("foobar"))
		Ω(DumpDefaultOptions()).Should(Or(Equal(`# default options for test

# [foo]
# Name = "bar"

# [bar]
# Name = "foobar"
`), Equal(`# default options for test

# [bar]
# Name = "foobar"

# [foo]
# Name = "bar"
`)))
	})

})

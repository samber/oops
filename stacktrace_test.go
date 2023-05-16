package oops

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func a() *oopsStacktrace {
	return b()
}

func b() *oopsStacktrace {
	return c()
}

func c() *oopsStacktrace {
	return d()
}

func d() *oopsStacktrace {
	return e()
}

func e() *oopsStacktrace {
	return f()
}

func f() *oopsStacktrace {
	return newStacktrace()
}

func TestStacktrace(t *testing.T) {
	is := assert.New(t)

	frames := a()

	is.NotNil(frames)

	if frames != nil {
		for _, f := range *frames {
			is.True(strings.Contains(f.file, "github.com/samber/oops/stacktrace_test.go"))
		}

		is.Len(*frames, 7, "expected 7 frames")

		if len(*frames) == 7 {
			is.Equal("f", (*frames)[0].function)
			is.Equal("e", (*frames)[1].function)
			is.Equal("d", (*frames)[2].function)
			is.Equal("c", (*frames)[3].function)
			is.Equal("b", (*frames)[4].function)
			is.Equal("a", (*frames)[5].function)
			is.Equal("TestStacktrace", (*frames)[6].function)
		}
	}
}

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
	return newStacktrace("1234")
}

func TestStacktrace(t *testing.T) {
	is := assert.New(t)

	st := a()

	is.NotNil(st)
	is.Equal("1234", st.span)

	if st.frames != nil {
		for _, f := range st.frames {
			is.True(strings.Contains(f.file, "github.com/samber/oops/stacktrace_test.go"))
		}

		is.Len(st.frames, 7, "expected 7 frames")

		if len(st.frames) == 7 {
			is.Equal("f", (st.frames)[0].function)
			is.Equal("e", (st.frames)[1].function)
			is.Equal("d", (st.frames)[2].function)
			is.Equal("c", (st.frames)[3].function)
			is.Equal("b", (st.frames)[4].function)
			is.Equal("a", (st.frames)[5].function)
			is.Equal("TestStacktrace", (st.frames)[6].function)
		}
	}
}

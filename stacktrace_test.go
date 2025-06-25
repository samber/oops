package oops

import (
	"runtime"
	"runtime/debug"
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

	bi, ok := debug.ReadBuildInfo()
	is.True(ok)

	path := strings.Replace(bi.Path, ".test", "", 1) // starting go1.24, go adds ".test" to the path when running tests

	if st.frames != nil {
		for _, f := range st.frames {
			is.Contains(f.file, path, "frame file %s should contain %s", f.file, path)
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

func TestShortFuncNameExtended(t *testing.T) {
	is := assert.New(t)

	// Test with a real function
	pc, _, _, ok := runtime.Caller(0)
	is.True(ok)
	f := runtime.FuncForPC(pc)
	is.NotNil(f)

	result := shortFuncName(f)
	is.NotEmpty(result)
	is.Contains(result, "TestShortFuncNameExtended")

	// Test with nil function
	result2 := shortFuncName(nil)
	is.Equal("", result2)
}

func TestOopsStacktraceError(t *testing.T) {
	is := assert.New(t)

	// Test stacktrace Error method - returns formatted stacktrace, not span
	st := &oopsStacktrace{span: "test"}
	err := st.Error()
	is.Equal("", err) // Empty because no frames

	// Test with frames
	frame := oopsStacktraceFrame{
		file:     "test.go",
		line:     10,
		function: "testFunc",
	}
	st2 := &oopsStacktrace{span: "test", frames: []oopsStacktraceFrame{frame}}
	err2 := st2.Error()
	is.Contains(err2, "test.go:10 testFunc()")
}

func TestOopsStacktraceString(t *testing.T) {
	is := assert.New(t)

	// Test with empty frames
	st := &oopsStacktrace{span: "test", frames: []oopsStacktraceFrame{}}
	result := st.String("")
	is.Empty(result)

	// Test with frames
	frame := oopsStacktraceFrame{
		file:     "test.go",
		line:     10,
		function: "testFunc",
	}
	st2 := &oopsStacktrace{span: "test", frames: []oopsStacktraceFrame{frame}}
	result2 := st2.String("")
	is.Contains(result2, "test.go:10 testFunc()")

	// Test with deepest frame
	result3 := st2.String("test.go:10 testFunc()")
	is.Empty(result3)
}

func TestOopsStacktraceFrameString(t *testing.T) {
	is := assert.New(t)

	// Test with function
	frame := &oopsStacktraceFrame{
		file:     "test.go",
		line:     10,
		function: "testFunc",
	}
	result := frame.String()
	is.Equal("test.go:10 testFunc()", result)

	// Test without function
	frame2 := &oopsStacktraceFrame{
		file: "test.go",
		line: 10,
	}
	result2 := frame2.String()
	is.Equal("test.go:10", result2)
}

package oops

import (
	"errors"
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
	return newStacktrace("1234", 0)
}

func TestStacktrace(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	st := a()

	is.NotNil(st)
	is.Equal("1234", st.span)

	bi, ok := debug.ReadBuildInfo()
	is.True(ok)

	if !strings.Contains(bi.Path, "github.com/samber/oops") {
		t.Skip("This test is meant to run on oops main repo")
	}

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
	t.Parallel()

	// Test with a real function
	pc, _, _, ok := runtime.Caller(0)
	is.True(ok)
	f := runtime.FuncForPC(pc)
	is.NotNil(f)

	result := shortFuncName(f.Name())
	is.NotEmpty(result)
	is.Contains(result, "TestShortFuncNameExtended")

	// Test with empty function name
	result2 := shortFuncName("")
	is.Empty(result2)
}

func TestOopsStacktraceError(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	// Test stacktrace Error method - returns formatted stacktrace, not span
	st := &oopsStacktrace{span: "test"}
	err := st.Error()
	is.Empty(err) // Empty because no frames

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
	t.Parallel()

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
	t.Parallel()

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

// helperWrap wraps oops.Wrap without any caller skip — it should appear in the stack trace.
func helperWrap(err error) error {
	return Wrap(err)
}

// helperWrapSkip wraps oops.CallerSkip(1).Wrap — it skips 1 user frame (itself),
// so helperWrapSkip itself should NOT appear in the stack trace.
// CallerSkip(1) means "skip 1 user frame", where the base offset already accounts
// for runtime.Callers → newStacktrace → builder_method (internalFrameDepth=3).
func helperWrapSkip(err error) error {
	return CallerSkip(1).Wrap(err)
}

// frameNames returns the list of short function names from an OopsError's stack trace,
// applying output-time FrameSkip filtering (via StackFrames()).
func frameNames(err error) []string {
	oopsErr, ok := err.(OopsError)
	if !ok {
		return nil
	}
	frames := oopsErr.StackFrames()
	names := make([]string, 0, len(frames))
	for _, f := range frames {
		names = append(names, f.Function)
	}
	return names
}

// --- CallerSkip tests ---

func TestCallerSkip_NoSkip(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	base := errors.New("base")
	err := helperWrap(base)

	names := frameNames(err)
	is.NotEmpty(names)
	is.Contains(names, "helperWrap", "helperWrap should appear in frames when no skip is applied")
}

func TestCallerSkip_SkipOne(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	base := errors.New("base")
	err := helperWrapSkip(base)

	names := frameNames(err)
	is.NotEmpty(names)
	is.NotContains(names, "helperWrapSkip", "helperWrapSkip should NOT appear in frames when CallerSkip(1) is used")
	// The test function itself should be the first user frame instead
	is.Contains(names, "TestCallerSkip_SkipOne", "the calling test function should appear in frames")
}

func TestCallerSkip_LastCallWins(t *testing.T) {
	is := assert.New(t)
	t.Parallel()

	base := errors.New("base")
	// Chain CallerSkip(100) then CallerSkip(0): the last call should win (skip=0).
	// With skip=0, no user frames are skipped, so the test function itself should appear.
	// With skip=100, all frames would be skipped and the test function would not appear.
	// The test function's presence here confirms that skip=0 (the last value) was used,
	// not skip=100.
	err := CallerSkip(100).CallerSkip(0).Wrap(base)

	names := frameNames(err)
	is.NotEmpty(names)
	is.Contains(names, "TestCallerSkip_LastCallWins", "with skip=0 the test function should still be in frames — confirms last CallerSkip wins")
}

// --- FrameSkip tests ---

// gWrapper is a helper used by TestFrameSkip_ByFunction to verify function-name filtering.
func gWrapper(err error) error {
	return Wrap(err)
}

func TestFrameSkip_ByFunction(t *testing.T) {
	// Modifies global framesSkip — do not run in parallel.
	is := assert.New(t)

	originalFramesSkip := framesSkip
	defer func() { framesSkip = originalFramesSkip }()

	// Confirm gWrapper appears before registering the skip.
	base := errors.New("base")
	errBefore := gWrapper(base)
	namesBefore := frameNames(errBefore)
	is.Contains(namesBefore, "gWrapper", "gWrapper should appear before FrameSkip is registered")

	// Create error BEFORE registering the skip — filtering should still apply at output time.
	errCreatedBeforeSkip := gWrapper(base)
	FrameSkip("", "gWrapper")
	namesBeforeSkip := frameNames(errCreatedBeforeSkip)
	is.NotContains(namesBeforeSkip, "gWrapper", "FrameSkip should apply to errors created before registration")

	errAfter := gWrapper(base)
	namesAfter := frameNames(errAfter)
	is.NotContains(namesAfter, "gWrapper", "gWrapper should NOT appear after FrameSkip(\"\", \"gWrapper\") is registered")
}

func TestFrameSkip_ByFile(t *testing.T) {
	// Modifies global framesSkip — do not run in parallel.
	is := assert.New(t)

	originalFramesSkip := framesSkip
	defer func() { framesSkip = originalFramesSkip }()

	// Get the raw file path as runtime.CallersFrames would report it.
	_, thisFile, _, _ := runtime.Caller(0)

	// Confirm helperWrap appears before registering the skip.
	base := errors.New("base")
	errBefore := helperWrap(base)
	is.Contains(frameNames(errBefore), "helperWrap", "helperWrap should appear before FrameSkip is registered")

	// Create error BEFORE registering the skip — filtering should still apply at output time.
	errCreatedBeforeSkip := helperWrap(base)
	// Register a skip for this test file (helperWrap lives in the same file).
	FrameSkip(thisFile, "")
	namesBeforeSkip := frameNames(errCreatedBeforeSkip)
	is.NotContains(namesBeforeSkip, "helperWrap", "FrameSkip should apply to errors created before registration")

	errAfter := helperWrap(base)
	namesAfter := frameNames(errAfter)
	is.NotContains(namesAfter, "helperWrap", "helperWrap should NOT appear after FrameSkip by file path")
}

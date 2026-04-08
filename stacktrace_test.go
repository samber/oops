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

// helperWrapSkip wraps oops.CallerSkip(3).Wrap — it skips itself and the oops internals,
// so helperWrapSkip itself should NOT appear in the stack trace.
// The call chain into runtime.Callers is: newStacktrace (frame 2) → Wrap (frame 3) → helperWrapSkip (frame 4).
// runtime.Callers(1+skip, ...) already skips frame 1 (runtime.Callers itself), so skip=3 means
// we start capturing at frame 4 (helperWrapSkip), which is the first user frame to be dropped.
func helperWrapSkip(err error) error {
	return CallerSkip(3).Wrap(err)
}

// frameNames returns the list of short function names from an OopsError's stack trace.
func frameNames(err error) []string {
	oopsErr, ok := err.(OopsError)
	if !ok {
		return nil
	}
	names := make([]string, 0, len(oopsErr.stacktrace.frames))
	for _, f := range oopsErr.stacktrace.frames {
		names = append(names, f.function)
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
	// Chain CallerSkip(100) then CallerSkip(1): the last call should win (skip=1).
	// With skip=1 the test function should still be present.
	// With skip=100 the test function would not appear (all frames skipped), so its presence
	// here confirms that skip=1 (the last value) was used and not skip=100.
	err := CallerSkip(100).CallerSkip(1).Wrap(base)

	names := frameNames(err)
	is.NotEmpty(names)
	is.Contains(names, "TestCallerSkip_LastCallWins", "with skip=1 the test function should still be in frames — confirms last CallerSkip wins")
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

	// Register a skip for the function name "gWrapper".
	FrameSkip("", "gWrapper")

	errAfter := gWrapper(base)
	namesAfter := frameNames(errAfter)
	is.NotContains(namesAfter, "gWrapper", "gWrapper should NOT appear after FrameSkip(\"\", \"gWrapper\") is registered")
}

func TestFrameSkip_ByFile(t *testing.T) {
	// Modifies global framesSkip — do not run in parallel.
	is := assert.New(t)

	originalFramesSkip := framesSkip
	defer func() { framesSkip = originalFramesSkip }()

	// Capture an error to discover the exact file path of this test file.
	base := errors.New("base")
	errSample := helperWrap(base)
	sampleNames := frameNames(errSample)
	is.NotEmpty(sampleNames)

	// Find the file path corresponding to "helperWrap" frame.
	oopsErr, ok := errSample.(OopsError)
	is.True(ok)

	var testFilePath string
	for _, fr := range oopsErr.stacktrace.frames {
		if fr.function == "helperWrap" {
			testFilePath = fr.file
			break
		}
	}
	is.NotEmpty(testFilePath, "should have found the file path for helperWrap")

	// Register a skip for that exact file path.
	FrameSkip(testFilePath, "")

	errAfter := helperWrap(base)
	namesAfter := frameNames(errAfter)
	is.NotContains(namesAfter, "helperWrap", "helperWrap should NOT appear after FrameSkip by file path")
}

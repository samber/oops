package oops

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

///
/// Inspired by palantir/stacktrace repo
/// -> https://github.com/palantir/stacktrace/blob/master/stacktrace.go
/// -> Apache 2.0 LICENSE
///

type fake struct{}

var (
	StackTraceMaxDepth int = 10

	packageName = reflect.TypeOf(fake{}).PkgPath()
)

type oopsStacktraceFrame struct {
	pc       uintptr
	file     string
	function string
	line     int
}

func (frame *oopsStacktraceFrame) String() string {
	currentFrame := fmt.Sprintf("%v:%v", frame.file, frame.line)
	if frame.function != "" {
		currentFrame = fmt.Sprintf("%v:%v %v()", frame.file, frame.line, frame.function)
	}

	return currentFrame
}

type oopsStacktrace struct {
	span   string
	frames []oopsStacktraceFrame
}

func (st *oopsStacktrace) Error() string {
	return st.String("")
}

func (st *oopsStacktrace) String(deepestFrame string) string {
	var str string

	newline := func() {
		if str != "" && !strings.HasSuffix(str, "\n") {
			str += "\n"
		}
	}

	for _, frame := range st.frames {
		if frame.file != "" {
			currentFrame := frame.String()
			if currentFrame == deepestFrame {
				break
			}

			newline()
			str += "  --- at " + currentFrame
		}
	}

	return str
}

func (st *oopsStacktrace) Source() (string, []string) {
	if len(st.frames) == 0 {
		return "", []string{}
	}

	firstFrame := st.frames[0]

	header := firstFrame.String()
	body := getSourceFromFrame(firstFrame)

	return header, body
}

func newStacktrace(span string) *oopsStacktrace {
	frames := []oopsStacktraceFrame{}

	// We loop until we have StackTraceMaxDepth frames or we run out of frames.
	// Frames from this package are skipped.
	for i := 1; len(frames) < StackTraceMaxDepth; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		file = removeGoPath(file)

		f := runtime.FuncForPC(pc)
		if f == nil {
			break
		}
		function := shortFuncName(f)

		packageNameExamples := packageName + "/examples/"

		isGoPkg := len(runtime.GOROOT()) > 0 && strings.Contains(file, runtime.GOROOT()) // skip frames in GOROOT if it's set
		isOopsPkg := strings.Contains(file, packageName)                                 // skip frames in this package
		isExamplePkg := strings.Contains(file, packageNameExamples)                      // do not skip frames in this package examples
		isTestPkg := strings.Contains(file, "_test.go")                                  // do not skip frames in tests

		if !isGoPkg && (!isOopsPkg || isExamplePkg || isTestPkg) {
			frames = append(frames, oopsStacktraceFrame{
				pc:       pc,
				file:     file,
				function: function,
				line:     line,
			})
		}
	}

	return &oopsStacktrace{
		span:   span,
		frames: frames,
	}
}

func shortFuncName(f *runtime.Func) string {
	// f.Name() is like one of these:
	// - "github.com/palantir/shield/package.FuncName"
	// - "github.com/palantir/shield/package.Receiver.MethodName"
	// - "github.com/palantir/shield/package.(*PtrReceiver).MethodName"
	longName := f.Name()

	withoutPath := longName[strings.LastIndex(longName, "/")+1:]
	withoutPackage := withoutPath[strings.Index(withoutPath, ".")+1:]

	shortName := withoutPackage
	shortName = strings.Replace(shortName, "(", "", 1)
	shortName = strings.Replace(shortName, "*", "", 1)
	shortName = strings.Replace(shortName, ")", "", 1)

	return shortName
}

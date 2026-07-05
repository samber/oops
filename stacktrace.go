package oops

import (
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/samber/lo"
)

///
/// Stack trace generation and processing functionality.
///
/// This module provides comprehensive stack trace capture and formatting
/// capabilities for oops.OopsError instances. It includes intelligent
/// frame filtering to exclude irrelevant stack frames and provides
/// formatted output suitable for debugging and logging.
///
/// Inspired by palantir/stacktrace repo
/// -> https://github.com/palantir/stacktrace/blob/master/stacktrace.go
/// -> Apache 2.0 LICENSE
///

// fake is a dummy struct used to determine the current package name
// for stack trace filtering. This allows the package to identify
// and filter out frames from its own code while preserving frames
// from user code and examples.
type fake struct{}

// internalFrameDepth is the number of raw frames between runtime.Callers and
// the first user frame in the standard call chain:
//
//	runtime.Callers → newStacktrace → builder_method → user_code
//
// Builder terminal methods (Wrap, Errorf, etc.) pass (internalFrameDepth-1)+callerSkip
// to newStacktrace so that CallerSkip(n) means "skip n user frames".
const internalFrameDepth = 3

// Global configuration for stack trace generation.
var (
	// StackTraceMaxDepth controls the maximum number of stack frames
	// to capture in a stack trace. This prevents stack traces from
	// becoming excessively long while still providing sufficient
	// context for debugging.
	//
	// The default value of 10 provides a good balance between
	// detail and readability. For deep call stacks, this will
	// capture the most recent 10 frames, which typically include
	// the most relevant debugging information.
	StackTraceMaxDepth = 10

	// framesSkip is a list of frame patterns used to filter out frames from stack traces.
	// Each entry's file and function fields are matched using strings.Contains against
	// the raw runtime.CallersFrames values. Register patterns via the FrameSkip() function.
	framesSkip []oopsStacktraceFrame

	// packageName stores the current package name for frame filtering.
	// This is determined at package initialization time and used
	// to identify frames that should be excluded from stack traces.
	packageName = reflect.TypeOf(fake{}).PkgPath()

	// packageNameExamples and goroot are filtering inputs that never change
	// at runtime; computing them once here keeps them out of the per-frame
	// resolution loop (runtime.GOROOT in particular re-reads an env var on
	// every call).
	packageNameExamples = packageName + "/examples/"
	goroot              = runtime.GOROOT()
)

// oopsStacktraceFrame represents a single frame in a stack trace.
// Each frame contains information about a function call in the
// call stack, including the program counter, file path, function
// name, and line number.
type oopsStacktraceFrame struct {
	pc          uintptr // Program counter for the function call
	file        string  // cleaned path via removeGoPath (for display)
	function    string  // short name via shortFuncName (for display)
	line        int     // Line number in the file where the call occurred
	rawFile     string  // raw frame.File from runtime.CallersFrames (for matching)
	rawFunction string  // raw frame.Function from runtime.CallersFrames (for matching)
}

// String returns a formatted string representation of the stack frame.
// The format follows the standard Go stack trace format: "file:line function()"
// or just "file:line" if no function name is available.
//
// This method is used for both individual frame display and as part
// of complete stack trace formatting.
//
// Example output:
//
//	"main.go:42 main()"
//	"handler.go:15 processRequest()"
func (frame *oopsStacktraceFrame) String() string {
	if frame.function == "" {
		return frame.file + ":" + strconv.Itoa(frame.line)
	}

	return frame.file + ":" + strconv.Itoa(frame.line) + " " + frame.function + "()"
}

// oopsStacktrace represents a complete stack trace with multiple frames.
// It contains a span identifier for correlation and an ordered list
// of stack frames representing the call hierarchy.
//
// Program counters are captured eagerly at error-creation time (the stack is
// gone afterwards), but resolving them into file/function/line strings is
// deferred to the first read: symbolization via runtime.CallersFrames is the
// dominant cost of error creation, and most errors are created, checked
// against nil, and discarded without their stack trace ever being formatted.
type oopsStacktrace struct {
	span string

	// pcs holds the raw program counters captured by runtime.Callers,
	// consumed (and set to nil) by resolve on first access.
	pcs []uintptr
	// maxDepth snapshots StackTraceMaxDepth at capture time so a later
	// change of the global does not alter already-captured traces.
	maxDepth int

	once   sync.Once
	frames []oopsStacktraceFrame // Ordered list of stack frames (most recent first)
}

// resolvedFrames symbolizes the captured program counters on first call and
// returns the filtered frames. Safe for concurrent use.
func (st *oopsStacktrace) resolvedFrames() []oopsStacktraceFrame {
	st.once.Do(st.resolve)
	return st.frames
}

// resolve converts raw program counters into filtered, display-ready frames.
// It applies the same filtering rules that previously ran at capture time:
// frames from GOROOT and from this package are excluded (except examples and
// tests), and at most maxDepth frames are kept.
func (st *oopsStacktrace) resolve() {
	if len(st.pcs) == 0 {
		st.pcs = nil
		return
	}

	capDepth := min(st.maxDepth, len(st.pcs))
	frames := make([]oopsStacktraceFrame, 0, capDepth)

	// Iterate over the captured frames
	iter := runtime.CallersFrames(st.pcs)
	for len(frames) < st.maxDepth {
		frame, more := iter.Next()

		// Clean up the file path by removing Go path prefixes
		file := removeGoPath(frame.File)

		// Apply frame filtering logic
		isGoPkg := len(goroot) > 0 && strings.Contains(file, goroot) // skip frames in GOROOT if it's set
		isOopsPkg := strings.Contains(file, packageName)             // skip frames in this package
		isExamplePkg := strings.Contains(file, packageNameExamples)  // do not skip frames in this package examples
		isTestPkg := strings.Contains(file, "_test.go")              // do not skip frames in tests

		// Include frame if it passes all filtering criteria
		if !isGoPkg && (!isOopsPkg || isExamplePkg || isTestPkg) {
			frames = append(frames, oopsStacktraceFrame{
				pc: frame.PC,
				// Extract a short, readable function name — only for frames
				// that are kept, since the runtime/oops frames filtered out
				// above would pay the string processing for nothing.
				function:    shortFuncName(frame.Function),
				file:        file,
				line:        frame.Line,
				rawFile:     frame.File,
				rawFunction: frame.Function,
			})
		}

		if !more {
			break
		}
	}

	st.frames = frames
	st.pcs = nil
}

// Error implements the error interface for stack traces.
// This allows stack traces to be used directly as errors if needed.
func (st *oopsStacktrace) Error() string {
	return st.String("")
}

// String returns a formatted string representation of the complete stack trace.
// The output includes all frames in the stack trace, formatted with proper
// indentation and structure for readability.
//
// The deepestFrame parameter is used to avoid duplicate frames when
// combining stack traces from nested errors. When a frame matches
// the deepestFrame, the formatting stops to prevent redundancy.
//
// Example output:
//
//	"  --- at main.go:42 main()
//	   --- at handler.go:15 processRequest()
//	   --- at server.go:123 handleHTTP()"
func (st *oopsStacktrace) String(deepestFrame string) string {
	var str strings.Builder

	// Helper function to add newlines between frames
	newline := func() {
		if str.Len() != 0 {
			tmpStr := str.String()
			if tmpStr[len(tmpStr)-1] != '\n' {
				str.WriteRune('\n')
			}
		}
	}

	// Iterate through all frames and format them
	for _, frame := range st.resolvedFrames() {
		if frame.file != "" {
			currentFrame := frame.String()

			// Stop if we've reached the deepest frame to avoid duplication
			if currentFrame == deepestFrame {
				break
			}

			newline()
			str.WriteString("  --- at ")
			str.WriteString(currentFrame)
		}
	}

	return str.String()
}

// Source returns the source code context for the first frame in the stack trace.
// This method provides both a header (file:line function()) and the actual
// source code lines around the error location for enhanced debugging.
//
// The source code includes a configurable number of lines before and after
// the error location, with the error line highlighted. This is particularly
// useful for understanding the context in which an error occurred.
//
// Performance: This method involves file I/O operations to read source code,
// which may have performance implications for frequently called code paths.
// The results are cached to minimize repeated file reads.
//
// Returns:
//   - header: Formatted string like "main.go:42 main()"
//   - body: Slice of strings containing source code lines with line numbers
func (st *oopsStacktrace) Source() (string, []string) {
	frames := st.resolvedFrames()
	if len(frames) == 0 {
		return "", []string{}
	}

	firstFrame := frames[0]

	header := firstFrame.String()
	body := getSourceFromFrame(firstFrame)

	return header, body
}

// newStacktrace creates a new stack trace by capturing the current call stack.
// This function walks up the call stack starting from the caller of this
// function and captures frame information while applying intelligent filtering.
//
// The function implements sophisticated frame filtering to provide relevant
// debugging information while excluding noise:
// - Excludes frames from the Go standard library (GOROOT)
// - Excludes frames from this package (except examples and tests)
// - Limits the number of frames to StackTraceMaxDepth
// - Includes frames from user code and package examples/tests
//
// Performance: This function has O(d) complexity where d is the depth
// of the call stack, with additional overhead for frame filtering and
// function name processing.
//
// Example usage:
//
//	stack := newStacktrace("span-123", 0)
//	fmt.Println(stack.String(""))
func newStacktrace(span string, skip int) *oopsStacktrace {
	// Capture all program counters in a single batch call.
	// The buffer must be large enough to hold the desired user frames PLUS the
	// oops-internal and runtime frames that will be filtered out during
	// resolution. Cap at 512 to avoid huge allocations when StackTraceMaxDepth
	// is set to a very large value.
	bufSize := min(StackTraceMaxDepth*3+20, 512)
	pcs := make([]uintptr, bufSize)
	n := runtime.Callers(1+skip, pcs)

	// Symbolization and filtering happen lazily in resolve, on first read.
	return &oopsStacktrace{
		span:     span,
		pcs:      pcs[:n],
		maxDepth: StackTraceMaxDepth,
	}
}

// shortFuncName extracts a short, readable function name from a full function
// name string (as returned by runtime.Frame.Function). This function processes
// the full function name (which includes package path and receiver information)
// and returns a simplified version suitable for display in stack traces.
//
// The function handles various function name formats:
// - Package functions: "github.com/user/pkg.FuncName" -> "FuncName"
// - Methods: "github.com/user/pkg.Receiver.MethodName" -> "MethodName"
// - Pointer methods: "github.com/user/pkg.(*PtrReceiver).MethodName" -> "MethodName"
//
// Example transformations:
//
//	"github.com/user/app.(*Handler).ProcessRequest" -> "ProcessRequest"
//	"main.main" -> "main"
//	"github.com/user/pkg.helper" -> "helper"

// ptrReceiverReplacer strips pointer receiver syntax characters from function names.
var ptrReceiverReplacer = strings.NewReplacer("(", "", "*", "", ")", "")

func shortFuncName(longName string) string {
	// longName is the full function name including package path
	// Examples of possible formats:
	// - "github.com/palantir/shield/package.FuncName"
	// - "github.com/palantir/shield/package.Receiver.MethodName"
	// - "github.com/palantir/shield/package.(*PtrReceiver).MethodName"

	// Remove the package path by finding the last "/" and taking everything after it
	withoutPath := longName[strings.LastIndex(longName, "/")+1:]

	// Remove the package name by finding the first "." and taking everything after it
	withoutPackage := withoutPath[strings.Index(withoutPath, ".")+1:]

	// Clean up the function name by removing parentheses and asterisks
	// that are part of pointer receiver syntax
	return ptrReceiverReplacer.Replace(withoutPackage)
}

// applyFrameSkip returns a copy of frames with any entries matching framesSkip patterns removed.
// Matching uses strings.Contains against the raw runtime.CallersFrames values stored in
// rawFile and rawFunction. An empty pattern field is a wildcard (matches anything).
func applyFrameSkip(frames []oopsStacktraceFrame) []oopsStacktraceFrame {
	if len(framesSkip) == 0 {
		return frames
	}
	filtered := make([]oopsStacktraceFrame, 0, len(frames))
	for _, f := range frames {
		skip := false
		for _, pattern := range framesSkip {
			fileMatch := pattern.file == "" || strings.Contains(f.rawFile, pattern.file)
			funcMatch := pattern.function == "" || strings.Contains(f.rawFunction, pattern.function)
			if fileMatch && funcMatch {
				skip = true
				break
			}
		}
		if !skip {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func framesToStacktraceBlocks(blocks []outputBlock) []string {
	output := make([]string, 0, len(blocks))
	shownFrames := make(map[string]bool)

	for _, e := range blocks {
		err := ""
		if e.err != nil {
			err = e.err.Error()
		}
		msg := coalesceOrEmpty(e.msg, err, "Error")

		// Build stacktrace for this error, avoiding already shown frames
		var frameLines []string
		firstFrame := true // we always show the first frame, because the PC of a recursive function might appear multiple time.
		for _, frame := range e.frames {
			frameStr := frame.String()
			if !shownFrames[frameStr] || firstFrame {
				frameLines = append(frameLines, "  --- at "+frameStr)
				shownFrames[frameStr] = true
			}
			firstFrame = false
		}

		output = append(output, msg+"\n"+strings.Join(frameLines, "\n"))
	}

	slices.Reverse(output)
	return output
}

func framesToSourceBlocks(blocks []outputBlock) []string {
	output := make([][]string, 0, len(blocks))

	for _, b := range blocks {
		st := oopsStacktrace{frames: b.frames}
		header, body := st.Source()

		if b.msg != "" {
			header = b.msg + "\n" + header
		}

		if header != "" && len(body) > 0 {
			output = append(output, append([]string{header}, body...))
		}
	}

	slices.Reverse(output)
	return lo.Map(output, func(items []string, _ int) string {
		return strings.Join(items, "\n")
	})
}

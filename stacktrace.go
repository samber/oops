package oops

import (
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"strings"

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
	currentFrame := fmt.Sprintf("%v:%v", frame.file, frame.line)
	if frame.function != "" {
		currentFrame = fmt.Sprintf("%v:%v %v()", frame.file, frame.line, frame.function)
	}

	return currentFrame
}

// oopsStacktrace represents a complete stack trace with multiple frames.
// It contains a span identifier for correlation and an ordered list
// of stack frames representing the call hierarchy.
type oopsStacktrace struct {
	span   string                // Unique identifier for the stack trace
	frames []oopsStacktraceFrame // Ordered list of stack frames (most recent first)
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
	for _, frame := range st.frames {
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
	if len(st.frames) == 0 {
		return "", []string{}
	}

	firstFrame := st.frames[0]

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
	frames := make([]oopsStacktraceFrame, 0, StackTraceMaxDepth)

	// Capture all program counters in a single batch call.
	// The buffer must be large enough to hold the desired user frames PLUS the
	// oops-internal and runtime frames that will be filtered out during iteration.
	// Cap at 512 to avoid huge allocations when StackTraceMaxDepth is set to a
	// very large value.
	bufSize := min(StackTraceMaxDepth*3+20, 512)
	pcs := make([]uintptr, bufSize)
	n := runtime.Callers(1+skip, pcs)
	pcs = pcs[:n]

	// Define package name patterns for filtering (computed once, outside the loop)
	packageNameExamples := packageName + "/examples/"
	goroot := runtime.GOROOT()

	// Iterate over the captured frames
	iter := runtime.CallersFrames(pcs)
	for len(frames) < StackTraceMaxDepth {
		frame, more := iter.Next()

		// Clean up the file path by removing Go path prefixes
		file := removeGoPath(frame.File)

		// Extract a short, readable function name
		function := shortFuncName(frame.Function)

		// Apply frame filtering logic
		isGoPkg := len(goroot) > 0 && strings.Contains(file, goroot) // skip frames in GOROOT if it's set
		isOopsPkg := strings.Contains(file, packageName)             // skip frames in this package
		isExamplePkg := strings.Contains(file, packageNameExamples)  // do not skip frames in this package examples
		isTestPkg := strings.Contains(file, "_test.go")              // do not skip frames in tests

		// Include frame if it passes all filtering criteria
		if !isGoPkg && (!isOopsPkg || isExamplePkg || isTestPkg) {
			frames = append(frames, oopsStacktraceFrame{
				pc:          frame.PC,
				file:        file,
				function:    function,
				line:        frame.Line,
				rawFile:     frame.File,
				rawFunction: frame.Function,
			})
		}

		if !more {
			break
		}
	}

	return &oopsStacktrace{
		span:   span,
		frames: frames,
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
	shortName := withoutPackage
	shortName = strings.Replace(shortName, "(", "", 1) // Remove opening parenthesis
	shortName = strings.Replace(shortName, "*", "", 1) // Remove asterisk
	shortName = strings.Replace(shortName, ")", "", 1) // Remove closing parenthesis

	return shortName
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

func framesToStacktraceBlocks(blocks []lo.Tuple3[error, string, []oopsStacktraceFrame]) []string {
	output := make([]string, 0, len(blocks))
	shownFrames := make(map[string]bool)

	for _, e := range blocks {
		err := lo.TernaryF(e.A != nil, func() string { return e.A.Error() }, func() string { return "" })
		msg := coalesceOrEmpty(e.B, err, "Error")

		// Build stacktrace for this error, avoiding already shown frames
		var frameLines []string
		firstFrame := true // we always show the first frame, because the PC of a recursive function might appear multiple time.
		for _, frame := range e.C {
			frameStr := frame.String()
			if !shownFrames[frameStr] || firstFrame {
				frameLines = append(frameLines, "  --- at "+frame.String())
				shownFrames[frameStr] = true
			}
			firstFrame = false
		}

		stacktraceStr := strings.Join(frameLines, "\n")
		block := fmt.Sprintf("%s\n%s", msg, stacktraceStr)

		output = append(output, block)
	}

	slices.Reverse(output)
	return output
}

func framesToSourceBlocks(blocks []lo.Tuple2[string, *oopsStacktrace]) []string {
	output := [][]string{}

	for _, e := range blocks {
		header, body := e.B.Source()

		if e.A != "" {
			header = fmt.Sprintf("%s\n%s", e.A, header)
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

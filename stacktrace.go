package oops

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
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

// Global configuration for stack trace generation
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
	StackTraceMaxDepth int = 10

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
	pc       uintptr // Program counter for the function call
	file     string  // File path where the call occurred
	function string  // Name of the function being called
	line     int     // Line number in the file where the call occurred
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
	var str string

	// Helper function to add newlines between frames
	newline := func() {
		if str != "" && !strings.HasSuffix(str, "\n") {
			str += "\n"
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
			str += "  --- at " + currentFrame
		}
	}

	return str
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
//	stack := newStacktrace("span-123")
//	fmt.Println(stack.String(""))
func newStacktrace(span string) *oopsStacktrace {
	frames := []oopsStacktraceFrame{}

	// Walk up the call stack starting from the caller of this function
	// Continue until we have enough frames or run out of stack frames
	for i := 1; len(frames) < StackTraceMaxDepth; i++ {
		// Get frame information from the runtime
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break // No more frames available
		}

		// Clean up the file path by removing Go path prefixes
		file = removeGoPath(file)

		// Get function information for the program counter
		f := runtime.FuncForPC(pc)
		if f == nil {
			break // Invalid program counter
		}

		// Extract a short, readable function name
		function := shortFuncName(f)

		// Define package name patterns for filtering
		packageNameExamples := packageName + "/examples/"

		// Apply frame filtering logic
		isGoPkg := len(runtime.GOROOT()) > 0 && strings.Contains(file, runtime.GOROOT()) // skip frames in GOROOT if it's set
		isOopsPkg := strings.Contains(file, packageName)                                 // skip frames in this package
		isExamplePkg := strings.Contains(file, packageNameExamples)                      // do not skip frames in this package examples
		isTestPkg := strings.Contains(file, "_test.go")                                  // do not skip frames in tests

		// Include frame if it passes all filtering criteria
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

// shortFuncName extracts a short, readable function name from a runtime.Func.
// This function processes the full function name (which includes package path
// and receiver information) and returns a simplified version suitable for
// display in stack traces.
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
func shortFuncName(f *runtime.Func) string {
	// f.Name() returns the full function name including package path
	// Examples of possible formats:
	// - "github.com/palantir/shield/package.FuncName"
	// - "github.com/palantir/shield/package.Receiver.MethodName"
	// - "github.com/palantir/shield/package.(*PtrReceiver).MethodName"
	longName := f.Name()

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

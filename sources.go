package oops

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/samber/lo"
)

///
/// Source code fragment extraction and caching functionality.
///
/// This module provides capabilities to extract and display source code
/// fragments around error locations for enhanced debugging. It includes
/// intelligent caching to minimize file I/O operations and provides
/// formatted output with line numbers and error highlighting.
///

// Global cache for source code fragments to minimize file I/O operations.
// The cache stores file contents as slices of strings (lines) and is
// protected by a read-write mutex for thread-safe concurrent access.
//
// Performance: The cache significantly reduces file I/O overhead for
// frequently accessed source files, especially in development environments
// where the same files are accessed repeatedly.
//
// Memory Usage: The cache stores entire file contents in memory, which
// may grow large for projects with many source files. The cache is
// currently unbounded, so memory usage should be monitored in production.
//
// @TODO: Implement LRU cache with size limits to prevent unbounded memory growth
var (
	mutex sync.RWMutex
	cache = map[string][]string{}
)

// Configuration constants for source code fragment extraction
const (
	// nbrLinesBefore specifies the number of source code lines to include
	// before the error location. This provides context about what led
	// to the error condition.
	nbrLinesBefore = 5

	// nbrLinesAfter specifies the number of source code lines to include
	// after the error location. This provides context about what follows
	// the error and may help understand the intended flow.
	nbrLinesAfter = 5
)

// readFileWithCache reads a file and returns its contents as a slice of strings,
// with caching to minimize repeated file I/O operations.
//
// This function implements a thread-safe caching mechanism that stores
// file contents in memory after the first read. Subsequent reads of the
// same file will return the cached content, significantly improving
// performance for frequently accessed source files.
//
// The function only processes .go files to avoid unnecessary processing
// of other file types. Files that cannot be read (due to permissions,
// non-existence, or other errors) are handled gracefully by returning
// false in the second return value.
//
// Performance: This function has O(1) complexity for cached files and
// O(n) complexity for uncached files, where n is the file size. The
// caching mechanism provides significant performance benefits for
// repeated access to the same files.
//
// Memory Management: The cache stores entire file contents in memory.
// For large projects, this may consume significant memory. Consider
// implementing cache eviction policies for production use.
//
// Example usage:
//
//	lines, ok := readFileWithCache("/path/to/main.go")
//	if ok {
//	  // Process the file lines
//	  for i, line := range lines {
//	    fmt.Printf("Line %d: %s\n", i+1, line)
//	  }
//	}
func readFileWithCache(path string) ([]string, bool) {
	// First, try to read from cache using a read lock
	mutex.RLock()
	lines, ok := cache[path]
	mutex.RUnlock()

	if ok {
		return lines, true
	}

	// Only process .go files to avoid unnecessary I/O
	if !strings.HasSuffix(path, ".go") {
		return nil, false
	}

	// Read the file from disk (this operation is not cached)
	// bearer:disable go_gosec_filesystem_filereadtaint
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	// Split the file content into lines
	lines = strings.Split(string(b), "\n")

	// Store the result in cache using a write lock
	mutex.Lock()
	cache[path] = lines
	mutex.Unlock()

	return lines, true
}

// getSourceFromFrame extracts source code fragments around a specific
// line in a file, providing context for debugging error locations.
//
// This function reads the source file (with caching) and extracts a
// configurable number of lines before and after the error location.
// The output includes line numbers and highlights the exact line where
// the error occurred with a visual indicator (caret characters).
//
// The function handles edge cases such as:
// - Files that cannot be read or don't exist
// - Line numbers that are out of bounds
// - Empty files or files with insufficient lines
// - Tab character handling for proper alignment
//
// Performance: This function benefits from the caching mechanism in
// readFileWithCache, making subsequent calls for the same file very fast.
// The line extraction and formatting operations are O(n) where n
// is the number of lines being extracted.
//
// Example output:
//
//	40	func processRequest(req *Request) error {
//	41	    if req == nil {
//	42	    ^^^^^^^^^^^^^
//	43	        return errors.New("request cannot be nil")
//	44	    }
//	45	    return nil
//	46	}
func getSourceFromFrame(frame oopsStacktraceFrame) []string {
	// Read the source file with caching
	lines, ok := readFileWithCache(frame.file)
	if !ok {
		return []string{}
	}

	// Validate that the requested line number is within bounds
	if len(lines) < frame.line {
		return []string{}
	}

	// Calculate the range of lines to extract
	current := frame.line - 1 // Convert to 0-based index
	start := lo.Max([]int{0, current - nbrLinesBefore})
	end := lo.Min([]int{len(lines) - 1, current + nbrLinesAfter})

	output := []string{}

	// Extract and format each line in the range
	for i := start; i <= end; i++ {
		if i < 0 || i >= len(lines) {
			continue
		}

		line := lines[i]

		// Format the line with line number
		message := fmt.Sprintf("%d\t%s", i+1, line)
		output = append(output, message)

		// Add visual indicator for the error line
		if i == current {
			// Calculate the position of the first non-whitespace character
			lenWithoutLeadingSpaces := len(strings.TrimLeft(line, " \t"))
			lenLeadingSpaces := len(line) - lenWithoutLeadingSpaces

			// Handle tab characters properly (tabs are typically 8 characters wide)
			nbrTabs := strings.Count(line[0:lenLeadingSpaces], "\t")
			firstCharIndex := lenLeadingSpaces + (8-1)*nbrTabs // 8 chars per tab

			// Create the visual indicator line
			sublinePrefix := string(lo.RepeatBy(firstCharIndex, func(_ int) byte { return ' ' }))
			subline := string(lo.RepeatBy(lenWithoutLeadingSpaces, func(_ int) byte { return '^' }))
			output = append(output, "\t"+sublinePrefix+subline)
		}
	}

	return output
}

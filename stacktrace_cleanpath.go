package oops

///
/// GOPATH path cleaning functionality for stacktraces.
///
/// This module provides utilities to clean up file paths in stacktraces
/// by removing GOPATH prefixes to make paths more readable and portable.
/// It includes logic to handle multiple GOPATH entries and ensures that
/// the longest matching prefix is removed for consistency.
///
/// Stolen from palantir/stacktrace repo
/// -> https://github.com/palantir/stacktrace/blob/master/cleanpath/gopath.go
/// -> Apache 2.0 LICENSE
///

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// removeGoPath makes a file path relative to one of the src directories
// in the $GOPATH environment variable, making stacktraces more readable
// and portable across different development environments.
//
// This function processes file paths from stacktraces and attempts to
// remove GOPATH prefixes to show paths relative to the Go workspace
// structure. This is particularly useful for:
// - Making stacktraces more readable by removing long absolute paths
// - Ensuring stacktraces are portable across different machines
// - Providing consistent path formatting regardless of GOPATH configuration
//
// The function handles several edge cases:
// - Empty or unset GOPATH environment variable
// - Paths that are not within any GOPATH src directory
// - Multiple GOPATH entries (uses the longest matching prefix)
// - Paths that would require traversing parent directories
//
// Performance: This function has O(n * m) complexity where n is the number
// of GOPATH entries and m is the average path length. For typical use cases
// with a few GOPATH entries, this is very fast.
//
// Example usage:
//
//	// With GOPATH="/home/user/go:/usr/local/go"
//	path := "/home/user/go/src/github.com/user/project/main.go"
//	clean := removeGoPath(path)
//	// Result: "github.com/user/project/main.go"
//
//	// Path not in GOPATH
//	path := "/usr/local/bin/program"
//	clean := removeGoPath(path)
//	// Result: "/usr/local/bin/program" (unchanged)
func removeGoPath(path string) string {
	// Get all GOPATH entries and split them into individual directories
	dirs := filepath.SplitList(os.Getenv("GOPATH"))

	// Sort directories in decreasing order by length so the longest
	// matching prefix is removed first. This ensures that if a path
	// matches multiple GOPATH entries, it's made relative to the
	// most specific (longest) one.
	sort.Stable(longestFirst(dirs))

	// Check each GOPATH directory for a match
	for _, dir := range dirs {
		// Construct the src directory path for this GOPATH entry
		srcdir := filepath.Join(dir, "src")

		// Try to make the path relative to this src directory
		rel, err := filepath.Rel(srcdir, path)

		// filepath.Rel can traverse parent directories (using ".."), which
		// we don't want. We only want paths that are actually within the
		// src directory structure.
		if err == nil && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return rel
		}
	}

	// If no GOPATH match is found, return the original path unchanged
	return path
}

// longestFirst is a custom sort.Interface implementation that sorts
// strings in decreasing order by length (longest first). This is used
// to ensure that when multiple GOPATH entries could match a path,
// the longest (most specific) match is chosen.
//
// This sorting strategy is important because:
// - It ensures consistent behavior when multiple GOPATH entries exist
// - It prefers the most specific GOPATH entry for path cleaning
// - It makes the path cleaning deterministic regardless of GOPATH order
//
// Performance: The sorting has O(n log n) complexity where n is the
// number of GOPATH entries. For typical use cases with few GOPATH
// entries, this overhead is negligible.
type longestFirst []string

// Len returns the number of strings in the slice.
func (strs longestFirst) Len() int { return len(strs) }

// Less returns true if the string at index i is longer than the string
// at index j. This implements the "longest first" sorting order.
func (strs longestFirst) Less(i, j int) bool { return len(strs[i]) > len(strs[j]) }

// Swap exchanges the strings at indices i and j.
func (strs longestFirst) Swap(i, j int) { strs[i], strs[j] = strs[j], strs[i] }

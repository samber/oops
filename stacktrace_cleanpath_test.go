package oops

///
/// Stolen from palantir/stacktrace repo
/// -> https://github.com/palantir/stacktrace/blob/master/cleanpath/gopath_test.go
/// -> Apache 2.0 LICENSE
///

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// resetGoPathDirsCache resets the sync.Once cache so tests can set GOPATH
// to different values between calls to removeGoPath.
func resetGoPathDirsCache() {
	goPathDirsOnce = sync.Once{}
	goPathDirs = nil
}

func TestRemoveGoPath(t *testing.T) { //nolint:paralleltest
	is := assert.New(t)

	// This test mutates os.Setenv("GOPATH") and the package-level cache, so it
	// must NOT run in parallel with other tests that use removeGoPath indirectly
	// (e.g. any test that creates an oops error with a stacktrace).
	originalGoPath := os.Getenv("GOPATH")
	t.Cleanup(func() {
		_ = os.Setenv("GOPATH", originalGoPath)
		resetGoPathDirsCache()
	})

	for _, testcase := range []struct {
		gopath   []string
		path     string
		expected string
	}{
		{
			// empty gopath
			gopath:   []string{},
			path:     "/some/dir/src/pkg/prog.go",
			expected: "/some/dir/src/pkg/prog.go",
		},
		{
			// single matching dir in gopath
			gopath:   []string{"/some/dir"},
			path:     "/some/dir/src/pkg/prog.go",
			expected: "pkg/prog.go",
		},
		{
			// nonmatching dir in gopath
			gopath:   []string{"/other/dir"},
			path:     "/some/dir/src/pkg/prog.go",
			expected: "/some/dir/src/pkg/prog.go",
		},
		{
			// multiple matching dirs in gopath, shorter first
			gopath:   []string{"/some", "/some/src/dir"},
			path:     "/some/src/dir/src/pkg/prog.go",
			expected: "pkg/prog.go",
		},
		{
			// multiple matching dirs in gopath, longer first
			gopath:   []string{"/some/src/dir", "/some"},
			path:     "/some/src/dir/src/pkg/prog.go",
			expected: "pkg/prog.go",
		},
	} {
		gopath := strings.Join(testcase.gopath, string(filepath.ListSeparator))
		err := os.Setenv("GOPATH", gopath)
		is.NoError(err, "error setting gopath")

		// Reset the cache so each test case reads the updated GOPATH.
		resetGoPathDirsCache()

		cleaned := removeGoPath(testcase.path)
		is.Equal(testcase.expected, cleaned, "testcase: %+v", testcase)
	}
}

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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveGoPath(t *testing.T) { //nolint:paralleltest
	is := assert.New(t)

	// This test mutates os.Setenv("GOPATH") so it must NOT run in parallel with
	// other tests that create oops errors (which call removeGoPath indirectly).
	// The GOPATH-keyed cache in cachedGoPathDirs invalidates automatically, so
	// no explicit cache reset is required between sub-cases.
	originalGoPath := os.Getenv("GOPATH")
	t.Cleanup(func() {
		_ = os.Setenv("GOPATH", originalGoPath)
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

		cleaned := removeGoPath(testcase.path)
		is.Equal(testcase.expected, cleaned, "testcase: %+v", testcase)
	}
}

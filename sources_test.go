package oops

import (
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadFileWithCache(t *testing.T) {
	_, currentFile, _, _ := runtime.Caller(0)

	t.Run("CacheMiss", func(t *testing.T) {
		// Ensure a clean cache before the test
		mutex.Lock()
		delete(cache, currentFile)
		mutex.Unlock()

		t.Cleanup(func() {
			mutex.Lock()
			cache = map[string][]string{}
			mutex.Unlock()
		})

		lines, ok := readFileWithCache(currentFile)
		assert.True(t, ok)
		assert.NotNil(t, lines)
		assert.Greater(t, len(lines), 0)
	})

	t.Run("CacheHit", func(t *testing.T) {
		// Pre-populate cache
		mutex.Lock()
		cache[currentFile] = []string{"line1", "line2"}
		mutex.Unlock()

		t.Cleanup(func() {
			mutex.Lock()
			cache = map[string][]string{}
			mutex.Unlock()
		})

		lines, ok := readFileWithCache(currentFile)
		assert.True(t, ok)
		assert.Equal(t, []string{"line1", "line2"}, lines)
	})

	t.Run("NonGoFile", func(t *testing.T) {
		t.Parallel()

		lines, ok := readFileWithCache("/some/path/file.txt")
		assert.False(t, ok)
		assert.Nil(t, lines)
	})

	t.Run("MissingFile", func(t *testing.T) {
		t.Parallel()

		lines, ok := readFileWithCache("/nonexistent/path/file.go")
		assert.False(t, ok)
		assert.Nil(t, lines)
	})

	t.Run("ContentCorrectness", func(t *testing.T) {
		// Ensure a clean cache for this file path
		mutex.Lock()
		delete(cache, currentFile)
		mutex.Unlock()

		t.Cleanup(func() {
			mutex.Lock()
			cache = map[string][]string{}
			mutex.Unlock()
		})

		lines, ok := readFileWithCache(currentFile)
		assert.True(t, ok)
		assert.NotNil(t, lines)

		// The test file itself must contain "package oops" on the first line
		assert.Equal(t, "package oops", lines[0])

		// The file should contain recognizable content from itself
		found := false
		for _, line := range lines {
			if strings.Contains(line, "TestReadFileWithCache") {
				found = true
				break
			}
		}
		assert.True(t, found, "expected to find 'TestReadFileWithCache' in file lines")
	})
}

func TestGetSourceFromFrame(t *testing.T) {
	_, currentFile, currentLine, _ := runtime.Caller(0)

	t.Run("ValidFrame", func(t *testing.T) {
		t.Cleanup(func() {
			mutex.Lock()
			cache = map[string][]string{}
			mutex.Unlock()
		})

		frame := oopsStacktraceFrame{
			file:     currentFile,
			line:     currentLine,
			function: "TestGetSourceFromFrame",
			pc:       0,
		}

		result := getSourceFromFrame(frame)
		assert.NotEmpty(t, result)

		// Should contain the line number in the output
		found := false
		for _, s := range result {
			if strings.Contains(s, "runtime.Caller") {
				found = true
				break
			}
		}
		assert.True(t, found, "expected source context to include the line with runtime.Caller")
	})

	t.Run("OutOfBoundsLine", func(t *testing.T) {
		t.Cleanup(func() {
			mutex.Lock()
			cache = map[string][]string{}
			mutex.Unlock()
		})

		frame := oopsStacktraceFrame{
			file:     currentFile,
			line:     999999,
			function: "TestGetSourceFromFrame",
			pc:       0,
		}

		result := getSourceFromFrame(frame)
		assert.Empty(t, result)
	})

	t.Run("MissingFile", func(t *testing.T) {
		t.Parallel()

		frame := oopsStacktraceFrame{
			file:     "/nonexistent/path/file.go",
			line:     1,
			function: "TestGetSourceFromFrame",
			pc:       0,
		}

		result := getSourceFromFrame(frame)
		assert.Empty(t, result)
	})

	t.Run("TabCharacterHandling", func(t *testing.T) {
		// Inject a fake file with a tab-indented line into the cache
		fakeFile := "/fake/tab_test_file.go"
		fakeLines := []string{
			"package fake",
			"\t\tfunc tabIndented() {",
			"\t\t\treturn",
			"\t\t}",
			"",
		}

		mutex.Lock()
		cache[fakeFile] = fakeLines
		mutex.Unlock()

		t.Cleanup(func() {
			mutex.Lock()
			cache = map[string][]string{}
			mutex.Unlock()
		})

		// Point at line 2: "\t\tfunc tabIndented() {"
		frame := oopsStacktraceFrame{
			file:     fakeFile,
			line:     2,
			function: "tabIndented",
			pc:       0,
		}

		result := getSourceFromFrame(frame)
		assert.NotEmpty(t, result)

		// Find the caret line (it starts with a tab)
		var caretLine string
		for _, s := range result {
			if strings.HasPrefix(s, "\t") && strings.Contains(s, "^") {
				caretLine = s
				break
			}
		}
		assert.NotEmpty(t, caretLine, "expected a caret indicator line")

		// 2 tabs => 2 * (8-1) = 14 extra spaces beyond the tab count itself
		// The caret prefix should account for the expanded tab width
		// prefix = repeat(" ", lenLeadingSpaces + (8-1)*nbrTabs)
		// lenLeadingSpaces = 2 (two tab chars), nbrTabs = 2
		// firstCharIndex = 2 + 14 = 16
		// So the caret line (after the leading \t) should have 16 spaces before ^
		caretContent := strings.TrimPrefix(caretLine, "\t")
		assert.True(t, strings.HasPrefix(caretContent, strings.Repeat(" ", 16)),
			"expected 16 leading spaces in caret line, got: %q", caretContent)
		assert.Contains(t, caretContent, "^")
	})
}

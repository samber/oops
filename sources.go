package oops

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/samber/lo"
)

var mutex sync.RWMutex
var cache = map[string][]string{}

const nbrLinesBefore = 5
const nbrLinesAfter = 5

func readFile(path string) ([]string, bool) {
	mutex.RLock()
	lines, ok := cache[path]
	mutex.RUnlock()

	if ok {
		return lines, true
	}

	if !strings.HasSuffix(path, ".go") {
		return nil, false
	}

	// bearer:disable go_gosec_filesystem_filereadtaint
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	lines = strings.Split(string(b), "\n")

	mutex.Lock()
	cache[path] = lines
	mutex.Unlock()

	return lines, true
}

func getSourceFromFrame(frame oopsStacktraceFrame) []string {
	lines, ok := readFile(frame.file)
	if !ok {
		return []string{}
	}

	if len(lines) < frame.line {
		return []string{}
	}

	current := frame.line - 1
	start := lo.Max([]int{0, current - nbrLinesBefore})
	end := lo.Min([]int{len(lines) - 1, current + nbrLinesAfter})

	output := []string{}

	for i := start; i <= end; i++ {
		if i < 0 || i >= len(lines) {
			continue
		}

		line := lines[i]
		message := fmt.Sprintf("%d\t%s", i+1, line)
		output = append(output, message)

		if i == current {
			lenWithoutLeadingSpaces := len(strings.TrimLeft(line, " \t"))
			lenLeadingSpaces := len(line) - lenWithoutLeadingSpaces
			nbrTabs := strings.Count(line[0:lenLeadingSpaces], "\t")
			firstCharIndex := lenLeadingSpaces + (8-1)*nbrTabs // 8 chars per tab

			sublinePrefix := string(lo.RepeatBy(firstCharIndex, func(_ int) byte { return ' ' }))
			subline := string(lo.RepeatBy(lenWithoutLeadingSpaces, func(_ int) byte { return '^' }))
			output = append(output, "\t"+sublinePrefix+subline)
		}
	}

	return output
}

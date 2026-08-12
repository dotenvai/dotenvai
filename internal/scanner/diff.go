package scanner

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"
)

var hunk = regexp.MustCompile(`^@@ -\d+(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

func ParseUnifiedDiff(input string) []Candidate {
	var result []Candidate
	file, line := "", 0
	s := bufio.NewScanner(strings.NewReader(input))
	buf := make([]byte, 64*1024)
	s.Buffer(buf, 2*1024*1024)
	for s.Scan() {
		text := s.Text()
		switch {
		case strings.HasPrefix(text, "+++ b/"):
			file = strings.TrimPrefix(text, "+++ b/")
		case hunk.MatchString(text):
			m := hunk.FindStringSubmatch(text)
			line, _ = strconv.Atoi(m[1])
		case strings.HasPrefix(text, "+") && !strings.HasPrefix(text, "+++"):
			result = append(result, Candidate{File: file, Line: line, Text: strings.TrimPrefix(text, "+")})
			line++
		case strings.HasPrefix(text, "-"):
		default:
			if file != "" {
				line++
			}
		}
	}
	return result
}

package audit

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dotenvai/dotenvai/internal/scanner"
	_ "modernc.org/sqlite"
)

type Finding struct {
	Agent       string          `json:"agent"`
	Surface     string          `json:"surface"`
	SessionFile string          `json:"session_file"`
	Record      string          `json:"record,omitempty"`
	Finding     scanner.Finding `json:"finding"`
}

type Result struct {
	FilesScanned int       `json:"files_scanned"`
	Findings     []Finding `json:"findings"`
	Warnings     []string  `json:"warnings,omitempty"`
}

func Run(ctx context.Context, agents []string) (Result, error) {
	for _, agent := range agents {
		if !Supported(agent) {
			return Result{}, fmt.Errorf("unsupported agent %q: use claude, codex, opencode, or cursor", agent)
		}
	}
	selected := Selected(agents)
	if len(selected) == 0 {
		return Result{}, fmt.Errorf("no supported agents selected")
	}
	result := Result{Findings: []Finding{}}
	for _, source := range selected {
		files, err := discover(source)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s: %v", source.Agent, err))
			continue
		}
		for _, path := range files {
			var candidates []scanner.Candidate
			var records []string
			switch source.Storage {
			case JSONL:
				candidates, err = readLines(path)
			case SQLite:
				candidates, records, err = readSQLite(ctx, path, source.Queries)
			case Files:
				candidates, err = readArtifact(path)
			}
			if err != nil {
				if errors.Is(err, errSkippedArtifact) {
					continue
				}
				result.Warnings = append(result.Warnings, fmt.Sprintf("%s: %s: %v", source.Agent, path, err))
				continue
			}
			result.FilesScanned++
			found, err := scanner.Scan(candidates)
			if err != nil {
				return result, err
			}
			for _, finding := range found {
				record := ""
				if finding.Line > 0 && finding.Line <= len(records) {
					record = records[finding.Line-1]
				}
				result.Findings = append(result.Findings, Finding{Agent: source.Agent, Surface: source.Surface, SessionFile: path, Record: record, Finding: finding})
			}
		}
	}
	return result, nil
}

func discover(source Source) ([]string, error) {
	seen := map[string]bool{}
	var files []string
	for _, pattern := range patterns(source) {
		root := globRoot(pattern)
		if _, err := os.Stat(root); errors.Is(err, fs.ErrNotExist) {
			continue
		}
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if entry.IsDir() {
				return nil
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return nil
			}
			matched, _ := filepath.Match(strings.ReplaceAll(pattern, "**/", ""), path)
			if strings.Contains(pattern, "**") {
				matched = strings.HasSuffix(path, strings.TrimPrefix(filepath.Ext(pattern), "*"))
			}
			if matched && !seen[path] {
				seen[path] = true
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(files)
	return files, nil
}

const maxArtifactSize = 10 << 20

var errSkippedArtifact = errors.New("artifact is binary, empty, or too large")

func readArtifact(path string) ([]scanner.Candidate, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() == 0 || info.Size() > maxArtifactSize {
		return nil, errSkippedArtifact
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	probe := make([]byte, 8192)
	n, err := file.Read(probe)
	file.Close()
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	if bytes.IndexByte(probe[:n], 0) >= 0 {
		return nil, errSkippedArtifact
	}
	return readLines(path)
}

func globRoot(pattern string) string {
	i := strings.IndexAny(pattern, "*?[")
	if i < 0 {
		return filepath.Dir(pattern)
	}
	root := filepath.Dir(pattern[:i])
	for strings.ContainsAny(root, "*?[") {
		root = filepath.Dir(root)
	}
	return root
}

func readLines(path string) ([]scanner.Candidate, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var result []scanner.Candidate
	s := bufio.NewScanner(file)
	s.Buffer(make([]byte, 64*1024), 8*1024*1024)
	for line := 1; s.Scan(); line++ {
		result = append(result, scanner.Candidate{File: path, Line: line, Text: s.Text()})
	}
	return result, s.Err()
}

func readSQLite(ctx context.Context, path string, queries []string) ([]scanner.Candidate, []string, error) {
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()
	var candidates []scanner.Candidate
	var records []string
	succeeded := 0
	var queryErrors []string
	for _, query := range queries {
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			queryErrors = append(queryErrors, err.Error())
			continue
		} // schema variants are expected
		succeeded++
		for rows.Next() {
			var record, text string
			if err := rows.Scan(&record, &text); err != nil {
				rows.Close()
				return nil, nil, err
			}
			records = append(records, record)
			candidates = append(candidates, scanner.Candidate{File: path, Line: len(records), Text: text})
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, nil, err
		}
		rows.Close()
	}
	if succeeded == 0 {
		return nil, nil, fmt.Errorf("no supported history schema: %s", strings.Join(queryErrors, "; "))
	}
	return candidates, records, nil
}

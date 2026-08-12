package audit

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Storage string

const (
	JSONL  Storage = "jsonl"
	SQLite Storage = "sqlite"
)

// Source is deliberately declarative: new agent layouts and database schema
// variants should normally require only another path or query.
type Source struct {
	Agent   string
	Storage Storage
	Paths   map[string][]string
	Queries []string
}

var Sources = []Source{
	{Agent: "claude", Storage: JSONL, Paths: map[string][]string{
		"all": {"~/.claude/projects/**/*.jsonl"},
	}},
	{Agent: "codex", Storage: JSONL, Paths: map[string][]string{
		"all": {"~/.codex/sessions/**/*.jsonl"},
	}},
	{Agent: "opencode", Storage: SQLite, Paths: map[string][]string{
		"all": {"${XDG_DATA_HOME:-~/.local/share}/opencode/opencode*.db"},
	}, Queries: []string{
		"SELECT 'message:' || id, data FROM message",
		"SELECT 'part:' || id, data FROM part",
	}},
	{Agent: "cursor", Storage: SQLite, Paths: map[string][]string{
		"linux":   {"~/.config/Cursor/User/globalStorage/state.vscdb", "~/.config/Cursor/User/workspaceStorage/*/state.vscdb"},
		"darwin":  {"~/Library/Application Support/Cursor/User/globalStorage/state.vscdb", "~/Library/Application Support/Cursor/User/workspaceStorage/*/state.vscdb"},
		"windows": {"${APPDATA}/Cursor/User/globalStorage/state.vscdb", "${APPDATA}/Cursor/User/workspaceStorage/*/state.vscdb"},
	}, Queries: []string{
		"SELECT key, value FROM ItemTable WHERE key LIKE '%composer%' OR key LIKE '%chat%' OR key LIKE '%aiService%'",
	}},
}

func Selected(names []string) []Source {
	if len(names) == 0 {
		return Sources
	}
	wanted := map[string]bool{}
	for _, name := range names {
		wanted[strings.ToLower(strings.TrimSpace(name))] = true
	}
	var selected []Source
	for _, source := range Sources {
		if wanted[source.Agent] {
			selected = append(selected, source)
		}
	}
	return selected
}

func Supported(name string) bool {
	for _, source := range Sources {
		if source.Agent == strings.ToLower(strings.TrimSpace(name)) {
			return true
		}
	}
	return false
}

func patterns(source Source) []string {
	result := append([]string{}, source.Paths["all"]...)
	result = append(result, source.Paths[runtime.GOOS]...)
	for i, pattern := range result {
		result[i] = expandPath(pattern)
	}
	return result
}

func expandPath(path string) string {
	home, _ := os.UserHomeDir()
	xdg := os.Getenv("XDG_DATA_HOME")
	if xdg == "" {
		xdg = filepath.Join(home, ".local", "share")
	}
	path = strings.ReplaceAll(path, "${XDG_DATA_HOME:-~/.local/share}", xdg)
	path = os.ExpandEnv(path)
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}

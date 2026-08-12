package audit

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadLinesFindsSecretWithoutReturningValue(t *testing.T) {
	secret := "AKIA" + "ABCDEFGHIJKLMNOP"
	path := filepath.Join(t.TempDir(), "session.jsonl")
	if err := os.WriteFile(path, []byte(`{"output":"`+secret+`"}`+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	candidates, err := readLines(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 1 || !strings.Contains(candidates[0].Text, secret) {
		t.Fatalf("candidates=%#v", candidates)
	}
}

func TestReadSQLite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	secret := "abcDEF" + "1234567890"
	if _, err := db.Exec(`CREATE TABLE part (id TEXT, data TEXT); INSERT INTO part VALUES ('p1', ?)`, `{"text":"TOKEN=`+secret+`"}`); err != nil {
		t.Fatal(err)
	}
	db.Close()
	candidates, records, err := readSQLite(context.Background(), path, []string{"SELECT id, data FROM part"})
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 1 || records[0] != "p1" {
		t.Fatalf("candidates=%#v records=%#v", candidates, records)
	}
}

func TestSelected(t *testing.T) {
	got := Selected([]string{"CODEX", "cursor"})
	if len(got) != 4 || got[0].Agent != "codex" || got[1].Agent != "codex" || got[2].Agent != "cursor" || got[3].Agent != "cursor" {
		t.Fatalf("got=%#v", got)
	}
}

func TestReadArtifactSkipsBinaryAndLargeFiles(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "binary")
	if err := os.WriteFile(binary, []byte{'a', 0, 'b'}, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readArtifact(binary); !errors.Is(err, errSkippedArtifact) {
		t.Fatalf("err=%v", err)
	}
	large := filepath.Join(dir, "large")
	if err := os.WriteFile(large, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Truncate(large, maxArtifactSize+1); err != nil {
		t.Fatal(err)
	}
	if _, err := readArtifact(large); !errors.Is(err, errSkippedArtifact) {
		t.Fatalf("err=%v", err)
	}
}

func TestAllSupportedAgentsHaveDiscoverySources(t *testing.T) {
	want := map[string]bool{"claude": true, "codex": true, "opencode": true, "cursor": true}
	for _, source := range Sources {
		delete(want, source.Agent)
		if len(source.Paths["all"])+len(source.Paths["linux"])+len(source.Paths["darwin"])+len(source.Paths["windows"]) == 0 {
			t.Errorf("%s has no discovery paths", source.Agent)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing sources: %v", want)
	}
}

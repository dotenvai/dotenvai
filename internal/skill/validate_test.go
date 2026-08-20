package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidate(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sample-skill")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	contents := "---\nname: sample-skill\ndescription: Use for tests.\n---\n\n# Test\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	metadata, err := Validate(dir)
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Name != "sample-skill" {
		t.Fatalf("name=%q", metadata.Name)
	}
}

func TestRejectsDirectoryMismatch(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "wrong")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	contents := "---\nname: right\ndescription: Use for tests.\n---\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Validate(dir); err == nil {
		t.Fatal("expected mismatch error")
	}
}

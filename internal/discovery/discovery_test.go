package discovery

import (
	"strings"
	"testing"
)

func TestResolveDomain(t *testing.T) {
	got, err := ResolveURL("example.com")
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://example.com/.well-known/agent-skills.json" {
		t.Fatalf("url=%q", got)
	}
}

func TestDecode(t *testing.T) {
	document := `{"schema_version":"0.1","publisher":{"name":"Example","url":"https://example.com"},"skills":[{"name":"example","description":"Example skill.","version":"1.0.0","source":"https://github.com/example/skills/tree/v1.0.0/example"}]}`
	if _, err := Decode(strings.NewReader(document)); err != nil {
		t.Fatal(err)
	}
}

func TestRejectsArchiveWithoutHash(t *testing.T) {
	document := Document{SchemaVersion: "0.1", Publisher: Publisher{Name: "Example", URL: "https://example.com"}, Skills: []Skill{{Name: "example", Description: "Example.", Version: "1.0.0", Source: "https://example.com/source", Archive: "https://example.com/archive.tgz"}}}
	if err := document.Validate(); err == nil {
		t.Fatal("expected hash error")
	}
}

func TestRejectsTrailingDocument(t *testing.T) {
	document := `{"schema_version":"0.1","publisher":{"name":"Example","url":"https://example.com"},"skills":[{"name":"example","description":"Example skill.","version":"1.0.0","source":"https://example.com/source"}]} {}`
	if _, err := Decode(strings.NewReader(document)); err == nil {
		t.Fatal("expected trailing content error")
	}
}

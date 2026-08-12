package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelp(t *testing.T) {
	var out bytes.Buffer
	if code := Run([]string{"--help"}, &out, &out); code != 0 || out.Len() == 0 {
		t.Fatalf("code=%d out=%q", code, out.String())
	}
}

func TestRootExposesCobraCommands(t *testing.T) {
	var out bytes.Buffer
	if code := Run([]string{"help"}, &out, &out); code != 0 {
		t.Fatalf("code=%d out=%q", code, out.String())
	}
	for _, command := range []string{"audit", "version", "completion"} {
		if !strings.Contains(out.String(), command) {
			t.Fatalf("help does not contain %q:\n%s", command, out.String())
		}
	}
}

func TestAuditRejectsUnknownAgent(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"audit", "--agent", "made-up"}, &stdout, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "unsupported agent") {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
}

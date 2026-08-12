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
	for _, command := range []string{"scan", "version", "completion"} {
		if !strings.Contains(out.String(), command) {
			t.Fatalf("help does not contain %q:\n%s", command, out.String())
		}
	}
}

func TestScanRejectsInvalidThreshold(t *testing.T) {
	var out bytes.Buffer
	if code := Run([]string{"scan", "--fail-on", "sometimes"}, &out, &out); code != 1 {
		t.Fatalf("code=%d out=%q", code, out.String())
	}
	if !strings.Contains(out.String(), "use any, high, or never") {
		t.Fatalf("out=%q", out.String())
	}
}

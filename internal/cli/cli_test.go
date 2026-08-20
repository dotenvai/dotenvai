package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelp(t *testing.T) {
	var output bytes.Buffer
	if code := Run(nil, &output, &output); code != 0 {
		t.Fatalf("code=%d", code)
	}
	if !strings.Contains(output.String(), "validate-manifest") {
		t.Fatalf("output=%q", output.String())
	}
}

func TestUnknownCommand(t *testing.T) {
	var output bytes.Buffer
	if code := Run([]string{"wat"}, &output, &output); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/dotenvai/dotenvai/internal/discovery"
	"github.com/dotenvai/dotenvai/internal/skill"
)

const usage = `dotenvai — discover and validate official Agent Skills

Usage:
  dotenvai validate <skill-directory>
  dotenvai validate-manifest <agent-skills.json>
  dotenvai discover <domain-or-https-url>
  dotenvai version
`

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprint(stdout, usage)
		return 0
	}
	switch args[0] {
	case "version":
		if len(args) != 1 {
			return fail(stderr, "version takes no arguments")
		}
		fmt.Fprintln(stdout, "dotenvai 0.1.0-dev")
		return 0
	case "validate":
		if len(args) != 2 {
			return fail(stderr, "usage: dotenvai validate <skill-directory>")
		}
		metadata, err := skill.Validate(args[1])
		if err != nil {
			return fail(stderr, err.Error())
		}
		fmt.Fprintf(stdout, "valid skill: %s\n", metadata.Name)
		return 0
	case "validate-manifest":
		if len(args) != 2 {
			return fail(stderr, "usage: dotenvai validate-manifest <agent-skills.json>")
		}
		document, err := discovery.ReadFile(args[1])
		if err != nil {
			return fail(stderr, err.Error())
		}
		fmt.Fprintf(stdout, "valid discovery document: %s (%d skills)\n", document.Publisher.Name, len(document.Skills))
		return 0
	case "discover":
		if len(args) != 2 {
			return fail(stderr, "usage: dotenvai discover <domain-or-https-url>")
		}
		document, resolved, err := discovery.Fetch(context.Background(), args[1])
		if err != nil {
			return fail(stderr, err.Error())
		}
		output, err := json.MarshalIndent(document, "", "  ")
		if err != nil {
			return fail(stderr, err.Error())
		}
		fmt.Fprintf(stdout, "discovered %s\n%s\n", resolved, output)
		return 0
	default:
		return fail(stderr, fmt.Sprintf("unknown command %q\n\n%s", args[0], usage))
	}
}

func fail(stderr io.Writer, message string) int {
	fmt.Fprintf(stderr, "dotenvai: %s\n", message)
	return 1
}

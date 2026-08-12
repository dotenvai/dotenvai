package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/dotenvai/dotenvai/internal/scanner"
	"github.com/spf13/cobra"
)

const ExitFindings = 2

var errFindings = errors.New("findings meet the failure threshold")

func Run(args []string, stdout, stderr io.Writer) int {
	command := NewRootCommand()
	command.SetArgs(args)
	command.SetOut(stdout)
	command.SetErr(stderr)
	if err := command.Execute(); err != nil {
		if errors.Is(err, errFindings) {
			return ExitFindings
		}
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "dotenvai",
		Short:         "Catch secrets before they escape",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.AddCommand(newScanCommand())
	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the dotenvai version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintln(cmd.OutOrStdout(), "dotenvai dev")
		},
	})
	return root
}

type scanOptions struct {
	base   string
	head   string
	staged bool
	format string
	failOn string
}

func newScanCommand() *cobra.Command {
	options := scanOptions{}
	command := &cobra.Command{
		Use:   "scan",
		Short: "Scan added Git diff lines for secrets",
		Long:  "Scan staged changes, a Git revision range, or the current working-tree diff without sending source code anywhere.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return options.run(cmd.OutOrStdout())
		},
	}
	flags := command.Flags()
	flags.StringVar(&options.base, "base", "", "base Git revision")
	flags.StringVar(&options.head, "head", "HEAD", "head Git revision")
	flags.BoolVar(&options.staged, "staged", false, "scan staged changes")
	flags.StringVarP(&options.format, "format", "f", "text", "output format: text, json, or github")
	flags.StringVar(&options.failOn, "fail-on", "any", "failure threshold: any, high, or never")
	command.MarkFlagsMutuallyExclusive("base", "staged")
	return command
}

func (o scanOptions) run(stdout io.Writer) error {
	if o.failOn != "any" && o.failOn != "high" && o.failOn != "never" {
		return fmt.Errorf("invalid --fail-on %q: use any, high, or never", o.failOn)
	}
	diffArgs := []string{"diff", "--no-ext-diff", "--unified=0"}
	if o.staged {
		diffArgs = append(diffArgs, "--cached")
	} else if o.base != "" {
		diffArgs = append(diffArgs, o.base+"..."+o.head)
	}
	cmd := exec.Command("git", diffArgs...)
	b, err := cmd.Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return fmt.Errorf("git diff: %s", strings.TrimSpace(string(exitError.Stderr)))
		}
		return fmt.Errorf("git diff: %w", err)
	}
	findings, err := scanner.Scan(scanner.ParseUnifiedDiff(string(b)))
	if err != nil {
		return err
	}
	if err := writeFindings(stdout, o.format, findings); err != nil {
		return err
	}
	for _, finding := range findings {
		if o.failOn == "any" || (o.failOn == "high" && finding.Severity == scanner.SeverityHigh) {
			return errFindings
		}
	}
	return nil
}

func writeFindings(w io.Writer, format string, findings []scanner.Finding) error {
	switch format {
	case "json":
		if findings == nil {
			findings = []scanner.Finding{}
		}
		return json.NewEncoder(w).Encode(struct {
			Findings []scanner.Finding `json:"findings"`
		}{findings})
	case "github":
		for _, f := range findings {
			fmt.Fprintf(w, "::error file=%s,line=%d,title=%s::%s\n", escapeProperty(f.File), f.Line, escapeProperty("dotenvai "+f.Rule), escapeCommand(fmt.Sprintf("%s (%s; fingerprint %s)", f.Description, f.Severity, f.Fingerprint)))
		}
	case "text":
		if len(findings) == 0 {
			fmt.Fprintln(w, "dotenvai: no secrets found")
			return nil
		}
		for _, f := range findings {
			fmt.Fprintf(w, "%s:%d: %s: %s [%s]\n", f.File, f.Line, f.Severity, f.Description, f.Rule)
		}
	default:
		return fmt.Errorf("invalid --format %q: use text, json, or github", format)
	}
	return nil
}

func escapeCommand(s string) string {
	r := strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A")
	return r.Replace(s)
}

func escapeProperty(s string) string {
	r := strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A", ":", "%3A", ",", "%2C")
	return r.Replace(s)
}

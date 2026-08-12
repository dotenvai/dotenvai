package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/dotenvai/dotenvai/internal/audit"
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
	root.AddCommand(newAuditCommand())
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

type auditOptions struct {
	agents []string
	format string
}

func newAuditCommand() *cobra.Command {
	options := auditOptions{}
	command := &cobra.Command{
		Use:   "audit",
		Short: "Find secrets retained in AI coding-agent transcripts",
		Long:  "Discover and scan local Claude Code, Codex, OpenCode, and Cursor session history without sending it anywhere.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			result, err := audit.Run(cmd.Context(), options.agents)
			if err != nil {
				return err
			}
			switch options.format {
			case "json":
				if err := json.NewEncoder(cmd.OutOrStdout()).Encode(result); err != nil {
					return err
				}
				if len(result.Findings) > 0 {
					return errFindings
				}
				return nil
			case "text":
				for _, warning := range result.Warnings {
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s\n", warning)
				}
				if len(result.Findings) == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "dotenvai: scanned %d agent history files; no secrets found\n", result.FilesScanned)
					return nil
				}
				fmt.Fprintf(cmd.OutOrStdout(), "dotenvai: found %d possible secret exposures in %d agent history files\n", len(result.Findings), result.FilesScanned)
				for _, finding := range result.Findings {
					location := fmt.Sprintf("%s:%d", finding.SessionFile, finding.Finding.Line)
					if finding.Record != "" {
						location += " (" + finding.Record + ")"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%s  %-8s %-10s %s\n  %s\n", strings.ToUpper(string(finding.Finding.Severity)), finding.Agent, finding.Surface, finding.Finding.Description, location)
				}
				return errFindings
			default:
				return fmt.Errorf("invalid --format %q: use text or json", options.format)
			}
		},
	}
	command.Flags().StringSliceVar(&options.agents, "agent", nil, "agents to scan: claude, codex, opencode, cursor (repeatable)")
	command.Flags().StringVarP(&options.format, "format", "f", "text", "output format: text or json")
	return command
}

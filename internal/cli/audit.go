package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/VIGOR-Digital-Solution/vigor-cli/internal/audit"
	// Side-effect import — checks register themselves with audit.
	_ "github.com/VIGOR-Digital-Solution/vigor-cli/internal/audit/checks"
)

func auditCmd() *cobra.Command {
	var (
		jsonOut bool
		failOn  string
	)

	cmd := &cobra.Command{
		Use:   "audit [path]",
		Short: "Run the Vigor standards audit on a directory",
		Long: `Walks the target directory (default: cwd) and runs every registered audit
check. Outputs a markdown table by default; pass --json for machine output.

Exit code:
  0 — all checks ≤ --fail-on severity
  1 — at least one check is at or above --fail-on (default: fail)

Severity ladder: skip < pass < warn < fail`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}
			absRoot, err := filepath.Abs(root)
			if err != nil {
				return fmt.Errorf("resolve root: %w", err)
			}

			report := audit.Audit(context.Background(), absRoot)

			if jsonOut {
				return report.WriteJSON(cmd.OutOrStdout())
			}
			report.WriteMarkdown(cmd.OutOrStdout())

			// Decide exit
			threshold := severityRank(audit.Severity(failOn))
			worst := 0
			for _, r := range report.Results {
				if rank := severityRank(r.Severity); rank > worst {
					worst = rank
				}
			}
			if worst >= threshold {
				// Print summary to stderr for CI legibility
				bad := lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4444")).Render
				fmt.Fprintln(os.Stderr, bad(fmt.Sprintf("audit: failed at severity %s (threshold %s)", rankSeverity(worst), failOn)))
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit a JSON report instead of markdown")
	cmd.Flags().StringVar(&failOn, "fail-on", "fail", "exit non-zero at this severity or above (skip|pass|warn|fail)")

	return cmd
}

func severityRank(s audit.Severity) int {
	switch s {
	case audit.SeveritySkip:
		return 0
	case audit.SeverityPass:
		return 1
	case audit.SeverityWarn:
		return 2
	case audit.SeverityFail:
		return 3
	}
	return 0
}

func rankSeverity(r int) audit.Severity {
	switch r {
	case 0:
		return audit.SeveritySkip
	case 1:
		return audit.SeverityPass
	case 2:
		return audit.SeverityWarn
	case 3:
		return audit.SeverityFail
	}
	return audit.SeveritySkip
}

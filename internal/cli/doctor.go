package cli

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// Required CLIs for the six-CLI Vigor stack (per ADR-0006 + ADR-0009).
var requiredCLIs = []cliCheck{
	{"gh", "GitHub CLI", "https://cli.github.com/"},
	{"vercel", "Vercel CLI", "https://vercel.com/docs/cli"},
	{"railway", "Railway CLI", "https://docs.railway.app/develop/cli"},
	{"supabase", "Supabase CLI", "https://supabase.com/docs/guides/cli"},
	{"resend", "Resend CLI", "https://resend.com/docs/dashboard/cli/introduction"},
	{"hostinger", "Hostinger CLI", "https://developers.hostinger.com/"},

	// Required toolchain
	{"node", "Node.js", "https://nodejs.org/"},
	{"pnpm", "pnpm", "https://pnpm.io/"},
	{"git", "Git", "https://git-scm.com/"},
}

type cliCheck struct {
	bin   string
	label string
	url   string
}

func doctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check that all CLIs in the Vigor stack are installed",
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()

			ok := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Render
			bad := lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4444")).Render
			muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#71717a")).Render

			missing := 0
			fmt.Fprintln(out, "Vigor CLI doctor")
			fmt.Fprintln(out, muted("(install: "), muted("https://github.com/VIGOR-Digital-Solution/vigor-boilerplate/blob/main/docs/adrs/0006-deployment-platforms.md"), muted(")"))
			fmt.Fprintln(out)

			for _, c := range requiredCLIs {
				path, err := exec.LookPath(c.bin)
				if err != nil {
					fmt.Fprintf(out, "  %s  %-12s %s\n", bad("✗"), c.bin, muted("(install: "+c.url+")"))
					missing++
					continue
				}
				fmt.Fprintf(out, "  %s  %-12s %s\n", ok("✓"), c.bin, muted(path))
			}

			fmt.Fprintln(out)
			if missing > 0 {
				fmt.Fprintf(out, "%s %d CLI(s) missing. Install the ones above to enable end-to-end automation.\n", bad("!"), missing)
				return fmt.Errorf("%d missing CLIs", missing)
			}
			fmt.Fprintln(out, ok("All CLIs present."))
			return nil
		},
	}
}

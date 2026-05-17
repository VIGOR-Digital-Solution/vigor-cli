package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/VIGOR-Digital-Solution/vigor-cli/internal/scaffold"
)

func scaffoldCmd() *cobra.Command {
	var (
		ref     string
		force   bool
		noInit  bool
		listAll bool
	)

	cmd := &cobra.Command{
		Use:   "scaffold <platform> <name>",
		Short: "Scaffold a new project from a Vigor template",
		Long: `Fetches the matching template from VIGOR-Digital-Solution/vigor-boilerplate,
copies it into ./<name>, replaces placeholders, and initialises git with the
three-branch model (production → staging → development).

Platforms: web, pwa, mobile, backend, ai, iot

Examples:
  vigor scaffold web my-app
  vigor scaffold backend my-api --ref=v0.11.0
  vigor scaffold ai my-ai --no-init`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if listAll {
				out := cmd.OutOrStdout()
				bold := lipgloss.NewStyle().Bold(true).Render
				muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#71717a")).Render
				for _, t := range scaffold.All() {
					fmt.Fprintf(out, "  %s  %s\n  %s  %s\n\n",
						bold(string(t.Platform)), muted("("+t.Language+")"),
						muted(""), t.Description)
				}
				return nil
			}

			if len(args) != 2 {
				return errors.New("usage: vigor scaffold <platform> <name>")
			}

			template, err := scaffold.Resolve(args[0])
			if err != nil {
				return err
			}
			name := args[1]
			dest, err := filepath.Abs(name)
			if err != nil {
				return fmt.Errorf("resolve dest: %w", err)
			}

			if _, err := os.Stat(dest); err == nil {
				if !force {
					return fmt.Errorf("%s already exists (use --force to overwrite)", dest)
				}
			}

			src := scaffold.DefaultSource()
			if ref != "" {
				src.Ref = ref
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "→ Fetching %s template from %s@%s…\n", template.Platform, src.Repo, src.Ref)
			if err := scaffold.Fetch(src, template, dest); err != nil {
				return fmt.Errorf("fetch: %w", err)
			}

			fmt.Fprintf(out, "→ Replacing placeholders (PROJECT_NAME=%s)…\n", name)
			if err := scaffold.Apply(dest, scaffold.Placeholders{ProjectName: name}); err != nil {
				return fmt.Errorf("apply placeholders: %w", err)
			}

			if !noInit {
				fmt.Fprintln(out, "→ Initialising git (production / staging / development)…")
				if err := scaffold.PostInit(dest); err != nil {
					return fmt.Errorf("post-init: %w", err)
				}
			}

			ok := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Render
			fmt.Fprintf(out, "\n%s scaffolded into %s\n\n", ok("✓"), dest)
			fmt.Fprintf(out, "Next:\n")
			fmt.Fprintf(out, "  cd %s\n", name)
			fmt.Fprintf(out, "  pnpm install   # or uv sync for ai template\n")
			fmt.Fprintf(out, "  pnpm init      # runs scripts/init.sh — gh + vercel + railway + supabase + resend + hostinger\n")
			return nil
		},
	}

	cmd.Flags().StringVar(&ref, "ref", "", "git ref (branch or tag) in vigor-boilerplate (default: main)")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing destination directory")
	cmd.Flags().BoolVar(&noInit, "no-init", false, "skip git init / first commit")
	cmd.Flags().BoolVar(&listAll, "list", false, "list available platforms and exit")

	return cmd
}

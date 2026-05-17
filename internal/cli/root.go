// Package cli wires the cobra command tree for the vigor binary.
package cli

import (
	"github.com/spf13/cobra"
)

const longDescription = `vigor — Vigor Digital's scaffolder + audit CLI.

Mirrors the /vigor-* slash commands available in Claude Code, for teammates
who work outside the editor. Built on the same templates, standards, and
ADRs shipped from VIGOR-Digital-Solution/vigor-boilerplate.

Common commands:
  vigor scaffold web my-app       Scaffold a new project from a template
  vigor audit                     Run the standards audit on the current dir
  vigor doctor                    Check that all required CLIs are installed
  vigor upgrade                   Check for a newer vigor release
`

// Root returns the root *cobra.Command. Sub-commands attach themselves via
// AddCommand in their init() — pattern matches gh / fly / kubectl.
func Root() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "vigor",
		Short:         "Vigor Digital scaffolder + audit CLI",
		Long:          longDescription,
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	cmd.PersistentFlags().Bool("quiet", false, "suppress non-essential output")
	cmd.PersistentFlags().Bool("verbose", false, "emit debug-level logs")

	cmd.AddCommand(
		scaffoldCmd(),
		auditCmd(),
		doctorCmd(),
		upgradeCmd(),
		versionCmd(),
	)

	return cmd
}

package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/VIGOR-Digital-Solution/vigor-cli/internal/version"
)

func versionCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version, commit, build date, and Go runtime",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if jsonOutput {
				fmt.Fprintf(cmd.OutOrStdout(),
					"{\"version\":%q,\"commit\":%q,\"date\":%q,\"go\":%q,\"os\":%q,\"arch\":%q}\n",
					version.Version(), version.Commit(), version.Date(),
					runtime.Version(), runtime.GOOS, runtime.GOARCH,
				)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"vigor %s (commit %s, built %s)\n  go %s %s/%s\n",
				version.Version(), version.Commit(), version.Date(),
				runtime.Version(), runtime.GOOS, runtime.GOARCH,
			)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "emit JSON")
	return cmd
}

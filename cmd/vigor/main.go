// Command vigor scaffolds and audits Vigor Digital projects. See `vigor --help`.
//
// Distributed via Homebrew (`brew install vigor-digital-solution/tap/vigor`),
// scoop (`scoop install vigor`), or the install script. Mirrors the Claude Code
// `/vigor-*` skill surface so non-Claude-Code workflows have the same tools.
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/VIGOR-Digital-Solution/vigor-cli/internal/cli"
	"github.com/VIGOR-Digital-Solution/vigor-cli/internal/version"
)

// Build-time injected via goreleaser ldflags.
var (
	versionStr = "dev"
	commit     = "unknown"
	date       = "unknown"
)

func main() {
	version.Set(versionStr, commit, date)

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	if err := cli.Root().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "vigor:", err)
		os.Exit(1)
	}
}

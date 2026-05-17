package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/VIGOR-Digital-Solution/vigor-cli/internal/version"
)

const releasesAPI = "https://api.github.com/repos/VIGOR-Digital-Solution/vigor-cli/releases/latest"

func upgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Check for a newer vigor release",
		Long: `Hits the GitHub Releases API for VIGOR-Digital-Solution/vigor-cli, compares
against the embedded version, and prints upgrade instructions for Homebrew,
scoop, or the install script. No phone-home telemetry is sent.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			ok := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Render
			warn := lipgloss.NewStyle().Foreground(lipgloss.Color("#f59e0b")).Render
			muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#71717a")).Render

			current := strings.TrimPrefix(version.Version(), "v")
			fmt.Fprintf(out, "Current: %s %s\n", current, muted("(commit "+version.Commit()+")"))

			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()

			latest, err := fetchLatest(ctx)
			if err != nil {
				return fmt.Errorf("check latest: %w", err)
			}
			fmt.Fprintf(out, "Latest:  %s\n\n", latest)

			if current == "dev" {
				fmt.Fprintln(out, warn("(running a dev build; install a release with one of the commands below)"))
				printUpgradeInstructions(out, latest)
				return nil
			}

			currentSemver, err := semver.NewVersion(current)
			if err != nil {
				fmt.Fprintln(out, warn("can't parse current version; cannot compare"))
				printUpgradeInstructions(out, latest)
				return nil
			}
			latestSemver, err := semver.NewVersion(strings.TrimPrefix(latest, "v"))
			if err != nil {
				return fmt.Errorf("parse latest %s: %w", latest, err)
			}

			if !currentSemver.LessThan(latestSemver) {
				fmt.Fprintln(out, ok("✓ You're on the latest release."))
				return nil
			}

			fmt.Fprintln(out, warn("⚠ Newer release available."))
			printUpgradeInstructions(out, latest)
			return nil
		},
	}
}

func fetchLatest(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releasesAPI, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github returned %d", resp.StatusCode)
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.TagName == "" {
		return "", fmt.Errorf("empty tag_name in response")
	}
	return payload.TagName, nil
}

func printUpgradeInstructions(out interface{ Write(p []byte) (int, error) }, latest string) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Upgrade:")
	fmt.Fprintln(out, "  Homebrew:   brew upgrade vigor")
	fmt.Fprintln(out, "  scoop:      scoop update vigor")
	fmt.Fprintln(out, "  Script:     curl -fsSL https://raw.githubusercontent.com/VIGOR-Digital-Solution/vigor-cli/main/install.sh | bash")
	fmt.Fprintln(out)
	fmt.Fprintf(out, "Release notes: https://github.com/VIGOR-Digital-Solution/vigor-cli/releases/tag/%s\n", latest)
}

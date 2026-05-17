package scaffold

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// PostInit runs `git init` + initial commit + prints a next-steps checklist.
// Skips git steps if the user is already inside a repo (e.g. they scaffolded
// into an existing project structure).
func PostInit(dest string) error {
	if _, err := exec.LookPath("git"); err != nil {
		return errors.New("git not found in PATH; cannot finish scaffold")
	}

	if isGitRepo(dest) {
		return nil
	}

	if err := run(dest, "git", "init", "-b", "production"); err != nil {
		return err
	}
	if err := run(dest, "git", "checkout", "-b", "development"); err != nil {
		return err
	}
	if err := run(dest, "git", "add", "."); err != nil {
		return err
	}
	if err := run(dest, "git", "commit", "-m", "chore: initial scaffold from vigor-boilerplate"); err != nil {
		return err
	}
	return nil
}

func run(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %v: %w", name, args, err)
	}
	return nil
}

func isGitRepo(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir
	return cmd.Run() == nil
}

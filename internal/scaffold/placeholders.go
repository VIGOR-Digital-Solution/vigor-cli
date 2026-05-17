package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Placeholders the templates embed for replacement at scaffold time.
type Placeholders struct {
	ProjectName string
}

// Apply walks `root` and rewrites every text file, replacing each placeholder
// with its concrete value. Binary files (detected via simple magic-byte heuristic)
// are skipped.
func Apply(root string, p Placeholders) error {
	replacements := []struct{ from, to string }{
		{"{{PROJECT_NAME}}", p.ProjectName},
	}

	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip vendored / generated dirs
			if name := d.Name(); name == "node_modules" || name == ".git" || name == "dist" || name == "build" {
				return fs.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("stat %s: %w", path, err)
		}
		// Skip large files (> 1MB) — almost always binary at template scaffold time
		if info.Size() > 1<<20 {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if isBinary(data) {
			return nil
		}

		modified := false
		out := string(data)
		for _, r := range replacements {
			if strings.Contains(out, r.from) {
				out = strings.ReplaceAll(out, r.from, r.to)
				modified = true
			}
		}
		if !modified {
			return nil
		}
		if err := os.WriteFile(path, []byte(out), info.Mode().Perm()); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("walk %s: %w", root, err)
	}
	return nil
}

// isBinary checks for a null byte in the first 8 KB — same heuristic git uses.
func isBinary(b []byte) bool {
	limit := len(b)
	if limit > 8000 {
		limit = 8000
	}
	for i := 0; i < limit; i++ {
		if b[i] == 0 {
			return true
		}
	}
	return false
}

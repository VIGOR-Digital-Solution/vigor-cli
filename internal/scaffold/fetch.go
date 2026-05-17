package scaffold

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultRepo = "VIGOR-Digital-Solution/vigor-boilerplate"
	defaultRef  = "main"
)

// Source is the GitHub coordinates for the template fetch. Override REPO / REF
// via env for testing or forks.
type Source struct {
	Repo string // owner/name
	Ref  string // tag or branch
}

func DefaultSource() Source {
	repo := os.Getenv("VIGOR_TEMPLATES_REPO")
	if repo == "" {
		repo = defaultRepo
	}
	ref := os.Getenv("VIGOR_TEMPLATES_REF")
	if ref == "" {
		ref = defaultRef
	}
	return Source{Repo: repo, Ref: ref}
}

// Fetch downloads the template subtree from GitHub as a tarball and extracts
// only the entries under templates/<platform>/ into `dest`. Skips the
// archive's top-level directory prefix (`<repo>-<sha>/`).
func Fetch(src Source, template Template, dest string) error {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return fmt.Errorf("create dest: %w", err)
	}

	url := fmt.Sprintf("https://codeload.github.com/%s/tar.gz/%s", src.Repo, src.Ref)
	client := &http.Client{Timeout: 60 * time.Second}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download tarball: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d fetching %s", resp.StatusCode, url)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer gz.Close()

	prefix := template.RepoPath + "/"

	tr := tar.NewReader(gz)
	files := 0
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("tar next: %w", err)
		}

		// Tarball entries look like `<repo>-<sha>/templates/web/file.ts`. Strip
		// the top-level directory prefix.
		parts := strings.SplitN(hdr.Name, "/", 2)
		if len(parts) < 2 {
			continue
		}
		rel := parts[1]
		if !strings.HasPrefix(rel, prefix) {
			continue
		}
		rel = strings.TrimPrefix(rel, prefix)
		if rel == "" {
			continue
		}

		// Guard against zip-slip
		full := filepath.Join(dest, rel)
		if !strings.HasPrefix(filepath.Clean(full)+string(os.PathSeparator), filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("unsafe path in tarball: %s", rel)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(full, 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", full, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
				return fmt.Errorf("mkdir parent of %s: %w", full, err)
			}
			out, err := os.OpenFile(full, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode)&0o777)
			if err != nil {
				return fmt.Errorf("create %s: %w", full, err)
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return fmt.Errorf("write %s: %w", full, err)
			}
			out.Close()
			files++
		}
	}

	if files == 0 {
		return fmt.Errorf("template %s not found in %s@%s", template.Platform, src.Repo, src.Ref)
	}
	return nil
}

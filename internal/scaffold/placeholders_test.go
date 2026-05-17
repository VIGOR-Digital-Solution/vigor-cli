package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApply_replacesProjectName(t *testing.T) {
	dir := t.TempDir()

	files := map[string]string{
		"package.json":  `{"name": "{{PROJECT_NAME}}"}`,
		"app/page.tsx":  `export const NAME = '{{PROJECT_NAME}}';`,
		"README.md":     `# {{PROJECT_NAME}}`,
		"deeply/nested": `{{PROJECT_NAME}} appears here`,
	}
	for path, content := range files {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := Apply(dir, Placeholders{ProjectName: "my-app"}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	for path := range files {
		out, err := os.ReadFile(filepath.Join(dir, path))
		if err != nil {
			t.Fatal(err)
		}
		if got := string(out); contains(got, "{{PROJECT_NAME}}") {
			t.Errorf("%s still contains placeholder: %s", path, got)
		}
		if got := string(out); !contains(got, "my-app") {
			t.Errorf("%s missing replacement: %s", path, got)
		}
	}
}

func TestApply_skipsBinary(t *testing.T) {
	dir := t.TempDir()
	binary := append([]byte{0x00, 0x01, 0x02}, []byte("{{PROJECT_NAME}}")...)
	path := filepath.Join(dir, "image.png")
	if err := os.WriteFile(path, binary, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Apply(dir, Placeholders{ProjectName: "x"}); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(string(out), "{{PROJECT_NAME}}") {
		t.Errorf("binary file was modified; expected to be skipped")
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

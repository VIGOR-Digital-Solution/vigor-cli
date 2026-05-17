// Package checks holds the individual audit checks. Each registers itself
// at init() so audit.Audit() picks them up automatically.
package checks

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/VIGOR-Digital-Solution/vigor-cli/internal/audit"
)

func init() {
	audit.Register(&tsconfigStrict{})
	audit.Register(&eslintPresent{})
	audit.Register(&prettierPresent{})
	audit.Register(&vitestOrJestConfig{})
	audit.Register(&playwrightConfig{})
	audit.Register(&githubActionsCI{})
	audit.Register(&envExample{})
	audit.Register(&conventionalCommits{})
	audit.Register(&lighthouseConfig{})
	audit.Register(&secretsNotCommitted{})
}

// ─── TS strict ───────────────────────────────────────────────────────────────
type tsconfigStrict struct{}

func (tsconfigStrict) ID() string       { return "code-quality.tsconfig-strict" }
func (tsconfigStrict) Category() string { return "Code quality" }
func (tsconfigStrict) Run(_ context.Context, root string) audit.Result {
	r := audit.Result{
		ID:       "code-quality.tsconfig-strict",
		Category: "Code quality",
		Standard: "docs/standards/code-quality.md",
	}
	path := filepath.Join(root, "tsconfig.json")
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		r.Severity = audit.SeveritySkip
		r.Message = "tsconfig.json not found (not a TS project?)"
		return r
	}
	if err != nil {
		r.Severity = audit.SeverityFail
		r.Message = err.Error()
		return r
	}
	// JSON-with-comments tolerated by stripping // lines before parsing.
	cleaned := stripJSONComments(string(data))
	var parsed struct {
		Extends         string                 `json:"extends"`
		CompilerOptions map[string]interface{} `json:"compilerOptions"`
	}
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		r.Severity = audit.SeverityWarn
		r.Message = "couldn't parse tsconfig.json (comments / JSONC?)"
		return r
	}
	if strings.Contains(parsed.Extends, "@vigor/tsconfig") {
		r.Severity = audit.SeverityPass
		r.Message = "extends @vigor/tsconfig (strict by inheritance)"
		return r
	}
	if strict, ok := parsed.CompilerOptions["strict"].(bool); ok && strict {
		r.Severity = audit.SeverityPass
		r.Message = "strict: true"
		return r
	}
	r.Severity = audit.SeverityFail
	r.Message = "strict mode not enabled"
	r.Suggestion = `Extend "@vigor/tsconfig/<variant>" or set compilerOptions.strict: true`
	return r
}

// ─── ESLint config present ──────────────────────────────────────────────────
type eslintPresent struct{}

func (eslintPresent) ID() string       { return "code-quality.eslint-config" }
func (eslintPresent) Category() string { return "Code quality" }
func (eslintPresent) Run(_ context.Context, root string) audit.Result {
	r := audit.Result{ID: "code-quality.eslint-config", Category: "Code quality"}
	candidates := []string{"eslint.config.js", "eslint.config.mjs", "eslint.config.ts", ".eslintrc.json", ".eslintrc.js"}
	for _, c := range candidates {
		if exists(filepath.Join(root, c)) {
			r.Severity = audit.SeverityPass
			r.Message = "found " + c
			return r
		}
	}
	r.Severity = audit.SeverityFail
	r.Message = "no eslint config found"
	r.Suggestion = "Add eslint.config.js extending @vigor/eslint-config/<variant>"
	return r
}

// ─── Prettier config ─────────────────────────────────────────────────────────
type prettierPresent struct{}

func (prettierPresent) ID() string       { return "code-quality.prettier-config" }
func (prettierPresent) Category() string { return "Code quality" }
func (prettierPresent) Run(_ context.Context, root string) audit.Result {
	r := audit.Result{ID: "code-quality.prettier-config", Category: "Code quality"}
	candidates := []string{".prettierrc", ".prettierrc.json", ".prettierrc.js", ".prettierrc.mjs", "prettier.config.js"}
	for _, c := range candidates {
		if exists(filepath.Join(root, c)) {
			r.Severity = audit.SeverityPass
			r.Message = "found " + c
			return r
		}
	}
	// Also accept "prettier" field in package.json
	if pkg, ok := readPackageJSON(root); ok {
		if _, present := pkg["prettier"]; present {
			r.Severity = audit.SeverityPass
			r.Message = `"prettier" field in package.json`
			return r
		}
	}
	r.Severity = audit.SeverityWarn
	r.Message = "no prettier config detected"
	r.Suggestion = `Add "prettier": "@vigor/prettier-config" to package.json`
	return r
}

// ─── Test config present ─────────────────────────────────────────────────────
type vitestOrJestConfig struct{}

func (vitestOrJestConfig) ID() string       { return "testing.test-runner-config" }
func (vitestOrJestConfig) Category() string { return "Testing" }
func (vitestOrJestConfig) Run(_ context.Context, root string) audit.Result {
	r := audit.Result{ID: "testing.test-runner-config", Category: "Testing"}
	candidates := []string{"vitest.config.ts", "vitest.config.js", "jest.config.ts", "jest.config.js", "jest.config.json"}
	for _, c := range candidates {
		if exists(filepath.Join(root, c)) {
			r.Severity = audit.SeverityPass
			r.Message = "found " + c
			return r
		}
	}
	if pkg, ok := readPackageJSON(root); ok {
		if _, hasJest := pkg["jest"]; hasJest {
			r.Severity = audit.SeverityPass
			r.Message = `"jest" field in package.json`
			return r
		}
	}
	if exists(filepath.Join(root, "pyproject.toml")) {
		// Trust pyproject for Python projects
		r.Severity = audit.SeverityPass
		r.Message = "pyproject.toml (assume pytest configured)"
		return r
	}
	r.Severity = audit.SeverityFail
	r.Message = "no test-runner config found"
	r.Suggestion = "Add vitest.config.ts / jest.config.ts (per Tier 1 of the testing matrix)"
	return r
}

// ─── Playwright config ──────────────────────────────────────────────────────
type playwrightConfig struct{}

func (playwrightConfig) ID() string       { return "testing.playwright-config" }
func (playwrightConfig) Category() string { return "Testing" }
func (playwrightConfig) Run(_ context.Context, root string) audit.Result {
	r := audit.Result{ID: "testing.playwright-config", Category: "Testing"}
	candidates := []string{"playwright.config.ts", "playwright.config.js"}
	for _, c := range candidates {
		if exists(filepath.Join(root, c)) {
			r.Severity = audit.SeverityPass
			r.Message = "found " + c
			return r
		}
	}
	// E2E is required for frontend templates only; skip for backend/AI
	if exists(filepath.Join(root, "next.config.ts")) || exists(filepath.Join(root, "next.config.js")) || exists(filepath.Join(root, "app.config.ts")) {
		r.Severity = audit.SeverityWarn
		r.Message = "frontend project without playwright.config.* — tier 4 E2E gap"
		r.Suggestion = "Add playwright.config.ts with viewport projects (see templates/web)"
		return r
	}
	r.Severity = audit.SeveritySkip
	r.Message = "no frontend detected; Playwright not required"
	return r
}

// ─── GitHub Actions CI ──────────────────────────────────────────────────────
type githubActionsCI struct{}

func (githubActionsCI) ID() string       { return "deployment.github-actions" }
func (githubActionsCI) Category() string { return "Deployment" }
func (githubActionsCI) Run(_ context.Context, root string) audit.Result {
	r := audit.Result{ID: "deployment.github-actions", Category: "Deployment"}
	dir := filepath.Join(root, ".github", "workflows")
	entries, err := os.ReadDir(dir)
	if err != nil {
		r.Severity = audit.SeverityFail
		r.Message = "no .github/workflows/ directory"
		r.Suggestion = "Add at least ci.yml (lint/typecheck/test/build)"
		return r
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && (strings.HasSuffix(e.Name(), ".yml") || strings.HasSuffix(e.Name(), ".yaml")) {
			count++
		}
	}
	if count == 0 {
		r.Severity = audit.SeverityFail
		r.Message = ".github/workflows/ exists but has no workflows"
		return r
	}
	r.Severity = audit.SeverityPass
	r.Message = "workflows present"
	return r
}

// ─── .env.example ────────────────────────────────────────────────────────────
type envExample struct{}

func (envExample) ID() string       { return "security.env-example" }
func (envExample) Category() string { return "Security" }
func (envExample) Run(_ context.Context, root string) audit.Result {
	r := audit.Result{ID: "security.env-example", Category: "Security"}
	if exists(filepath.Join(root, ".env.example")) {
		r.Severity = audit.SeverityPass
		r.Message = ".env.example committed"
		return r
	}
	r.Severity = audit.SeverityWarn
	r.Message = "no .env.example found"
	r.Suggestion = "Add .env.example so contributors know what env vars to set"
	return r
}

// ─── commitlint / conventional-pre-commit ───────────────────────────────────
type conventionalCommits struct{}

func (conventionalCommits) ID() string       { return "code-quality.commitlint" }
func (conventionalCommits) Category() string { return "Code quality" }
func (conventionalCommits) Run(_ context.Context, root string) audit.Result {
	r := audit.Result{ID: "code-quality.commitlint", Category: "Code quality"}
	candidates := []string{"commitlint.config.mjs", "commitlint.config.js", ".commitlintrc", ".commitlintrc.json", ".pre-commit-config.yaml"}
	for _, c := range candidates {
		if exists(filepath.Join(root, c)) {
			r.Severity = audit.SeverityPass
			r.Message = "found " + c
			return r
		}
	}
	r.Severity = audit.SeverityWarn
	r.Message = "no commit-message linter detected"
	r.Suggestion = "Add commitlint.config.mjs + husky commit-msg hook (or .pre-commit-config.yaml for Python)"
	return r
}

// ─── Lighthouse config ───────────────────────────────────────────────────────
type lighthouseConfig struct{}

func (lighthouseConfig) ID() string       { return "performance.lighthouse-config" }
func (lighthouseConfig) Category() string { return "Performance" }
func (lighthouseConfig) Run(_ context.Context, root string) audit.Result {
	r := audit.Result{ID: "performance.lighthouse-config", Category: "Performance"}
	if !exists(filepath.Join(root, "next.config.ts")) && !exists(filepath.Join(root, "next.config.js")) {
		r.Severity = audit.SeveritySkip
		r.Message = "not a Next.js project"
		return r
	}
	if exists(filepath.Join(root, "lighthouserc.json")) || exists(filepath.Join(root, "lighthouserc.js")) {
		r.Severity = audit.SeverityPass
		r.Message = "lighthouserc present"
		return r
	}
	r.Severity = audit.SeverityWarn
	r.Message = "Next.js project without Lighthouse CI config"
	r.Suggestion = "Add lighthouserc.json with perf/a11y/best-practices budgets"
	return r
}

// ─── No secrets committed (env files in tree) ───────────────────────────────
type secretsNotCommitted struct{}

func (secretsNotCommitted) ID() string       { return "security.no-committed-secrets" }
func (secretsNotCommitted) Category() string { return "Security" }
func (secretsNotCommitted) Run(_ context.Context, root string) audit.Result {
	r := audit.Result{ID: "security.no-committed-secrets", Category: "Security"}
	risky := []string{".env", ".env.local", ".env.production", ".env.development", "google-service-account.json"}
	found := []string{}
	for _, f := range risky {
		if exists(filepath.Join(root, f)) {
			found = append(found, f)
		}
	}
	if len(found) == 0 {
		r.Severity = audit.SeverityPass
		r.Message = "no risky env / credential files at root"
		return r
	}
	r.Severity = audit.SeverityFail
	r.Message = "files present that may contain secrets: " + strings.Join(found, ", ")
	r.Suggestion = "Move to a secret store and add to .gitignore. Rotate immediately if these were ever committed."
	return r
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func readPackageJSON(root string) (map[string]any, bool) {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return nil, false
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, false
	}
	return out, true
}

// stripJSONComments removes `//` line comments. Good enough for tsconfig.json.
func stripJSONComments(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if idx := strings.Index(line, "//"); idx >= 0 {
			// don't strip inside strings — simple heuristic: only at start or after whitespace
			if idx == 0 || strings.TrimSpace(line[:idx]) == "" || !strings.Contains(line[:idx], "\"") {
				lines[i] = line[:idx]
			}
		}
	}
	return strings.Join(lines, "\n")
}

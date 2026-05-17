# vigor-cli

Single-binary Vigor Digital scaffolder + audit. Mirrors the Claude Code `/vigor-*` slash commands for teams who work outside the editor.

Conforms to [ADR-0005 — Go as a scoped tier](https://github.com/VIGOR-Digital-Solution/vigor-boilerplate/blob/main/docs/adrs/0005-go-as-scoped-tier.md). CLI tooling is one of the three slots where Go is the right tool for the job.

## Install

```bash
# macOS / Linux — Homebrew tap
brew install vigor-digital-solution/tap/vigor

# Windows — scoop bucket
scoop bucket add vigor https://github.com/VIGOR-Digital-Solution/scoop-bucket
scoop install vigor

# Any POSIX — install script
curl -fsSL https://raw.githubusercontent.com/VIGOR-Digital-Solution/vigor-cli/main/install.sh | bash

# Docker
docker run --rm ghcr.io/vigor-digital-solution/vigor-cli:latest --help
```

## Commands

| Command                            | What it does                                                                                                                 |
| ---------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| `vigor scaffold <platform> <name>` | Fetch a template, replace placeholders, `git init` with three-branch model                                                   |
| `vigor scaffold --list`            | List available platforms                                                                                                     |
| `vigor audit [path]`               | Run the standards audit (10 checks across code-quality / testing / security / deployment / performance)                      |
| `vigor audit --json`               | Emit a machine-readable JSON report (good for CI)                                                                            |
| `vigor doctor`                     | Verify all six CLIs in the Vigor stack are installed (gh, vercel, railway, supabase, resend, hostinger) plus node, pnpm, git |
| `vigor upgrade`                    | Compare installed version against the latest GitHub release; print upgrade instructions                                      |
| `vigor version`                    | Print version + commit + build date + Go runtime                                                                             |

## Scaffolding

```bash
vigor scaffold web my-app
vigor scaffold backend my-api --ref v0.11.0
vigor scaffold ai my-ai --no-init
```

The CLI fetches the matching `templates/<platform>/` from `VIGOR-Digital-Solution/vigor-boilerplate` (override via `VIGOR_TEMPLATES_REPO` / `VIGOR_TEMPLATES_REF`), extracts only that subdirectory, replaces `{{PROJECT_NAME}}` everywhere (text files only — binaries skipped), and initialises git on the `production` branch with a `development` branch checked out.

## Audit

```bash
$ vigor audit
# Vigor Standards Audit
...
| Severity | Category    | Check                              | ...
| ✓        | Code quality| `code-quality.tsconfig-strict`     | ...
| ✗        | Testing     | `testing.playwright-config`        | ...
```

Exit code:

- `0` — every check is below the `--fail-on` threshold (default `fail`)
- `1` — at least one check at or above the threshold

CI usage:

```yaml
- run: vigor audit --fail-on=warn # stricter: warn or worse fails the build
```

JSON output for tooling:

```bash
vigor audit --json > audit.json
jq '.results[] | select(.severity == "fail")' audit.json
```

## Doctor

```bash
$ vigor doctor
Vigor CLI doctor

  ✓  gh           /opt/homebrew/bin/gh
  ✗  vercel       (install: https://vercel.com/docs/cli)
  ✓  railway      /opt/homebrew/bin/railway
  ✓  supabase     /opt/homebrew/bin/supabase
  ✗  resend       (install: https://resend.com/docs/dashboard/cli/introduction)
  ✓  hostinger    /opt/homebrew/bin/hostinger
  ✓  node         /opt/homebrew/bin/node
  ✓  pnpm         /opt/homebrew/bin/pnpm
  ✓  git          /usr/bin/git

! 2 CLI(s) missing.
```

Non-zero exit if any required CLI is missing. Use in onboarding to verify a fresh laptop is ready.

## Layout

```
tools/vigor-cli/
├── cmd/vigor/main.go            entry — wires ldflags into version pkg
├── internal/
│   ├── cli/                     cobra commands (root, scaffold, audit, doctor, upgrade, version)
│   ├── scaffold/                fetch (codeload tarball) + placeholders + post-init git
│   ├── audit/                   audit registry + Report + markdown / JSON writers
│   │   └── checks/              individual checks; each registers itself in init()
│   ├── version/                 build-time identifiers
│   └── log/                     (reserved for future)
├── .goreleaser.yaml             cross-compile + brew tap + scoop bucket + ghcr docker
├── install.sh                   POSIX one-liner
├── Dockerfile                   distroless static — ~15 MB
└── Makefile
```

## Build locally

```bash
make build          # ./bin/vigor with embedded version metadata
make test           # go test -race ./...
make lint           # golangci-lint
make snapshot       # goreleaser snapshot (no publish)
make install        # cp bin/vigor → $GOPATH/bin
```

## Releasing

Tag a release:

```bash
git tag v0.1.0
git push --tags
```

The `release.yml` workflow runs GoReleaser, which:

1. Cross-compiles for `linux/darwin/windows × amd64/arm64`
2. Uploads tarballs + zip + checksums to GitHub Releases
3. Updates the Homebrew tap (`VIGOR-Digital-Solution/homebrew-tap`)
4. Updates the scoop bucket (`VIGOR-Digital-Solution/scoop-bucket`)
5. Pushes a docker image to `ghcr.io/vigor-digital-solution/vigor-cli`

## Configuration

| Env var                | Default                                    | Purpose                            |
| ---------------------- | ------------------------------------------ | ---------------------------------- |
| `VIGOR_TEMPLATES_REPO` | `VIGOR-Digital-Solution/vigor-boilerplate` | Source repo for `scaffold`         |
| `VIGOR_TEMPLATES_REF`  | `main`                                     | Branch or tag                      |
| `VIGOR_INSTALL_DIR`    | `/usr/local/bin`                           | Where `install.sh` puts the binary |

## See also

- [`vigor-boilerplate`](https://github.com/VIGOR-Digital-Solution/vigor-boilerplate) — the standards + templates + Claude Code plugin
- [ADR-0005 — Go scoped tier](https://github.com/VIGOR-Digital-Solution/vigor-boilerplate/blob/main/docs/adrs/0005-go-as-scoped-tier.md)
- [Claude Code skills](https://github.com/VIGOR-Digital-Solution/vigor-boilerplate/tree/main/skills) — same surface, AI-routed

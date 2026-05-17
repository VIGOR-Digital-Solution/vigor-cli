// Package version exposes build-time identifiers injected via -ldflags.
package version

var (
	value  = "dev"
	commit = "unknown"
	date   = "unknown"
)

// Set is called once from main() with the goreleaser-injected values.
func Set(v, c, d string) {
	if v != "" {
		value = v
	}
	if c != "" {
		commit = c
	}
	if d != "" {
		date = d
	}
}

func Version() string { return value }
func Commit() string  { return commit }
func Date() string    { return date }

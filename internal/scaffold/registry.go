// Package scaffold owns the "fetch + replace + post-init" flow for new projects.
package scaffold

import "fmt"

// Platform identifies a template inside vigor-boilerplate.
type Platform string

const (
	PlatformWeb     Platform = "web"
	PlatformPWA     Platform = "pwa"
	PlatformMobile  Platform = "mobile"
	PlatformBackend Platform = "backend"
	PlatformAI      Platform = "ai"
	PlatformIoT     Platform = "iot"
)

// Template describes one entry in the platform → template registry.
type Template struct {
	Platform    Platform
	Description string
	// Subdir under the vigor-boilerplate repo
	RepoPath string
	// Default language tag printed in the doctor / help output
	Language string
}

var registry = map[Platform]Template{
	PlatformWeb: {
		Platform:    PlatformWeb,
		Description: "Next.js 16 web app (App Router, Tailwind v4, Supabase Auth, fumadocs + Swagger UI)",
		RepoPath:    "templates/web",
		Language:    "TypeScript",
	},
	PlatformPWA: {
		Platform:    PlatformPWA,
		Description: "Next.js 16 PWA — web template + Serwist + Web Push + Dexie + install UX",
		RepoPath:    "templates/pwa",
		Language:    "TypeScript",
	},
	PlatformMobile: {
		Platform:    PlatformMobile,
		Description: "Expo SDK 52 + React Native 0.76 with Expo Router + NativeWind + EAS profiles",
		RepoPath:    "templates/mobile",
		Language:    "TypeScript",
	},
	PlatformBackend: {
		Platform:    PlatformBackend,
		Description: "NestJS 11 backend with Drizzle + Zod + Passport JWT + Railway deploy",
		RepoPath:    "templates/backend",
		Language:    "TypeScript",
	},
	PlatformAI: {
		Platform:    PlatformAI,
		Description: "Python 3.12 AI service — FastAPI + Pydantic AI + Vercel AI Gateway + Langfuse + eval gate",
		RepoPath:    "templates/ai",
		Language:    "Python",
	},
	PlatformIoT: {
		Platform:    PlatformIoT,
		Description: "ESP-IDF firmware + Go edge gateway + NestJS ingestion + Grafana dashboards",
		RepoPath:    "templates/iot",
		Language:    "C/C++, Go, TypeScript",
	},
}

// Resolve returns the Template for a platform string, or an error listing valid options.
func Resolve(name string) (Template, error) {
	t, ok := registry[Platform(name)]
	if !ok {
		return Template{}, fmt.Errorf("unknown platform %q (valid: %s)", name, supported())
	}
	return t, nil
}

// All returns the registry sorted by platform name.
func All() []Template {
	order := []Platform{PlatformWeb, PlatformPWA, PlatformMobile, PlatformBackend, PlatformAI, PlatformIoT}
	out := make([]Template, 0, len(order))
	for _, p := range order {
		out = append(out, registry[p])
	}
	return out
}

func supported() string {
	all := All()
	s := ""
	for i, t := range all {
		if i > 0 {
			s += ", "
		}
		s += string(t.Platform)
	}
	return s
}

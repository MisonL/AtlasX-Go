package blueprint

import "strings"

const (
	productName = "AtlasX"
	targetHost  = "Intel x64 macOS"
)

type Product struct {
	Name         string
	ControlPlane string
	DataPlane    string
	Phases       []string
}

var lines = []string{
	productName + " Blueprint",
	"",
	"Positioning",
	"- Atlas-like desktop browser rebuilt for " + targetHost,
	"- Go owns the control plane",
	"- Chromium owns the browser data plane",
	"",
	"Current Scope",
	"- Diagnostics and local control plane bootstrap",
	"- Chrome webapp fallback launch with isolated or shared profile mode",
	"- Profile and config state rooted in Application Support",
	"",
	"Deferred Scope",
	"- CDP control",
	"- Import pipelines",
	"- Agent and browser memories",
}

func Default() Product {
	return Product{
		Name:         productName,
		ControlPlane: "Go",
		DataPlane:    "Chromium",
		Phases: []string{
			"Phase 0: Fallback",
			"Phase 1: Go Control Plane",
			"Phase 2: Browser Capability Takeover",
			"Phase 3: Managed Chromium Runtime",
			"Phase 4: Intelligence Layer",
		},
	}
}

func Render() string {
	return strings.Join(lines, "\n") + "\n"
}

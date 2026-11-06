// Package gfxaudit examines graphics-related macOS imports separately from
// Linux host graphics support. All runtime checks are read-only.
package gfxaudit

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/XenoTechWorks/Astelyn/internal/frameworkresolve"
	"github.com/XenoTechWorks/Astelyn/internal/inspection"
)

type Row struct {
	Framework          string   `json:"framework"`
	Role               string   `json:"role"`
	Import             string   `json:"import"`
	UsedBy             []string `json:"used_by"`
	State              string   `json:"state"`
	HostPath           string   `json:"host_path,omitempty"`
	ComparisonState    string   `json:"comparison_state,omitempty"`
	ComparisonHostPath string   `json:"comparison_host_path,omitempty"`
}

type HostAPI struct {
	API        string `json:"api"`
	Available  bool   `json:"available"`
	Diagnostic string `json:"diagnostic"`
	Detail     string `json:"detail"`
}

type Report struct {
	Archive                string    `json:"archive"`
	DarlingRoot            string    `json:"darling_root"`
	ComparisonRoot         string    `json:"comparison_root,omitempty"`
	GraphicsImports        []Row     `json:"graphics_imports"`
	Missing                int       `json:"missing"`
	Present                int       `json:"present"`
	ImprovedInComparison   []string  `json:"improved_in_comparison,omitempty"`
	HostAPIs               []HostAPI `json:"host_apis"`
	LoaderBlocked          bool      `json:"loader_blocked"`
	LoaderClassification   string    `json:"loader_classification,omitempty"`
	MissingLoaderLibraries []string  `json:"missing_loader_libraries,omitempty"`
	Notes                  []string  `json:"notes"`
}

var roles = map[string]string{
	"AppKit.framework":       "windowing",
	"Cocoa.framework":        "app-framework",
	"CoreGraphics.framework": "2d-graphics",
	"CoreText.framework":     "text",
	"CoreVideo.framework":    "frame-timing",
	"IOSurface.framework":    "surface-sharing",
	"Metal.framework":        "gpu-rendering",
	"OpenGL.framework":       "gpu-rendering",
	"QuartzCore.framework":   "compositing",
	"VideoToolbox.framework": "video-decoding",
}

func validateRoot(root string) error {
	if !filepath.IsAbs(root) || strings.ContainsAny(root, "\x00\n\r") {
		return errors.New("root must be absolute without control characters")
	}
	st, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("runtime root: %w", err)
	}
	if !st.IsDir() {
		return errors.New("runtime root is not a directory")
	}
	return nil
}

// Analyze accepts the existing static import inventory. HostAPIs are supplied
// separately by the CLI, never inferred from macOS framework file presence.
func Analyze(deps inspection.DependencyReport, root, comparison string, host []HostAPI, loaderClass string, missingLoader []string) (Report, error) {
	if err := validateRoot(root); err != nil {
		return Report{}, err
	}
	if comparison != "" {
		if err := validateRoot(comparison); err != nil {
			return Report{}, err
		}
	}
	p := frameworkresolve.Analyze(deps, root)
	result := Report{Archive: deps.Archive, DarlingRoot: root, ComparisonRoot: comparison, GraphicsImports: []Row{}, HostAPIs: append([]HostAPI{}, host...), LoaderClassification: loaderClass, MissingLoaderLibraries: append([]string{}, missingLoader...), Notes: []string{
		"Host Vulkan/OpenGL diagnostics do not prove macOS Metal, OpenGL, Cocoa, or Roblox compatibility.",
		"Framework path presence is not evidence of symbols, ABI/API behavior or working graphics.",
		"Only direct framework imports are covered; dyld dlopen and indirect dependencies may be absent.",
		"No frameworks, Roblox files, renderers, or Darling libraries are installed, replaced or patched.",
	}}
	cmp := map[string]frameworkresolve.Entry{}
	if comparison != "" {
		for _, e := range frameworkresolve.Analyze(deps, comparison).Entries {
			cmp[e.Import] = e
		}
	}
	for _, e := range p.Entries {
		role, ok := roles[e.Framework]
		if !ok {
			continue
		}
		row := Row{Framework: e.Framework, Role: role, Import: e.Import, UsedBy: append([]string{}, e.UsedBy...), State: e.State, HostPath: e.HostPath}
		if e.State == "missing" {
			result.Missing++
		} else {
			result.Present++
		}
		if comparison != "" {
			c, ok := cmp[e.Import]
			if !ok {
				row.ComparisonState = "missing"
			} else {
				row.ComparisonState = c.State
				row.ComparisonHostPath = c.HostPath
			}
			if e.State == "missing" && (row.ComparisonState == "exact" || row.ComparisonState == "alternate") {
				result.ImprovedInComparison = append(result.ImprovedInComparison, e.Framework)
			}
		}
		result.GraphicsImports = append(result.GraphicsImports, row)
	}
	sort.Slice(result.GraphicsImports, func(i, j int) bool { return result.GraphicsImports[i].Framework < result.GraphicsImports[j].Framework })
	sort.Strings(result.ImprovedInComparison)
	sort.Strings(result.MissingLoaderLibraries)
	if loaderClass == "missing-dependency-error" || len(missingLoader) > 0 {
		result.LoaderBlocked = true
		result.Notes = append(result.Notes, "Observed dyld dependency failure occurs before a graphics initialization test can be interpreted.")
	}
	return result, nil
}

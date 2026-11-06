// Package frameworkcheck inventories directly imported macOS frameworks.
// Finding a file does not establish that Darling implements its APIs.
package frameworkcheck

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/XenoTechWorks/Astelyn/internal/inspection"
)

type Entry struct {
	Import    string   `json:"import"`
	Framework string   `json:"framework"`
	UsedBy    []string `json:"used_by"`
	GuestPath string   `json:"guest_path,omitempty"`
	State     string   `json:"state"`
}

type Report struct {
	Archive    string  `json:"archive"`
	Root       string  `json:"darling_root"`
	Entries    []Entry `json:"entries"`
	Disclaimer string  `json:"disclaimer"`
}

const disclaimer = "File presence is not proof of Darling framework/API compatibility; imports are direct dependencies only."

// Analyze checks only safe, absolute guest install names against the given root.
// @rpath, @loader_path, and @executable_path require loader context and are not resolved.
func Analyze(deps inspection.DependencyReport, root string) Report {
	result := Report{Archive: deps.Archive, Root: root, Entries: []Entry{}, Disclaimer: disclaimer}
	for _, dep := range deps.Dependencies {
		at := strings.Index(dep.Name, ".framework/")
		if at < 0 {
			continue
		}
		pre := dep.Name[:at]
		framework := filepath.Base(pre) + ".framework"
		if framework == ".framework" {
			continue
		}
		entry := Entry{Import: dep.Name, Framework: framework, UsedBy: append([]string(nil), dep.UsedBy...), State: "unresolved"}
		if strings.HasPrefix(dep.Name, "@") || !filepath.IsAbs(dep.Name) || strings.ContainsRune(dep.Name, '\x00') {
			result.Entries = append(result.Entries, entry)
			continue
		}
		cleaned := filepath.Clean(dep.Name)
		// Only probe Apple's standard framework locations. This is not a dyld resolver.
		if !(strings.HasPrefix(cleaned, "/System/Library/Frameworks/") || strings.HasPrefix(cleaned, "/Library/Frameworks/")) || cleaned != dep.Name {
			result.Entries = append(result.Entries, entry)
			continue
		}
		host := filepath.Join(root, strings.TrimPrefix(cleaned, "/"))
		rootAbs, rootErr := filepath.Abs(root)
		hostAbs, hostErr := filepath.Abs(host)
		if rootErr != nil || hostErr != nil {
			result.Entries = append(result.Entries, entry)
			continue
		}
		rel, err := filepath.Rel(rootAbs, hostAbs)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			result.Entries = append(result.Entries, entry)
			continue
		}
		entry.GuestPath = cleaned
		entry.State = "missing"
		if info, err := os.Stat(hostAbs); err == nil && info.Mode().IsRegular() {
			entry.State = "file-present"
		}
		result.Entries = append(result.Entries, entry)
	}
	sort.Slice(result.Entries, func(i, j int) bool { return result.Entries[i].Import < result.Entries[j].Import })
	return result
}

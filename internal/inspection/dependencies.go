package inspection

import (
	"sort"
	"strings"
)

// Dependency groups linked Mach-O install names by their general origin.
// These labels are inventory categories, not claims that a library is supported.
type Dependency struct {
	Name     string   `json:"name"`
	Category string   `json:"category"`
	UsedBy   []string `json:"used_by"`
}

type DependencyReport struct {
	Archive      string       `json:"archive"`
	Binaries     int          `json:"binaries"`
	Dependencies []Dependency `json:"dependencies"`
}

func category(name string) string {
	switch {
	case strings.HasPrefix(name, "@rpath/"), strings.HasPrefix(name, "@loader_path/"), strings.HasPrefix(name, "@executable_path/"):
		return "relative-or-bundled"
	case strings.Contains(name, ".framework/"):
		return "apple-framework"
	case strings.HasPrefix(name, "/usr/lib/"), strings.HasPrefix(name, "/System/Library/"):
		return "system-library"
	default:
		return "other"
	}
}

// AnalyzeDependencies aggregates exactly what the static Mach-O dependency
// parser reports. It does not resolve paths or judge runtime compatibility.
func AnalyzeDependencies(report Report) DependencyReport {
	index := make(map[string]*Dependency)
	for _, binary := range report.Binaries {
		for _, name := range binary.Libraries {
			dep, ok := index[name]
			if !ok {
				dep = &Dependency{Name: name, Category: category(name), UsedBy: []string{}}
				index[name] = dep
			}
			found := false
			for _, path := range dep.UsedBy {
				if path == binary.Path {
					found = true
					break
				}
			}
			if !found {
				dep.UsedBy = append(dep.UsedBy, binary.Path)
			}
		}
	}
	result := DependencyReport{Archive: report.Archive, Binaries: len(report.Binaries), Dependencies: make([]Dependency, 0, len(index))}
	for _, dep := range index {
		sort.Strings(dep.UsedBy)
		result.Dependencies = append(result.Dependencies, *dep)
	}
	sort.Slice(result.Dependencies, func(i, j int) bool { return result.Dependencies[i].Name < result.Dependencies[j].Name })
	return result
}

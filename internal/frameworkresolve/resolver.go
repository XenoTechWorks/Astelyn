// Package frameworkresolve checks host-side Darling framework paths.
// File discovery does not establish that a framework is implemented or usable.
package frameworkresolve

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
	State     string   `json:"state"`
	HostPath  string   `json:"host_path,omitempty"`
}

type Report struct {
	Archive     string  `json:"archive"`
	DarlingRoot string  `json:"darling_root"`
	Entries     []Entry `json:"entries"`
	Disclaimer  string  `json:"disclaimer"`
}

// Analyze resolves direct macOS framework imports against a Darling root.
// It does not load binaries, install frameworks, or assess API compatibility.
func Analyze(deps inspection.DependencyReport, root string) Report {
	report := Report{
		Archive: deps.Archive, DarlingRoot: root,
		Entries:    []Entry{},
		Disclaimer: "File presence, including alternate versions, does not prove ABI, API or runtime compatibility. Only direct imports are examined.",
	}
	for _, dep := range deps.Dependencies {
		dir, framework, suffix, ok := parseImport(dep.Name)
		if !ok {
			continue
		}
		state, host := resolve(root, dir, framework, suffix)
		report.Entries = append(report.Entries, Entry{
			Import: dep.Name, Framework: framework,
			UsedBy: append([]string{}, dep.UsedBy...), State: state, HostPath: host,
		})
	}
	return report
}

// Only recognized macOS framework roots are considered, to prevent imported
// paths from escaping the configured Darling root.
func parseImport(name string) (dir, framework, suffix string, ok bool) {
	var root string
	switch {
	case strings.HasPrefix(name, "/System/Library/Frameworks/"):
		root = "/System/Library/Frameworks/"
	case strings.HasPrefix(name, "/Library/Frameworks/"):
		root = "/Library/Frameworks/"
	default:
		return "", "", "", false
	}
	remaining := strings.TrimPrefix(name, root)
	i := strings.Index(remaining, ".framework/")
	if i <= 0 {
		return "", "", "", false
	}
	framework = remaining[:i] + ".framework"
	suffix = remaining[i+len(".framework/"):]
	if framework == ".framework" || strings.ContainsAny(remaining, "\\\x00") ||
		strings.Contains(framework, "/") || suffix == "" {
		return "", "", "", false
	}
	for _, part := range strings.Split(suffix, "/") {
		if part == "" || part == "." || part == ".." {
			return "", "", "", false
		}
	}
	return strings.TrimSuffix(root, "/"), framework, suffix, true
}

func resolve(root, dir, framework, suffix string) (string, string) {
	folder := filepath.Join(root, strings.TrimPrefix(dir, "/"), framework)
	candidates := []string{filepath.Join(folder, filepath.FromSlash(suffix))}
	binary := filepath.Base(framework[:len(framework)-len(".framework")])
	if strings.HasPrefix(suffix, "Versions/") {
		candidates = append(candidates, filepath.Join(folder, binary))
		candidates = append(candidates, filepath.Join(folder, "Versions", "Current", binary))
	} else {
		candidates = append(candidates, filepath.Join(folder, "Versions", "Current", binary))
	}
	versions, err := os.ReadDir(filepath.Join(folder, "Versions"))
	if err == nil {
		names := []string{}
		for _, v := range versions {
			if v.Name() != "Current" && !strings.Contains(v.Name(), ".") {
				names = append(names, v.Name())
			}
		}
		sort.Strings(names)
		for _, version := range names {
			candidates = append(candidates, filepath.Join(folder, "Versions", version, binary))
		}
	}
	for i, candidate := range candidates {
		if safeRegularFile(root, candidate) {
			if i == 0 {
				return "exact", candidate
			}
			return "alternate", candidate
		}
	}
	return "missing", ""
}

func safeRegularFile(root, candidate string) bool {
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return false
	}
	realFile, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(realRoot, realFile)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return false
	}
	info, err := os.Stat(realFile)
	return err == nil && info.Mode().IsRegular()
}

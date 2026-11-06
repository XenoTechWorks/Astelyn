// Package dyldreport classifies errors from a captured Darling launch.
// Analysis is read-only and never attempts to repair or bypass missing libraries.
package dyldreport

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const MaxInputBytes = 2 << 20

type Launch struct {
	HostExecutable  string `json:"host_executable"`
	GuestExecutable string `json:"guest_executable"`
	ExitCode        int    `json:"exit_code"`
	TimedOut        bool   `json:"timed_out"`
	Stdout          string `json:"stdout"`
	Stderr          string `json:"stderr"`
	Error           string `json:"error"`
	Note            string `json:"note"`
}

type Dependency struct {
	InstallName    string `json:"install_name"`
	Framework      string `json:"framework,omitempty"`
	ReferencedFrom string `json:"referenced_from,omitempty"`
	Reason         string `json:"reason,omitempty"`
	HostPath       string `json:"host_path,omitempty"`
	State          string `json:"state"`
}

type Report struct {
	DarlingRoot        string       `json:"darling_root"`
	ExitCode           int          `json:"exit_code"`
	TimedOut           bool         `json:"timed_out"`
	Classification     string       `json:"classification"`
	SharedCacheWarning bool         `json:"shared_cache_warning"`
	Dependencies       []Dependency `json:"dependencies"`
	Notes              []string     `json:"notes"`
}

var (
	missingPattern    = regexp.MustCompile(`(?m)(?:dyld:\s*)?Library not loaded:\s*(\S+)`)
	reasonPattern     = regexp.MustCompile(`(?m)^\s*Reason:\s*([^\n\r]+)`)
	referencedPattern = regexp.MustCompile(`(?m)^\s*Referenced from:\s*([^\n\r]+)`)
)

func Decode(r io.Reader) (Launch, error) {
	var launch Launch
	d := json.NewDecoder(io.LimitReader(r, MaxInputBytes+1))
	d.DisallowUnknownFields() // strict so accidental wrong report shapes are surfaced
	if err := d.Decode(&launch); err != nil {
		return launch, fmt.Errorf("read launch JSON: %w", err)
	}
	var trailing interface{}
	if err := d.Decode(&trailing); err != io.EOF {
		return launch, errors.New("expected exactly one launch JSON object")
	}
	return launch, nil
}

func frameworkName(install string) string {
	for _, part := range strings.Split(install, "/") {
		if strings.HasSuffix(part, ".framework") {
			return strings.TrimSuffix(part, ".framework")
		}
	}
	return ""
}

func inspect(root, install string) (string, string) {
	// @rpath and other loader expansions cannot be resolved statically.
	if !filepath.IsAbs(install) || strings.ContainsAny(install, "\x00\r\n") || filepath.Clean(install) != install {
		return "unresolved", ""
	}
	rootResolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "root-unavailable", ""
	}
	candidate := filepath.Join(root, strings.TrimPrefix(install, "/"))
	resolved, err := filepath.EvalSymlinks(candidate)
	if errors.Is(err, os.ErrNotExist) {
		return "missing", ""
	}
	if err != nil {
		return "unreadable", ""
	}
	rel, err := filepath.Rel(rootResolved, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "unsafe-symlink", ""
	}
	st, err := os.Stat(resolved)
	if err != nil {
		return "unreadable", ""
	}
	if !st.Mode().IsRegular() {
		return "not-regular", ""
	}
	return "present-file", resolved
}

// Analyze considers only errors observed in the supplied launch log.
// It does not enumerate every dependency in the application.
func Analyze(launch Launch, root string) (Report, error) {
	if !filepath.IsAbs(root) {
		return Report{}, errors.New("Darling root must be an absolute path")
	}
	info, err := os.Stat(root)
	if err != nil {
		return Report{}, fmt.Errorf("Darling root: %w", err)
	}
	if !info.IsDir() {
		return Report{}, errors.New("Darling root is not a directory")
	}
	text := launch.Stdout + "\n" + launch.Stderr
	result := Report{DarlingRoot: root, ExitCode: launch.ExitCode, TimedOut: launch.TimedOut, Dependencies: []Dependency{}, Notes: []string{}}
	result.SharedCacheWarning = strings.Contains(text, "No shared cache present")
	seen := map[string]struct{}{}
	for _, m := range missingPattern.FindAllStringSubmatchIndex(text, -1) {
		name := text[m[2]:m[3]]
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		// Bind the reason and referencing image to this particular loader message only.
		tail := text[m[1]:]
		if next := missingPattern.FindStringIndex(tail); next != nil {
			tail = tail[:next[0]]
		}
		dep := Dependency{InstallName: name, Framework: frameworkName(name)}
		if ref := referencedPattern.FindStringSubmatch(tail); ref != nil {
			dep.ReferencedFrom = strings.TrimSpace(ref[1])
		}
		if why := reasonPattern.FindStringSubmatch(tail); why != nil {
			dep.Reason = strings.TrimSpace(why[1])
		}
		dep.State, dep.HostPath = inspect(root, name)
		result.Dependencies = append(result.Dependencies, dep)
	}
	sort.Slice(result.Dependencies, func(i, j int) bool { return result.Dependencies[i].InstallName < result.Dependencies[j].InstallName })
	switch {
	case len(result.Dependencies) > 0:
		result.Classification = "missing-dependency-error"
	case launch.TimedOut:
		result.Classification = "timeout"
	case launch.ExitCode == 0 && launch.Error == "":
		result.Classification = "exited-zero-no-loader-error"
	default:
		result.Classification = "unclassified-launch-failure"
	}
	result.Notes = append(result.Notes, "Observed dyld failures only; additional missing dependencies may appear after this blocker is addressed.")
	if result.SharedCacheWarning {
		result.Notes = append(result.Notes, "dyld reported no shared cache; this is recorded separately and not treated as proof of the primary failure.")
	}
	result.Notes = append(result.Notes, "File existence does not demonstrate ABI/API compatibility. No libraries are installed, replaced, or bypassed.")
	return result, nil
}

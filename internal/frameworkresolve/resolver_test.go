package frameworkresolve

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/XenoTechWorks/Astelyn/internal/inspection"
)

func write(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("stub"), 0644); err != nil {
		t.Fatal(err)
	}
}

func reportFor(names ...string) inspection.DependencyReport {
	deps := inspection.DependencyReport{Archive: "test.zip"}
	for _, name := range names {
		deps.Dependencies = append(deps.Dependencies, inspection.Dependency{Name: name, UsedBy: []string{"App"}})
	}
	return deps
}

func TestExactAlternateAndMissing(t *testing.T) {
	root := t.TempDir()
	base := filepath.Join(root, "System/Library/Frameworks")
	write(t, filepath.Join(base, "CoreFoundation.framework/Versions/A/CoreFoundation"))
	write(t, filepath.Join(base, "AppKit.framework/Versions/A/AppKit"))
	r := Analyze(reportFor(
		"/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation",
		"/System/Library/Frameworks/AppKit.framework/Versions/C/AppKit",
		"/System/Library/Frameworks/Metal.framework/Versions/A/Metal",
	), root)
	if len(r.Entries) != 3 {
		t.Fatalf("entries: %+v", r.Entries)
	}
	for i, want := range []string{"exact", "alternate", "missing"} {
		if r.Entries[i].State != want {
			t.Errorf("entry %d: got %q want %q", i, r.Entries[i].State, want)
		}
	}
}

func TestSymlinksAndEscape(t *testing.T) {
	root := t.TempDir()
	base := filepath.Join(root, "System/Library/Frameworks", "Foundation.framework")
	target := filepath.Join(base, "Versions", "A", "Foundation")
	write(t, target)
	if err := os.Symlink("A", filepath.Join(base, "Versions", "Current")); err != nil {
		t.Fatal(err)
	}
	if got, _ := resolve(root, "/System/Library/Frameworks", "Foundation.framework", "Versions/C/Foundation"); got != "alternate" {
		t.Fatal(got)
	}
	outside := filepath.Join(t.TempDir(), "Outside")
	write(t, outside)
	if err := os.MkdirAll(filepath.Join(root, "Library", "Frameworks", "Escape.framework"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "Library", "Frameworks", "Escape.framework", "Escape")); err != nil {
		t.Fatal(err)
	}
	if got, _ := resolve(root, "/Library/Frameworks", "Escape.framework", "Escape"); got != "missing" {
		t.Fatalf("escaped root: %s", got)
	}
}

func TestInvalidImportsIgnored(t *testing.T) {
	root := t.TempDir()
	r := Analyze(reportFor("@rpath/AppKit.framework/AppKit", "/System/Library/Frameworks/../evil.framework/X", "/System/Library/Frameworks/Foo.framework/../bad", "/System/Library/Frameworks/X.framework/Versions/A/X"), root)
	if len(r.Entries) != 1 || r.Entries[0].State != "missing" {
		t.Fatalf("unexpected entries: %+v", r.Entries)
	}
}

package frameworkcheck

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/XenoTechWorks/Astelyn/internal/inspection"
)

func TestAnalyzeFrameworkFiles(t *testing.T) {
	root := t.TempDir()
	guest := "/System/Library/Frameworks/Foundation.framework/Versions/C/Foundation"
	host := filepath.Join(root, "System/Library/Frameworks/Foundation.framework/Versions/C/Foundation")
	if err := os.MkdirAll(filepath.Dir(host), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(host, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	deps := inspection.DependencyReport{Archive: "app.zip", Dependencies: []inspection.Dependency{
		{Name: guest, UsedBy: []string{"App"}},
		{Name: "/System/Library/Frameworks/AppKit.framework/Versions/C/AppKit", UsedBy: []string{"App"}},
		{Name: "/usr/lib/libSystem.B.dylib", UsedBy: []string{"App"}},
	}}
	report := Analyze(deps, root)
	if len(report.Entries) != 2 {
		t.Fatalf("entries: %+v", report.Entries)
	}
	if report.Entries[0].Framework != "AppKit.framework" || report.Entries[0].State != "missing" {
		t.Fatalf("missing: %+v", report.Entries[0])
	}
	if report.Entries[1].Framework != "Foundation.framework" || report.Entries[1].State != "file-present" || report.Entries[1].GuestPath != guest {
		t.Fatalf("present: %+v", report.Entries[1])
	}
}

func TestUnresolvedFrameworkPaths(t *testing.T) {
	deps := inspection.DependencyReport{Dependencies: []inspection.Dependency{
		{Name: "@rpath/Metal.framework/Metal"},
		{Name: "/System/Library/Frameworks/../PrivateFrameworks/Hidden.framework/Hidden"},
		{Name: "relative/Example.framework/Example"},
	}}
	report := Analyze(deps, t.TempDir())
	if len(report.Entries) != 3 {
		t.Fatalf("entries: %+v", report.Entries)
	}
	for _, entry := range report.Entries {
		if entry.State != "unresolved" || entry.GuestPath != "" {
			t.Fatalf("unsafe resolution: %+v", entry)
		}
	}
}

func TestFileDirectoryNotPresent(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "System/Library/Frameworks/Example.framework/Example")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	deps := inspection.DependencyReport{Dependencies: []inspection.Dependency{{Name: "/System/Library/Frameworks/Example.framework/Example"}}}
	if state := Analyze(deps, root).Entries[0].State; state != "missing" {
		t.Fatalf("directory was considered a library: %q", state)
	}
}

package gfxaudit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/XenoTechWorks/Astelyn/internal/inspection"
)

func TestAnalyzeGraphicsAndCompare(t *testing.T) {
	root := t.TempDir()
	comparison := t.TempDir()
	dest := filepath.Join(comparison, "System/Library/Frameworks/Metal.framework/Versions/A")
	if err := os.MkdirAll(dest, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "Metal"), []byte("file presence only"), 0600); err != nil {
		t.Fatal(err)
	}
	deps := inspection.DependencyReport{Archive: "Roblox.zip", Dependencies: []inspection.Dependency{
		{Name: "/System/Library/Frameworks/Metal.framework/Versions/A/Metal", UsedBy: []string{"RobloxPlayer"}},
		{Name: "/System/Library/Frameworks/CoreML.framework/Versions/A/CoreML"},
		{Name: "/System/Library/Frameworks/AppKit.framework/Versions/C/AppKit"},
	}}
	r, err := Analyze(deps, root, comparison, []HostAPI{{API: "vulkan", Available: true}}, "missing-dependency-error", []string{"CoreML"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Missing != 2 || r.Present != 0 || len(r.GraphicsImports) != 2 {
		t.Fatalf("unexpected report %+v", r)
	}
	if len(r.ImprovedInComparison) != 1 || r.ImprovedInComparison[0] != "Metal.framework" {
		t.Fatalf("comparison %+v", r.ImprovedInComparison)
	}
	if !r.LoaderBlocked || !r.HostAPIs[0].Available {
		t.Fatalf("expected blocked host-independent report: %+v", r)
	}
}
func TestRejectInvalidRoot(t *testing.T) {
	_, err := Analyze(inspection.DependencyReport{}, "relative", "", nil, "", nil)
	if err == nil {
		t.Fatal("relative root accepted")
	}
}

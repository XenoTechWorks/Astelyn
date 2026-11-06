package runtimecompare

import (
	"strings"
	"testing"
	"time"

	"github.com/XenoTechWorks/Astelyn/internal/dyldreport"
)

func TestComparison(t *testing.T) {
	a := dyldreport.Report{Classification: "missing-dependency-error", Dependencies: []dyldreport.Dependency{{InstallName: "/System/Library/Frameworks/CoreML.framework/Versions/A/CoreML"}}}
	b := dyldreport.Report{Classification: "missing-dependency-error", Dependencies: []dyldreport.Dependency{{InstallName: "/System/Library/Frameworks/AppKit.framework/Versions/C/AppKit"}}}
	got := Compare(a, b)
	if got.SameFirstBlocker {
		t.Fatal("different first blocker reported as same")
	}
	if len(got.BaselineLibraries) != 1 || len(got.CandidateLibraries) != 1 {
		t.Fatalf("missing comparison info: %+v", got)
	}
	if !Compare(a, a).SameFirstBlocker {
		t.Fatal("identical first blocker not detected")
	}
}
func TestValidateBadInputs(t *testing.T) {
	tests := []Request{
		{Launcher: "missing-launcher-astelyn-should-not-exist", Root: "/tmp", App: "/tmp/foo", Timeout: time.Second},
		{Launcher: "/bin/true", Root: "relative", App: "/tmp/foo", Timeout: time.Second},
		{Launcher: "/bin/true", Root: "/tmp", App: "/tmp/foo", Timeout: 0},
	}
	for _, r := range tests {
		if _, err := Validate(r, false); err == nil {
			t.Fatalf("expected validation failure for %+v", r)
		}
	}
}
func TestLaunchJSON(t *testing.T) {
	j := `{"host_executable":"/a","exit_code":1,"stdout":"Library not loaded: /CoreML","note":"safe"}`
	x, err := DecodeLaunch(strings.NewReader(j))
	if err != nil {
		t.Fatal(err)
	}
	if x.ExitCode != 1 {
		t.Fatal("decode exit")
	}
}

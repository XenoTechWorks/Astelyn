package inspection

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAnalyzeDependencies(t *testing.T) {
	r := Report{Archive: "sample.zip", Binaries: []Binary{
		{Path: "App/Contents/MacOS/Main", Libraries: []string{"/System/Library/Frameworks/AppKit.framework/Versions/C/AppKit", "@rpath/libexample.dylib", "/usr/lib/libSystem.B.dylib"}},
		{Path: "App/Contents/MacOS/Helper", Libraries: []string{"/usr/lib/libSystem.B.dylib", "/usr/lib/libSystem.B.dylib"}},
	}}
	got := AnalyzeDependencies(r)
	if got.Binaries != 2 || len(got.Dependencies) != 3 {
		t.Fatalf("unexpected report: %+v", got)
	}
	if got.Dependencies[0].Category != "apple-framework" {
		t.Fatal(got.Dependencies[0])
	}
	for _, d := range got.Dependencies {
		if strings.Contains(d.Name, "libSystem") && (d.Category != "system-library" || len(d.UsedBy) != 2) {
			t.Fatalf("unexpected dependency: %+v", d)
		}
	}
	if _, err := json.Marshal(got); err != nil {
		t.Fatal(err)
	}
}
func TestCategory(t *testing.T) {
	for _, tc := range []struct{ name, want string }{
		{"@loader_path/libfoo.dylib", "relative-or-bundled"},
		{"/System/Library/Frameworks/Metal.framework/Metal", "apple-framework"},
		{"/usr/lib/libobjc.A.dylib", "system-library"},
		{"libcustom.dylib", "other"},
	} {
		if got := category(tc.name); got != tc.want {
			t.Fatalf("%s: got %s", tc.name, got)
		}
	}
}

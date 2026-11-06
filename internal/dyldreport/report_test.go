package dyldreport

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCoreMLRealWorldLog(t *testing.T) {
	root := t.TempDir()
	launch := Launch{ExitCode: 1, Stdout: "abort_with_payload: reason: dyld: No shared cache present\nLibrary not loaded: /System/Library/Frameworks/CoreML.framework/Versions/A/CoreML\n  Referenced from: /Volumes/SystemRoot/home/animated/Games/Astelyn/RobloxPlayer.app/Contents/MacOS/RobloxPlayer\n  Reason: image not found; code: 1\n", Stderr: "dyld: Library not loaded: /System/Library/Frameworks/CoreML.framework/Versions/A/CoreML\n  Reason: image not found\n"}
	r, err := Analyze(launch, root)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Dependencies) != 1 || r.Dependencies[0].Framework != "CoreML" || r.Dependencies[0].State != "missing" || r.Dependencies[0].ReferencedFrom == "" || !r.SharedCacheWarning {
		t.Fatalf("wrong report: %#v", r)
	}
}
func TestPresentAndUnsafeSymlink(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "usr", "lib")
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "libok.dylib"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "libsecret.dylib")
	if err := os.WriteFile(outside, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(path, "libbad.dylib")); err != nil {
		t.Fatal(err)
	}
	log := "Library not loaded: /usr/lib/libok.dylib\n  Reason: test\nLibrary not loaded: /usr/lib/libbad.dylib\n  Reason: test"
	r, err := Analyze(Launch{Stdout: log}, root)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Dependencies) != 2 || r.Dependencies[0].State != "unsafe-symlink" || r.Dependencies[1].State != "present-file" {
		t.Fatalf("unexpected: %+v", r.Dependencies)
	}
}
func TestDecode(t *testing.T) {
	if _, err := Decode(strings.NewReader(`{"exit_code":1,"stdout":"hello"}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := Decode(strings.NewReader(`{"exit_code":1}{}`)); err == nil {
		t.Fatal("expected trailing JSON error")
	}
	if _, err := Analyze(Launch{}, "relative/path"); err == nil {
		t.Fatal("expected root error")
	}
}

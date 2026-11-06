package objcprobe

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMissingLibraries(t *testing.T) {
	root := t.TempDir()
	r, e := Analyze(root)
	if e != nil {
		t.Fatal(e)
	}
	if len(r.Libraries) != 3 {
		t.Fatalf("got %d libraries", len(r.Libraries))
	}
	for _, l := range r.Libraries {
		if l.Present {
			t.Fatalf("unexpected library: %+v", l)
		}
	}
}
func TestRejectInvalidRoot(t *testing.T) {
	if _, e := Analyze("relative"); e == nil {
		t.Fatal("accepted relative root")
	}
}
func TestRejectSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "usr/lib")
	if err := os.MkdirAll(p, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/etc/passwd", filepath.Join(p, "libobjc.A.dylib")); err != nil {
		t.Fatal(err)
	}
	l := inspect(root, "/usr/lib/libobjc.A.dylib")
	if l.Present || l.Error == "" {
		t.Fatalf("unsafe link: %+v", l)
	}
}
func TestRejectFileAsRoot(t *testing.T) {
	p := filepath.Join(t.TempDir(), "file")
	os.WriteFile(p, []byte("x"), 0600)
	if _, e := Analyze(p); e == nil {
		t.Fatal("accepted file")
	}
}

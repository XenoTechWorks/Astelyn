package runtimecheck

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyze(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "usr/lib/libobjc.A.dylib")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte{0x7f, 'E', 'L', 'F'}, 0600); err != nil {
		t.Fatal(err)
	}
	r, err := Analyze(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Components) != len(targets) {
		t.Fatalf("wrong count %d", len(r.Components))
	}
	if !r.Components[0].Present || r.Components[0].Format != "ELF" {
		t.Fatalf("unexpected component: %+v", r.Components[0])
	}
	if r.Components[10].Present {
		t.Fatal("AppKit should be missing")
	}
}
func TestSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "lib")
	if err := os.WriteFile(outside, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(root, "usr/lib")
	if err := os.MkdirAll(dest, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dest, "libobjc.A.dylib")); err != nil {
		t.Fatal(err)
	}
	r, err := Analyze(root)
	if err != nil {
		t.Fatal(err)
	}
	if r.Components[0].Present {
		t.Fatal("escaped file marked present")
	}
}
func TestInvalidRoot(t *testing.T) {
	if _, err := Analyze(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected error")
	}
}

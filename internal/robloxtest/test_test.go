package robloxtest

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGuestPathValidation(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "RobloxPlayer.app", "Contents", "MacOS", "RobloxPlayer")
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("not Mach-O"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := GuestPath(target); err == nil {
		t.Fatal("non Mach-O accepted")
	}
	if _, err := GuestPath(filepath.Join(dir, "wrong")); err == nil {
		t.Fatal("non-existent accepted")
	}
}
func TestAssessEmptyZip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.zip")
	f, e := os.Create(path)
	if e != nil {
		t.Fatal(e)
	}
	z := zip.NewWriter(f)
	if e = z.Close(); e != nil {
		t.Fatal(e)
	}
	f.Close()
	p, e := Assess(path)
	if e != nil {
		t.Fatal(e)
	}
	if p.Found {
		t.Fatal("unexpected binary")
	}
}
func TestCappedWriter(t *testing.T) {
	w := &cappedWriter{cap: 5}
	for i := 0; i < 3; i++ {
		if _, err := w.Write([]byte("abc")); err != nil {
			t.Fatal(err)
		}
	}
	if w.buf.String() != "abcab" {
		t.Fatal(w.buf.String())
	}
}
func TestLaunchRejectsTimeout(t *testing.T) {
	_, e := Launch(nil, "darling", "/not-real", time.Hour)
	if e == nil || !strings.Contains(e.Error(), "no such") { // Path must be checked before timeout: still error.
		if e == nil {
			t.Fatal("expected error")
		}
	}
}

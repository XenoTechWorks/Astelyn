package darling

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestReadinessCompleteFixture(t *testing.T) {
	root := t.TempDir()
	bash := filepath.Join(root, "bin", "bash")
	if err := os.MkdirAll(filepath.Dir(bash), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bash, []byte{0xcf, 0xfa, 0xed, 0xfe}, 0600); err != nil {
		t.Fatal(err)
	}
	launcher := filepath.Join(t.TempDir(), "darling")
	if err := os.WriteFile(launcher, []byte("#!/bin/sh\nexit 0\n"), 0700); err != nil {
		t.Fatal(err)
	}
	result := CheckReadiness(launcher, root)
	if !result.LauncherFound || !result.SystemRootFound || !result.GuestBashMachO {
		t.Fatalf("unexpected report: %+v", result)
	}
	if result.ReadyForSmokeTest != (runtime.GOOS == "linux") {
		t.Fatalf("unexpected readiness: %+v", result)
	}
}

func TestReadinessMissingPaths(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing")
	result := CheckReadiness(filepath.Join(root, "darling"), root)
	if result.LauncherFound || result.SystemRootFound || result.GuestBashMachO || result.ReadyForSmokeTest {
		t.Fatalf("false positive: %+v", result)
	}
	if len(result.Notes) < 2 {
		t.Fatalf("missing diagnostics: %+v", result)
	}
}

func TestReadinessRejectsELFGuest(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "bin"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "bin", "bash"), []byte("\x7fELF"), 0600); err != nil {
		t.Fatal(err)
	}
	result := CheckReadiness(filepath.Join(root, "absent"), root)
	if result.GuestBashMachO {
		t.Fatal("accepted ELF as Mach-O")
	}
	if !strings.Contains(strings.Join(result.Notes, " "), "Mach-O") {
		t.Fatal("missing header diagnostic")
	}
}

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPreflightMissingPaths(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	if got := runPreflight([]string{"--root", missing, "--launcher", missing}); got != 1 {
		t.Fatalf("expected diagnostic failure exit code 1, got %d", got)
	}
}

func TestPreflightUsage(t *testing.T) {
	if got := runPreflight([]string{"--not-a-flag"}); got != 2 {
		t.Fatalf("expected usage exit code 2, got %d", got)
	}
	if got := runPreflight([]string{"unexpected"}); got != 2 {
		t.Fatalf("expected usage exit code 2, got %d", got)
	}
}

func TestPreflightFixtureJSON(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "bin"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "bin", "bash"), []byte{0xcf, 0xfa, 0xed, 0xfe}, 0600); err != nil {
		t.Fatal(err)
	}
	launcher := filepath.Join(t.TempDir(), "darling")
	if err := os.WriteFile(launcher, []byte("#!/bin/sh\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if got := runPreflight([]string{"--json", "--root", root, "--launcher", launcher}); got != 0 {
		t.Fatalf("expected preflight success on Linux, got %d", got)
	}
}

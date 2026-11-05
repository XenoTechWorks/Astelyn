package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOverlayPersistence(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if n := runGraphicsExtra([]string{"overlay", "set", "on"}); n != 0 {
		t.Fatalf("set %d", n)
	}
	if n := runGraphicsExtra([]string{"overlay", "status"}); n != 0 {
		t.Fatalf("status %d", n)
	}
	p := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "astelyn", "performance.json")
	if _, e := os.Stat(p); e != nil {
		t.Fatal(e)
	}
	if n := runGraphicsExtra([]string{"overlay", "set", "banana"}); n != 2 {
		t.Fatalf("invalid args %d", n)
	}
}
func TestNativeRunExitCode(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	script := filepath.Join(t.TempDir(), "test-command")
	if e := os.WriteFile(script, []byte("#!/bin/sh\nexit 17\n"), 0700); e != nil {
		t.Fatal(e)
	}
	if got := runNativeTest([]string{"--", script}); got != 17 {
		t.Fatalf("exit %d", got)
	}
	if got := runNativeTest([]string{"--"}); got != 2 {
		t.Fatalf("usage %d", got)
	}
}

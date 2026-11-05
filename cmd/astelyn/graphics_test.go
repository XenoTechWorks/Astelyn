package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGraphicsCLI(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	if n := run([]string{"graphics", "set", "opengl"}); n != 0 {
		t.Fatalf("set: %d", n)
	}
	if n := run([]string{"graphics", "get"}); n != 0 {
		t.Fatalf("get: %d", n)
	}
	if n := run([]string{"graphics", "set", "metal"}); n != 1 {
		t.Fatalf("invalid backend: %d", n)
	}
	if _, err := os.Stat(filepath.Join(dir, "astelyn", "graphics.json")); err != nil {
		t.Fatal(err)
	}
	if n := run([]string{"graphics", "set"}); n != 2 {
		t.Fatalf("usage: %d", n)
	}
}

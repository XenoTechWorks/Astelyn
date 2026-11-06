package hostwindow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDriverEnv(t *testing.T) {
	input := []string{"HOME=/tmp", "SDL_VIDEODRIVER=old", "DISPLAY=:1"}
	got, err := DriverEnv("wayland", input)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(got, "|") != "HOME=/tmp|DISPLAY=:1|SDL_VIDEODRIVER=wayland" {
		t.Fatalf("got: %v", got)
	}
	auto, err := DriverEnv("auto", input)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(auto, "|"), "SDL_VIDEODRIVER=") {
		t.Fatal(auto)
	}
	if _, err := DriverEnv("cocoa", input); err == nil {
		t.Fatal("allowed unsupported driver")
	}
}
func TestBinaryPath(t *testing.T) {
	if got := BinaryPath("/tmp/example"); got != "/tmp/example/.astelyn/build/astelyn-host-window" {
		t.Fatal(got)
	}
}
func TestRunRejected(t *testing.T) {
	if err := Run(t.TempDir(), "auto", 0); err == nil {
		t.Fatal("accepted invalid seconds")
	}
	if err := Run(t.TempDir(), "cocoa", 1); err == nil {
		t.Fatal("accepted driver")
	}
	if err := Run(t.TempDir(), "auto", 1); err == nil {
		t.Fatal("ran missing executable")
	}
}
func TestRunSmokeStub(t *testing.T) {
	// A fake executable confirms argv and successful exit; does not test a display.
	root := t.TempDir()
	path := BinaryPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\ntest \"$1\" = --seconds && test \"$2\" = 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := Run(root, "x11", 1); err != nil {
		t.Fatal(err)
	}
}

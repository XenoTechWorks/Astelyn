// Package hostwindow manages native Linux window test prerequisites.
// It does not implement AppKit, NSWindow, a macOS graphics driver, or Roblox rendering.
package hostwindow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Status struct {
	Display          string `json:"display"`
	WaylandAvailable bool   `json:"wayland_available"`
	X11Available     bool   `json:"x11_available"`
	SDL2Development  bool   `json:"sdl2_development"`
	PkgConfig        bool   `json:"pkg_config"`
	Note             string `json:"note"`
}

func Detect() Status {
	_, pkgErr := exec.LookPath("pkg-config")
	out := Status{
		Display:          "none",
		WaylandAvailable: os.Getenv("WAYLAND_DISPLAY") != "" && os.Getenv("XDG_RUNTIME_DIR") != "",
		X11Available:     os.Getenv("DISPLAY") != "",
		PkgConfig:        pkgErr == nil,
		Note:             "Host Linux display only; macOS Cocoa/AppKit bridging is not implemented",
	}
	if out.WaylandAvailable {
		out.Display = "wayland"
	} else if out.X11Available {
		out.Display = "x11"
	}
	if out.PkgConfig {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		out.SDL2Development = exec.CommandContext(ctx, "pkg-config", "--exists", "sdl2").Run() == nil
	}
	return out
}

func DriverEnv(driver string, inherited []string) ([]string, error) {
	if driver != "auto" && driver != "wayland" && driver != "x11" {
		return nil, fmt.Errorf("unsupported driver %q (use auto, wayland, or x11)", driver)
	}
	env := make([]string, 0, len(inherited)+1)
	for _, v := range inherited {
		if !strings.HasPrefix(v, "SDL_VIDEODRIVER=") {
			env = append(env, v)
		}
	}
	if driver != "auto" {
		env = append(env, "SDL_VIDEODRIVER="+driver)
	}
	return env, nil
}

func BinaryPath(root string) string {
	return filepath.Join(root, ".astelyn", "build", "astelyn-host-window")
}

// Run launches only the project's fixed host test executable; no arbitrary command execution.
func Run(root, driver string, seconds int) error {
	if seconds < 1 || seconds > 300 {
		return errors.New("seconds must be 1..300")
	}
	env, err := DriverEnv(driver, os.Environ())
	if err != nil {
		return err
	}
	path := BinaryPath(root)
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("window binary unavailable; build it first: %w", err)
	}
	if !info.Mode().IsRegular() {
		return errors.New("window binary is not a regular file")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(seconds+8)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, "--seconds", fmt.Sprint(seconds))
	cmd.Env, cmd.Stdin, cmd.Stdout, cmd.Stderr = env, os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("window process exceeded time limit: %w", ctx.Err())
		}
		return fmt.Errorf("window process failed: %w", err)
	}
	return nil
}

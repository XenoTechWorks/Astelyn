// Package graphics manages Astelyn's renderer preference and reports host API tools.
// A preference does not force Roblox or Darling to use that API.
package graphics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Backend string

const (
	Auto   Backend = "auto"
	Vulkan Backend = "vulkan"
	OpenGL Backend = "opengl"
)

func Valid(b Backend) bool { return b == Auto || b == Vulkan || b == OpenGL }

type Settings struct {
	Backend Backend `json:"backend"`
}

type Device struct {
	Backend   Backend `json:"backend"`
	Tool      string  `json:"tool"`
	Available bool    `json:"available"`
	Details   string  `json:"details"`
}

// ConfigPath follows XDG_CONFIG_HOME, with ~/.config as its fallback.
func ConfigPath() (string, error) {
	root := os.Getenv("XDG_CONFIG_HOME")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		root = filepath.Join(home, ".config")
	}
	if !filepath.IsAbs(root) {
		return "", errors.New("XDG_CONFIG_HOME must be an absolute path")
	}
	return filepath.Join(root, "astelyn", "graphics.json"), nil
}

func Load(path string) (Settings, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Settings{Backend: Auto}, nil
	}
	if err != nil {
		return Settings{}, err
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return Settings{}, fmt.Errorf("invalid graphics settings: %w", err)
	}
	if !Valid(s.Backend) {
		return Settings{}, fmt.Errorf("invalid saved graphics backend %q", s.Backend)
	}
	return s, nil
}

func Save(path string, b Backend) error {
	if !Valid(b) {
		return fmt.Errorf("unsupported backend %q; use auto, vulkan, or opengl", b)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".graphics-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err := f.Chmod(0600); err != nil {
		f.Close()
		return err
	}
	if err := json.NewEncoder(f).Encode(Settings{Backend: b}); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

// Probe tests only that the relevant diagnostic can query the active display/driver.
// This is not a test of Darwin/Roblox renderer support.
func Probe(b Backend) Device {
	var name string
	var args []string
	switch b {
	case Vulkan:
		name = "vulkaninfo"
		args = []string{"--summary"}
	case OpenGL:
		name = "glxinfo"
		args = []string{"-B"}
	default:
		return Device{Backend: b, Details: "not a concrete renderer"}
	}
	if _, err := exec.LookPath(name); err != nil {
		return Device{Backend: b, Tool: name, Details: "diagnostic tool not installed"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	data, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		reason := "diagnostic failed"
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			reason = "diagnostic timed out"
		} else if len(data) > 0 {
			reason = strings.TrimSpace(string(data))
			if len(reason) > 200 {
				reason = reason[:200]
			}
		}
		return Device{Backend: b, Tool: name, Details: reason}
	}
	details := "diagnostic succeeded"
	if b == OpenGL {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "OpenGL renderer string:") {
				details = strings.TrimSpace(strings.TrimPrefix(line, "OpenGL renderer string:"))
				break
			}
		}
	} else {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, "deviceName") {
				details = strings.TrimSpace(line)
				break
			}
		}
	}
	return Device{Backend: b, Tool: name, Available: true, Details: details}
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/XenoTechWorks/Astelyn/internal/gfxconfig"
	"os"
	"os/exec"
	"strings"
)

// runNativeTest launches an explicitly supplied *native Linux* command.
// It intentionally does not wrap Darling, macOS apps or Roblox.
func runNativeTest(args []string) int {
	if len(args) < 1 {
		return nativeUsage()
	}
	if args[0] == "--" {
		args = args[1:]
	}
	if len(args) < 1 || strings.HasPrefix(args[0], "-") {
		return nativeUsage()
	}
	path, err := gfxconfig.Path()
	if err != nil {
		return graphicsError(err)
	}
	conf, err := gfxconfig.Load(path)
	if err != nil {
		return graphicsError(err)
	}
	name, err := exec.LookPath(args[0])
	if err != nil {
		return graphicsError(err)
	}
	cmd := exec.Command(name, args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if conf.MangoHud {
		if _, err := exec.LookPath("mangohud"); err != nil {
			return graphicsError(fmt.Errorf("MangoHud is enabled but not installed: %w", err))
		}
		// MANGOHUD=1 activates the Vulkan overlay on supported Linux applications.
		// The OpenGL alternative can be tested with `mangohud glxgears` directly.
		cmd.Env = append(cmd.Env, "MANGOHUD=1")
	}
	err = cmd.Run()
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return graphicsError(err)
}
func nativeUsage() int {
	fmt.Fprintln(os.Stderr, "Usage: astelyn graphics native-run -- <linux-executable> [arguments...]")
	return 2
}
func runGPUs() int {
	return graphicsErrorIf(json.NewEncoder(os.Stdout).Encode(gfxconfig.DiscoverGPUs()))
}
func graphicsErrorIf(err error) int {
	if err != nil {
		return graphicsError(err)
	}
	return 0
}

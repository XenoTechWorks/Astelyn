package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/XenoTechWorks/Astelyn/internal/windowbridge"
)

// runWindowBridge demonstrates an explicit, host-only IPC control surface.
// It does not load Objective-C classes or implement a macOS window framework.
func runWindowBridge(args []string) int {
	if len(args) == 1 && args[0] == "build" {
		cmd := exec.Command("sh", "experiments/window-bridge/build.sh")
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Astelyn:", err)
			return 1
		}
		return 0
	}
	flags := flag.NewFlagSet("window-bridge", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	execute := flags.Bool("execute", false, "launch host SDL2 window")
	driver := flags.String("driver", "auto", "SDL video driver: auto, wayland or x11")
	jsonOut := flags.Bool("json", false, "emit protocol events as JSON lines")
	if len(args) == 0 || args[0] != "demo" {
		fmt.Fprintln(os.Stderr, "Usage: astelyn window-bridge build | demo [--driver auto|wayland|x11] [--json] [--execute]")
		return 2
	}
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if flags.NArg() != 0 || (*driver != "auto" && *driver != "wayland" && *driver != "x11") {
		fmt.Fprintln(os.Stderr, "invalid window-bridge options")
		return 2
	}
	if !*execute {
		fmt.Println("Dry run: native Linux SDL2 IPC demo. Add --execute (not Cocoa/Roblox).")
		return 0
	}
	if err := launchBridge(*driver, *jsonOut, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "Astelyn bridge:", err)
		return 1
	}
	return 0
}

func launchBridge(driver string, jsonOut bool, output io.Writer) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	path := filepath.Join(root, ".astelyn", "build", "astelyn-window-bridge")
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("build window bridge first: %w", err)
	}
	if !info.Mode().IsRegular() {
		return errors.New("bridge binary must be a regular file")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path)
	cmd.Env = os.Environ()
	if driver != "auto" {
		env := make([]string, 0, len(cmd.Env)+1)
		for _, e := range cmd.Env {
			if !strings.HasPrefix(e, "SDL_VIDEODRIVER=") {
				env = append(env, e)
			}
		}
		cmd.Env = append(env, "SDL_VIDEODRIVER="+driver)
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 512), 8192)
	ready, pong, closed := false, false, false
	for scanner.Scan() {
		e, parseErr := windowbridge.ParseEvent(scanner.Text())
		if parseErr != nil {
			break
		}
		if jsonOut {
			if err := json.NewEncoder(output).Encode(e); err != nil {
				break
			}
		} else {
			fmt.Fprintf(output, "HOST_EVENT %s %s\n", e.Name, e.Value)
		}
		if e.Name == "READY" && !ready {
			ready = true
			for _, c := range []windowbridge.Command{{Op: "TITLE", Title: "Astelyn v0.2.8 - Host Bridge"}, {Op: "SIZE", Width: 800, Height: 480}, {Op: "PING"}} {
				message, err := windowbridge.Encode(c)
				if err != nil {
					break
				}
				if _, err = io.WriteString(stdin, message); err != nil {
					break
				}
			}
		}
		if e.Name == "PONG" && !pong {
			pong = true
			// Give the user a chance to see the native window before closing.
			time.Sleep(2 * time.Second)
			message, _ := windowbridge.Encode(windowbridge.Command{Op: "QUIT"})
			io.WriteString(stdin, message)
		}
		if e.Name == "CLOSED" {
			closed = true
		}
	}
	stdin.Close()
	scanErr := scanner.Err()
	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		return fmt.Errorf("native bridge timed out: %w", ctx.Err())
	}
	if scanErr != nil {
		return scanErr
	}
	if waitErr != nil {
		return waitErr
	}
	if !ready || !pong || !closed {
		return fmt.Errorf("incomplete native bridge handshake (ready=%t pong=%t closed=%t)", ready, pong, closed)
	}
	return nil
}

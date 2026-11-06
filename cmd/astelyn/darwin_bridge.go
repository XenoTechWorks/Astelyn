package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/XenoTechWorks/Astelyn/internal/darwinbridge"
	"github.com/XenoTechWorks/Astelyn/internal/windowbridge"
)

// runDarwinBridge is an opt-in host-runner: Darling produces a bounded request,
// then Astelyn translates it into its existing native Linux SDL2 bridge.
func runDarwinBridge(args []string) int {
	if len(args) == 1 && args[0] == "build" {
		for _, script := range []string{"experiments/window-bridge/build.sh", "experiments/darwin-bridge/build.sh"} {
			cmd := exec.Command("sh", script)
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintln(os.Stderr, "Astelyn:", err)
				return 1
			}
		}
		return 0
	}
	if len(args) == 0 || args[0] != "demo" {
		fmt.Fprintln(os.Stderr, "Usage: astelyn darwin-bridge build | demo [--driver auto|wayland|x11] [--execute]")
		return 2
	}
	flags := flag.NewFlagSet("darwin-bridge demo", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	execute := flags.Bool("execute", false, "launch Darwin guest and native host window")
	driver := flags.String("driver", "auto", "auto, wayland or x11")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if flags.NArg() != 0 || (*driver != "auto" && *driver != "wayland" && *driver != "x11") {
		fmt.Fprintln(os.Stderr, "invalid options")
		return 2
	}
	if !*execute {
		fmt.Println("Dry run: Darwin guest request -> verified Go protocol -> Linux SDL2 window. Add --execute.")
		return 0
	}
	if err := runDarwinWindowDemo(*driver, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "Astelyn Darwin bridge:", err)
		return 1
	}
	return 0
}

func guestWindowRequest(ctx context.Context, root string) (darwinbridge.Request, error) {
	guestHost := filepath.Join(root, ".astelyn", "build", "astelyn-darwin-window-request")
	if err := verifyMachO(guestHost); err != nil {
		return darwinbridge.Request{}, err
	}
	launcher, err := exec.LookPath("darling")
	if err != nil {
		return darwinbridge.Request{}, err
	}
	// This path was previously verified on the user's Void/Darling installation.
	guestPath := "/Volumes/SystemRoot" + filepath.ToSlash(guestHost)
	cmd := exec.CommandContext(ctx, launcher, "shell", guestPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return darwinbridge.Request{}, fmt.Errorf("Darling guest request failed: %w; stderr=%s", err, strings.TrimSpace(stderr.String()))
	}
	request, err := darwinbridge.Parse(bytes.NewReader(output))
	if err != nil {
		return darwinbridge.Request{}, err
	}
	return request, nil
}

func verifyMachO(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("build v0.2.9 first: %w", err)
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return err
	}
	if !st.Mode().IsRegular() {
		return errors.New("guest is not a regular file")
	}
	var magic [4]byte
	if _, err := io.ReadFull(f, magic[:]); err != nil {
		return err
	}
	if magic != [4]byte{0xcf, 0xfa, 0xed, 0xfe} {
		return errors.New("guest executable is not an x86-64 Mach-O")
	}
	return nil
}

func runDarwinWindowDemo(driver string, out io.Writer) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	req, err := guestWindowRequest(ctx, root)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Darwin guest requested window: %q %dx%d\n", req.Title, req.Width, req.Height)
	commands, err := req.Commands()
	if err != nil {
		return err
	}
	host := filepath.Join(root, ".astelyn", "build", "astelyn-window-bridge")
	st, err := os.Stat(host)
	if err != nil {
		return fmt.Errorf("build v0.2.8 host bridge first: %w", err)
	}
	if !st.Mode().IsRegular() {
		return errors.New("host bridge is not a regular executable")
	}
	cmd := exec.CommandContext(ctx, host)
	cmd.Env = os.Environ()
	if driver != "auto" {
		filtered := make([]string, 0, len(cmd.Env)+1)
		for _, e := range cmd.Env {
			if !strings.HasPrefix(e, "SDL_VIDEODRIVER=") {
				filtered = append(filtered, e)
			}
		}
		cmd.Env = append(filtered, "SDL_VIDEODRIVER="+driver)
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
	var protocolErr error
	for scanner.Scan() {
		event, err := windowbridge.ParseEvent(scanner.Text())
		if err != nil {
			protocolErr = err
			break
		}
		fmt.Fprintf(out, "HOST_EVENT %s %s\n", event.Name, event.Value)
		if event.Name == "READY" && !ready {
			ready = true
			for _, line := range commands {
				if _, err = io.WriteString(stdin, line); err != nil {
					protocolErr = err
					break
				}
			}
			if protocolErr != nil {
				break
			}
		}
		if event.Name == "PONG" && !pong {
			pong = true
			// Show the result briefly. Never block longer than context timeout.
			time.Sleep(time.Second)
			if _, err = io.WriteString(stdin, "QUIT\n"); err != nil {
				protocolErr = err
				break
			}
		}
		if event.Name == "CLOSED" {
			closed = true
		}
	}
	stdin.Close()
	scanErr := scanner.Err()
	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		return fmt.Errorf("Darwin window bridge timed out: %w", ctx.Err())
	}
	if protocolErr != nil {
		return protocolErr
	}
	if scanErr != nil {
		return scanErr
	}
	if waitErr != nil {
		return waitErr
	}
	if !ready || !pong || !closed {
		return fmt.Errorf("incomplete host handshake: ready=%t pong=%t closed=%t", ready, pong, closed)
	}
	fmt.Fprintln(out, "ASTELYN_DARWIN_HOST_BRIDGE_OK")
	return nil
}

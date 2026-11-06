// Package robloxtest provides opt-in, bounded compatibility launch diagnostics.
// It does not modify, inject into, or circumvent protections in Roblox binaries.
package robloxtest

import (
	"bytes"
	"context"
	"debug/macho"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/XenoTechWorks/Astelyn/internal/inspection"
)

const marker = "RobloxPlayer.app/Contents/MacOS/RobloxPlayer"

// Plan is an inventory of a ZIP, not a runtime compatibility guarantee.
type Plan struct {
	Archive       string   `json:"archive"`
	Executable    string   `json:"executable,omitempty"`
	Architectures []string `json:"architectures,omitempty"`
	DirectImports int      `json:"direct_imports"`
	Found         bool     `json:"found"`
	Note          string   `json:"note"`
}

func Assess(zipPath string) (Plan, error) {
	report, err := inspection.InspectZip(zipPath)
	if err != nil {
		return Plan{}, err
	}
	p := Plan{Archive: zipPath, Note: "Static ZIP inspection only; missing macOS GUI frameworks may prevent launch. Does not extract or execute Roblox."}
	for _, b := range report.Binaries {
		if strings.HasSuffix(b.Path, "/"+marker) || b.Path == marker {
			p.Found = true
			p.Executable = b.Path
			p.Architectures = b.Architectures
			p.DirectImports = len(b.Libraries)
			return p, nil
		}
	}
	return p, nil
}

type LaunchReport struct {
	HostExecutable  string `json:"host_executable"`
	GuestExecutable string `json:"guest_executable"`
	ExitCode        int    `json:"exit_code"`
	TimedOut        bool   `json:"timed_out"`
	Stdout          string `json:"stdout,omitempty"`
	Stderr          string `json:"stderr,omitempty"`
	Error           string `json:"error,omitempty"`
	Note            string `json:"note"`
}

func GuestPath(host string) (string, error) {
	abs, err := filepath.Abs(host)
	if err != nil {
		return "", err
	}
	abs, err = filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(abs) || strings.ContainsAny(abs, "\x00\n\r") {
		return "", errors.New("invalid executable path")
	}
	if !strings.HasSuffix(filepath.ToSlash(abs), "/RobloxPlayer.app/Contents/MacOS/RobloxPlayer") {
		return "", errors.New("executable must be RobloxPlayer.app/Contents/MacOS/RobloxPlayer")
	}
	stat, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !stat.Mode().IsRegular() {
		return "", errors.New("executable is not a regular file")
	}
	f, err := macho.Open(abs)
	if err != nil {
		return "", fmt.Errorf("expected Mach-O executable: %w", err)
	}
	f.Close()
	return "/Volumes/SystemRoot" + filepath.ToSlash(abs), nil
}

// CappedWriter retains at most cap bytes while avoiding unbounded buffering.
type cappedWriter struct {
	buf bytes.Buffer
	cap int
}

func (w *cappedWriter) Write(p []byte) (int, error) {
	n := len(p)
	if w.buf.Len() < w.cap {
		take := w.cap - w.buf.Len()
		if take > n {
			take = n
		}
		_, _ = w.buf.Write(p[:take])
	}
	return n, nil
}

// Launch runs only when explicitly invoked by the caller after validation.
// The supplied timeout should be short: this is diagnostic, not a game launcher.
func Launch(ctx context.Context, launcher, host string, timeout time.Duration) (LaunchReport, error) {
	guest, err := GuestPath(host)
	if err != nil {
		return LaunchReport{}, err
	}
	if timeout <= 0 || timeout > 30*time.Second {
		return LaunchReport{}, errors.New("timeout must be > 0 and <= 30 seconds")
	}
	if _, err := exec.LookPath(launcher); err != nil {
		return LaunchReport{}, err
	}
	child, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(child, launcher, "shell", guest)
	cmd.Stdin = bytes.NewReader(nil)
	out, errout := &cappedWriter{cap: 64 * 1024}, &cappedWriter{cap: 64 * 1024}
	cmd.Stdout, cmd.Stderr = out, errout
	cmd.WaitDelay = 500 * time.Millisecond
	runErr := cmd.Run()
	r := LaunchReport{HostExecutable: host, GuestExecutable: guest, ExitCode: -1, Stdout: out.buf.String(), Stderr: errout.buf.String(), Note: "Experimental Darling launch. Exit code, logs and timeout do not establish graphical compatibility; no anti-cheat bypass or Roblox modifications."}
	if runErr == nil {
		r.ExitCode = 0
	} else {
		var e *exec.ExitError
		if errors.As(runErr, &e) {
			r.ExitCode = e.ExitCode()
		} else {
			r.Error = runErr.Error()
		}
	}
	if errors.Is(child.Err(), context.DeadlineExceeded) {
		r.TimedOut = true
		r.Error = "diagnostic timeout reached"
	}
	return r, nil
}

// Encode formats the report with standard Go JSON encoding.
func Encode(w io.Writer, r interface{}) error { return json.NewEncoder(w).Encode(r) }

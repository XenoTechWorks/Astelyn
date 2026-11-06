// Package runtimecompare launches a verified Mach-O through an explicitly chosen
// Darling executable and reports only the resulting loader diagnostics.
package runtimecompare

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/XenoTechWorks/Astelyn/internal/dyldreport"
	"github.com/XenoTechWorks/Astelyn/internal/robloxtest"
)

const maxLog = 64 * 1024

type Request struct {
	Launcher string        `json:"launcher"`
	Root     string        `json:"root"`
	Prefix   string        `json:"prefix"`
	App      string        `json:"app"`
	Timeout  time.Duration `json:"-"`
}

type Comparison struct {
	SameFirstBlocker        bool     `json:"same_first_blocker"`
	BaselineLibraries       []string `json:"baseline_libraries"`
	CandidateLibraries      []string `json:"candidate_libraries"`
	BaselineClassification  string   `json:"baseline_classification"`
	CandidateClassification string   `json:"candidate_classification"`
	Note                    string   `json:"note"`
}

type Report struct {
	Launcher        string                   `json:"launcher"`
	Root            string                   `json:"root"`
	Prefix          string                   `json:"prefix"`
	GuestExecutable string                   `json:"guest_executable"`
	Command         []string                 `json:"command"`
	Executed        bool                     `json:"executed"`
	Launch          *robloxtest.LaunchReport `json:"launch,omitempty"`
	Diagnosis       *dyldreport.Report       `json:"diagnosis,omitempty"`
	Comparison      *Comparison              `json:"comparison,omitempty"`
	Notes           []string                 `json:"notes"`
}

func Validate(r Request, execute bool) (Report, error) {
	if r.Timeout <= 0 || r.Timeout > 30*time.Second {
		return Report{}, errors.New("timeout must be between 1ns and 30s")
	}
	if r.Launcher == "" {
		return Report{}, errors.New("--launcher is required")
	}
	exe, err := exec.LookPath(r.Launcher)
	if err != nil {
		return Report{}, fmt.Errorf("launcher: %w", err)
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return Report{}, err
	}
	info, err := os.Stat(exe)
	if err != nil {
		return Report{}, err
	}
	if !info.Mode().IsRegular() {
		return Report{}, errors.New("launcher must be a regular executable file")
	}
	if !filepath.IsAbs(r.Root) {
		return Report{}, errors.New("--root must be an absolute Darling runtime root")
	}
	info, err = os.Stat(r.Root)
	if err != nil {
		return Report{}, fmt.Errorf("runtime root: %w", err)
	}
	if !info.IsDir() {
		return Report{}, errors.New("runtime root must be a directory")
	}
	if execute && r.Prefix == "" {
		return Report{}, errors.New("--prefix is required when executing to keep runtime state explicit")
	}
	if r.Prefix != "" {
		if !filepath.IsAbs(r.Prefix) || strings.ContainsAny(r.Prefix, "\x00\r\n") || filepath.Clean(r.Prefix) != r.Prefix || r.Prefix == "/" {
			return Report{}, errors.New("--prefix must be a clean absolute directory path")
		}
		if home, err := os.UserHomeDir(); err == nil && filepath.Clean(r.Prefix) == filepath.Join(home, ".darling") {
			return Report{}, errors.New("refusing regular Darling default prefix; select an isolated test prefix")
		}
	}
	app := r.App
	if !strings.HasSuffix(app, "/Contents/MacOS/RobloxPlayer") {
		app = filepath.Join(app, "Contents", "MacOS", "RobloxPlayer")
	}
	guest, err := robloxtest.GuestPath(app)
	if err != nil {
		return Report{}, err
	}
	return Report{Launcher: exe, Root: r.Root, Prefix: r.Prefix, GuestExecutable: guest, Command: []string{exe, "shell", guest}, Notes: []string{"This selects a Darling CLI binary; --root is for diagnostic lookup only, not an automatic sysroot override.", "Do not run regular Darling and CiderDarling concurrently against the same prefix.", "Launch results do not establish gameplay, graphics compatibility, or anti-cheat support."}}, nil
}

type limitedWriter struct{ b bytes.Buffer }

func (w *limitedWriter) Write(p []byte) (int, error) {
	n := len(p)
	if w.b.Len() < maxLog {
		k := maxLog - w.b.Len()
		if k > len(p) {
			k = len(p)
		}
		_, _ = w.b.Write(p[:k])
	}
	return n, nil
}

func Run(ctx context.Context, r Request, baseline *dyldreport.Launch) (Report, error) {
	result, err := Validate(r, true)
	if err != nil {
		return Report{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, result.Launcher, "shell", result.GuestExecutable)
	env := []string{}
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "DPREFIX=") {
			env = append(env, e)
		}
	}
	cmd.Env = append(env, "DPREFIX="+r.Prefix)
	cmd.Stdin = bytes.NewReader(nil)
	stdout, stderr := &limitedWriter{}, &limitedWriter{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.WaitDelay = 500 * time.Millisecond
	runErr := cmd.Run()
	launch := robloxtest.LaunchReport{HostExecutable: r.App, GuestExecutable: result.GuestExecutable, ExitCode: -1, Stdout: stdout.b.String(), Stderr: stderr.b.String(), Note: "Explicit isolated Darling runtime launch; no Roblox files modified."}
	if runErr == nil {
		launch.ExitCode = 0
	} else {
		var x *exec.ExitError
		if errors.As(runErr, &x) {
			launch.ExitCode = x.ExitCode()
		} else {
			launch.Error = runErr.Error()
		}
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		launch.TimedOut = true
		launch.Error = "diagnostic timeout reached"
	}
	result.Executed = true
	result.Launch = &launch
	observed := dyldreport.Launch{HostExecutable: launch.HostExecutable, GuestExecutable: launch.GuestExecutable, ExitCode: launch.ExitCode, TimedOut: launch.TimedOut, Stdout: launch.Stdout, Stderr: launch.Stderr, Error: launch.Error, Note: launch.Note}
	diag, err := dyldreport.Analyze(observed, r.Root)
	if err != nil {
		return result, fmt.Errorf("diagnose candidate: %w", err)
	}
	result.Diagnosis = &diag
	if baseline != nil {
		bd, err := dyldreport.Analyze(*baseline, r.Root)
		if err != nil {
			return result, fmt.Errorf("diagnose baseline: %w", err)
		}
		comparison := Compare(bd, diag)
		result.Comparison = &comparison
	}
	return result, nil
}

func Compare(baseline, candidate dyldreport.Report) Comparison {
	c := Comparison{BaselineClassification: baseline.Classification, CandidateClassification: candidate.Classification, BaselineLibraries: []string{}, CandidateLibraries: []string{}, Note: "Only first-observed loader errors can be compared. Different paths and loader roots may produce distinct failures; no compatibility gain is implied."}
	for _, v := range baseline.Dependencies {
		c.BaselineLibraries = append(c.BaselineLibraries, v.InstallName)
	}
	for _, v := range candidate.Dependencies {
		c.CandidateLibraries = append(c.CandidateLibraries, v.InstallName)
	}
	c.SameFirstBlocker = len(c.BaselineLibraries) > 0 && len(c.CandidateLibraries) > 0 && c.BaselineLibraries[0] == c.CandidateLibraries[0]
	return c
}

func DecodeLaunch(r io.Reader) (dyldreport.Launch, error) { return dyldreport.Decode(r) }

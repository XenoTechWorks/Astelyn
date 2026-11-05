package darling

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"
	"unicode/utf8"
)

type ProbeResult struct {
	GuestPath    string   `json:"guest_path"`
	Arguments    []string `json:"arguments"`
	Stdout       string   `json:"stdout,omitempty"`
	Stderr       string   `json:"stderr,omitempty"`
	StdoutBase64 string   `json:"stdout_base64,omitempty"`
	StderrBase64 string   `json:"stderr_base64,omitempty"`
	ExitCode     int      `json:"exit_code"`
	TimedOut     bool     `json:"timed_out"`
	Error        string   `json:"error,omitempty"`
	StdoutBytes  []byte   `json:"-"`
	StderrBytes  []byte   `json:"-"`
}

// Probe uses empty stdin by default.
func Probe(launcher, guest string, timeout time.Duration, args ...string) ProbeResult {
	return ProbeWithInput(launcher, guest, timeout, bytes.NewReader(nil), args...)
}

// ProbeWithInput captures binary output without data loss.
// The caller must validate guest paths before execution.
func ProbeWithInput(launcher, guest string, timeout time.Duration, input io.Reader, args ...string) ProbeResult {
	result := ProbeResult{
		GuestPath: guest,
		Arguments: append([]string{}, args...),
		ExitCode:  -1,
	}

	if timeout <= 0 {
		result.Error = "timeout must be positive"
		return result
	}
	if input == nil {
		input = bytes.NewReader(nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	command := append([]string{"shell", guest}, args...)
	cmd := exec.CommandContext(ctx, launcher, command...)
	cmd.Stdin = input
	cmd.WaitDelay = 250 * time.Millisecond

	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	err := cmd.Run()

	result.StdoutBytes = out.Bytes()
	result.StderrBytes = errOut.Bytes()

	if utf8.Valid(result.StdoutBytes) {
		result.Stdout = string(result.StdoutBytes)
	} else {
		result.StdoutBase64 = base64.StdEncoding.EncodeToString(result.StdoutBytes)
	}

	if utf8.Valid(result.StderrBytes) {
		result.Stderr = string(result.StderrBytes)
	} else {
		result.StderrBase64 = base64.StdEncoding.EncodeToString(result.StderrBytes)
	}

	if err == nil {
		result.ExitCode = 0
		return result
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		result.TimedOut = true
		result.Error = fmt.Sprintf("execution timed out after %s", timeout)
		return result
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
		return result
	}

	result.Error = err.Error()
	return result
}

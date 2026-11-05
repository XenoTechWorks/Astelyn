// Package darling provides a conservative experimental bridge to an installed Darling CLI.
// It does not provide Darwin compatibility itself and never downloads or patches binaries.
package darling

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Find locates an existing Darling launcher on PATH.
func Find() (string, error) { return exec.LookPath("darling") }

// ValidateMachO rejects app directories, ELF binaries, and arbitrary scripts.
func ValidateMachO(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("not a regular executable file: %s", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	var magic [4]byte
	if _, err := io.ReadFull(file, magic[:]); err != nil {
		return err
	}
	switch magic {
	case [4]byte{0xfe, 0xed, 0xfa, 0xce}, [4]byte{0xce, 0xfa, 0xed, 0xfe},
		[4]byte{0xfe, 0xed, 0xfa, 0xcf}, [4]byte{0xcf, 0xfa, 0xed, 0xfe},
		[4]byte{0xca, 0xfe, 0xba, 0xbe}, [4]byte{0xbe, 0xba, 0xfe, 0xca},
		[4]byte{0xca, 0xfe, 0xba, 0xbf}, [4]byte{0xbf, 0xba, 0xfe, 0xca}:
		return nil
	default:
		return fmt.Errorf("not a Mach-O executable: %s", path)
	}
}

// SystemRoot is the default installed Darling guest root on Linux.
const SystemRoot = "/usr/local/libexec/darling"

// ResolveGuest maps a Mach-O binary in Darling's installed root to a guest path.
// Host files outside that root cannot be assumed visible inside Darling.
func ResolveGuest(hostPath, systemRoot string) (string, error) {
	absolute, err := filepath.Abs(hostPath)
	if err != nil {
		return "", err
	}
	if err := ValidateMachO(absolute); err != nil {
		return "", err
	}
	root, err := filepath.Abs(systemRoot)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, absolute)
	if err != nil {
		return "", err
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("host file %q is outside Darling root %q; copy it into a guest-visible directory or use --guest", absolute, root)
	}
	return "/" + filepath.ToSlash(rel), nil
}

// ValidateGuest checks an explicit Darling guest path against its host-side root.
// Existence verifies the path, not runtime compatibility.
func ValidateGuest(path, systemRoot string) error {
	if !filepath.IsAbs(path) || strings.ContainsAny(path, "\n\r\x00") {
		return fmt.Errorf("guest path must be absolute and contain no control characters: %q", path)
	}
	clean := filepath.Clean(path)
	file := filepath.Join(systemRoot, strings.TrimPrefix(clean, "/"))
	if err := ValidateMachO(file); err != nil {
		return fmt.Errorf("guest %q (%s): %w", path, file, err)
	}
	return nil
}

// Execute runs a verified Darling guest command without invoking a host shell.
func Execute(darlingPath, guestPath string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := append([]string{"shell", guestPath}, args...)
	cmd := exec.CommandContext(ctx, darlingPath, command...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("Darling test timed out after 20s: %w", err)
		}
		return fmt.Errorf("Darling failed: %w", err)
	}
	return nil
}

const SmokeMarker = "ASTELYN_DARLING_OK"

// SmokeTest checks that Darling can execute a guest Bash builtin and return output.
func SmokeTest(darlingPath, systemRoot string) error {
	if err := ValidateGuest("/bin/bash", systemRoot); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, darlingPath, "shell", "/bin/bash", "--noprofile", "--norc", "-c", "printf '%s\n' ASTELYN_DARLING_OK")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("Darling smoke test timed out: %w", err)
		}
		return fmt.Errorf("Darling smoke test failed: %w; output: %s", err, strings.TrimSpace(string(output)))
	}
	if !strings.Contains(string(output), SmokeMarker) {
		return fmt.Errorf("Darling exited successfully but did not print test marker; output: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// Run preserves the original API while mapping system-root host paths safely.
func Run(darlingPath, machoPath string) error {
	guest, err := ResolveGuest(machoPath, SystemRoot)
	if err != nil {
		return err
	}
	return Execute(darlingPath, guest)
}

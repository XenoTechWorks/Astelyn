package darling

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func fakeProbeLauncher(t *testing.T, script string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "darling")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script), 0700); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestProbeBinaryStdinRoundTrip(t *testing.T) {
	launcher := fakeProbeLauncher(t, `
[ "$1" = shell ] || exit 81
[ "$2" = /bin/bzip2 ] || exit 82
[ "$3" = -c ] || exit 83
cat
`)
	payload := []byte{'B', 'Z', 'h', '9', 0, 0xff, 0x80, '\n'}
	result := ProbeWithInput(launcher, "/bin/bzip2", time.Second, bytes.NewReader(payload), "-c")
	if result.Error != "" || result.ExitCode != 0 || result.TimedOut {
		t.Fatalf("execution failed: %+v", result)
	}
	if !bytes.Equal(result.StdoutBytes, payload) || result.Stdout != "" {
		t.Fatalf("binary data altered: %+v", result)
	}
	decoded, err := base64.StdEncoding.DecodeString(result.StdoutBase64)
	if err != nil || !bytes.Equal(decoded, payload) {
		t.Fatalf("base64 not lossless: %v", err)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var report map[string]any
	if err := json.Unmarshal(encoded, &report); err != nil {
		t.Fatal(err)
	}
	if report["stdout_base64"] != result.StdoutBase64 {
		t.Fatalf("JSON lost binary data: %s", encoded)
	}
}

func TestProbeTextAndGuestExitCode(t *testing.T) {
	launcher := fakeProbeLauncher(t, `
[ "$1" = shell ] || exit 81
printf 'guest output\n'
printf 'guest warning\n' >&2
exit 17
`)
	result := Probe(launcher, "/bin/zsh", time.Second)
	if result.ExitCode != 17 || result.Error != "" || result.TimedOut ||
		result.Stdout != "guest output\n" || result.Stderr != "guest warning\n" {
		t.Fatalf("wrong guest exit result: %+v", result)
	}
}

func TestProbeDefaultsToEmptyStdin(t *testing.T) {
	launcher := fakeProbeLauncher(t, `
[ "$1" = shell ] || exit 81
cat
`)
	result := Probe(launcher, "/bin/bash", time.Second)
	if result.Error != "" || result.ExitCode != 0 || len(result.StdoutBytes) != 0 {
		t.Fatalf("default stdin was not empty: %+v", result)
	}
}

func TestProbeRejectsInvalidTimeout(t *testing.T) {
	result := Probe("/not/used", "/bin/bash", 0)
	if result.ExitCode != -1 || result.Error == "" {
		t.Fatalf("expected timeout validation: %+v", result)
	}
}

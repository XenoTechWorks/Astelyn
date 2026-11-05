package darling

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProbe(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "darling")

	script := `#!/bin/sh
[ "$1" = shell ] || exit 88
[ "$2" = /bin/echo ] || exit 89
shift 2
printf '%s\n' "$1"
printf 'warning\n' >&2
exit 7
`
	if err := os.WriteFile(exe, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}

	r := Probe(exe, "/bin/echo", time.Second, "hello")

	if r.ExitCode != 7 ||
		r.Stdout != "hello\n" ||
		r.Stderr != "warning\n" ||
		r.Error != "" ||
		r.TimedOut {
		t.Fatalf("unexpected result: %+v", r)
	}
}

func TestProbeMissingLauncher(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "missing")
	r := Probe(exe, "/bin/echo", time.Second)

	if r.ExitCode != -1 || r.Error == "" {
		t.Fatalf("unexpected result: %+v", r)
	}
}

func TestProbeTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("process timeout test")
	}

	exe := filepath.Join(t.TempDir(), "darling")

	if err := os.WriteFile(exe,
		[]byte("#!/bin/sh\nsleep 2\n"), 0700); err != nil {
		t.Fatal(err)
	}

	r := Probe(exe, "/bin/echo", 20*time.Millisecond)

	if !r.TimedOut || !strings.Contains(r.Error, "timed out") {
		t.Fatalf("unexpected result: %+v", r)
	}
}

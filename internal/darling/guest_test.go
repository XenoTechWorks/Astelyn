package darling

import (
	"os"
	"path/filepath"
	"testing"
)

func makeMachO(t *testing.T, root string) string {
	t.Helper()
	path := filepath.Join(root, "bin", "bash")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte{0xcf, 0xfa, 0xed, 0xfe, 0, 0, 0, 0}, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
func TestGuestMapping(t *testing.T) {
	root := t.TempDir()
	host := makeMachO(t, root)
	got, err := ResolveGuest(host, root)
	if err != nil || got != "/bin/bash" {
		t.Fatalf("got %q, %v", got, err)
	}
	if err := ValidateGuest("/bin/bash", root); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"bin/bash", "/missing", "/bin/bash\n"} {
		if err := ValidateGuest(path, root); err == nil {
			t.Errorf("accepted guest path %q", path)
		}
	}
	other := makeMachO(t, t.TempDir())
	if _, err := ResolveGuest(other, root); err == nil {
		t.Fatal("accepted host file outside guest root")
	}
}
func TestSmokeTestFakeDarling(t *testing.T) {
	root := t.TempDir()
	makeMachO(t, root)
	fake := filepath.Join(t.TempDir(), "darling")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\nprintf 'ASTELYN_DARLING_OK\\n'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := SmokeTest(fake, root); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fake, []byte("#!/bin/sh\necho nope\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := SmokeTest(fake, root); err == nil {
		t.Fatal("missing marker should fail")
	}
}

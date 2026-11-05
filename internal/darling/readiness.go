package darling

import (
	"encoding/binary"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Readiness is a static, non-executing assessment of the Darling environment.
// In particular, ReadyForSmokeTest does not mean that Darling is functional.
type Readiness struct {
	HostOS            string   `json:"host_os"`
	HostArchitecture  string   `json:"host_architecture"`
	Launcher          string   `json:"launcher,omitempty"`
	LauncherFound     bool     `json:"launcher_found"`
	SystemRoot        string   `json:"system_root"`
	SystemRootFound   bool     `json:"system_root_found"`
	GuestBash         string   `json:"guest_bash"`
	GuestBashMachO    bool     `json:"guest_bash_macho"`
	ReadyForSmokeTest bool     `json:"ready_for_smoke_test"`
	Notes             []string `json:"notes"`
}

// CheckReadiness discovers paths and examines file headers. It never runs
// Darling, executes untrusted binaries, or writes to the host filesystem.
// launcher and systemRoot may be injected to assess custom installations.
func CheckReadiness(launcher, systemRoot string) Readiness {
	result := Readiness{
		HostOS:           runtime.GOOS,
		HostArchitecture: runtime.GOARCH,
		SystemRoot:       systemRoot,
		GuestBash:        "/bin/bash",
		Notes:            []string{},
	}
	if launcher == "" {
		var err error
		launcher, err = exec.LookPath("darling")
		if err != nil {
			result.Notes = append(result.Notes, "Darling launcher was not found on PATH")
		}
	}
	if launcher != "" {
		if file, err := os.Stat(launcher); err == nil && file.Mode().IsRegular() && file.Mode().Perm()&0111 != 0 {
			result.LauncherFound = true
			result.Launcher = launcher
		} else {
			result.Notes = append(result.Notes, "Darling launcher is missing or not executable")
		}
	}
	if info, err := os.Stat(systemRoot); err == nil && info.IsDir() {
		result.SystemRootFound = true
	} else {
		result.Notes = append(result.Notes, "Darling system root directory was not found")
	}
	if result.SystemRootFound {
		path := filepath.Join(systemRoot, "bin", "bash")
		if looksLikeMachO(path) {
			result.GuestBashMachO = true
		} else {
			result.Notes = append(result.Notes, "Guest /bin/bash is missing or is not recognized as Mach-O")
		}
	}
	if runtime.GOOS != "linux" {
		result.Notes = append(result.Notes, "Darling execution is intended for Linux hosts")
	}
	result.ReadyForSmokeTest = runtime.GOOS == "linux" && result.LauncherFound && result.SystemRootFound && result.GuestBashMachO
	result.Notes = append(result.Notes, "Static inspection only; use darling test to verify actual execution")
	return result
}

func looksLikeMachO(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 4 {
		return false
	}
	var magic [4]byte
	if _, err := io.ReadFull(file, magic[:]); err != nil {
		return false
	}
	v := binary.BigEndian.Uint32(magic[:])
	switch v {
	case 0xfeedface, 0xcefaedfe, 0xfeedfacf, 0xcffaedfe, 0xcafebabe, 0xbebafeca, 0xcafebabf, 0xbfbafeca:
		return true
	}
	return false
}

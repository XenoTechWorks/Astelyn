// Package runtimecheck inventories runtime components in an installed Darling root.
// It does not load libraries or assert that macOS frameworks are implemented.
package runtimecheck

import (
	"debug/elf"
	"debug/macho"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Component struct {
	Name         string `json:"name"`
	Kind         string `json:"kind"`
	GuestPath    string `json:"guest_path"`
	Present      bool   `json:"present"`
	HostPath     string `json:"host_path,omitempty"`
	Format       string `json:"format,omitempty"`
	Architecture string `json:"architecture,omitempty"`
	Note         string `json:"note,omitempty"`
}
type Report struct {
	DarlingRoot string      `json:"darling_root"`
	Components  []Component `json:"components"`
	Disclaimer  string      `json:"disclaimer"`
}

type target struct{ name, kind, path, note string }

var targets = []target{
	{"Objective-C runtime", "darwin-runtime", "/usr/lib/libobjc.A.dylib", "Required for Objective-C messaging; presence alone does not verify behavior"},
	{"Objective-C compatibility path", "darwin-runtime", "/usr/lib/libobjc.dylib", "May be an alias for libobjc.A.dylib"},
	{"Objective-C trampolines", "darwin-runtime", "/usr/lib/libobjc-trampolines.dylib", "Runtime support; not an AppKit implementation"},
	{"Grand Central Dispatch", "darwin-runtime", "/usr/lib/system/libdispatch.dylib", "Dispatch library; not Foundation"},
	{"Swift Foundation overlay", "swift-overlay", "/usr/lib/swift/libswiftFoundation.dylib", "Swift overlay is not Foundation.framework"},
	{"Swift Dispatch overlay", "swift-overlay", "/usr/lib/swift/libswiftDispatch.dylib", "Swift overlay is not a complete Dispatch API test"},
	{"Swift AppKit overlay", "swift-overlay", "/usr/lib/swift/libswiftAppKit.dylib", "Swift overlay is not AppKit.framework"},
	{"Swift Metal overlay", "swift-overlay", "/usr/lib/swift/libswiftMetal.dylib", "Swift overlay is not Metal.framework"},
	{"CoreFoundation framework", "framework", "/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation", "Presence does not prove API compatibility"},
	{"Foundation framework", "framework", "/System/Library/Frameworks/Foundation.framework/Versions/C/Foundation", "Version-specific check; see framework-resolve for alternate versions"},
	{"AppKit framework", "framework", "/System/Library/Frameworks/AppKit.framework/Versions/C/AppKit", "Version-specific check; see framework-resolve for alternate versions"},
	{"Metal framework", "framework", "/System/Library/Frameworks/Metal.framework/Versions/A/Metal", "Version-specific check; see framework-resolve for alternate versions"},
}

func Analyze(root string) (Report, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return Report{}, err
	}
	abs, err = filepath.EvalSymlinks(abs)
	if err != nil {
		return Report{}, fmt.Errorf("Darling root: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return Report{}, err
	}
	if !info.IsDir() {
		return Report{}, fmt.Errorf("Darling root is not a directory: %s", abs)
	}
	out := Report{DarlingRoot: abs, Components: make([]Component, 0, len(targets)), Disclaimer: "Static file inventory only. Swift overlays do not replace macOS frameworks. No Objective-C class loading, symbol resolution, GUI, or Roblox execution is tested."}
	for _, t := range targets {
		c := Component{Name: t.name, Kind: t.kind, GuestPath: t.path, Note: t.note}
		candidate := filepath.Join(abs, strings.TrimPrefix(t.path, "/"))
		resolved, e := filepath.EvalSymlinks(candidate)
		if e == nil {
			rel, e := filepath.Rel(abs, resolved)
			if e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
				if stat, e := os.Stat(resolved); e == nil && stat.Mode().IsRegular() {
					c.Present = true
					c.HostPath = resolved
					c.Format, c.Architecture = detect(resolved)
				}
			} else {
				c.Note += "; symlink resolves outside Darling root"
			}
		} else if !os.IsNotExist(e) {
			c.Note += "; cannot resolve file: " + e.Error()
		}
		out.Components = append(out.Components, c)
	}
	return out, nil
}

func detect(path string) (string, string) {
	f, err := os.Open(path)
	if err != nil {
		return "unreadable", ""
	}
	defer f.Close()
	var b [4]byte
	if _, err := f.Read(b[:]); err != nil {
		return "unknown", ""
	}
	if string(b[:]) == "\x7fELF" {
		if ef, e := elf.Open(path); e == nil {
			defer ef.Close()
			return "ELF", ef.Machine.String()
		}
		return "ELF", ""
	}
	n := binary.BigEndian.Uint32(b[:])
	le := binary.LittleEndian.Uint32(b[:])
	if n == 0xcafebabe || n == 0xcafebabf || le == 0xcafebabe || le == 0xcafebabf {
		return "Mach-O universal", "multiple"
	}
	if mf, e := macho.Open(path); e == nil {
		defer mf.Close()
		return "Mach-O", mf.Cpu.String()
	}
	return "unknown", ""
}

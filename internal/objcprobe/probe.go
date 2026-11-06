// Package objcprobe inspects the installed Darling Objective-C runtime without loading it.
package objcprobe

import (
	"debug/macho"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Library struct {
	GuestPath    string          `json:"guest_path"`
	Present      bool            `json:"present"`
	Format       string          `json:"format"`
	Architecture string          `json:"architecture,omitempty"`
	SymbolTable  string          `json:"symbol_table"`
	Symbols      map[string]bool `json:"symbols,omitempty"`
	Error        string          `json:"error,omitempty"`
}
type Report struct {
	DarlingRoot string    `json:"darling_root"`
	Libraries   []Library `json:"libraries"`
	Disclaimer  string    `json:"disclaimer"`
}

var targetSymbols = []string{"_objc_getClass", "_objc_getMetaClass", "_sel_registerName", "_objc_msgSend"}
var libraryPaths = []string{"/usr/lib/libobjc.A.dylib", "/usr/lib/libobjc.dylib", "/usr/lib/system/libdispatch.dylib"}

func insideRoot(root, guest string) (string, error) {
	if !filepath.IsAbs(root) || !filepath.IsAbs(guest) || strings.ContainsAny(guest, "\x00\n\r") || filepath.Clean(guest) != guest {
		return "", errors.New("root and guest must be absolute clean paths")
	}
	rootReal, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	candidate := filepath.Join(root, strings.TrimPrefix(guest, "/"))
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(rootReal, resolved)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errors.New("symlink escapes Darling root")
	}
	return resolved, nil
}
func inspect(root, guest string) Library {
	out := Library{GuestPath: guest, Format: "unknown", SymbolTable: "unknown"}
	path, err := insideRoot(root, guest)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			out.SymbolTable = "not-applicable"
			return out
		}
		out.Error = err.Error()
		return out
	}
	st, err := os.Stat(path)
	if err != nil {
		out.Error = err.Error()
		return out
	}
	if !st.Mode().IsRegular() {
		out.Error = "not a regular file"
		return out
	}
	out.Present = true
	f, err := os.Open(path)
	if err != nil {
		out.Error = err.Error()
		return out
	}
	defer f.Close()
	var magic [4]byte
	if _, err = f.Read(magic[:]); err != nil {
		out.Error = err.Error()
		return out
	}
	if binary.LittleEndian.Uint32(magic[:]) == 0x464c457f {
		out.Format = "ELF (not Mach-O)"
		return out
	}
	mf, err := macho.Open(path)
	if err != nil {
		out.Error = fmt.Sprintf("Mach-O parse failed: %v", err)
		return out
	}
	defer mf.Close()
	out.Format = "Mach-O"
	out.Architecture = mf.Cpu.String()
	if mf.Symtab == nil {
		out.SymbolTable = "absent or stripped"
		return out
	}
	out.SymbolTable = "present"
	out.Symbols = make(map[string]bool, len(targetSymbols))
	for _, s := range targetSymbols {
		out.Symbols[strings.TrimPrefix(s, "_")] = false
	}
	for _, s := range mf.Symtab.Syms {
		for _, t := range targetSymbols {
			if s.Name == t || s.Name == strings.TrimPrefix(t, "_") {
				out.Symbols[strings.TrimPrefix(t, "_")] = true
			}
		}
	}
	return out
}
func Analyze(root string) (Report, error) {
	if !filepath.IsAbs(root) {
		return Report{}, errors.New("Darling root must be absolute")
	}
	st, err := os.Stat(root)
	if err != nil {
		return Report{}, err
	}
	if !st.IsDir() {
		return Report{}, errors.New("Darling root is not a directory")
	}
	report := Report{DarlingRoot: root, Disclaimer: "Static library and symbol-table inspection only. Missing symbols may be stripped or exported by other mechanisms. Library presence does not prove Objective-C, Cocoa or Roblox functionality."}
	for _, p := range libraryPaths {
		report.Libraries = append(report.Libraries, inspect(root, p))
	}
	return report, nil
}

// Package inspection statically inspects macOS Mach-O binaries in ZIP archives.
// It never executes, patches, or extracts the application.
package inspection

import (
	"archive/zip"
	"bytes"
	"debug/macho"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

const maxBinarySize = 512 << 20 // Protect against unexpectedly large archive entries.

type Binary struct {
	Path          string   `json:"path"`
	Size          uint64   `json:"size_bytes"`
	Architectures []string `json:"architectures"`
	Libraries     []string `json:"libraries"`
}
type Report struct {
	Archive  string   `json:"archive"`
	Entries  int      `json:"entries"`
	Binaries []Binary `json:"binaries"`
}

func InspectZip(path string) (Report, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return Report{}, err
	}
	defer reader.Close()
	result := Report{Archive: path, Entries: len(reader.File), Binaries: make([]Binary, 0)}
	for _, file := range reader.File {
		if file.FileInfo().IsDir() || !strings.Contains(file.Name, "/Contents/MacOS/") || strings.HasSuffix(file.Name, ".dylib") || strings.HasSuffix(file.Name, ".so") {
			continue
		}
		if file.UncompressedSize64 > maxBinarySize {
			return Report{}, fmt.Errorf("archive entry too large: %s", file.Name)
		}
		r, err := file.Open()
		if err != nil {
			return Report{}, err
		}
		payload, readErr := io.ReadAll(io.LimitReader(r, maxBinarySize+1))
		closeErr := r.Close()
		if readErr != nil {
			return Report{}, readErr
		}
		if closeErr != nil {
			return Report{}, closeErr
		}
		if len(payload) > maxBinarySize {
			return Report{}, fmt.Errorf("archive entry too large: %s", file.Name)
		}
		binaries := parseMachO(payload)
		if len(binaries.Architectures) == 0 {
			continue
		}
		binaries.Path = filepath.ToSlash(file.Name)
		binaries.Size = file.UncompressedSize64
		result.Binaries = append(result.Binaries, binaries)
	}
	return result, nil
}

func parseMachO(data []byte) Binary {
	result := Binary{Architectures: make([]string, 0), Libraries: make([]string, 0)}
	if fat, err := macho.NewFatFile(bytes.NewReader(data)); err == nil {
		defer fat.Close()
		for _, arch := range fat.Arches {
			result.Architectures = appendUnique(result.Architectures, arch.Cpu.String())
			libs, _ := arch.ImportedLibraries()
			for _, lib := range libs {
				result.Libraries = appendUnique(result.Libraries, lib)
			}
		}
		return result
	} else if !errors.Is(err, macho.ErrNotFat) {
		return result
	}
	thin, err := macho.NewFile(bytes.NewReader(data))
	if err != nil {
		return result
	}
	defer thin.Close()
	result.Architectures = append(result.Architectures, thin.Cpu.String())
	result.Libraries, _ = thin.ImportedLibraries()
	return result
}
func appendUnique(items []string, value string) []string {
	for _, x := range items {
		if x == value {
			return items
		}
	}
	return append(items, value)
}

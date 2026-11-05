package darling

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateMachO(t *testing.T) {
	dir := t.TempDir()
	tests := []struct {
		name string
		data []byte
		good bool
	}{
		{"thin64", []byte{0xcf, 0xfa, 0xed, 0xfe, 0, 0, 0, 0}, true},
		{"fat", []byte{0xca, 0xfe, 0xba, 0xbe, 0, 0, 0, 0}, true},
		{"elf", []byte{0x7f, 'E', 'L', 'F'}, false},
		{"script", []byte("#!/bin/sh"), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := filepath.Join(dir, tc.name)
			if err := os.WriteFile(p, tc.data, 0600); err != nil {
				t.Fatal(err)
			}
			err := ValidateMachO(p)
			if (err == nil) != tc.good {
				t.Fatalf("ValidateMachO: %v", err)
			}
		})
	}
	if err := ValidateMachO(dir); err == nil {
		t.Fatal("directory should be rejected")
	}
	if err := ValidateMachO(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("missing file should be rejected")
	}
}

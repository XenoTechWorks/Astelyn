package inspection

import (
	"archive/zip"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestInspectThinMachOAndDependency(t *testing.T) {
	// Minimal 64-bit little-endian x86_64 Mach-O with one LC_LOAD_DYLIB.
	data := make([]byte, 32+48)
	binary.LittleEndian.PutUint32(data[0:4], 0xfeedfacf)
	binary.LittleEndian.PutUint32(data[4:8], 0x01000007)
	binary.LittleEndian.PutUint32(data[8:12], 3)
	binary.LittleEndian.PutUint32(data[12:16], 2)
	binary.LittleEndian.PutUint32(data[16:20], 1)
	binary.LittleEndian.PutUint32(data[20:24], 48)
	binary.LittleEndian.PutUint32(data[32:36], 0xc)
	binary.LittleEndian.PutUint32(data[36:40], 48)
	binary.LittleEndian.PutUint32(data[40:44], 24)
	copy(data[56:], []byte("/usr/lib/libA.dylib\x00"))
	dir := t.TempDir()
	path := filepath.Join(dir, "test.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	z := zip.NewWriter(f)
	w, err := z.Create("RobloxPlayer.app/Contents/MacOS/RobloxPlayer")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = w.Write(data); err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	report, err := InspectZip(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Binaries) != 1 {
		t.Fatalf("expected 1 binary, got %d", len(report.Binaries))
	}
	b := report.Binaries[0]
	if len(b.Architectures) != 1 || b.Architectures[0] != "CpuAmd64" {
		t.Fatalf("unexpected arch: %v", b.Architectures)
	}
	if len(b.Libraries) != 1 || b.Libraries[0] != "/usr/lib/libA.dylib" {
		t.Fatalf("unexpected libraries: %v", b.Libraries)
	}
}
func TestNonMachOIgnored(t *testing.T) {
	got := parseMachO([]byte("not a mach-o"))
	if len(got.Architectures) != 0 {
		t.Fatalf("expected none: %v", got.Architectures)
	}
}

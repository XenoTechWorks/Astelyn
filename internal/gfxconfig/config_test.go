package gfxconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultAndRoundtrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config", "performance.json")
	c, e := Load(p)
	if e != nil || c.MangoHud || c.GPU != "auto" {
		t.Fatalf("default %#v %v", c, e)
	}
	c.MangoHud = true
	if e := Save(p, c); e != nil {
		t.Fatal(e)
	}
	got, e := Load(p)
	if e != nil || got != c {
		t.Fatalf("roundtrip %#v %v", got, e)
	}
	info, e := os.Stat(p)
	if e != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("permissions %v %v", info, e)
	}
}
func TestRejectGPUOverride(t *testing.T) {
	if e := Save(filepath.Join(t.TempDir(), "x"), Config{GPU: "1"}); e == nil {
		t.Fatal("accepted GPU selection without implementation")
	}
}
func TestPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	p, e := Path()
	if e != nil || filepath.Base(p) != "performance.json" {
		t.Fatalf("%s %v", p, e)
	}
	t.Setenv("XDG_CONFIG_HOME", "relative")
	if _, e = Path(); e == nil {
		t.Fatal("relative path accepted")
	}
}
func TestCorruptFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "p.json")
	if e := os.WriteFile(p, []byte("{"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := Load(p); e == nil {
		t.Fatal("invalid JSON accepted")
	}
}

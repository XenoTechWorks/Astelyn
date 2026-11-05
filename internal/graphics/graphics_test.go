package graphics

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSettingsRoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config", "graphics.json")
	s, err := Load(p)
	if err != nil || s.Backend != Auto {
		t.Fatalf("default: %#v %v", s, err)
	}
	if err := Save(p, Vulkan); err != nil {
		t.Fatal(err)
	}
	s, err = Load(p)
	if err != nil || s.Backend != Vulkan {
		t.Fatalf("saved: %#v %v", s, err)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("permissions: %v", info.Mode())
	}
	if err := Save(p, "metal"); err == nil {
		t.Fatal("accepted invalid backend")
	}
}
func TestInvalidConfig(t *testing.T) {
	p := filepath.Join(t.TempDir(), "graphics.json")
	if err := os.WriteFile(p, []byte(`{"backend":"metal"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p); err == nil {
		t.Fatal("accepted invalid stored backend")
	}
}
func TestConfigPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	p, err := ConfigPath()
	if err != nil || filepath.Base(p) != "graphics.json" {
		t.Fatalf("%s %v", p, err)
	}
	t.Setenv("XDG_CONFIG_HOME", "relative")
	if _, err := ConfigPath(); err == nil {
		t.Fatal("accepted relative XDG path")
	}
}

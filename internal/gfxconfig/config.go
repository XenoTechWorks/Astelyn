// Package gfxconfig stores experimental host graphics settings.
// Settings do not change Darling, Metal or Roblox rendering.
package gfxconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	MangoHud bool   `json:"mangohud"`
	GPU      string `json:"gpu"`
}

func Default() Config { return Config{GPU: "auto"} }

func Path() (string, error) {
	root := os.Getenv("XDG_CONFIG_HOME")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		root = filepath.Join(home, ".config")
	}
	if !filepath.IsAbs(root) {
		return "", errors.New("XDG_CONFIG_HOME must be absolute")
	}
	return filepath.Join(root, "astelyn", "performance.json"), nil
}

func Validate(c Config) error {
	if c.GPU == "" || c.GPU == "auto" {
		return nil
	}
	return fmt.Errorf("GPU override %q not supported yet; use auto", c.GPU)
}

func Load(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Config{}, err
	}
	c := Default()
	if err = json.Unmarshal(b, &c); err != nil {
		return Config{}, fmt.Errorf("invalid performance config: %w", err)
	}
	if err = Validate(c); err != nil {
		return Config{}, err
	}
	return c, nil
}

func Save(path string, c Config) error {
	if err := Validate(c); err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".performance-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = f.Chmod(0600); err != nil {
		f.Close()
		return err
	}
	if err = json.NewEncoder(f).Encode(c); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

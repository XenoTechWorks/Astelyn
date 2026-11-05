package main

import (
	"encoding/json"
	"fmt"
	"github.com/XenoTechWorks/Astelyn/internal/gfxconfig"
	"os"
	"os/exec"
)

func runGraphicsExtra(args []string) int {
	if len(args) < 1 {
		return 2
	}
	switch args[0] {
	case "gpus":
		if len(args) == 1 {
			return runGPUs()
		}
	case "native-run":
		return runNativeTest(args[1:])
	case "overlay":
		if len(args) == 2 && args[1] == "status" {
			p, e := gfxconfig.Path()
			if e != nil {
				return graphicsError(e)
			}
			c, e := gfxconfig.Load(p)
			if e != nil {
				return graphicsError(e)
			}
			path, e := exec.LookPath("mangohud")
			info := struct {
				Enabled    bool   `json:"enabled"`
				Installed  bool   `json:"installed"`
				Executable string `json:"executable,omitempty"`
				Scope      string `json:"scope"`
			}{c.MangoHud, e == nil, path, "native-linux-only; not connected to Darling/Roblox"}
			if e = json.NewEncoder(os.Stdout).Encode(info); e != nil {
				return graphicsError(e)
			}
			return 0
		}
		if len(args) == 3 && args[1] == "set" && (args[2] == "on" || args[2] == "off") {
			p, e := gfxconfig.Path()
			if e != nil {
				return graphicsError(e)
			}
			c, e := gfxconfig.Load(p)
			if e != nil {
				return graphicsError(e)
			}
			c.MangoHud = args[2] == "on"
			if e = gfxconfig.Save(p, c); e != nil {
				return graphicsError(e)
			}
			fmt.Printf("MangoHud preference: %s (native Linux tests only)\n", args[2])
			return 0
		}
	case "performance":
		if len(args) == 2 && args[1] == "get" {
			p, e := gfxconfig.Path()
			if e != nil {
				return graphicsError(e)
			}
			c, e := gfxconfig.Load(p)
			if e != nil {
				return graphicsError(e)
			}
			if e = json.NewEncoder(os.Stdout).Encode(c); e != nil {
				return graphicsError(e)
			}
			return 0
		}
	}
	fmt.Fprintln(os.Stderr, "Usage: astelyn graphics overlay <status|set on|set off> | performance get")
	return 2
}
func graphicsError(e error) int { fmt.Fprintln(os.Stderr, "Astelyn:", e); return 1 }

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/XenoTechWorks/Astelyn/internal/hostwindow"
	"os"
	"os/exec"
	"path/filepath"
)

func runHostWindow(args []string) int {
	usage := func() {
		fmt.Fprintln(os.Stderr, "Usage: astelyn host-window <status|build|run> [--driver auto|wayland|x11] [--seconds 12] [--execute]")
	}
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "status":
		flags := flag.NewFlagSet("host-window status", flag.ContinueOnError)
		flags.SetOutput(os.Stderr)
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 {
			usage()
			return 2
		}
		if err := json.NewEncoder(os.Stdout).Encode(hostwindow.Detect()); err != nil {
			return 1
		}
		return 0
	case "build":
		if len(args) != 1 {
			usage()
			return 2
		}
		cmd := exec.Command("sh", "experiments/host-window/build.sh")
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Astelyn:", err)
			return 1
		}
		return 0
	case "run":
		flags := flag.NewFlagSet("host-window run", flag.ContinueOnError)
		flags.SetOutput(os.Stderr)
		driver := flags.String("driver", "auto", "SDL2 Linux video driver")
		seconds := flags.Int("seconds", 12, "maximum window lifetime (1..300)")
		execute := flags.Bool("execute", false, "show the native Linux test window")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 || *seconds < 1 || *seconds > 300 || (*driver != "auto" && *driver != "wayland" && *driver != "x11") {
			usage()
			return 2
		}
		if !*execute {
			fmt.Printf("Dry run: %s (--driver %s, %d seconds). Add --execute.\n", filepath.Join(".astelyn", "build", "astelyn-host-window"), *driver, *seconds)
			return 0
		}
		root, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Astelyn:", err)
			return 1
		}
		if err := hostwindow.Run(root, *driver, *seconds); err != nil {
			fmt.Fprintln(os.Stderr, "Astelyn:", err)
			return 1
		}
		return 0
	default:
		usage()
		return 2
	}
}

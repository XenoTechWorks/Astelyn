package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/XenoTechWorks/Astelyn/internal/darling"
	"github.com/XenoTechWorks/Astelyn/internal/objcprobe"
	"os"
	"strings"
	"time"
)

const objcSmokeMarker = "ASTELYN_OBJC_OK"

func runObjCProbe(args []string) int {
	if len(args) > 0 && args[0] == "run" {
		return runObjCSmoke(args[1:])
	}
	flags := flag.NewFlagSet("objc-probe", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	jsonOut := flags.Bool("json", false, "JSON report")
	root := flags.String("root", darling.SystemRoot, "Darling installation root")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "Usage: astelyn objc-probe [--json] [--root path] OR objc-probe run --guest /path [--execute]")
		return 2
	}
	report, err := objcprobe.Analyze(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Astelyn:", err)
		return 1
	}
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			return 1
		}
		return 0
	}
	fmt.Println("Darling Objective-C runtime libraries:")
	for _, l := range report.Libraries {
		status := "missing"
		if l.Present {
			status = l.Format
		}
		fmt.Printf("  %-35s %s", l.GuestPath, status)
		if l.Architecture != "" {
			fmt.Print(" (", l.Architecture, ")")
		}
		if l.Error != "" {
			fmt.Print(" [", l.Error, "]")
		}
		fmt.Println()
	}
	fmt.Println(report.Disclaimer)
	return 0
}

// runObjCSmoke only executes a caller-supplied *guest* Mach-O on explicit --execute.
// It never builds or downloads binaries, and it doesn't modify the Darling root.
func runObjCSmoke(args []string) int {
	flags := flag.NewFlagSet("objc-probe run", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	guest := flags.String("guest", "", "absolute path of compiled Mach-O in Darling guest root")
	root := flags.String("root", darling.SystemRoot, "Darling installation root")
	execute := flags.Bool("execute", false, "actually run the guest program")
	jsonOut := flags.Bool("json", false, "JSON report")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *guest == "" || flags.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "Usage: astelyn objc-probe run --guest /path [--root path] [--execute] [--json]")
		return 2
	}
	if err := darling.ValidateGuest(*guest, *root); err != nil {
		fmt.Fprintln(os.Stderr, "Astelyn:", err)
		return 1
	}
	if !*execute {
		fmt.Printf("Ready to probe guest Mach-O %s (dry-run; add --execute)\n", *guest)
		return 0
	}
	launcher, err := darling.Find()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Astelyn:", err)
		return 1
	}
	result := darling.Probe(launcher, *guest, 15*time.Second)
	passed := result.ExitCode == 0 && result.Error == "" && strings.Contains(result.Stdout, objcSmokeMarker)
	if *jsonOut {
		out := struct {
			Passed bool                `json:"passed"`
			Result darling.ProbeResult `json:"result"`
		}{passed, result}
		json.NewEncoder(os.Stdout).Encode(out)
	} else {
		if result.Stdout != "" {
			fmt.Print(result.Stdout)
		}
		if result.Stderr != "" {
			fmt.Fprint(os.Stderr, result.Stderr)
		}
		if passed {
			fmt.Println("Objective-C smoke marker verified")
		} else {
			fmt.Fprintln(os.Stderr, "Objective-C smoke marker missing, command failed, or nonzero exit")
		}
	}
	if passed {
		return 0
	}
	return 1
}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/XenoTechWorks/Astelyn/internal/darling"
)

func runPreflight(args []string) int {
	flags := flag.NewFlagSet("preflight", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	jsonOut := flags.Bool("json", false, "output a machine-readable JSON report")
	root := flags.String("root", darling.SystemRoot, "Darling system root directory")
	launcher := flags.String("launcher", "", "Darling launcher path (default: locate on PATH)")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "Usage: astelyn preflight [--json] [--root path] [--launcher path]")
		return 2
	}
	report := darling.CheckReadiness(*launcher, *root)
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	} else {
		fmt.Printf("Linux host: %s/%s\n", report.HostOS, report.HostArchitecture)
		fmt.Printf("Darling launcher found: %t\n", report.LauncherFound)
		fmt.Printf("Darling root found: %t\n", report.SystemRootFound)
		fmt.Printf("Darling guest Bash is Mach-O: %t\n", report.GuestBashMachO)
		fmt.Printf("Prepared for smoke test: %t\n", report.ReadyForSmokeTest)
		for _, note := range report.Notes {
			fmt.Println("Note:", note)
		}
	}
	if !report.ReadyForSmokeTest {
		return 1
	}
	return 0
}

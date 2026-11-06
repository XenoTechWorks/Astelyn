package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/XenoTechWorks/Astelyn/internal/darling"
	"github.com/XenoTechWorks/Astelyn/internal/runtimecheck"
)

func runRuntimeCheck(args []string) int {
	flags := flag.NewFlagSet("runtime-check", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	jsonOutput := flags.Bool("json", false, "print JSON report")
	root := flags.String("root", darling.SystemRoot, "host path to Darling root")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "Usage: astelyn runtime-check [--json] [--root path]")
		return 2
	}
	report, err := runtimecheck.Analyze(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Astelyn:", err)
		return 1
	}
	if *jsonOutput {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			return 1
		}
		return 0
	}
	fmt.Printf("Darling runtime inventory: %s\n", report.DarlingRoot)
	for _, item := range report.Components {
		state := "missing"
		if item.Present {
			state = "present"
		}
		fmt.Printf("  [%s] %s: %s", state, item.Name, item.GuestPath)
		if item.Present {
			fmt.Printf(" (%s %s)", item.Format, item.Architecture)
		}
		fmt.Println()
	}
	fmt.Println(report.Disclaimer)
	return 0
}

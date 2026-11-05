package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/XenoTechWorks/Astelyn/internal/darling"
	"github.com/XenoTechWorks/Astelyn/internal/frameworkcheck"
	"github.com/XenoTechWorks/Astelyn/internal/inspection"
)

func runFrameworks(args []string) int {
	flags := flag.NewFlagSet("frameworks", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	jsonOut := flags.Bool("json", false, "JSON report")
	root := flags.String("root", darling.SystemRoot, "host path to Darling system root")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "Usage: astelyn frameworks [--json] [--root path] <app.zip>")
		return 2
	}
	inspected, err := inspection.InspectZip(flags.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Astelyn:", err)
		return 1
	}
	report := frameworkcheck.Analyze(inspection.AnalyzeDependencies(inspected), *root)
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return 0
	}
	fmt.Printf("Framework imports: %d\n", len(report.Entries))
	for _, e := range report.Entries {
		fmt.Printf("  [%s] %s\n", e.State, e.Import)
	}
	fmt.Println(report.Disclaimer)
	return 0
}

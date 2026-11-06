package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/XenoTechWorks/Astelyn/internal/darling"
	"github.com/XenoTechWorks/Astelyn/internal/dyldreport"
	"io"
	"os"
)

func runRobloxAnalyze(args []string) int {
	flags := flag.NewFlagSet("roblox-test analyze", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	input := flags.String("input", "", "path to captured roblox-test run --json output, or - for stdin")
	root := flags.String("root", darling.SystemRoot, "Darling installation root")
	jsonOut := flags.Bool("json", false, "emit structured analysis JSON")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 || *input == "" {
		fmt.Fprintln(os.Stderr, "Usage: astelyn roblox-test analyze --input report.json| - [--root path] [--json]")
		return 2
	}
	var reader io.Reader = os.Stdin
	if *input != "-" {
		f, err := os.Open(*input)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Astelyn:", err)
			return 1
		}
		defer f.Close()
		reader = f
	}
	launch, err := dyldreport.Decode(reader)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Astelyn:", err)
		return 1
	}
	report, err := dyldreport.Analyze(launch, *root)
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
	fmt.Printf("Roblox dyld analysis: %s (exit=%d, timeout=%t)\n", report.Classification, report.ExitCode, report.TimedOut)
	for _, d := range report.Dependencies {
		fmt.Printf("  [%s] %s\n", d.State, d.InstallName)
		if d.ReferencedFrom != "" {
			fmt.Printf("    Referenced from: %s\n", d.ReferencedFrom)
		}
		if d.Reason != "" {
			fmt.Printf("    Reason: %s\n", d.Reason)
		}
	}
	if report.SharedCacheWarning {
		fmt.Println("  Notice: dyld reported no shared cache")
	}
	for _, note := range report.Notes {
		fmt.Println("Note:", note)
	}
	return 0
}

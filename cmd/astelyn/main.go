package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/XenoTechWorks/Astelyn/internal/inspection"
)

func main() { os.Exit(run(os.Args[1:])) }
func run(args []string) int {
	if len(args) < 1 || (args[0] != "inspect" && args[0] != "deps") {
		fmt.Fprintln(os.Stderr, "Usage: astelyn <inspect|deps> [--json] <RobloxPlayer.zip>")
		return 2
	}
	mode := args[0]

	flags := flag.NewFlagSet(mode, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	jsonOut := flags.Bool("json", false, "print JSON report")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if flags.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "Expected one ZIP archive path")
		return 2
	}
	report, err := inspection.InspectZip(flags.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Astelyn:", err)
		return 1
	}
	if mode == "deps" {
		deps := inspection.AnalyzeDependencies(report)
		if *jsonOut {
			if err := json.NewEncoder(os.Stdout).Encode(deps); err != nil {
				return 1
			}
			return 0
		}
		fmt.Printf("Dependency inventory: %d binaries, %d unique imports\n", deps.Binaries, len(deps.Dependencies))
		for _, dep := range deps.Dependencies {
			fmt.Printf("  [%s] %s (%d binaries)\n", dep.Category, dep.Name, len(dep.UsedBy))
		}
		fmt.Println("Static imports only; categories do not indicate compatibility.")
		return 0
	}
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			return 1
		}
		return 0
	}
	fmt.Printf("Archive: %s (%d entries)\n", report.Archive, report.Entries)
	for _, binary := range report.Binaries {
		fmt.Printf("\n%s — %s\n", binary.Path, strings.Join(binary.Architectures, ", "))
		for _, lib := range binary.Libraries {
			fmt.Println("  ", lib)
		}
	}
	fmt.Println("\nStatic inspection only. Astelyn cannot launch Roblox yet.")
	return 0
}

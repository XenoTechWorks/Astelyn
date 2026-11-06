package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/XenoTechWorks/Astelyn/internal/darling"
	"github.com/XenoTechWorks/Astelyn/internal/dyldreport"
	"github.com/XenoTechWorks/Astelyn/internal/gfxaudit"
	"github.com/XenoTechWorks/Astelyn/internal/graphics"
	"github.com/XenoTechWorks/Astelyn/internal/inspection"
)

func runGraphicsAudit(args []string) int {
	flags := flag.NewFlagSet("graphics-audit", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	root := flags.String("root", darling.SystemRoot, "Darling root")
	compare := flags.String("compare-root", "", "optional second Darling root")
	launch := flags.String("launch", "", "optional JSON report from roblox-test run")
	host := flags.Bool("host", false, "run native Linux Vulkan/OpenGL diagnostics")
	jsonOut := flags.Bool("json", false, "print JSON report")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "Usage: astelyn graphics-audit [--json] [--host] [--root path] [--compare-root path] [--launch report.json] <RobloxPlayer.zip>")
		return 2
	}
	inspected, err := inspection.InspectZip(flags.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Astelyn:", err)
		return 1
	}
	hostAPIs := []gfxaudit.HostAPI{}
	if *host {
		for _, backend := range []graphics.Backend{graphics.Vulkan, graphics.OpenGL} {
			d := graphics.Probe(backend)
			hostAPIs = append(hostAPIs, gfxaudit.HostAPI{API: string(backend), Available: d.Available, Diagnostic: d.Tool, Detail: d.Details})
		}
	}
	loaderClass := ""
	missing := []string{}
	if *launch != "" {
		f, e := os.Open(*launch)
		if e != nil {
			fmt.Fprintln(os.Stderr, "Astelyn:", e)
			return 1
		}
		captured, e := dyldreport.Decode(f)
		f.Close()
		if e != nil {
			fmt.Fprintln(os.Stderr, "Astelyn:", e)
			return 1
		}
		analyzed, e := dyldreport.Analyze(captured, *root)
		if e != nil {
			fmt.Fprintln(os.Stderr, "Astelyn:", e)
			return 1
		}
		loaderClass = analyzed.Classification
		for _, d := range analyzed.Dependencies {
			if d.State == "missing" {
				missing = append(missing, d.InstallName)
			}
		}
	}
	report, err := gfxaudit.Analyze(inspection.AnalyzeDependencies(inspected), *root, *compare, hostAPIs, loaderClass, missing)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Astelyn:", err)
		return 1
	}
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return 0
	}
	fmt.Printf("Astelyn graphics audit: %d graphics imports (%d missing, %d files present)\n", len(report.GraphicsImports), report.Missing, report.Present)
	for _, r := range report.GraphicsImports {
		fmt.Printf("  %-10s %-22s %s\n", r.State, r.Framework, r.Role)
	}
	if report.LoaderBlocked {
		fmt.Printf("Loader blocked before graphics initialization: %v\n", report.MissingLoaderLibraries)
	}
	for _, a := range report.HostAPIs {
		fmt.Printf("Host %s: available=%t (%s)\n", a.API, a.Available, a.Detail)
	}
	if *compare != "" {
		fmt.Printf("Additional framework files in comparison root: %v\n", report.ImprovedInComparison)
	}
	fmt.Println("Read-only inventory. Host rendering availability is not guest rendering support.")
	return 0
}

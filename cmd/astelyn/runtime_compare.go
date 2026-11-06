package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/XenoTechWorks/Astelyn/internal/darling"
	"github.com/XenoTechWorks/Astelyn/internal/dyldreport"
	"github.com/XenoTechWorks/Astelyn/internal/runtimecompare"
)

func runRuntimeCompare(args []string) int {
	flags := flag.NewFlagSet("runtime-compare", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	launcher := flags.String("launcher", "darling", "Darling CLI path; use alternate CiderDarling executable here")
	root := flags.String("root", darling.SystemRoot, "Darling runtime filesystem root used for diagnosis only")
	prefix := flags.String("prefix", "", "isolated Darling prefix (required for execution)")
	app := flags.String("app", "", "RobloxPlayer.app path")
	timeout := flags.Duration("timeout", 10*time.Second, "maximum 30s")
	execute := flags.Bool("execute", false, "opt-in executable launch")
	baselineFile := flags.String("baseline", "", "optional prior roblox-test run --json capture")
	jsonOut := flags.Bool("json", false, "JSON output")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 || *app == "" {
		fmt.Fprintln(os.Stderr, "Usage: astelyn runtime-compare --app /path/RobloxPlayer.app [--launcher /path/to/darling] [--root /path/to/sysroot] [--prefix /absolute/test-prefix] [--baseline old-launch.json] [--execute] [--json]")
		return 2
	}
	req := runtimecompare.Request{Launcher: *launcher, Root: *root, Prefix: *prefix, App: *app, Timeout: *timeout}
	var report runtimecompare.Report
	var err error
	if !*execute {
		report, err = runtimecompare.Validate(req, false)
	} else {
		var baseline *dyldreport.Launch
		if *baselineFile != "" {
			f, e := os.Open(*baselineFile)
			if e != nil {
				fmt.Fprintln(os.Stderr, "Astelyn:", e)
				return 1
			}
			l, e := runtimecompare.DecodeLaunch(f)
			_ = f.Close()
			if e != nil {
				fmt.Fprintln(os.Stderr, "Astelyn:", e)
				return 1
			}
			baseline = &l
		}
		report, err = runtimecompare.Run(context.Background(), req, baseline)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Astelyn:", err)
		return 1
	}
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, "Astelyn:", err)
			return 1
		}
		return 0
	}
	fmt.Printf("Runtime CLI: %s\nRuntime root (lookup): %s\nIsolated prefix: %s\nGuest: %s\n", report.Launcher, report.Root, report.Prefix, report.GuestExecutable)
	if !report.Executed {
		fmt.Println("DRY RUN: add --execute (requires --prefix) to try the selected runtime.")
		return 0
	}
	fmt.Printf("Launch exit=%d timeout=%t\n", report.Launch.ExitCode, report.Launch.TimedOut)
	fmt.Printf("Loader classification: %s\n", report.Diagnosis.Classification)
	for _, d := range report.Diagnosis.Dependencies {
		fmt.Printf("  %s (%s)\n", d.InstallName, d.State)
	}
	if report.Comparison != nil {
		fmt.Printf("Same first loader blocker: %t\n", report.Comparison.SameFirstBlocker)
	}
	fmt.Println("No compatibility or gameplay success is implied.")
	return 0
}

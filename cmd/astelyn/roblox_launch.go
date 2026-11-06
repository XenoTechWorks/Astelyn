package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/XenoTechWorks/Astelyn/internal/robloxtest"
)

func runRobloxTest(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: astelyn roblox-test plan [--json] <RobloxPlayer.zip> | run --app /path/RobloxPlayer.app [--timeout 10s] [--execute] [--json]")
		return 2
	}
	switch args[0] {
	case "analyze":
		return runRobloxAnalyze(args[1:])
	case "plan":
		flags := flag.NewFlagSet("roblox-test plan", flag.ContinueOnError)
		jsonOut := flags.Bool("json", false, "JSON report")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 1 {
			return 2
		}
		p, err := robloxtest.Assess(flags.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, "Astelyn:", err)
			return 1
		}
		if *jsonOut {
			if err := robloxtest.Encode(os.Stdout, p); err != nil {
				return 1
			}
		} else {
			fmt.Printf("Roblox executable found: %t\nPath: %s\nArchitectures: %v\nDirect imports: %d\n%s\n", p.Found, p.Executable, p.Architectures, p.DirectImports, p.Note)
		}
		if !p.Found {
			return 1
		}
		return 0
	case "run":
		flags := flag.NewFlagSet("roblox-test run", flag.ContinueOnError)
		app := flags.String("app", "", "path to extracted RobloxPlayer.app directory")
		timeout := flags.Duration("timeout", 10*time.Second, "diagnostic timeout, maximum 30s")
		execute := flags.Bool("execute", false, "actually start Roblox through Darling")
		jsonOut := flags.Bool("json", false, "JSON report")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 || *app == "" || *timeout <= 0 || *timeout > 30*time.Second {
			return 2
		}
		executable := *app
		if len(executable) < len("/Contents/MacOS/RobloxPlayer") || executable[len(executable)-len("/Contents/MacOS/RobloxPlayer"):] != "/Contents/MacOS/RobloxPlayer" {
			executable = executable + "/Contents/MacOS/RobloxPlayer"
		}
		guest, err := robloxtest.GuestPath(executable)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Astelyn:", err)
			return 1
		}
		if !*execute {
			fmt.Printf("DRY RUN: darling shell %s (timeout %s). Add --execute to launch; this does not bypass Roblox protections.\n", guest, *timeout)
			return 0
		}
		launcher, err := exec.LookPath("darling")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Astelyn: Darling missing:", err)
			return 1
		}
		report, err := robloxtest.Launch(context.Background(), launcher, executable, *timeout)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Astelyn:", err)
			return 1
		}
		if *jsonOut {
			if err := robloxtest.Encode(os.Stdout, report); err != nil {
				return 1
			}
		} else {
			fmt.Print(report.Stdout)
			fmt.Fprint(os.Stderr, report.Stderr)
			fmt.Printf("Roblox diagnostic: exit=%d timeout=%t error=%q\n%s\n", report.ExitCode, report.TimedOut, report.Error, report.Note)
		}
		// An app still running at the timeout is not a failed diagnostic capture.
		return 0
	default:
		fmt.Fprintln(os.Stderr, "Unknown roblox-test subcommand")
		return 2
	}
}

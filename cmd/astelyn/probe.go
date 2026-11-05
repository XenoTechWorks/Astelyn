package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/XenoTechWorks/Astelyn/internal/darling"
)

func runProbe(args []string) int {
	flags := flag.NewFlagSet("probe", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	jsonOut := flags.Bool("json", false, "machine-readable result")
	stdinPath := flags.String("stdin", "", "input file or - for stdin")
	launcher := flags.String("launcher", "", "Darling launcher")
	root := flags.String("root", darling.SystemRoot, "Darling system root")
	timeout := flags.Duration("timeout", 20*time.Second, "execution timeout")

	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() < 1 || *timeout <= 0 {
		fmt.Fprintln(os.Stderr,
			"Usage: astelyn probe [--json] [--stdin file|-] [--timeout 20s] /guest/program [-- args...]")
		return 2
	}

	guest := flags.Arg(0)
	if err := darling.ValidateGuest(guest, *root); err != nil {
		fmt.Fprintln(os.Stderr, "Astelyn:", err)
		return 1
	}

	exe := *launcher
	if exe == "" {
		var err error
		exe, err = darling.Find()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Astelyn:", err)
			return 1
		}
	}

	var input io.Reader
	if *stdinPath == "-" {
		input = os.Stdin
	} else if *stdinPath != "" {
		f, err := os.Open(*stdinPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Astelyn:", err)
			return 1
		}
		defer f.Close()
		input = f
	}

	forwarded := guestArguments(flags.Args()[1:])
	result := darling.ProbeWithInput(exe, guest, *timeout, input, forwarded...)

	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	} else {
		if _, err := os.Stdout.Write(result.StdoutBytes); err != nil {
			fmt.Fprintln(os.Stderr, "Astelyn stdout:", err)
			return 1
		}
		if _, err := os.Stderr.Write(result.StderrBytes); err != nil {
			fmt.Fprintln(os.Stderr, "Astelyn stderr:", err)
			return 1
		}

		fmt.Fprintf(os.Stderr, "Astelyn probe: exit=%d timeout=%t\n",
			result.ExitCode, result.TimedOut)

		if result.Error != "" {
			fmt.Fprintln(os.Stderr, "Astelyn:", result.Error)
		}
	}

	if result.TimedOut {
		return 124
	}
	if result.Error != "" || result.ExitCode < 0 {
		return 1
	}
	return result.ExitCode
}

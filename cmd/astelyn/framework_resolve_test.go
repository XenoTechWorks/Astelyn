package main

import "testing"

func TestFrameworkResolveUsage(t *testing.T) {
	for _, args := range [][]string{{}, {"--bad-option"}, {"first.zip", "second.zip"}} {
		if got := runFrameworkResolve(args); got != 2 {
			t.Fatalf("runFrameworkResolve(%q) = %d; want 2", args, got)
		}
	}
}

func TestFrameworkResolveMissingArchive(t *testing.T) {
	if got := runFrameworkResolve([]string{"--json", "/this-file-does-not-exist-astelyn.zip"}); got != 1 {
		t.Fatalf("missing archive returned %d; want 1", got)
	}
}

package main

import "testing"

func TestWindowBridgeArgs(t *testing.T) {
	for _, args := range [][]string{{}, {"oops"}, {"demo", "--driver", "cocoa"}, {"demo", "--json", "unused"}} {
		if got := runWindowBridge(args); got != 2 {
			t.Errorf("runWindowBridge(%v)=%d want 2", args, got)
		}
	}
	if got := runWindowBridge([]string{"demo"}); got != 0 {
		t.Fatalf("dry run code %d", got)
	}
}

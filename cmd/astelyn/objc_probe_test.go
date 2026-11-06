package main
import "testing"
func TestObjCProbeRejectsExtraArgs(t *testing.T){if got:=runObjCProbe([]string{"unexpected"});got!=2{t.Fatalf("got %d, want 2",got)}}
func TestObjCSmokeRequiresGuest(t *testing.T){if got:=runObjCProbe([]string{"run","--execute"});got!=2{t.Fatalf("got %d, want 2",got)}}

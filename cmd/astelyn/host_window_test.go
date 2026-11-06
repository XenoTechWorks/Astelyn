package main
import "testing"
func TestHostWindowBadArguments(t *testing.T) {
    for _, args := range [][]string{{},{"nope"},{"run","--seconds","0"},{"run","--driver","cocoa"},{"status","extra"}} {
        if got := runHostWindow(args); got != 2 {t.Fatalf("args=%v, got=%d",args,got)}
    }
}

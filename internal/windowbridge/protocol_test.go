package windowbridge

import "testing"

func TestEncode(t *testing.T) {
	cases := []struct {
		c    Command
		want string
		ok   bool
	}{
		{Command{Op: "PING"}, "PING\n", true},
		{Command{Op: "QUIT"}, "QUIT\n", true},
		{Command{Op: "SIZE", Width: 800, Height: 480}, "SIZE 800 480\n", true},
		{Command{Op: "TITLE", Title: "Astelyn Bridge"}, "TITLE Astelyn Bridge\n", true},
		{Command{Op: "TITLE", Title: "bad\nQUIT"}, "", false},
		{Command{Op: "SIZE", Width: 9999, Height: 5}, "", false},
		{Command{Op: "BOGUS"}, "", false},
	}
	for _, tc := range cases {
		got, err := Encode(tc.c)
		if (err == nil) != tc.ok || got != tc.want {
			t.Errorf("%+v: got %q err=%v", tc.c, got, err)
		}
	}
}
func TestParseEvent(t *testing.T) {
	for _, s := range []string{"READY", "DRIVER wayland", "RESIZED 720 420", "PONG", "KEY escape", "CLOSED", "QUIT"} {
		if _, err := ParseEvent(s); err != nil {
			t.Errorf("%s: %v", s, err)
		}
	}
	for _, s := range []string{"", "RESIZED x 2", "RESIZED 1", "DRIVER ", "HACK foo"} {
		if _, err := ParseEvent(s); err == nil {
			t.Errorf("accepted %q", s)
		}
	}
}

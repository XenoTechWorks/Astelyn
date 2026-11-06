package darwinbridge

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	r, err := Parse(strings.NewReader(Header + "\nTITLE Astelyn guest request\nSIZE 840 520\n"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Title != "Astelyn guest request" || r.Width != 840 || r.Height != 520 {
		t.Fatalf("%+v", r)
	}
	commands, err := r.Commands()
	if err != nil {
		t.Fatal(err)
	}
	if len(commands) != 3 || commands[2] != "PING\n" {
		t.Fatalf("%v", commands)
	}
}
func TestReject(t *testing.T) {
	for _, sample := range []string{
		"TITLE Bad\nSIZE 800 480\n",
		Header + "\nTITLE hi\nSIZE 2 4\n",
		Header + "\nTITLE hi\nSIZE 800 480 extra\n",
		Header + "\nTITLE hi\nSIZE 800 480\nQUIT\n",
		Header + "\nTITLE bad\\n\nSIZE 800 480\n",
		Header + "\nTITLE hi\nSIZE 800 480\n" + strings.Repeat("a", 5000),
	} {
		if _, err := Parse(strings.NewReader(sample)); err == nil {
			t.Fatalf("accepted %q", sample)
		}
	}
}

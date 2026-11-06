// Package darwinbridge translates a tiny guest-originated request format into
// Astelyn's existing Linux host-window protocol. It does NOT implement Cocoa.
package darwinbridge

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/XenoTechWorks/Astelyn/internal/windowbridge"
)

const Header = "ASTELYN_GUEST_WINDOW_V1"

type Request struct {
	Title  string `json:"title"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// Parse accepts only a fixed, strictly bounded request. Guest output is
// untrusted; it must not turn into shell commands or arbitrary host operations.
func Parse(r io.Reader) (Request, error) {
	s := bufio.NewScanner(io.LimitReader(r, 4097))
	s.Buffer(make([]byte, 128), 4096)
	lines := make([]string, 0, 4)
	for s.Scan() {
		lines = append(lines, strings.TrimSuffix(s.Text(), "\r"))
		if len(lines) > 3 {
			return Request{}, errors.New("too many guest request lines")
		}
	}
	if err := s.Err(); err != nil {
		return Request{}, fmt.Errorf("guest request too large or unreadable: %w", err)
	}
	if len(lines) != 3 || lines[0] != Header {
		return Request{}, errors.New("expected guest window protocol header, TITLE and SIZE")
	}
	if !strings.HasPrefix(lines[1], "TITLE ") {
		return Request{}, errors.New("missing TITLE request")
	}
	title := strings.TrimPrefix(lines[1], "TITLE ")
	size := strings.TrimPrefix(lines[2], "SIZE ")
	var w, h int
	var extra string
	if n, err := fmt.Sscanf(size, "%d %d %s", &w, &h, &extra); err != io.EOF || n != 2 {
		// Sscanf returns EOF after successfully scanning exactly two values.
		return Request{}, errors.New("invalid SIZE request")
	}
	if _, err := windowbridge.Encode(windowbridge.Command{Op: "TITLE", Title: title}); err != nil {
		return Request{}, fmt.Errorf("invalid guest title: %w", err)
	}
	if _, err := windowbridge.Encode(windowbridge.Command{Op: "SIZE", Width: w, Height: h}); err != nil {
		return Request{}, fmt.Errorf("invalid guest size: %w", err)
	}
	return Request{Title: title, Width: w, Height: h}, nil
}

// Commands returns validated commands to deliver to the existing SDL2 window.
func (r Request) Commands() ([]string, error) {
	out := make([]string, 0, 3)
	for _, c := range []windowbridge.Command{{Op: "TITLE", Title: r.Title}, {Op: "SIZE", Width: r.Width, Height: r.Height}, {Op: "PING"}} {
		v, err := windowbridge.Encode(c)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

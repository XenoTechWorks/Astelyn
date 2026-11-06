// Package windowbridge defines an intentionally small host-side window control protocol.
// It is NOT Cocoa, NSWindow, CoreGraphics, or a macOS API implementation.
package windowbridge

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Command struct {
	Op     string
	Width  int
	Height int
	Title  string
}

// Encode validates the command before encoding one ASCII line for the native host.
// No raw newlines or delimiters are permitted from user-supplied strings.
func Encode(c Command) (string, error) {
	switch c.Op {
	case "PING", "QUIT":
		if c.Width != 0 || c.Height != 0 || c.Title != "" {
			return "", errors.New("unexpected command fields")
		}
		return c.Op + "\n", nil
	case "SIZE":
		if c.Width < 200 || c.Width > 1920 || c.Height < 150 || c.Height > 1080 || c.Title != "" {
			return "", errors.New("size out of range (200..1920 x 150..1080)")
		}
		return fmt.Sprintf("SIZE %d %d\n", c.Width, c.Height), nil
	case "TITLE":
		if len(c.Title) == 0 || len(c.Title) > 100 || c.Width != 0 || c.Height != 0 {
			return "", errors.New("invalid title length or fields")
		}
		for _, r := range c.Title {
			if r < ' ' || r > 126 || r == '|' || r == '\\' {
				return "", errors.New("title must be printable ASCII without | or backslash")
			}
		}
		return "TITLE " + c.Title + "\n", nil
	default:
		return "", fmt.Errorf("unknown operation %q", c.Op)
	}
}

type Event struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}

func ParseEvent(line string) (Event, error) {
	line = strings.TrimSuffix(line, "\r")
	if line == "READY" || line == "PONG" || line == "CLOSED" || line == "QUIT" {
		return Event{Name: line}, nil
	}
	for _, prefix := range []string{"DRIVER ", "RESIZED ", "KEY ", "ERROR "} {
		if strings.HasPrefix(line, prefix) {
			value := strings.TrimPrefix(line, prefix)
			if len(value) == 0 || len(value) > 400 {
				return Event{}, errors.New("invalid event value")
			}
			if prefix == "RESIZED " {
				a := strings.Fields(value)
				if len(a) != 2 {
					return Event{}, errors.New("invalid resize event")
				}
				for _, v := range a {
					n, e := strconv.Atoi(v)
					if e != nil || n < 1 || n > 32768 {
						return Event{}, errors.New("invalid resize dimension")
					}
				}
			}
			return Event{Name: strings.TrimSpace(prefix), Value: value}, nil
		}
	}
	return Event{}, fmt.Errorf("unrecognized window event %q", line)
}

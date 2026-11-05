package main

import (
	"reflect"
	"testing"
)

func TestGuestArguments(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			"separator",
			[]string{"--", "--noprofile", "-c", "echo hello"},
			[]string{"--noprofile", "-c", "echo hello"},
		},
		{
			"no separator",
			[]string{"--norc", "-c", "echo hello"},
			[]string{"--norc", "-c", "echo hello"},
		},
		{
			"empty",
			nil,
			nil,
		},
		{
			"preserve later separator",
			[]string{"-c", "echo hello", "--"},
			[]string{"-c", "echo hello", "--"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := guestArguments(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

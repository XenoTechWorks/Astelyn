package main

// guestArguments removes the optional CLI separator.
// All remaining arguments are passed unchanged to the guest.
func guestArguments(args []string) []string {
	if len(args) > 0 && args[0] == "--" {
		return args[1:]
	}
	return args
}

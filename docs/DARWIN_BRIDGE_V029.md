# Astelyn v0.2.9: Darwin-originated window request

This is **not** Cocoa/AppKit/NSWindow/Metal compatibility. A small macOS Mach-O
runs inside Darling and prints a bounded request on stdout. A Go parser validates
its protocol and applies a title/size to the existing Linux SDL2 window bridge.
The guest program makes **no GUI API calls**. The two processes are run by the
Astelyn Go orchestrator; there is no production cross-process macOS framework API.

Tested paths are based on the Void Linux + Darling layout used for v0.2.5.

## Void Linux requirements

- `clang-21`, `ld64.lld` (`lld` package)
- Darling installed at `/usr/local/libexec/darling` and on `PATH`
- `SDL2-devel`, `pkg-config`, compiler `cc`

```sh
go test ./...
go run ./cmd/astelyn darwin-bridge build
go run ./cmd/astelyn darwin-bridge demo --driver wayland --execute
```

The guest reports the versioned header `ASTELYN_GUEST_WINDOW_V1`, a TITLE,
and a SIZE, then exits. The Go program requires the exact protocol and bounds
specified by `internal/windowbridge` before sending commands to native SDL2.
This prevents arbitrary guest data from becoming a shell command.

Successful output ends with `ASTELYN_DARWIN_HOST_BRIDGE_OK`.

### v0.3.0–v0.3.5

First Roblox **diagnostic tests** are planned for this range, not guaranteed
playability. AppKit, Metal, window semantics, game input, and other frameworks
remain blockers. Don't work around Roblox security or anti-cheat measures.

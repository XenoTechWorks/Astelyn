# Astelyn v0.2.8 — Host Window IPC Bridge

This is an **experimental native Linux window**, NOT a macOS `NSWindow`, Cocoa, AppKit, Metal, QuartzCore, or Roblox implementation.

The Go CLI starts a local SDL2 process and sends bounded ASCII line commands through stdin; the C process emits event lines on stdout. No socket, network port, privileges, or changes to Darling are needed.

## Void Linux prerequisites

`sudo xbps-install -S SDL2-devel pkg-config`

## Build and run

```
go run ./cmd/astelyn window-bridge build
go run ./cmd/astelyn window-bridge demo --driver wayland --execute
# or: --driver x11
# or: --json for JSON event lines
```

The CLI waits for `READY`, sets title and window size, sends `PING`, and waits for `PONG`. It then requests `QUIT`, and expects `CLOSED`. The session is time-limited to 15 seconds. Press Escape or close the window to terminate early; a manually closed window before the full handshake may produce an incomplete-session error.

### Protocol (host-only v1)

Commands: `PING`, `SIZE width height` (limited to 200..1920 x 150..1080), `TITLE printable ASCII`, `QUIT`. Events include `DRIVER wayland`, `READY`, `RESIZED 800 480`, `PONG`, `KEY escape`, `QUIT`, `CLOSED`, `ERROR ...`.

This is an intentionally small demonstration. A future **Darwin-side API bridge** would need to implement real Objective-C classes and methods, lifecycle handling, threading, event dispatch, and graphics translation. A working SDL window by itself does not let macOS applications display windows.

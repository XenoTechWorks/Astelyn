# Astelyn v0.2.7 — Native Linux Host Window

This is **not** an AppKit or Cocoa implementation. It creates an ordinary Linux SDL2
window to verify the host-side Wayland or X11/XWayland path. It does not load a
macOS application or translate NSWindow, CoreGraphics, or Metal calls.

## Void Linux prerequisites

```sh
sudo xbps-install -S SDL2-devel pkg-config gcc
```

Use an existing compiler if already installed. Build:

```sh
go run ./cmd/astelyn host-window status
go run ./cmd/astelyn host-window build
go run ./cmd/astelyn host-window run --driver wayland --seconds 12 --execute
```

Use `--driver x11` to test XWayland when `DISPLAY` and XWayland are available.
Use `--driver auto` for SDL2 backend selection. The test window remains open up
to the specified seconds or until Escape / window close. Its terminal output
reports the **actually selected** SDL2 video backend, not an inferred one.

Expected startup line: `ASTELYN_HOST_WINDOW_CREATED backend=wayland` (or `x11`).
Closing normally prints `ASTELYN_HOST_WINDOW_CLOSED`.

## Architecture boundary

Go orchestrates and validates; the tiny C SDL2 program creates a native host
window. Future integration would require separate implementations of macOS
window semantics, Obj-C classes and event translation. SDL2 does not provide
any of these automatically. Keep GUI functionality clearly separated from
Darling and the macOS framework resolver.

The compiled binary is stored at `.astelyn/build/astelyn-host-window`.
Do not commit that executable. Add `.astelyn/` to `.gitignore` if necessary.

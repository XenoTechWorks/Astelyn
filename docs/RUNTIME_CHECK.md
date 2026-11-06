# v0.2.3 — Darling runtime inventory

Use `astelyn runtime-check --json` to examine key Objective-C, Dispatch, Swift overlay, and framework library files in the installed Darling root. Override with `--root /path/to/test/root` for fixture validation.

This diagnostic reads host filesystem metadata and checks binary format headers. It **does not execute** the libraries, load Objective-C classes, perform ABI compatibility checks, verify the framework APIs, or claim graphical macOS compatibility. `libswiftAppKit.dylib` is a Swift overlay and does not substitute for `AppKit.framework`.

Run `go test ./...` after applying. Do not change or symlink installed Darling libraries based on this report.

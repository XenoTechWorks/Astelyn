# Astelyn v0.3.0 — First Roblox launch diagnostics

This is an experimental diagnostic path and **not** a working Roblox compatibility layer. The macOS GUI frameworks and graphics translation remain incomplete. The launch may fail immediately, show a dyld error, or time out. Do not bypass Roblox protections or modify its binaries.

## Read-only assessment

```sh
go run ./cmd/astelyn roblox-test plan --json ~/version-2fc79d515be6449d-RobloxPlayer.zip
```

## Optional controlled launch

Extract your legitimate macOS RobloxPlayer ZIP yourself into a **user-writable** directory, without replacing Darling system libraries. Choose the resulting `RobloxPlayer.app` path (e.g. `~/Downloads/RobloxPlayer.app`). Astelyn does **not** automatically extract the archive.

```sh
go run ./cmd/astelyn roblox-test run --app "$HOME/Downloads/RobloxPlayer.app"
go run ./cmd/astelyn roblox-test run --app "$HOME/Downloads/RobloxPlayer.app" --timeout 10s --execute --json
```

The executable must be an actual Mach-O file at `RobloxPlayer.app/Contents/MacOS/RobloxPlayer`. The host path is mapped into Darling through the previously tested `/Volumes/SystemRoot` mount. Logs are capped at 64 KiB per stream, and execution is limited to 30 seconds. Diagnostic execution is **opt-in**. The program may still try normal application networking during a run; do not use `--execute` if that is undesirable.

## Interpreting reports

- `exit_code` is the guest command exit code (or -1 if no normal exit code was available).
- `timed_out` means the bounded diagnostic limit expired; it is **not** proof of success or failure of the app.
- `stderr` often contains Darling or dyld loader failures. It may contain personal paths or other diagnostic data; review it before posting publicly.
- No GUI/Cocoa window, Metal or anti-cheat compatibility is provided by this update.

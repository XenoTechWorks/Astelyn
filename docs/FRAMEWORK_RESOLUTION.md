# Astelyn v0.2.2 — Darling framework resolution

```sh
go run ./cmd/astelyn framework-resolve --json ~/RobloxPlayer.zip
go run ./cmd/astelyn framework-resolve --root /usr/local/libexec/darling ~/RobloxPlayer.zip
```

This is **read-only**. It compares directly imported, absolute macOS framework install names with files in the configured Darling system root.

- `exact`: an existing regular file at the imported framework path.
- `alternate`: an existing regular file under the same framework using an unversioned entry, `Versions/Current`, or another installed version.
- `missing`: no candidate found under the checked framework directory.

All `host_path` values are on the **Linux host**, and symlink targets that escape the configured Darling root are rejected.

The resolver intentionally does not handle `@rpath`, `@loader_path`, or nested/transitive imports. An alternate version **is not necessarily ABI-compatible**. File presence is not proof of available classes, symbols, Metal/AppKit implementations, or a runnable Roblox Player. Swift overlay dylibs do not replace macOS system frameworks.

This command is separate from the original `frameworks` inventory so its results can be compared without breaking the existing interface.

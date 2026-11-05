# Framework inventory (v0.2 groundwork)

Astelyn can compare **directly imported** Mach-O framework paths against files in
an installed Darling root. This is a read-only, preliminary inventory.

```sh
go run ./cmd/astelyn frameworks --json ~/RobloxPlayer.zip
go run ./cmd/astelyn frameworks --root /usr/local/libexec/darling ~/RobloxPlayer.zip
```

Statuses:
- `file-present`: corresponding *regular file* exists in Darling's root; this is **not** evidence of functional API support.
- `missing`: an absolute path under `/System/Library/Frameworks` or `/Library/Frameworks` has no matching regular file.
- `unresolved`: import uses dyld-relative addressing or is otherwise not safely resolvable.

This command does not inspect transitive imports, dynamic loading, exported symbols,
Objective-C selectors, app bundles outside the current ZIP inspector scope, or
runtime compatibility. It does not install frameworks, execute binaries, or
attempt Roblox authentication or security circumvention.

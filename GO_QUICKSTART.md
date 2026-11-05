# Go CLI — dependency inspection (v0.1.1 development)

Astelyn's new Go CLI inspects Mach-O architectures and linked dylibs/frameworks in an official macOS Roblox ZIP. It is read-only and does not extract or run Roblox.

```bash
go test ./...
go run ./cmd/astelyn inspect ~/version-2fc79d515be6449d-RobloxPlayer.zip
go run ./cmd/astelyn inspect --json ~/version-2fc79d515be6449d-RobloxPlayer.zip
```

The older `python3 -m astelyn` commands remain available. Runtime loading, Darling integration, and graphical rendering are **not implemented**.

## Dependency inventory (v0.1.2)

```bash
go run ./cmd/astelyn deps ~/version-2fc79d515be6449d-RobloxPlayer.zip
go run ./cmd/astelyn deps --json ~/version-2fc79d515be6449d-RobloxPlayer.zip
```

`deps` groups Mach-O imports by path type, deduplicates them across the inspected executables, and reports which executables reference each import in JSON. It does not resolve dependencies, assess Darling support, or execute macOS code.

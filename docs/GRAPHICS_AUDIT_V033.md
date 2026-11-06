# Astelyn v0.3.3 — Graphics / runtime gap audit

This Go command separates **Linux host diagnostics** (Mesa Vulkan/OpenGL)
from the **macOS framework dependencies** directly imported by Roblox Player.

```sh
go run ./cmd/astelyn graphics-audit --json ~/version-2fc79d515be6449d-RobloxPlayer.zip

go run ./cmd/astelyn graphics-audit --host --launch /tmp/astelyn-roblox-launch.json --json ~/version-2fc79d515be6449d-RobloxPlayer.zip
```

Use `--compare-root /absolute/path` to compare a separately installed runtime.
**Do not overwrite your existing Darling installation.** Comparing a root is
filesystem-only; it does not attest compatibility, symbol correctness, licensing,
or gameplay support. The `--launch` option analyzes a preexisting v0.3.0 JSON
capture; it does **not** launch Roblox.

The audit deliberately omits unknown unverified replacements for Metal, CoreML,
Cocoa, and other Apple APIs. The reported class of renderer (Metal/OpenGL) and
framework file presence alone cannot prove a working renderer. A CoreML dyld
error occurs before rendering begins and should not be interpreted as a GPU
failure.

No third-party code is copied, and nothing is injected or bypassed.

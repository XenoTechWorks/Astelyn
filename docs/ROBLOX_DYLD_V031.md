# Astelyn v0.3.1: dyld launch failure analyzer

Use `astelyn roblox-test analyze --input <capture.json>` to review an existing `roblox-test run --json` report.
The analyzer is **read-only** and doesn't launch Roblox.

## Workflow

```sh
go run ./cmd/astelyn roblox-test run --app "$HOME/Games/Astelyn/RobloxPlayer.app" --timeout 10s --execute --json > /tmp/astelyn-roblox-launch.json
go run ./cmd/astelyn roblox-test analyze --input /tmp/astelyn-roblox-launch.json --json
```

You can test the analyzer without launching anything using `--input examples/roblox-launch-coreml.json`.

The report shows only missing dependencies actually named in dyld logs; it is not a complete dependency list. A `present-file` classification means a file exists at the requested location, **not** that it can be loaded or provides correct APIs. The shared-cache warning is reported separately. `@rpath` requests are unresolved, and symlinks escaping the Darling root are marked unsafe.

No libraries are installed, rewritten, redirected, or faked. This release does not implement CoreML or Roblox compatibility.

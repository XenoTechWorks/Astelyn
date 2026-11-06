# Astelyn v0.3.2 — Compatibility Runtime Audit

This is a read-only comparison of direct macOS framework imports from a Roblox ZIP against one or two *existing* Darling system roots. It does **not** install or execute Roblox, add missing frameworks, or guarantee binary compatibility.

## Commands

```sh
go run ./cmd/astelyn compat-audit --json "$HOME/version-2fc79d515be6449d-RobloxPlayer.zip"
```

Optional comparison against another installed runtime root **only if you already have one**:

```sh
go run ./cmd/astelyn compat-audit --json --root /usr/local/libexec/darling --compare-root /opt/other-runtime/libexec/darling "$HOME/version-2fc79d515be6449d-RobloxPlayer.zip"
```

The comparison root is not autodetected: supply the actual filesystem directory containing `System/Library/Frameworks`. A runtime command path such as `/usr/bin/darling` is **not** a system root. Do not point `--compare-root` at a live or user-private Darling prefix. Comparison is file-level only.

States: `exact`, `alternate`, and `missing`. A matching file may still lack required symbols or crash at load time. Relative install names, `dlopen` dependencies and transitive dependencies are out of scope.

## Source investigation

- Quadra: https://github.com/QuadraHQ/Quadra
- Quadra notes: https://github.com/QuadraHQ/Quadra/blob/main/docs/NOTES.md
- Quadra's CiderDarling runtime: https://github.com/QuadraHQ/CiderDarling
- Original Mac O' Blox: https://github.com/narezy/MacOBlox

Quadra documents the extra runtime fixes and `libQuadraShims.dylib`, but these notes do not verify a specific implementation of CoreML. This package contains no third-party source code or shims. Review licensing and runtime security before adopting or redistributing components. Quadra states GPLv3 licensing.

# Astelyn v0.3.4 — Opt-in Darling Runtime Comparison

`astelyn runtime-compare` is a read-only diagnostic until `--execute` is explicitly provided. It supports selecting a different Darling CLI executable and an isolated `DPREFIX`, then comparing the resulting loader errors to a previously captured Roblox launch JSON.

## Why runtime isolation matters

Quadra's current documentation states regular Darling uses `~/.darling` and CiderDarling defaults to `~/.ciderdarling`, with CiderDarling installed under `/opt/ciderdarling/<revision>` when using its installer. Avoid sharing prefix directories between different binaries and do not run multiple Darling versions against one prefix.

This release does **not** install Quadra, CiderDarling, Roblox, macOS frameworks, or third-party code. The user must supply an existing, executable Darling-compatible CLI. `--root` names the inspected sysroot; it does not force an arbitrary launcher to use it internally. If `--launcher` is a different build, supply its corresponding root explicitly.

## Commands

Dry run with the current Darling:

```sh
astelyn runtime-compare --app "$HOME/Games/Astelyn/RobloxPlayer.app" --json
```

Compare using regular Darling with a separate test prefix:

```sh
astelyn runtime-compare \
  --app "$HOME/Games/Astelyn/RobloxPlayer.app" \
  --launcher /usr/local/bin/darling \
  --root /usr/local/libexec/darling \
  --prefix "$HOME/.astelyn-darling-v034" \
  --baseline /tmp/astelyn-roblox-launch.json \
  --timeout 10s --execute --json > /tmp/astelyn-runtime-v034.json
```

For a CiderDarling installation, replace `--launcher`, `--root`, and `--prefix` with its actual installed paths. Use `find /opt/ciderdarling -type f -name darling` (if installed) to locate a candidate launcher; do not assume its path.

A different first missing dependency is not proof of usable frameworks, successful initialization or Roblox playability. A crash or other launch failure may also appear. Review third-party runtime licensing and build provenance independently.

Upstream references:
- https://github.com/QuadraHQ/Quadra
- https://github.com/QuadraHQ/CiderDarling

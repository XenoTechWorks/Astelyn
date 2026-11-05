# Darling bridge: guest execution

Astelyn v0.1.5 integrates a local Darling installation. It **does not run Roblox**.

## Verify your setup

```sh
go run ./cmd/astelyn darling status
go run ./cmd/astelyn darling test
```

`darling test` executes a Bash builtin inside Darling, checks for a known marker,
and reports an error if the process fails. On some minimal Darling installations
`/etc/profile: cp: command not found` may still appear during startup; this is
not fatal if the smoke test succeeds.

## Explicit guest paths

Linux `/bin/bash` is an ELF binary, but Darling's `/bin/bash` is Mach-O.
Use `--guest` to resolve the guest path against the Darling system root:

```sh
go run ./cmd/astelyn darling run --guest /bin/bash
go run ./cmd/astelyn darling run --execute --guest /bin/bash -- --noprofile --norc -c 'echo Astelyn'
```

The first command is a dry run. `--execute` is required to run anything.
Arguments after the executable are forwarded without a host shell.

For a host Mach-O path, Astelyn checks the binary and maps it to its guest path
if it resides under `/usr/local/libexec/darling/`. Other host paths cannot be
assumed visible inside the guest. To run an external program, make it available
in Darling's filesystem and use `--guest` to provide the guest-visible path.

This validation checks only file format, **not** trustworthiness, code signing,
or compatibility. Darling is not a security sandbox; do not run untrusted
executables. The runtime still lacks verified Roblox GUI/graphics support.

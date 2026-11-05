# Runtime preflight diagnostics

Astelyn's `preflight` checks the host and Darling installation without running any guest code:

```sh
go run ./cmd/astelyn preflight
go run ./cmd/astelyn preflight --json
```

You can inspect custom installations via `--root /path/to/root --launcher /path/to/darling`.
A successful exit means **only** that the launcher, system root, and a Mach-O-looking guest Bash are present on a Linux host. It does **not** establish that Darling can start, that the guest system is complete, that Apple frameworks are implemented, or that Roblox is playable.

Next test (explicitly executes guest code):

```sh
go run ./cmd/astelyn darling test
```

Only test known, trusted executables. Darling is not a security sandbox.

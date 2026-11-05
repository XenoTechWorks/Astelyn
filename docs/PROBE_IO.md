# Astelyn v0.1.5 — Probe I/O verification

The `probe` command can feed standard input into an installed Darling guest executable and capture its output and exit status. It is **not** a macOS framework implementation and does not establish Roblox compatibility.

## Real bzip2 round-trip

From the repository root, on a machine with Darling's Mach-O `/bin/bzip2` installed:

```sh
go test ./...
go build -o /tmp/astelyn ./cmd/astelyn
printf 'Astelyn v0.1.5 input\n' | /tmp/astelyn probe --stdin - /bin/bzip2 -- -c > /tmp/astelyn-probe-test.bz2
bzip2 -dc /tmp/astelyn-probe-test.bz2
```

The last command should print `Astelyn v0.1.5 input`.

Diagnostics go to standard error, leaving standard output safe for file redirection. The `--json` mode uses `stdout_base64` / `stderr_base64` for non-UTF-8 output, otherwise UTF-8 strings. It is **not** appropriate for writing compressed binary data directly to a `.bz2` file.

## Boundaries

The probe currently collects stdout and stderr in memory, so use small bounded test inputs. It has a timeout but not yet a maximum output-size limit. The default stdin is empty, unless you specify `--stdin -` or `--stdin PATH`.

Unit tests use a fake local launcher to verify input/output behavior without needing a macOS binary. Passing them does not prove that an actual Darling installation can execute a chosen binary.

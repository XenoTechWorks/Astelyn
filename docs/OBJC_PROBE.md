# Astelyn v0.2.4 — Objective-C runtime experiment

`astelyn objc-probe --json` examines installed Darling runtime *files*, and may inspect the Mach-O symbol table. A stripped symbol table does not imply a symbol is missing at runtime. It does not load or execute the libraries.

`astelyn objc-probe run --guest /absolute/guest/path` verifies an existing guest Mach-O file and defaults to dry-run. Add `--execute` to use the existing Darling runner. A successful smoke test requires process exit 0 and `ASTELYN_OBJC_OK` on stdout. This marker must come from a trusted compiled source; it is not a security or compatibility proof.

The sample at `experiments/objc-runtime/objc_smoke.c` calls Objective-C runtime functions `sel_registerName` and `objc_getClass` without requiring Foundation headers. Building it **requires a Darwin Mach-O cross-compiler and a compatible linker/runtime library**, not ordinary Linux clang. Build, link, and stage the resulting Mach-O in the Darling guest filesystem using your own tested toolchain; Astelyn does not automatically install toolchains, copy files or alter your guest root.

No AppKit/Cocoa window, Metal translation, Roblox execution, or anti-cheat compatibility is provided here. Never symlink Swift overlays to missing frameworks to make the resolver pass.

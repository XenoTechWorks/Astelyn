# Astelyn compatibility research

## Current state

Astelyn can inspect a ZIP containing macOS Mach-O binaries through its Go CLI:

- `inspect`: CPU architectures and directly linked shared libraries per executable
- `deps`: deduplicated direct-import inventory and referring binaries

These commands do not evaluate whether an Apple framework is implemented by Darling, and do not launch or patch Roblox. Transitive dependencies, dynamically loaded frameworks, ObjC class selectors, graphics paths, and sandbox/security policies remain unexamined.

## Next runtime experiment

1. Establish a reproducible Darling installation in a disposable test environment.
2. Run a legally redistributable, tiny macOS CLI executable and record logs.
3. If that succeeds, test a minimal GUI application separately.
4. Scope a Zig component only once its ABI boundary, input, and expected output are specified.

Do not use Roblox binaries as the first runtime test or make anti-cheat bypass a development goal.

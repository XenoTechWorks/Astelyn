# Astelyn

Experimental open-source research into running unmodified macOS desktop applications on Linux, with Roblox Player as a long-term target.

**Status:** Early prototype. Does not run macOS apps or Roblox yet.

## Quick start

```bash
python3 -m astelyn doctor
python3 -m astelyn doctor --json
python3 -m unittest discover -s tests -v
```

Astelyn's initial doctor is read-only. It inspects host tooling; it does not install or run binaries.

## Roadmap

1. Validate Linux prerequisites and investigate Darling.
2. Run a redistributable Mach-O command-line test program.
3. Prototype Cocoa window and input integration.
4. Research Metal-to-Vulkan requirements.
5. Investigate Roblox Player dependencies without modifying or bypassing security protections.

Astelyn does not distribute Apple or Roblox binaries and does not guarantee account safety.

## Inspect the macOS Roblox ZIP

```bash
python3 -m astelyn inspect-roblox ~/version-2fc79d515be6449d-RobloxPlayer.zip
```

See [research notes](docs/ROBLOX_RESEARCH.md).

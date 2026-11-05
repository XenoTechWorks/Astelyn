# Inspecting the macOS Roblox Player

Use an official macOS RobloxPlayer ZIP obtained separately. Astelyn does not
redistribute Roblox binaries.

From the repository root:

```bash
python3 -m astelyn inspect-roblox ~/version-2fc79d515be6449d-RobloxPlayer.zip
python3 -m astelyn inspect-roblox ~/version-2fc79d515be6449d-RobloxPlayer.zip --json
```

This command reads ZIP metadata and Mach-O headers in memory without
extracting or executing the app. It identifies CPU architectures for
candidate executables in `Contents/MacOS/`.

This is a reconnaissance tool, **not** a macOS loader, runtime, or anti-cheat bypass.
Framework dependency inspection and actual execution are future milestones.

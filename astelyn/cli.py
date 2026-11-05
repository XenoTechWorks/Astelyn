"""Read-only Linux prerequisite diagnostics."""
import argparse
import json
import platform
import shutil

TOOLS = ("darling", "vulkaninfo", "glxinfo", "wayland-info", "clang", "cmake")


def inspect_host(which=shutil.which, system=platform.system, machine=platform.machine):
    return {
        "platform": system(),
        "architecture": machine(),
        "tools": {tool: which(tool) for tool in TOOLS},
        "notes": [
            "Installed tools do not establish macOS compatibility.",
            "Astelyn cannot currently launch Roblox or macOS apps.",
        ],
    }


def main(argv=None):
    parser = argparse.ArgumentParser(prog="astelyn")
    commands = parser.add_subparsers(dest="command", required=True)
    doctor = commands.add_parser("doctor", help="Inspect host prerequisites")
    doctor.add_argument("--json", action="store_true")
    research = commands.add_parser("inspect-roblox", help="Inspect a macOS Roblox ZIP without extracting")
    research.add_argument("zip_path", help="Path to RobloxPlayer.zip")
    research.add_argument("--json", action="store_true")
    args = parser.parse_args(argv)
    if args.command == "inspect-roblox":
        from .research import print_report
        import sys
        import zipfile
        try:
            print_report(args.zip_path, args.json)
        except (OSError, ValueError, zipfile.BadZipFile) as exc:
            print(f"Astelyn: {exc}", file=sys.stderr)
            return 1
        return 0
    report = inspect_host()
    if args.json:
        print(json.dumps(report, indent=2, sort_keys=True))
    else:
        print(f"Astelyn host: {report['platform']} / {report['architecture']}")
        for tool, path in report["tools"].items():
            print(f"  {tool}: {path or 'not found'}")
        for note in report["notes"]:
            print(f"Note: {note}")
    return 0 if report["platform"] == "Linux" else 1

"""Read-only inspection of macOS Roblox ZIP bundles and Mach-O headers."""
import json
import struct
import zipfile
from pathlib import Path

ARCHES = {0x01000007: "x86_64", 0x0100000C: "arm64", 7: "i386", 12: "arm"}
THIN = {b"\xcf\xfa\xed\xfe": "<", b"\xfe\xed\xfa\xcf": ">"}
FAT = {b"\xca\xfe\xba\xbe": (">", False),
       b"\xbe\xba\xfe\xca": ("<", False),
       b"\xca\xfe\xba\xbf": (">", True),
       b"\xbf\xba\xfe\xca": ("<", True)}

def macho_architectures(data):
    """Return Mach-O CPU architectures, or [] for an unknown format."""
    magic = data[:4]
    if magic in THIN and len(data) >= 8:
        cpu = struct.unpack(THIN[magic] + "I", data[4:8])[0]
        return [ARCHES.get(cpu, hex(cpu))]
    if magic in FAT and len(data) >= 8:
        endian, is64 = FAT[magic]
        count = struct.unpack(endian + "I", data[4:8])[0]
        if count > 64:
            return []
        stride = 32 if is64 else 20
        result = []
        for i in range(count):
            offset = 8 + i * stride
            if len(data) < offset + 4:
                return []
            cpu = struct.unpack(endian + "I", data[offset:offset+4])[0]
            name = ARCHES.get(cpu, hex(cpu))
            if name not in result:
                result.append(name)
        return result
    return []

def inspect_zip(path):
    """Inspect ZIP metadata and small Mach-O headers without extracting."""
    path = Path(path)
    with zipfile.ZipFile(path) as archive:
        names = archive.namelist()
        executables = []
        for item in archive.infolist():
            name = item.filename
            if item.is_dir() or "/Contents/MacOS/" not in name:
                continue
            # Ignore nested resources and libraries; report likely app binaries.
            if name.endswith(".dylib") or name.endswith(".so"):
                continue
            if item.file_size < 8:
                continue
            with archive.open(item) as stream:
                header = stream.read(8 + 64 * 32)
            arch = macho_architectures(header)
            if arch:
                executables.append({"path": name, "architectures": arch,
                                    "size_bytes": item.file_size})
        return {"archive": str(path), "files": len(names),
                "executables": executables,
                "note": "Static inspection only; this does not run Roblox."}

def print_report(path, as_json=False):
    report = inspect_zip(path)
    if as_json:
        print(json.dumps(report, indent=2))
    else:
        print(f"Archive: {report['archive']} ({report['files']} entries)")
        for exe in report["executables"]:
            print(f"  {exe['path']}: {', '.join(exe['architectures'])} ({exe['size_bytes']} bytes)")
        if not report["executables"]:
            print("No Mach-O executables identified in Contents/MacOS.")
        print(report["note"])

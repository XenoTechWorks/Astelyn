import io
import struct
import tempfile
import unittest
import zipfile
from pathlib import Path
from contextlib import redirect_stdout
from astelyn.cli import main
from astelyn.research import macho_architectures, inspect_zip

class MachOTests(unittest.TestCase):
    def test_thin_arm64(self):
        self.assertEqual(macho_architectures(b"\xcf\xfa\xed\xfe" + struct.pack("<I", 0x0100000C)), ["arm64"])

    def test_fat_universal(self):
        data = b"\xca\xfe\xba\xbe" + struct.pack(">I", 2)
        data += struct.pack(">I", 0x01000007) + bytes(16)
        data += struct.pack(">I", 0x0100000C) + bytes(16)
        self.assertEqual(macho_architectures(data), ["x86_64", "arm64"])

    def test_zip_inspection(self):
        with tempfile.TemporaryDirectory() as d:
            path = Path(d) / "RobloxPlayer.zip"
            with zipfile.ZipFile(path, "w") as z:
                z.writestr("RobloxPlayer.app/Contents/MacOS/RobloxPlayer",
                           b"\xcf\xfa\xed\xfe" + struct.pack("<I", 0x01000007) + bytes(32))
            result = inspect_zip(path)
            self.assertEqual(result["executables"][0]["architectures"], ["x86_64"])
            with redirect_stdout(io.StringIO()):
                self.assertEqual(main(["inspect-roblox", str(path)]), 0)

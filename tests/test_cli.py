import io
import json
import unittest
from contextlib import redirect_stdout
from astelyn.cli import inspect_host, main


class DoctorTests(unittest.TestCase):
    def test_inspect(self):
        result = inspect_host(which=lambda _: None, system=lambda: "Linux", machine=lambda: "x86_64")
        self.assertEqual(result["architecture"], "x86_64")
        self.assertTrue(all(value is None for value in result["tools"].values()))

    def test_json(self):
        out = io.StringIO()
        with redirect_stdout(out):
            code = main(["doctor", "--json"])
        self.assertIn("tools", json.loads(out.getvalue()))
        self.assertIn(code, (0, 1))

import hashlib
import importlib.util
import json
import pathlib
import subprocess
import tempfile
import unittest

SCRIPT = pathlib.Path(__file__).with_name("verify_historical_baseline.py")
SPEC = importlib.util.spec_from_file_location("baseline_verifier", SCRIPT)
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)
verify = MODULE.verify


def git(root, *args, input_bytes=None):
    return subprocess.run(
        ["git", "-C", str(root), *args],
        input=input_bytes,
        check=True,
        stdout=subprocess.PIPE,
    ).stdout.decode().strip()


class HistoricalBaselineTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.root = pathlib.Path(self.temp.name)
        git(self.root, "init", "-q")
        git(self.root, "config", "user.email", "test@example.com")
        git(self.root, "config", "user.name", "Test")
        git(self.root, "config", "commit.gpgSign", "false")
        (self.root / ".gitleaks.toml").write_bytes(b"baseline\n")
        git(self.root, "add", ".gitleaks.toml")
        git(self.root, "commit", "-qm", "baseline")
        self.baseline = git(self.root, "rev-parse", "HEAD")
        (self.root / "later.txt").write_bytes(b"later\n")
        git(self.root, "add", "later.txt")
        git(self.root, "commit", "-qm", "later")
        self.manifest = self.root / "manifest.json"
        self.write_manifest()

    def tearDown(self):
        self.temp.cleanup()

    def write_manifest(self, **updates):
        manifest = {
            "schema_version": 1,
            "commit": self.baseline,
            "files": {
                ".gitleaks.toml": hashlib.sha256(b"baseline\n").hexdigest(),
            },
        }
        manifest.update(updates)
        self.manifest.write_text(json.dumps(manifest), encoding="utf-8")

    def test_accepts_ancestor_with_exact_raw_bytes(self):
        self.assertEqual(verify(self.root, self.manifest), [])

    def test_rejects_missing_or_tampered_baseline(self):
        self.write_manifest(commit="f" * 40)
        self.assertIn("baseline commit is missing", verify(self.root, self.manifest))
        self.write_manifest(files={".gitleaks.toml": "0" * 64})
        self.assertIn(
            "raw-byte hash mismatch: .gitleaks.toml",
            verify(self.root, self.manifest),
        )

    def test_rejects_non_ancestor(self):
        git(self.root, "checkout", "--orphan", "unrelated")
        git(self.root, "rm", "-q", "-rf", ".")
        (self.root / "unrelated.txt").write_text("unrelated\n", encoding="utf-8")
        git(self.root, "add", "unrelated.txt")
        git(self.root, "commit", "-qm", "unrelated")
        self.assertIn(
            "baseline commit is not an ancestor of HEAD",
            verify(self.root, self.manifest),
        )


if __name__ == "__main__":
    unittest.main()

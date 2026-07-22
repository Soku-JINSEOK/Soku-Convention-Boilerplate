"""Tests for the public control-plane provider action contract."""

from __future__ import annotations

import importlib.util
import tempfile
import unittest
from pathlib import Path


MODULE_PATH = Path(__file__).with_name("validate-config.py")
SPEC = importlib.util.spec_from_file_location("validate_config", MODULE_PATH)
if SPEC is None or SPEC.loader is None:
    raise RuntimeError("unable to load public provider validator")
VALIDATOR = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(VALIDATOR)

VALID_REF = "0123456789abcdef0123456789abcdef01234567"


class ValidateConfigTest(unittest.TestCase):
    """Prove that the public action exposes only reviewed portable values."""

    def setUp(self) -> None:
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.workspace = Path(self.temporary.name)
        source = MODULE_PATH.parents[2] / "providers" / "ci-cd-control-plane-v1"
        self.configuration = self.workspace / ".github" / "soku" / "provider.yml"
        self.configuration.parent.mkdir(parents=True)
        self.configuration.write_bytes((source / "example-config.yml").read_bytes())

    def validate(self, **overrides: str) -> dict[str, str]:
        values = {
            "configuration_path": ".github/soku/provider.yml",
            "integration_ref": VALID_REF,
            "integration_source": VALIDATOR.EXPECTED_SOURCE,
        }
        values.update(overrides)
        return VALIDATOR.validate_contract(self.workspace, **values)

    def test_reviewed_configuration_returns_only_cli_arguments(self) -> None:
        self.assertEqual(
            self.validate(),
            {
                "integration-source": VALIDATOR.EXPECTED_SOURCE,
                "integration-ref": VALID_REF,
                "integration-config": ".github/soku/provider.yml",
            },
        )

    def test_rejects_mutable_or_uppercase_ref(self) -> None:
        for value in ("main", VALID_REF.upper()):
            with self.subTest(value=value), self.assertRaisesRegex(
                ValueError, "lowercase full commit SHA"
            ):
                self.validate(integration_ref=value)

    def test_rejects_unsafe_configuration_paths(self) -> None:
        for value in ("../provider.yml", "/tmp/provider.yml", "provider.json"):
            with self.subTest(value=value), self.assertRaisesRegex(
                ValueError, "safe relative YAML path"
            ):
                self.validate(configuration_path=value)

    def test_rejects_literal_byte_tampering(self) -> None:
        self.configuration.write_bytes(self.configuration.read_bytes() + b"# changed\n")
        with self.assertRaisesRegex(ValueError, "reviewed contract"):
            self.validate()

    def test_rejects_non_utf8_configuration(self) -> None:
        self.configuration.write_bytes(b"\xff\xfe")
        with self.assertRaisesRegex(ValueError, "UTF-8"):
            self.validate()

    def test_rejects_unknown_source(self) -> None:
        with self.assertRaisesRegex(ValueError, "public bundle source"):
            self.validate(integration_source="github:example/private-provider")

    def test_rejects_symlinked_configuration(self) -> None:
        target = self.workspace / "target.yml"
        target.write_bytes(self.configuration.read_bytes())
        self.configuration.unlink()
        self.configuration.symlink_to(target)
        with self.assertRaisesRegex(ValueError, "non-symlink regular file"):
            self.validate()


if __name__ == "__main__":
    unittest.main()

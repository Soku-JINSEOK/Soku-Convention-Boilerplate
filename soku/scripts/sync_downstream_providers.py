#!/usr/bin/env python3
"""Mirror reviewed control-plane Provider API v1 bundles into public paths."""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path
from typing import Any

CENTRAL_MANIFEST = Path("integrations/soku/provider-api-v1/downstream-manifest-v1.json")
PUBLIC_ROOT = Path("soku/providers")
PROVENANCE = PUBLIC_ROOT / "provenance/registered-downstream-v1.json"


def load_json(path: Path) -> dict[str, Any]:
    """Load a JSON object."""
    value = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(value, dict):
        raise ValueError(f"{path} must contain an object")
    return value


def digest(data: bytes) -> str:
    """Return raw-byte SHA-256."""
    return hashlib.sha256(data).hexdigest()


def json_bytes(value: Any) -> bytes:
    """Serialize stable JSON."""
    return (json.dumps(value, ensure_ascii=False, indent=2) + "\n").encode()


def expected(source_root: Path, commit: str) -> dict[Path, bytes]:
    """Return exact public bytes and their provenance ledger."""
    manifest_path = source_root / CENTRAL_MANIFEST
    manifest_bytes = manifest_path.read_bytes()
    manifest = load_json(manifest_path)
    output: dict[Path, bytes] = {}
    bundles = []
    for bundle in manifest["bundles"]:
        provider_id = bundle["provider_id"]
        public_files = []
        for entry in bundle["files"]:
            central_path = Path(entry["path"])
            data = (source_root / central_path).read_bytes()
            if digest(data) != entry["sha256"]:
                raise ValueError(f"central byte hash mismatch: {central_path}")
            relative = central_path.relative_to(bundle["source_path"])
            public_path = PUBLIC_ROOT / provider_id / relative
            if relative.as_posix() == "provider-v1.json":
                metadata = json.loads(data)
                metadata["source"] = (
                    "github:Soku-JINSEOK/Soku-Convention-Boilerplate/"
                    f"soku/providers/{provider_id}"
                )
                data = json_bytes(metadata)
            output[public_path] = data
            public_files.append(
                {
                    "path": public_path.as_posix(),
                    "sha256": digest(data),
                    "central_sha256": entry["sha256"],
                }
            )
        bundles.append(
            {
                "project_id": bundle["project_id"],
                "provider_id": provider_id,
                "files": public_files,
            }
        )
    ledger = {
        "schema_version": 1,
        "hash_algorithm": "sha256-raw-bytes",
        "control_plane": {
            "repository": "https://github.com/Soku-JINSEOK/ci-cd-control-plane",
            "merge_commit": commit,
            "manifest_path": CENTRAL_MANIFEST.as_posix(),
            "manifest_sha256": digest(manifest_bytes),
        },
        "source_rewrite": "provider-v1.json:source-only",
        "bundles": bundles,
        "delivery_enabled": False,
    }
    output[PROVENANCE] = json_bytes(ledger)
    return output


def sync(repository: Path, source_root: Path, commit: str, check: bool) -> None:
    """Write the public mirror or fail if committed bytes differ."""
    mismatches = []
    for relative, data in expected(source_root, commit).items():
        target = repository / relative
        if check:
            if not target.is_file() or target.read_bytes() != data:
                mismatches.append(relative.as_posix())
        else:
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_bytes(data)
    if mismatches:
        raise ValueError("public provider mirror drift: " + ", ".join(mismatches))


def main() -> None:
    """Parse arguments and synchronize reviewed provider bytes."""
    parser = argparse.ArgumentParser()
    parser.add_argument("--source-root", type=Path, required=True)
    parser.add_argument("--control-plane-commit", required=True)
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    if len(args.control_plane_commit) != 40 or any(
        character not in "0123456789abcdef"
        for character in args.control_plane_commit
    ):
        raise ValueError("control-plane commit must be a lowercase full SHA")
    sync(
        Path(__file__).resolve().parents[2],
        args.source_root.resolve(),
        args.control_plane_commit,
        args.check,
    )
    action = "verified" if args.check else "synchronized"
    print(f"Registered downstream provider mirror {action}")


if __name__ == "__main__":
    main()

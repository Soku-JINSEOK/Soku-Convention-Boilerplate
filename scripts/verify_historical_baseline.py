#!/usr/bin/env python3
"""Verify that the reviewed historical security baseline is immutable."""

import hashlib
import json
import pathlib
import subprocess
import sys


def git(root, *arguments):
    return subprocess.run(
        ["git", "-C", str(root), *arguments],
        check=False,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )


def verify(root, manifest_path):
    errors = []
    try:
        manifest_bytes = pathlib.Path(manifest_path).read_bytes()
        manifest = json.loads(manifest_bytes)
    except (OSError, json.JSONDecodeError) as error:
        return [f"baseline manifest is unreadable: {error}"]

    if manifest.get("schema_version") != 1:
        errors.append("baseline schema_version must be 1")
    commit = manifest.get("commit", "")
    if not isinstance(commit, str) or len(commit) != 40:
        return errors + ["baseline commit must be a full SHA"]
    exists = git(root, "cat-file", "-e", f"{commit}^{{commit}}")
    if exists.returncode != 0:
        errors.append("baseline commit is missing")
        return errors
    ancestor = git(root, "merge-base", "--is-ancestor", commit, "HEAD")
    if ancestor.returncode != 0:
        errors.append("baseline commit is not an ancestor of HEAD")

    files = manifest.get("files")
    if not isinstance(files, dict) or not files:
        errors.append("baseline files must be a non-empty object")
        return errors
    for path, expected_hash in sorted(files.items()):
        if (
            not isinstance(path, str)
            or path.startswith("/")
            or ".." in pathlib.PurePosixPath(path).parts
        ):
            errors.append(f"unsafe baseline path: {path!r}")
            continue
        baseline_file = git(root, "show", f"{commit}:{path}")
        if baseline_file.returncode != 0:
            errors.append(f"baseline file is missing: {path}")
            continue
        actual_hash = hashlib.sha256(baseline_file.stdout).hexdigest()
        if actual_hash != expected_hash:
            errors.append(f"raw-byte hash mismatch: {path}")
    return errors


def main():
    if len(sys.argv) not in (1, 3):
        print(
            "usage: verify_historical_baseline.py [<repository> <manifest>]",
            file=sys.stderr,
        )
        return 2
    if len(sys.argv) == 3:
        root, manifest = pathlib.Path(sys.argv[1]), pathlib.Path(sys.argv[2])
    else:
        root = pathlib.Path(__file__).resolve().parent.parent
        manifest = root / "security/historical-baseline.json"
    errors = verify(root, manifest)
    if errors:
        for error in errors:
            print(f"error: {error}", file=sys.stderr)
        return 1
    print("Historical security baseline ancestry and raw bytes verified.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

#!/usr/bin/env python3
"""Validate the public provider configuration and emit bounded action outputs."""

from __future__ import annotations

import hashlib
import os
import re
from pathlib import Path, PurePosixPath

EXPECTED_CONFIGURATION_HASH = (
    "56aa1ba994c95e320da869998ffb64ce634fdcfb0d457c03a02f770ef4f3bf4d"
)
EXPECTED_SOURCE = (
    "github:Soku-JINSEOK/Soku-Convention-Boilerplate/"
    "soku/providers/ci-cd-control-plane-v1"
)
LOWER_COMMIT = re.compile(r"^[0-9a-f]{40}$")
SAFE_PATH = re.compile(r"^[A-Za-z0-9._/-]+$")


def normalized_sha256(data: bytes) -> str:
    """Hash bounded UTF-8 after soku-compatible newline normalization."""
    text = data.decode("utf-8")
    normalized = text.replace("\r\n", "\n").replace("\r", "\n")
    return hashlib.sha256(normalized.encode()).hexdigest()


def resolve_configuration(workspace: Path, value: str) -> tuple[Path, str]:
    """Resolve a safe regular YAML file inside the caller workspace."""
    if (
        not value
        or not SAFE_PATH.fullmatch(value)
        or "\\" in value
        or PurePosixPath(value).is_absolute()
        or ".." in PurePosixPath(value).parts
        or PurePosixPath(value).suffix.lower() not in {".yml", ".yaml"}
    ):
        raise ValueError("configuration path must be a safe relative YAML path")
    candidate = workspace / Path(*PurePosixPath(value).parts)
    if candidate.is_symlink() or not candidate.is_file():
        raise ValueError("configuration path must be a non-symlink regular file")
    resolved_workspace = workspace.resolve()
    resolved_candidate = candidate.resolve()
    if resolved_workspace not in resolved_candidate.parents:
        raise ValueError("configuration path escapes the workspace")
    return resolved_candidate, PurePosixPath(value).as_posix()


def validate_contract(
    workspace: Path,
    configuration_path: str,
    integration_ref: str,
    integration_source: str,
) -> dict[str, str]:
    """Validate exact configuration bytes, source, and full action ref."""
    if not LOWER_COMMIT.fullmatch(integration_ref):
        raise ValueError("integration ref must be a lowercase full commit SHA")
    if integration_source != EXPECTED_SOURCE:
        raise ValueError("integration source must equal the public bundle source")
    path, portable_path = resolve_configuration(workspace, configuration_path)
    data = path.read_bytes()
    if not data or len(data) > 64 * 1024:
        raise ValueError("configuration must be non-empty and no larger than 64 KiB")
    try:
        digest = normalized_sha256(data)
    except UnicodeDecodeError as error:
        raise ValueError("configuration must be UTF-8") from error
    if digest != EXPECTED_CONFIGURATION_HASH:
        raise ValueError("configuration bytes do not match the reviewed contract")
    return {
        "integration-source": integration_source,
        "integration-ref": integration_ref,
        "integration-config": portable_path,
    }


def append_outputs(path: Path, outputs: dict[str, str]) -> None:
    """Append validated single-line values to the GitHub output file."""
    with path.open("a", encoding="utf-8", newline="\n") as output:
        for key in (
            "integration-source",
            "integration-ref",
            "integration-config",
        ):
            output.write(f"{key}={outputs[key]}\n")


def main() -> None:
    """Validate action environment and publish outputs."""
    output_path = os.environ.get("GITHUB_OUTPUT", "")
    if not output_path:
        raise ValueError("GITHUB_OUTPUT is required")
    outputs = validate_contract(
        Path(os.environ.get("GITHUB_WORKSPACE", "")),
        os.environ.get("CONFIGURATION_PATH", ""),
        os.environ.get("INTEGRATION_REF", ""),
        os.environ.get("INTEGRATION_SOURCE", ""),
    )
    append_outputs(Path(output_path), outputs)


if __name__ == "__main__":
    main()

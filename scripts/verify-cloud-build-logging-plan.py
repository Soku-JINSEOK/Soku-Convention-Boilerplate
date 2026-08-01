#!/usr/bin/env python3
"""Fail closed unless a Terraform plan creates only the logging boundary."""

import json
import pathlib
import sys

EXPECTED = {
    "google_logging_project_bucket_config.cloud_build_validation": {
        "type": "google_logging_project_bucket_config",
        "checks": {
            "bucket_id": "cloud-build-validation",
            "location": "asia-northeast1",
            "retention_days": 30,
        },
    },
    "google_logging_project_sink.cloud_build_validation": {
        "type": "google_logging_project_sink",
        "checks": {
            "filter": 'resource.type="build"',
            "name": "cloud-build-validation",
            "unique_writer_identity": False,
        },
    },
    "google_logging_project_exclusion.default_disabled": {
        "type": "google_logging_project_exclusion",
        "checks": {
            "disabled": True,
            "filter": 'resource.type="build"',
            "name": "_Default",
        },
    },
}


def verify_plan(plan):
    errors = []
    changes = plan.get("resource_changes")
    if not isinstance(changes, list):
        return ["resource_changes must be a list"]

    observed = {}
    for change in changes:
        address = change.get("address", "")
        if address in observed:
            errors.append(f"duplicate resource change: {address}")
            continue
        observed[address] = change

    unexpected = sorted(set(observed) - set(EXPECTED))
    missing = sorted(set(EXPECTED) - set(observed))
    if unexpected:
        errors.append(f"unexpected resources: {', '.join(unexpected)}")
    if missing:
        errors.append(f"missing resources: {', '.join(missing)}")

    for address, expected in EXPECTED.items():
        change = observed.get(address)
        if not change:
            continue
        if change.get("type") != expected["type"]:
            errors.append(f"{address}: unexpected resource type")
        actions = change.get("change", {}).get("actions")
        if actions != ["create"]:
            errors.append(f"{address}: actions must be exactly create")
        after = change.get("change", {}).get("after") or {}
        for key, value in expected["checks"].items():
            if after.get(key) != value:
                errors.append(f"{address}: {key} must be {value!r}")
        rendered = json.dumps(after, sort_keys=True)
        if "iam" in address.lower() or "iam" in rendered.lower():
            errors.append(f"{address}: IAM is forbidden")
        if after.get("name") == "_Required" or "_Required" in rendered:
            errors.append(f"{address}: _Required is forbidden")

    bucket = observed.get("google_logging_project_bucket_config.cloud_build_validation", {})
    sink = observed.get("google_logging_project_sink.cloud_build_validation", {})
    exclusion = observed.get("google_logging_project_exclusion.default_disabled", {})
    bucket_after = bucket.get("change", {}).get("after") or {}
    sink_after = sink.get("change", {}).get("after") or {}
    exclusion_after = exclusion.get("change", {}).get("after") or {}
    project = bucket_after.get("project")
    if not isinstance(project, str) or not project:
        errors.append("logging resources require a known non-empty project")
    else:
        for address, after in (("sink", sink_after), ("exclusion", exclusion_after)):
            if after.get("project") != project:
                errors.append(f"{address}: project must match the bucket project")
        expected_destination = (
            "logging.googleapis.com/projects/"
            f"{project}/locations/asia-northeast1/buckets/cloud-build-validation"
        )
        if sink_after.get("destination") != expected_destination:
            errors.append("sink: destination must be the dedicated regional bucket")

    return errors


def main():
    if len(sys.argv) != 2:
        print("usage: verify-cloud-build-logging-plan.py <plan.json>", file=sys.stderr)
        return 2
    path = pathlib.Path(sys.argv[1])
    try:
        plan = json.loads(path.read_bytes())
    except (OSError, json.JSONDecodeError) as error:
        print(f"invalid plan JSON: {error}", file=sys.stderr)
        return 2
    errors = verify_plan(plan)
    if errors:
        for error in errors:
            print(f"error: {error}", file=sys.stderr)
        return 1
    print("Cloud Build logging plan accepted: 3 create, 0 update, 0 delete.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

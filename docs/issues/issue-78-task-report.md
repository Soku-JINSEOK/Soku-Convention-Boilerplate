# Issue #78 Task Report — Registered Provider Public Mirror

## Scope

Mirror the four downstream Provider API v1 bundles reviewed in control-plane
Issue #47 into stable public Boilerplate paths. The central source is squash
merge `ea24298d8081d8108f3b5d280a9e401c6b54df47`.

## Delivered Contract

- CutVi, archviz, report-hub, and SOKU-PR-site each receive a distinct public
  provider path and exact configuration schema.
- The synchronization tool verifies every central raw-byte hash before copying
  and rewrites only `provider-v1.json`'s source to the public repository path.
- A new provenance ledger records the central manifest hash, merge SHA, source
  rewrite rule, central hashes, and public raw-byte hashes.
- Generic lifecycle tests connect every bundle through the production loader
  and prove a cross-project configuration cannot connect.
- Existing providers, callers, catalogs, CLI behavior, and releases remain
  unchanged. Every literal configuration keeps delivery disabled.

## Verification

- [x] Central-to-public synchronization and repeatable `--check` pass.
- [x] Targeted public provider and lifecycle tests pass.
- [x] Go test/race/vet, lifecycle, package reproducibility, Node governance,
  Python action, Markdown/YAML, and whitespace checks pass locally.
- [x] Hosted Linux/macOS/Windows, runtime-template, package, security, CodeQL,
  governance, and aggregate Validation Gate checks pass.

## Security and Delivery Boundary

The bundles are declarative data. This work adds no credentials, permissions,
private Actions access, remote execution, caller activation, cloud resources,
delivery, deployment, tag, or Release.

## AI Assistance

- **Planning, implementation, tests, and drafting:** OpenAI Codex

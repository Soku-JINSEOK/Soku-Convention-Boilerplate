# Issue 163 implementation amendment

This note records the execution scope authorized by the current task. It
supersedes the earlier task-report sequencing where that report separated the
decision contract, planner, and installer into three independent changes.

The implementation is intentionally delivered in three reviewable pull
requests:

1. PR A owns the strict `ci-cd-decision-v1` contract, deterministic
   `soku ci-cd plan`, and the parser/help-only `soku ci-cd init` surface. The
   init mutation remains unavailable until the later transaction change.
2. PR B owns the immutable adapter provenance, the three validation-only
   mappings, trusted renderers, and semantic conformance gates. It closes
   Issue 198.
3. PR C owns transactional `soku ci-cd init` and closes Issue 163 only after
   the complete downstream mapping and lifecycle evidence is available.

All three changes are CI-only. They do not enable delivery, publication,
release, deployment, Cloud triggers, IAM, credentials, runners, billing, or
paid capacity. Planning is read-only and does not retain repository names,
absolute paths, timestamps, or machine inventory.

PR B records the public engine merge and byte hashes for the versioned schemas,
Local/GCP/Jenkins synthetic fixtures, and adapter descriptors. Its catalog is
strictly limited to `ci-only + github-hosted`, `ci-only + gcp-managed`, and
`ci-only + github-self-hosted`; each renderer is validation-only and binds
`delivery_authority: none`.

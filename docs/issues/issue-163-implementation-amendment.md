# Issue 163 implementation amendment

This note records the execution scope authorized by the current task. It
supersedes the earlier task-report sequencing where that report separated the
decision contract, planner, and installer into three independent changes.

The implementation is intentionally delivered in three reviewable pull
requests:

1. PR A (#212) owns the strict `ci-cd-decision-v1` contract, deterministic
   `soku ci-cd plan`, and the parser/help-only `soku ci-cd init` surface.
2. PR B owns the immutable adapter provenance, the three validation-only
   mappings, trusted renderers, and semantic conformance gates. It is PR #214
   and closes Issue 198 after current-head validation.
3. PR C (#215) owns transactional `soku ci-cd init` and closes Issue 163 only
   after the complete downstream mapping and lifecycle evidence is available.

All three changes are CI-only. They do not enable delivery, publication,
release, deployment, Cloud triggers, IAM, credentials, runners, billing, or
paid capacity. Planning is read-only and does not retain repository names,
absolute paths, timestamps, or machine inventory.

PR B records the public engine merge and byte hashes for the versioned schemas,
Local/GCP/Jenkins synthetic fixtures, and adapter descriptors. Its catalog is
strictly limited to `ci-only + github-hosted`, `ci-only + gcp-managed`, and
`ci-only + github-self-hosted`; each renderer is validation-only and binds
`delivery_authority: none`.

PR C completes the installer without adding a manifest schema version. It
accepts exactly one of `--dry-run` and `--yes`, requires an existing manifest,
recomputes the plan, worktree identity, catalog, renderer, and manifest before
writing, and rechecks those identities at the write boundary. It reuses the
shared backup, journal, atomic path write, manifest-last, rollback, and recovery
transaction. Manifest v1 migrates to v2; v2 and v3 remain at their existing
versions. The existing component and managed-file records carry the selected
configuration path, mapping binding, renderer hash, and caller baseline hash;
no raw configuration, command output, inventory, endpoint, credential, or
absolute path is stored. A root `.gitattributes` rule preserves the exact
hash-locked conformance bytes on Windows checkouts.

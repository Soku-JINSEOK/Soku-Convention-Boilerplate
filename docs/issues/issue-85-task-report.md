# Issue #85 Task Report — Dependabot Schema Repair

## Scope

Repair the invalid major-version ignore filters that make Dependabot reject all
four configured package ecosystems. Preserve the update schedule, grouping,
security update declarations, ownership metadata, and labels.

## Implementation

- Use Dependabot's documented `version-update:semver-major` enum value.
- Add a focused regression test that checks all four entries and rejects the
  invalid legacy value.
- Run the regression in the existing repository governance test job.

## Acceptance

- [x] All four configured ecosystems use a schema-valid major update filter.
- [x] Existing schedules, groups, assignees, labels, and security update
  declarations remain unchanged.
- [x] The focused Node regression, YAML parsing, and whitespace checks pass.
- [ ] Hosted validation passes after the current Actions account gate clears.

## Security and Approval Boundary

This change does not modify workflow permissions, repository settings,
credentials, delivery, deployment, releases, or cloud resources. The pull
request remains Draft and uses a non-closing Issue reference.

## AI Assistance

- **Audit, implementation, tests, and drafting:** OpenAI Codex

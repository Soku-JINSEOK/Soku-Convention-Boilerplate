# July 2026 Dependency Updates (Issue #69)

## Outcome

This report records the three dependency updates reviewed under
[Issue #69](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/69).
The updates are intentionally merged in order so every successor is validated
against the dependency state already accepted on `main`.

## Dependency Ledger

| PR | Dependency | Change | Upstream evidence | Repository evidence | Status |
| --- | --- | --- | --- | --- | --- |
| #60 | `actions/checkout` | 7.0.0 → 7.0.1 | Release `v7.0.1`; upstream commit `3d3c42e5aac5ba805825da76410c181273ba90b1` | All governed workflow references use the full release commit SHA; merged as `d9031926628130e3956f668db2859994318dc015` | Merged |
| #61 | `platformdirs` | 4.10.0 → 4.10.1 | Release `4.10.1`; upstream commit `2e740132ddbe62a3ff4030653bf41b54902949e1` | `templates/python/requirements-lock.txt` contains `platformdirs==4.10.1` | In review |
| #66 | `ruff` | 0.15.21 → 0.15.22 | Release `0.15.22`; upstream commit `0177a7e0d2c4a3805afa1960f106d72878766cbc` | `.github/workflows/templates-ci.template.yml` will retain the exact reviewed version change | Pending sequential rebase |

The original Dependabot release notes, changelog excerpts, and commit links are
preserved in each pull-request body.

## Verification

- [x] PR #60 passed hosted repository, template, dependency/license, and
  security validation before merge.
- [x] Every `actions/checkout` reference changed by PR #60 is pinned to full
  SHA `3d3c42e5aac5ba805825da76410c181273ba90b1`.
- [ ] PR #61 passes repository, Python template, dependency/license, and
  security validation on top of PR #60.
- [ ] PR #66 passes the same governed validation on top of PR #61.
- [ ] Issue #69 is closed only by the final PR after all three entries are
  complete and the Project item is `Done`.

## Security Boundary

These changes update declared dependencies only. They do not enable delivery,
change credentials, mutate cloud resources, or execute unreviewed remote
provider code. Workflow actions remain pinned to immutable full commit SHAs.

## Notes

PR #60 was merged before this report file was added. PR #61 therefore carries
the corrective shared report onto `main`; the report still preserves the exact
PR #60 merge and upstream pin evidence rather than presenting it as pending.

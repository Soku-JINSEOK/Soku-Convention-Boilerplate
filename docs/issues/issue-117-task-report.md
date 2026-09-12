# Issue 117 — preserve code validation evidence

## Current review scope

The owner requested verification of actual GitHub incorporation and correction
of discovered defects. This partial source correction supports Issue #117; it
does not perform its branch-protection or Hosted Full rollout.

## Finding and correction

Code run 34682376195 on PR #229 failed while later metadata-only runs skipped
all code groups and reported success with the same required check names.
Concurrency isolation did not prevent that overwrite. Both aggregate job names
now depend on the event: code events retain `CI Quick Gate` and
`Validation Gate`, while metadata events use `CI Quick Metadata Only` and
`Validation Metadata Only`. Names do not depend on job success or cancellation.

## Validation

Regression tests evaluate the actual workflow expressions for code and metadata
events, including a base edit. They execute the actual aggregate shell and
require failure for failed, cancelled or unexpectedly skipped code groups.
No extra job, trigger, permission or heavy validation execution is added.

The current security workflow, its trusted base/head separation and immutable
history baseline remain unchanged. PR Metadata Gate and repository required
contexts are unchanged. Skipping a required job is not used as a fix because
GitHub treats skipped checks as passing.

## Remaining Issue 117 scope

PR #158 still requires a current-main forward-port of its Hosted Full caller:
pass trusted `base-sha` and candidate `head-sha` to the current security workflow,
retain full-history/baseline guards and permissions, then validate the complete
caller contract. The comparison period, source acceptance and branch-protection
rollout are not claimed complete. PR #159/#179 and the broader #160 integration
remain separately reviewable pipeline work.

## AI assistance

- Provider: OpenAI
- Model: Codex (exact backend model identifier not exposed)
- Usage Summary: GitHub job review, false-green reproduction, narrow workflow
  correction and regression verification.

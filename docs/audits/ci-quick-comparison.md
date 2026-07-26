# CI Quick Comparison

## Purpose

This audit is the authoritative comparison record for Issue #116. It measures
the sharded `CI Quick Gate` against the hosted full validation path without
weakening the existing required checks during observation.

## Activation

- **Implementation:** detector-driven dynamic matrix from
  `verification/profiles.yml`
- **Activation pull request:** [#149](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/149)
- **Activation commit:** `9539413a08e545a2e0d91383709cacc7b8a385de`
- **Activation time:** `2026-07-26T14:13:58Z`
- **Earliest time-based completion:** `2026-08-09T14:13:58Z`
- **Minimum sample:** 10 merged code pull requests after activation

The earlier window that began at `2026-07-26T12:09:01Z` measured the serial
Quick implementation and is retained only as historical context. It does not
count toward this audit. Any later change to Quick behavior or coverage resets
both the time and pull-request counters.

## Inclusion Rules

Include merged code pull requests and Dependabot pull requests whose head commit
ran both Quick and Full validation after activation. Exclude metadata-only
events, synthetic observation pull requests, runs before activation, and runs
whose workflow implementation differs from the active comparison version.

## Completion Criteria

All of the following must hold for the complete sample:

- at least 14 elapsed days and 10 qualifying merged pull requests
- zero Quick-pass / Full-fail misses
- zero unresolved flaky results
- every scope-detector and shard-planner fixture passes
- median Quick critical duration is at most 50% of Full critical duration
- aggregate Quick runner-seconds are at least 40% lower than Full

## Samples

| PR | Merge commit | Quick run | Full run | Quick critical duration | Full critical duration | Quick runner-seconds | Full runner-seconds | Result |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- |
| Pending | Pending | Pending | Pending | Pending | Pending | Pending | Pending | Pending |

## Decision

**Observation active.** The sharded implementation merged in PR #149 at the
activation time above. Issue #117 must not change required contexts until every
criterion above is supported by linked Actions evidence.

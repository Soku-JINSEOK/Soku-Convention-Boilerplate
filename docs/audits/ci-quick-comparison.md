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

## Measurement Method

Use the GitHub Actions Jobs API for the final successful Validation run on each
merged head commit.

- Quick jobs are successful jobs named `Quick validation / ...` plus
  `CI Quick Gate`.
- Full jobs are successful jobs named `Full repository validation / ...`,
  `Full runtime-template validation / ...`, or `Security validation / ...`,
  plus `Validation Gate`.
- Critical duration is the interval from the earliest selected job start to
  the latest selected job completion.
- Runner-seconds are the sum of selected job durations.
- Critical ratio is Quick critical duration divided by Full critical duration.
- Runner reduction is one minus Quick runner-seconds divided by Full
  runner-seconds.

## Samples

| PR | Merge commit | Quick run | Full run | Quick critical duration | Full critical duration | Quick runner-seconds | Full runner-seconds | Result |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- |
| [#142](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/142) | `f483e56e9c10ebfb9caa4bf0b0c43ec595282aca` | [CI Quick Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30207657214/job/89808575715) | [Validation Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30207657214/job/89808585847) | 101 s | 109 s | 299 s | 894 s | Pass; 92.66% critical ratio; 66.55% runner reduction |
| [#152](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/152) | `67cbaf5456fbf3c069be2ddda01566531ca59a25` | [CI Quick Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30208730766/job/89811298032) | [Validation Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30208730766/job/89811404736) | 55 s | 108 s | 16 s | 879 s | Pass; 50.93% critical ratio; 98.18% runner reduction |

Current qualifying sample count: **2 of 10**.

## Decision

**Observation active.** The sharded implementation merged in PR #149 at the
activation time above. Issue #117 must not change required contexts until every
criterion above is supported by linked Actions evidence.

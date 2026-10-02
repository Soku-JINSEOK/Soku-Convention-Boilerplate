# CI Quick Comparison

> **Document purpose:** Comparison evidence. Records the Quick-versus-Full observation criteria and measurements for Issue #116.
>
> **Key point:** Required full checks remain until the documented transition criteria and ruleset decision are satisfied.

## Purpose

This audit is the authoritative comparison record for Issue #116. It measures
the sharded `CI Quick Gate` against the hosted full validation path without
weakening the existing required checks during observation.

## Previous Activation

- **Implementation:** automatic Quick, Full, and trusted Security validation
  restored by
  [#199](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/199)
- **Activation pull request:**
  [#199](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/199)
- **Activation commit:** `44a4106cab9e51f1568665e27d022a4d293e89c2`
- **Activation time:** `2026-08-08T02:16:45Z`
- **Earliest time-based completion:** `2026-08-22T02:16:45Z`
- **Minimum sample:** 10 merged code pull requests after activation

The sharded window that began at `2026-07-26T14:13:58Z` and the restarted
window after PR #176 are historical only. PR #181 removed automatic Quick and
Full validation from pull request and `main` events, invalidating those
observation windows. PR #199 restored the automatic caller and its merge commit
started the previous observation window at the timestamp above. PR #199 and this
activation-record pull request are excluded from the sample. Any later change
to Quick or Full behavior or coverage resets both counters again.

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

The `Validation metrics` completion workflow collects each Validation run
attempt, including failures and cancellations, using read-only Actions access.
It executes the default branch collector, never measured PR code. JSON
artifacts are named by run ID and attempt. Their retention is declared in
`.github/workflows/validation-metrics.yml`; export qualifying evidence before
expiry. Collection becomes automatic after that workflow reaches the default
branch. For an existing attempt:

```bash
node scripts/collect-validation-metrics.mjs OWNER/REPO RUN_ID ATTEMPT > metrics.json
```

- Quick jobs are named `Quick validation / ...` plus `CI Quick Gate`.
- Full jobs are named `Full repository validation / ...`,
  `Full runtime-template validation / ...`, or `Security validation / ...`,
  plus `Validation Gate`.
- Critical duration spans the earliest selected job start through the latest
  completion, including intervening waits. Runner-seconds sum job elapsed times;
  they are an execution-cost proxy, not a billing measurement.
- Reports retain job counts, conclusions, timestamps, run SHA, and attempt.
  Missing gates or cancellations are incomplete observations, not passes.
- Compare outcomes across all attempts before selecting timing samples. A
  Quick-pass / Full-fail result is a possible miss requiring investigation,
  including whether the failure is flaky, environmental, or revision-related.
- Critical ratio divides Quick critical duration by Full critical duration;
  runner reduction is one minus Quick runner-seconds divided by Full's.
- Cache hit rate and changed-scope relevance are `null` until instrumented.
  Unclassified jobs are counted so new job naming cannot silently hide cost.

Reports do not certify sample eligibility. Apply the inclusion rules, verify
which checkout revisions each workflow validated (PR head versus merge ref),
and exclude incompatible coverage epochs. Metrics collection is advisory and
cannot alter a required gate.

### Coverage changes in this revision

The expanded Python Full matrix changes coverage and starts a new observation
epoch when merged. Previous activation dates and samples below remain historical;
record the new merge SHA and activation time before counting natural samples.
The independent Hosted Full path and required check names remain unchanged.

## Observed adverse evidence

Read back from the GitHub run/attempt Jobs API during this operational follow-up:

| Run / attempt | Source head | Quick | Full | Classification |
| --- | --- | --- | --- | --- |
| [36806174238 / 1](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/36806174238) | `a6303bd17614a69212f80f16c505f5cbc62ba8ca` | Pass; 4 jobs; 49 job-seconds; 38 s critical interval | Fail; 24 jobs; 691 job-seconds; 125 s critical interval | Possible miss; known dependency coverage discrepancy in Issue #116 |

The API returned all 29 jobs; one unclassified job is excluded from both group
costs and remains visible in the report. This is an open dependency PR from a
previous coverage epoch, not a qualifying merged sample. Neither the favorable
cost ratio nor a later passing run erases this adverse result. Keep Issue #117
blocked until the coverage decision and new observation criteria are satisfied.

## Previous Window Samples (Historical)

| PR                                                                           | Merge commit                               | Quick run                                                                                                             | Full run                                                                                                                | Quick critical duration | Full critical duration | Quick runner-seconds | Full runner-seconds | Result                                                |
| ---------------------------------------------------------------------------- | ------------------------------------------ | --------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ----------------------: | ---------------------: | -------------------: | ------------------: | ----------------------------------------------------- |
| [#142](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/142) | `f483e56e9c10ebfb9caa4bf0b0c43ec595282aca` | [CI Quick Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30207657214/job/89808575715) | [Validation Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30207657214/job/89808585847) |                   101 s |                  109 s |                299 s |               894 s | Pass; 92.66% critical ratio; 66.55% runner reduction  |
| [#152](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/152) | `67cbaf5456fbf3c069be2ddda01566531ca59a25` | [CI Quick Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30208730766/job/89811298032) | [Validation Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30208730766/job/89811404736) |                    55 s |                  108 s |                 16 s |               879 s | Pass; 50.93% critical ratio; 98.18% runner reduction  |
| [#143](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/143) | `676257413f9abee68363ab782034dbd2092edaf0` | [CI Quick Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30236422620/job/89885082764) | [Validation Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30236422620/job/89885184954) |                    59 s |                  109 s |                 50 s |               917 s | Pass; 54.13% critical ratio; 94.55% runner reduction  |
| [#144](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/144) | `f551eb8ed6e81140d31676157d04c21e6313a90c` | [CI Quick Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30237198651/job/89887392129) | [Validation Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30237198651/job/89887375236) |                   113 s |                  109 s |                282 s |               934 s | Pass; 103.67% critical ratio; 69.81% runner reduction |
| [#145](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/145) | `1b38dd21b13fdd271d5b1e427cb8052e114ea33e` | [CI Quick Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30237433460/job/89887894749) | [Validation Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30237433460/job/89888003349) |                    52 s |                  106 s |                 51 s |               883 s | Pass; 49.06% critical ratio; 94.22% runner reduction  |
| [#161](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/161) | `3a54d92b331a3e1f4e63ef6f95a5180a39464585` | [CI Quick Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30332114671/job/90189626062) | [Validation Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30332114671/job/90189609276) |                   126 s |                  117 s |                268 s |               867 s | Pass; 107.69% critical ratio; 69.09% runner reduction |
| [#172](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/172) | `b15f5fbc216e0c82b884107ea68f788b0b7e5e9f` | [CI Quick Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30592105008/job/91036588410) | [Validation Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30592105008/job/91036703268) |                    62 s |                  106 s |                 24 s |               925 s | Pass; 58.49% critical ratio; 97.41% runner reduction  |
| [#166](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/166) | `6e4745b32c874a2ae6dcc15864a81b83cd50b3be` | [CI Quick Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30594257423/job/91043477252) | [Validation Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30594257423/job/91043366015) |                   178 s |                  124 s |                260 s |               890 s | Pass; 143.55% critical ratio; 70.79% runner reduction |
| [#168](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/168) | `5b8219f030c33d4d0070f093d26d32b592115df9` | [CI Quick Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30594889073/job/91045293772) | [Validation Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30594889073/job/91045339733) |                    80 s |                  138 s |                 67 s |               960 s | Pass; 57.97% critical ratio; 93.02% runner reduction  |
| [#167](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/167) | `64977f4c85728515c77537c26afbeb6dcbcf1824` | [CI Quick Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30596581097/job/91050320879) | [Validation Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30596581097/job/91050394753) |                    81 s |                  115 s |                 79 s |               851 s | Pass; 70.43% critical ratio; 90.72% runner reduction  |
| [#171](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/171) | `1a3295ea1355afa4cf43889cb68d51f47e3d67a1` | [CI Quick Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30597350157/job/91052691834) | [Validation Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30597350157/job/91052777913) |                    84 s |                  122 s |                 61 s |               828 s | Pass; 68.85% critical ratio; 92.63% runner reduction  |
| [#169](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/169) | `9c90b384eb7ec10bec1b80e4be07c67f0ff3752a` | [CI Quick Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30598030207/job/91054739375) | [Validation Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30598030207/job/91054830937) |                    69 s |                  126 s |                 63 s |               870 s | Pass; 54.76% critical ratio; 92.76% runner reduction  |
| [#170](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/170) | `64cd62bfc1f8fb762fa72435985fbd0ced1b2eb5` | [CI Quick Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30598309382/job/91055565611) | [Validation Gate](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30598309382/job/91055611274) |                    69 s |                   93 s |                 60 s |               870 s | Pass; 74.19% critical ratio; 93.10% runner reduction  |

Historical qualifying sample count: **13**.

Historical median critical-duration ratio: **68.85%**.

Historical aggregate runner reduction: **86.34%** (1,580 Quick runner-seconds
compared with 11,568 Full runner-seconds).

## Current Window Samples

No qualifying natural samples exist for a current activation epoch.

Current qualifying sample count: **0**.

Current median critical-duration ratio: **Not available**.

Current aggregate runner reduction: **Not available**.

## Decision

**A new observation epoch must be recorded after the coverage change merges.**
Issue #117 must not
change required contexts until every criterion is supported by linked Actions
evidence from this epoch.

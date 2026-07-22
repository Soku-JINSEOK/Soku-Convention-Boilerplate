# GitHub Governance Hardening for Boilerplate (Issue #55)

## Outcome

This report captures the metadata normalization and governance hardening work tracked by
[Issue #55](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/55) after execution.

## Implementation Snapshot

The repository-local implementation was rebased onto `origin/main` at
`2aad7940c4ed4883f69452cb41899cd563007c5c`. GitHub mutations use a fresh read
immediately before each batch. Closed pull requests remain closed and unmerged.

### Exact relationship and label mutation manifest

| PR | State | Preserved relationship/content | Minimal addition | Label addition | Before body SHA-256 |
| --- | --- | --- | --- | --- | --- |
| #51 | closed | `Tracks #44` and full body | `Related to #44` | none | `6344be872e2d74a9fa84a4acb6d2ed23ef63e0463dc3dd7b51bd1f82db49601c` |
| #58 | closed-unmerged | Dependabot body | `Related to #69` | `area:ci` | `798847ac5afe7627dc4bc00772ab7863207a66deeb230e500a342dbb7269c128` |
| #59 | closed-unmerged | Dependabot body | `Related to #69` | `area:ci` | `919d9c41fe8adba25999096dde6b9bb00d15722fbedbbe9714db584dc7e0deb8` |
| #60 | open | full governed Dependabot body | none (`Related to #69` exists) | `area:ci` | `9ff78f702df69828278412e8c9fbc5206d41ad166406d58794c5019c8a479afa` |
| #61 | open | full governed Dependabot body | none (`Related to #69` exists) | `area:templates` | `9dfcbd67b7c154e373cccdf52e8fb45f239883ef76213ed3688615e59afaf49a` |
| #62 | closed-unmerged | Dependabot body | `Related to #69` | `area:ci` | `991e7bcd6229df2727e2e3d05f6a3409c9dae27ed5d09b3fe02d333c9097cf96` |
| #63 | closed-unmerged | Dependabot body | `Related to #69` | none | `bf4e0f28edce920dc441f14f2c04cc67c4c6b7ec1428fa6f9ac8d4c57bef04ac` |
| #64 | closed-unmerged | Dependabot body | `Related to #69` | none | `2c57412216b60f534455fbc76008aef462f3709f37c85f7ed4fe486861098171` |
| #65 | closed-unmerged | Dependabot body | `Related to #69` | none | `4c9e971bd114ef8d885b55904002c2b96ff2b56288ff2585f00dd34b521f4d8e` |
| #66 | open | full governed Dependabot body | none (`Related to #69` exists) | `area:templates` | `4665fd39aff6989c3e1c387b9238d0e01caf5a8b0be2ed60867a04ba166470d2` |
| #68 | closed | control-plane #9 relationship and full body | `Related to #55` | none | `d9a3d174a50e198c155d4885be01529ba7728a72d6974ab490bd87992328c2a3` |

Project #2 was verified before mutation with exactly 17 fields and 8 views. The
native auto-add target is changed separately, with explicit approval, to
`repo:Soku-JINSEOK/Soku-Convention-Boilerplate is:issue is:open`.

### Relationship mutation verification (post-mutation)

- `#51`: `9e888907c824096c409a8effbd1172015ad4b52e5f2d8ad3c9daa4e96386b161`
- `#58`: `c5666d040f5cf1028fb561b0bfd4184f5e10647d6275777c86578d6b53f90400`
- `#59`: `c840827151fd2b57cc63a9337959a8fa0dd95e17c8a14298a6546634a806c080`
- `#60`: unchanged, `9ff78f702df69828278412e8c9fbc5206d41ad166406d58794c5019c8a479afa`
- `#61`: unchanged, `9dfcbd67b7c154e373cccdf52e8fb45f239883ef76213ed3688615e59afaf49a`
- `#62`: `1c97bb647d1b19c73bae2d8edf5977bf4e947c9af8253aa750c7c1f9d15bac85`
- `#63`: `48db8ac97a2ed41d8333c718df428e429974a5d709578b80064c8bb36e7461a6`
- `#64`: `ca69b726e65fabdd1479f40f6c19b1ea7ec92d97bf44c5773edb165ee7db3bac`
- `#65`: `5f2c9c69a003e11363b7eedf8e5bb27743c9909ed8552f883d346650b355724f`
- `#66`: unchanged, `4665fd39aff6989c3e1c387b9238d0e01caf5a8b0be2ed60867a04ba166470d2`
- `#68`: `f19b5cc78d948e19c1c792fc8bfd34a3f1ab9a449e8cd48c7d5b9ff81a50d059`

Fresh reads verified that `#51`, `#58`, `#59`, `#62`-`#65`, and `#68` remain
closed. Existing labels were preserved; only the manifest additions were made.

### Completion-audit metadata normalization

A fresh completion audit found eight historical pull requests that still had a
canonical `type:*` label but no canonical `area:*` label. The following exact
label-only manifest was declared in the execution log before the supplemental
batch and copied into this report after the mutations completed. The selected
areas match each pull request's linked Issue labels and changed-file scope; the
table is therefore an audit record, not evidence that this file was committed
before execution.

| PR | Preserved state | Exact label additions | Post-mutation body SHA-256 |
| --- | --- | --- | --- |
| #7 | closed-merged | `area:templates` | `3dc06b07583135b034061047426d01394cf4943f311da05273a2b01a0f7ba6b7` |
| #26 | closed-merged | `area:docs`, `area:tooling` | `dada719087742fc2b9fe9f42e4fd28504a002f7b4190d8a63a835dd73297ce79` |
| #29 | closed-merged | `area:docs`, `area:tooling`, `area:ci`, `area:templates` | `709af30765914d00f81aa22a4c1f9167e8518a2d3f8891f1b04e7400d9b74a7b` |
| #30 | closed-merged | `area:docs`, `area:tooling`, `area:templates` | `3c298237bd4485f1a43ef8cc39f6dfd4ebe05d2cf0f577498ce41e437a8c1f3c` |
| #31 | closed-merged | `area:tooling`, `area:ci`, `area:templates` | `a2f6ec4238072312ed5622dfec25676f78fb31e4c6bac823f2a8ef350a514da2` |
| #32 | closed-merged | `area:docs`, `area:tooling`, `area:templates` | `d88b1398f9c9038555ec20b40a9c2a25bdd94e6ad25a7d073ff79c206610be8d` |
| #33 | closed-merged | `area:docs`, `area:tooling`, `area:ci`, `area:templates` | `504d8bd863e324483e996eb2db747fc4b09d7c4bdc837ef4c8709c958f2646aa` |
| #50 | closed-merged | `area:docs`, `area:tooling`, `area:ci` | `b121629fd163eeeca14757d0617c91389405e58389c830cc2c392918a109dfc7` |

The batch used only `gh pr edit --add-label`; it supplied no body, title, state,
merge, or review arguments. Post-mutation reads verified that all eight pull
requests remain closed and merged, retain `assignee=Soku-JINSEOK`, and have the
listed canonical labels. The body hashes above are legacy evidence with mixed
newline handling. They use the exact output of
`gh api repos/Soku-JINSEOK/Soku-Convention-Boilerplate/pulls/<n> --jq .body`
including the CLI-added output newline, and are retained rather than rewritten.

The implementation pull request `#70` was also assigned to `Soku-JINSEOK`
without changing its body, labels, state, or Project membership.

### Issue / PR changes applied

| Target | Before | After | Note |
| --- | --- | --- | --- |
| Issue #2 | `fix(templates): ...` title only, normalized type/area completed later | `🐛 fix(templates): ...` | Added `type:bug`, `area:templates`; kept canonical `assignee=Soku-JINSEOK`, body preserved |
| Issue #3 | `chore(repo-hygiene): ...` title format | `🔧 chore(repo-hygiene): ...` | Added `type:chore`, `area:docs`, `area:tooling` |
| Issue #4 | `chore(documentation): ...` title format | `🔧 chore(documentation): ...` | Added `type:chore`, `area:docs`, `area:tooling` |
| Issue #56 | `security(release): ...` title format | `🔒️ security(release): ...` | Added `type:chore`, `area:docs`, `area:tooling`, `area:ci`; body preserved |
| PR #1 | `feat(boilerplate): ...` | `✨ feat(boilerplate): ...` | title normalized to governance format |
| PR #5 | `chore(convention): ...` | `🔧 chore(convention): ...` | title normalized; kept canonical labels |
| PR #34 | `docs(roadmap): ...` | `📚 docs(roadmap): ...` | title normalized |
| PR #37 | `fix(soku): ...` | `🐛 fix(soku): ...` | title normalized |
| PR #38 | `docs(release): ...` | `📚 docs(release): ...` | title normalized |
| PR #58 | dependency title | `📦 build(github-actions): ...` | title normalized to governance format |
| PR #59 | dependency title | `📦 build(github-actions): ...` | title normalized to governance format |
| PR #62 | dependency title | `📦 build(ci): ...` | title normalized to governance format |
| PR #63 | dependency title | `📦 build(python): ...` | title normalized to governance format |
| PR #64 | dependency title | `📦 build(templates): ...` | title normalized to governance format |
| PR #65 | dependency title | `📦 build(templates): ...` | title normalized to governance format |
| PR #57, #58, #59, #62, #63, #64, #65, #52, #53 | missing/legacy `type:*` in some cases | `type:bug` or `type:chore` added as applicable | required canonical `type:` labels now present |
| PR #52, #53, #57, #58, #59, #62, #63, #64, #65 | unassigned / mixed | `Soku-JINSEOK` assigned | assignee normalization applied |
| PR #7, #26, #29-#33, #50 | canonical `type:*` but no canonical `area:*` label | exact `area:*` additions in the completion-audit manifest | bodies, titles, and closed-merged states preserved |
| PR #70 | unassigned implementation pull request | `Soku-JINSEOK` assigned | body, labels, open state, and Project exclusion preserved |
| Issues #16–#24, #40, #44 | Open/closed project rows present before | removed from Project #2 | closed-only cleanup completed |
| PR #60, #61, #66 | PR-only changelog body | wrapped under full issue template; original release/changelog/commit evidence retained | Added `## 🔗 Common Metadata`, `English/KO/JP` sections, checklist, and AI section |
| Issues #41, #54, #55 | status labels mixed; project metadata inconsistent | Project fields now canonicalized | canonicalized through single source fields only |
| Issue #41 | In Project: priority/status labels in legacy set + missing `workstream` consistency | `Status: In progress`, `Priority: P1`, `Size: L`, `Workstream: Delivery`; canonical `type/area` retained | fixed |
| Issue #54 | In Project with no blocker notes | `Status: Blocked`, `Priority: P2`, `Size: M`, `Workstream: Engineering`; body updated with upstream blockers `ci-cd-control-plane#19/#25` | explicit boundary notes and blocker terms added |
| Issue #55 | No metadata lock-in fields | `Status: In progress`, `Priority: P1`, `Size: L`, `Workstream: Governance`, `Target date: 2026-07-31` | execution issue |
| Aggregate Issue #69 | not present in Project metadata initially | Added to Project #2: `Status: In progress`, `Priority: P2`, `Size: M`, `Workstream: Engineering`, `Target date: 2026-07-31`; links to #60/#61/#66 with `Related to` | tracks monthly dependency rollup |

### Open item membership check (post-run)

Open issues in repo: `#41`, `#54`, `#55`, `#69` — all are in Project #2. Closed issues are not in Project #2.

Issue #69 was updated to explicitly separate open (#60, #61, #66) and closed-unmerged (#58, #59, #62–#65, #68) PRs, and to record the corresponding hash ledger.

### Project #2 canonical fields for active Boilerplate issues

- `#41`: `In progress` / `P1` / `L` / `Delivery`
- `#54`: `Blocked` / `P2` / `M` / `Engineering`
- `#55`: `In progress` / `P1` / `L` / `Governance` (`Target date: 2026-07-31`)
- `#69`: `In progress` / `P2` / `M` / `Engineering` (`Target date: 2026-07-31`)

### Legacy body hash ledger (SHA-256, mixed newline handling)

These values are retained as originally recorded. Some were calculated from
raw body bytes and some from CLI-formatted output, so they must not be compared
as one canonical dataset.

- `#41`: `02c450f236f8d37debc5417d2469bb2d07a30e40d10aa45d718b50720cf550de`
- `#54`: `37f4fc2957ebf95095e19d6196c52466380c8196604e479ab09c177f30e9f5d5`
- `#55`: `d40327e4b1cfb912b810e794e0da75f632acd4344fc2c2db839305ce7caec3bd`
- `#69`: `bbba3774046753fed4a87b8d1609d550e8e32ae787df105eb80c780368075e37`
- `#60`: `9ff78f702df69828278412e8c9fbc5206d41ad166406d58794c5019c8a479afa`
- `#61`: `9dfcbd67b7c154e373cccdf52e8fb45f239883ef76213ed3688615e59afaf49a`
- `#66`: `4665fd39aff6989c3e1c387b9238d0e01caf5a8b0be2ed60867a04ba166470d2`

### Canonical current-state body ledger

The canonical procedure hashes only the raw UTF-8 bytes of the JSON `body`
value, without an output newline:

```bash
curl -fsSL "https://api.github.com/repos/<owner>/<repo>/<issues|pulls>/<n>" \
  | jq -j '.body // ""' \
  | shasum -a 256
```

Fresh public API reads on 2026-07-22 produced:

| Target | Canonical body SHA-256 |
| --- | --- |
| Issue #41 | `e4b408db77dde80280573c1080b82a1a155026bc247f16b2453c1fbe09087e81` |
| Issue #54 | `c530ea93d5283d96e77c4a786ee475360dcf71ab7a5d538754c97161a9104f80` |
| Issue #55 | `2469649c2a7136ec8bd99647c0e9f00e3b7299da915efe70670880e9cddc409f` |
| Issue #69 | `bbba3774046753fed4a87b8d1609d550e8e32ae787df105eb80c780368075e37` |
| PR #7 | `b46c6869bcad7bdd5fb543097ad0060d95f65bc8362f82e889b2549fa6e6d4b0` |
| PR #26 | `db1b3cd7bfeb3ef35121967c3bd97f664b69c494e9f3bbbcd7467600cd4d9d21` |
| PR #29 | `c56fb13aef6213ab2324e063b2bad913e7c5001c36059f5c4dc11e4349352dbf` |
| PR #30 | `7b237ec2bec19d36625eaf0e8c3ec30499d8e2de4347872a4611a4a9f9f11021` |
| PR #31 | `3e887ab8004bd9b881c2b9b3f5d70f2462bf3c634adac26fb4fbd51dfd609a7e` |
| PR #32 | `af31b8857f6d486eaa028bd180e2169043ec59f87a127192e597b6321dc9a3f9` |
| PR #33 | `b88c307847cc53e0aa04be9f7d61c6ee5fcb924852b21f05d6847c2dac7d1c56` |
| PR #50 | `e9524334f93ad06959cf5214e3dfef105ddf428341083f1168f12ec2ead0152b` |
| PR #60 | `b6216e81898bb429044ecb1d7c35eaf37f20b239e7b1b21674bc600bb74f35fd` |
| PR #61 | `eb75b1c206c6e35e2100ff4aefaf16c7e8cdc89fd20a5274e43f50d76554707a` |
| PR #66 | `4a56ac49f49ef33fd2efd475a6f2dced3e1c377d66defe7fb2909b1a04420f85` |

### Project workflow evidence

Project #2 workflow #7 was created as the native auto-add workflow for open
issues in this repository. A fresh GraphQL read on 2026-07-22 returned
`name=Auto-add to project`, `enabled=true`, and
`updatedAt=2026-07-21T20:14:04Z`. The workflow filter is not exposed by the
`ProjectV2Workflow` GraphQL type; the execution-time record is
`repo:Soku-JINSEOK/Soku-Convention-Boilerplate is:issue is:open`. This report
distinguishes that recorded configuration from the API-observed activation
state and does not claim that the workflow existed before the mutation manifest
was declared.

### Canonical label reconciliation

The catalog remained authoritative. After a fresh read, the approved exact
mutation manifest updated only these four existing labels through
`scripts/sync-labels.sh`; no custom labels were removed or rewritten:

| Label | Color | Description |
| --- | --- | --- |
| `area:standards` | `bfd4f2` | Standards, governance, and policy conventions |
| `area:templates` | `d4c5f9` | Copyable starter configuration sets under templates/ |
| `area:automation` | `c8f7c5` | Automation workflows and operational orchestration |
| `area:security` | `a5b6f7` | Security boundaries, safeguards, and secret handling |

A post-mutation read on 2026-07-22 matched each name, color, and description to
`.github/labels.yml` exactly.

### Verification commands

- `node --test templates/_shared/commitlint/*.test.mjs .github/validate-pr-governance.test.mjs .github/issue-form-order.test.mjs .github/validation-workflow.test.mjs`
  plus the rendered `scripts/contribution-title.test.mjs` and
  `scripts/pull-request-policy.test.mjs` ✓ pass, 63 tests
- `npx --yes yaml-lint@1.7.0 .github/*.yml .github/**/*.yml` ✓ pass
- Python YAML parse over `.github/**/*.yml` ✓ pass, 16 files
- `npx --yes markdownlint-cli2@0.22.1 --config .markdownlint.jsonc "**/*.md" "#**/node_modules/**"`
  ✓ pass, 63 files and 0 errors
- `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.10` ✓ pass
- `scripts/verify-release-tag_test.sh` ✓ pass
- `bash -n scripts/*.sh soku/scripts/*.sh` ✓ pass
- `git diff --check` ✓ clean
- `shellcheck` and PowerShell sync parity remain unavailable locally; hosted
  validation is required before merge.

## Notes

- The follow-up policy requires canonical catalog labels, assignee, a single
  exact Issue relation, a matching existing task report, the rendered profile,
  ordered headings, checked verification, and explicit AI disclosure.
- Code-bearing events create `Validation Gate` and `PR Metadata Gate`.
  Metadata-only events skip the heavy groups and dynamically name the full gate
  check `Full Validation Not Required`, while metadata validation reads the
  current pull request through the GitHub API.
- PR bodies for `#60`, `#61`, `#66` preserve original release evidence and must
  be updated to the strict metadata contract before merge.
- Post-mutation hashes are recorded only after a fresh verification read; no expected hash is presented as observed evidence.

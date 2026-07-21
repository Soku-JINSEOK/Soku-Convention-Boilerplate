# GitHub Governance Hardening for Boilerplate (Issue #55)

## Outcome

This report captures the metadata normalization and governance hardening work tracked by
[Issue #55](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/55) after execution.

## Implementation Snapshot

The repository-local implementation was rebased onto `origin/main` at
`ea8b3a5ac6c3692f9535a570ead15813b53ca8f4`. GitHub mutations use a fresh read
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
| Issues #16–#24, #40, #44 | Open/closed project rows present before | removed from Project #2 | closed-only cleanup completed |
| PR #60, #61, #66 | PR-only changelog body | wrapped under full issue template; original release/changelog/commit evidence retained | Added `## 🔗 Common Metadata`, `English/KO/JP` sections, checklist, and AI section |
| Issues #41, #54, #55 | status labels mixed; project metadata inconsistent | Project fields now canonicalized | canonicalized through single source fields only |
| Issue #41 | In Project: priority/status labels in legacy set + missing `workstream` consistency | `Status: In progress`, `Priority: P1`, `Size: L`, `Workstream: Delivery`; canonical `type/area` retained | fixed |
| Issue #54 | In Project with no blocker notes | `Status: Blocked`, `Priority: P2`, `Size: M`, `Workstream: Engineering`; body updated with upstream blockers `ci-cd-control-plane#19/#25` | explicit boundary notes and blocker terms added |
| Issue #55 | No metadata lock-in fields | `Status: In progress`, `Priority: P1`, `Size: L`, `Workstream: Governance`, `Target date: 2026-07-31` | execution issue |
| Aggregate Issue #69 | not present in Project metadata initially | Added to Project #2: `Status: In progress`, `Priority: P2`, `Size: M`, `Workstream: Engineering`, `Target date: 2026-07-31`; links to #60/#61/#66 with `Related to` | tracks monthly dependency rollup |

### Open item membership check (post-run)

Open issues in repo: `#41`, `#54`, `#55`, `#69` — all are in Project #2. Closed issues are not in Project #2.

### Project #2 canonical fields for active Boilerplate issues

- `#41`: `In progress` / `P1` / `L` / `Delivery`
- `#54`: `Blocked` / `P2` / `M` / `Engineering`
- `#55`: `In progress` / `P1` / `L` / `Governance` (`Target date: 2026-07-31`)
- `#69`: `In progress` / `P2` / `M` / `Engineering` (`Target date: 2026-07-31`)

### Body hash ledger (SHA-256, post-mutation)

- `#41`: `02c450f236f8d37debc5417d2469bb2d07a30e40d10aa45d718b50720cf550de`
- `#54`: `37f4fc2957ebf95095e19d6196c52466380c8196604e479ab09c177f30e9f5d5`
- `#55`: `d40327e4b1cfb912b810e794e0da75f632acd4344fc2c2db839305ce7caec3bd`
- `#69`: `f1da285e008dd6d07e864a356a08900c80f2d844d0822c71bf1fb9e3acdadbde`
- `#60`: `9ff78f702df69828278412e8c9fbc5206d41ad166406d58794c5019c8a479afa`
- `#61`: `9dfcbd67b7c154e373cccdf52e8fb45f239883ef76213ed3688615e59afaf49a`
- `#66`: `4665fd39aff6989c3e1c387b9238d0e01caf5a8b0be2ed60867a04ba166470d2`

### Verification commands

- `npx --yes yaml-lint@1.7.0 .github/*.yml .github/**/*.yml` ✓ pass
- `npx --yes markdownlint-cli2@0.22.1 --config .markdownlint.jsonc "**/*.md" "#**/node_modules/**"` ✓ pass
- `node --test templates/_shared/commitlint/*.test.mjs` ✓ pass
- `GITHUB_EVENT_PATH=/tmp/gh-event-55-pr-60.json node .github/validate-pr-governance.mjs` ✓ pass
- `git diff --check` ✓ clean

## Notes

- PR bodies for `#60`, `#61`, `#66` preserve original release evidence and now satisfy required heading/metadata/checklist order requirements.
- Post-mutation hashes are recorded only after a fresh verification read; no expected hash is presented as observed evidence.

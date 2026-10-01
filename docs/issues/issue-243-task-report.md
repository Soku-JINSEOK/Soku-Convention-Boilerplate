# Task report: design and readiness documentation

## Goal and Background

Related to #243. First-time adopters need to connect requirements, runtime/data
boundaries, security, verification and delivery using the existing documents.
The owner requested improvements to those documents and visual review aids,
rather than a separate step-by-step learning guide.

## Proposed Approach

Keep existing document authority and responsibilities. Add a review entrypoint
to USAGE_MANUAL and the three README overviews. Place detailed boundary,
permission, acceptance and delivery checks in PROJECT_STRUCTURE,
SECURITY_POLICY, VERIFICATION_GUIDE and CICD_STANDARDS. Extend the task-report
template and proportional applicability guidance.

Static SVG figures live under docs/assets and have descriptive text alternatives.
They are repository-owned source illustrations, not exported Figma/tldraw
artifacts or deployed-infrastructure claims. review-evidence.mmd provides an
editable flow representation. The document tables are usable without an image
or an external design account.

## Planned Implementation

- Add scope, quality, runtime/data, access, network, acceptance and recovery review
  questions to existing adoption guidance.
- Record owner, decision, evidence and Pass / Fail / Blocked / N/A semantics.
- Review web, connected desktop and offline/local applicability.
- Extend the existing task-report template with design/risk and verification/
  delivery planning, including its contiguous Korean block.
- Add three static diagrams and a Mermaid source under docs/assets.
- Preserve runtime code, templates, CI configuration, releases and delivery state.

## Acceptance Criteria

- Existing original document content and authoritative relationships are retained.
- Applicable review items have concrete decisions and expected evidence.
- Missing/skipped checks are not represented as passes; unused components need
  an N/A reason and owner.
- Product acceptance is separated from this repository's template verification.
- App rollback is separated from data migration/restoration.
- Overview links, relative image links and changed-target anchors resolve.
- Figures have no scripts, embedded credentials or external resources.

## Approval

- **Status:** Approved
- **Approved by:** Soku-JINSEOK
- **Basis:** The owner's explicit 2026-10-01 request to improve existing documents
  with visual review content and proceed with GitHub modifications.
- **Boundary:** Scoped document/source-illustration changes and a reviewable PR.
  Merge, release and deployment remain separate.

## Approved extension: document architecture and CI recovery

On 2026-10-01 the owner explicitly requested document-specific visualizations
that communicate engineering conventions, plus remediation of CI/CD failures
associated with commits, issues and pull requests.

- Visualize authority, code/data boundaries, access control, cloud exposure,
  verification, supply-chain inputs and the two release axes in their owning
  documents. Keep examples distinguishable from implemented infrastructure.
- Generate editable FigJam material and embed maintainable Mermaid diagrams
  directly in GitHub documents, with a plain-text explanation and checks.
- Patch the failing pinned dependencies: brace-expansion 5.0.9 to 5.0.12 and
  fast-uri 3.1.7 to 3.1.8, preserving the current major versions.
- Verify manifest/lock consistency, affected ecosystem behavior and hosted
  security audits. Inspect current results rather than rewriting old failures.
- Preserve required gates and trust boundaries. Historical cancelled runs are
  superseded evidence, not passed runs. Merge and release remain separate.

**Status:** Approved by the owner's explicit request. This extends the earlier
documentation-only scope to the dependency and validation fixes above.

## Issue and PR failure remediation

The same authorized CI/CD remediation scope includes the observed Project sync
secondary rate limit (run 36798003322) and PR #240's missing Issue relation.

The API client now serializes requests and retries only rate-limited GETs,
at most twice, honoring Retry-After/reset headers with a bounded wait.
Permission failures and mutations are not replayed; an exhausted limit fails.
Queued requests remain blocked until the server cooldown expires.
The distributed Soku asset uses the same implementation. Eight stub-based
regression cases cover serialization, timing, exhaustion and failure behavior.
They passed against the extracted client in an in-memory JavaScript harness;
the actual Node suites must also pass in hosted CI.

PR #240 now links its existing dependency tracking issue #242. Historical
failures remain intact. No credential or trust boundary was broadened.

## Current extension status

Added fourteen document-specific Mermaid diagrams, three README indexes,
an editable FigJam board with authority/security/CI diagrams, and the two
patch dependency updates. A read-only hosted job installs the capture runner
lockfile and runs its existing type and unit checks; browser E2E remains
explicitly opt-in. Full security and repository gates remain unchanged.

Local shell provisioning failed before command execution. Lockfile edits use
matching integrity metadata cross-checked against public upstream consumer
lockfiles and the libraries' tagged package manifests. Hosted npm installation
and audits must confirm these inputs; local execution is not claimed.

Static validation checked 296 relative links and 51 changed-target anchors
without missing targets. The FigJam board was read back and visually inspected.
Manifest and lock-root pins agree. CI evidence and remaining blockers are
tracked in PR #244 and issue #243.
Results in the earlier Verification section below describe the previous
implementation, not the newly extended patch.

## Implementation Status

Implemented eleven existing-document updates, three SVG figures and one
Mermaid source in PR #244. Sixteen committed files were read back and matched
the prepared contents. No separate learning guide, new governance profile or universal
application topology was introduced.

## Verification

Actually completed against the prepared change set:

- Relative local file links: 248 checked, no missing targets.
- Anchors to changed documents: 28 checked, no missing anchors.
- Original content order/preservation: eleven existing documents checked.
- Added Markdown lines: 274 checked for trailing whitespace and heading/table
  separation; no errors.
- Three SVG figures: XML element nesting, accessible title/description and
  prohibited script/external-resource checks passed.
- All three SVGs rendered and visually reviewed; overlapping branch labels
  were shortened, and the recovery-to-review path was made explicit.
- Hosted checks on `6df44c49b5a118cc1f382c8afc2af02b56580890`:
  Repository Hygiene (including Markdown lint and documentation regressions),
  PR Metadata Gate and CI Quick Gate passed.

Limitations:

- Full Validation reports existing `brace-expansion` and `fast-uri` security
  findings in unchanged lockfiles. Full repository verification is not a pass.
- The implementation commits are unsigned; signing must be addressed before
  merge under the repository's contribution rules.
- tldraw editing was blocked by the execution environment's approval policy.
- The earlier Figma selection limitation was resolved for this extension.
  The [editable board](https://www.figma.com/board/SJgcvEV1HZqYwHM5Nt5HWE)
  is available.
  The original SVGs remain repository-authored illustrations.
- Results above identify the checked implementation commit. Any later report
  edits require their own hosted checks before merge; no checks or branch rules
  are relaxed.

## Public Disclosure Review

- [x] No credentials, tokens, private keys, or credential-bearing URLs
- [x] No private repository, project, or product names
- [x] No cloud project IDs, account numbers, service URLs, image URIs, or revision
      identifiers
- [x] No personal billing, subscription, budget, or payment-status information
- [x] No personal email, phone, address, or local absolute path
- [x] No private Issue, PR, Project, or control-plane identifiers

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

## Follow-up: explainable decisions and CI workload

The owner requested additional implementation rationale and CI optimization in
Issue #245. Its [supplemental report](./issue-245-task-report.md) records approval,
alternatives, expected savings and validation limits. PR #244 retains this
original report as its Common Metadata record.

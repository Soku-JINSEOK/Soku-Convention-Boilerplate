# 🐙 GitHub Standards

## 🎯 Purpose

This document defines repository collaboration standards for GitHub-based workflows.

It exists to ensure that issues, pull requests, reviews, and automation remain consistent across projects built on this boilerplate.

## 📐 Principles

GitHub workflows should optimize for:

- clear communication
- low review friction
- explicit decision history
- predictable collaboration patterns

## 🌐 Collaboration Language

Document language (which content is Korean/English/Japanese vs. English-only) is governed separately by the [Language Policy in BLUEPRINT.md](../../BLUEPRINT.md#language-policy). That policy does not decide what language commit messages, issues, and pull requests are written in — collaboration language is a distinct, per-project choice, because contributor makeup varies (some projects work entirely in Korean, some entirely in English, some entirely in Japanese, some mixed).

Rules:

- Each downstream project must pick one collaboration language convention — Korean-only, English-only, Japanese-only, or an explicitly documented mix — and record it in that project's own `CONTRIBUTING.md`.
- This boilerplate's own `.gitmessage` and commit examples use English text as a default illustration of the Conventional Commits + gitmoji *structure*. The structure (type, scope, gitmoji, footer format) should be kept; the human-language description text should follow whatever the project has chosen.
- Do not assume a downstream project's collaboration language from this boilerplate's defaults — confirm it explicitly (see `docs/guides/INIT_GUIDE.md` for the AI-agent bootstrap checklist that includes this step).

## 🐞 Issue Standards

Issues should capture a problem, request, or decision point clearly enough that another contributor can act on them without extra guesswork.

Every issue should make the following as clear as possible:

- background
- goal
- scope
- constraints
- definition of done

### 📋 Recommended Issue Types

- bug report
- feature request
- refactor proposal
- documentation update
- chore or maintenance task

## 🔀 Pull Request Standards

Pull requests should be structured to help reviewers understand the change quickly.

Every PR should answer:

- What changed?
- Why was it needed?
- How was it validated?
- What risks or tradeoffs exist?
- What follow-up work remains?

## 🔍 Review Standards

> **Applies to:** Team — see [`docs/guides/APPLICABILITY.md`](../guides/APPLICABILITY.md). This section assumes a second person reviewing the PR; a solo project has no reviewer to apply it to.

Reviews should focus on:

- behavioral correctness
- architectural fit
- maintainability
- test quality
- clarity of intent

Review noise should be minimized.  
Formatting concerns should be delegated to tooling whenever possible.

## 🌿 Branching Guidance

Repositories may choose their own branching strategy, but the strategy should be documented and consistently applied.

At minimum, teams should define:

- default branch policy
- feature branch naming pattern
- hotfix handling
- release tagging approach

## 🔄 Release And Sync

Releases should be explicit and repeatable.

- Use semantic-style tags in the form `vMAJOR.MINOR.PATCH`.
- Pin downstream repositories to a specific boilerplate tag.
- Record the imported tag in the downstream README or setup notes.
- Use `scripts/sync-boilerplate.sh` (Linux/macOS) or `scripts/sync-boilerplate.ps1` (Windows) to copy convention-owned files into a downstream repository.
- Keep downstream application code separate from boilerplate updates.

See [RELEASE_AND_SYNC.md](./RELEASE_AND_SYNC.md) for the full operating contract.

### 🔏 Signed Commits and Tags (Verified)

To guarantee the integrity and authorship of the convention updates, all committers and maintainers should sign their commits and release tags. Signed commits get a **Verified** badge on GitHub.

#### 1. SSH Key Signing (Recommended)
If you already use an SSH key for GitHub, you can use it to sign commits:
```bash
# Configure Git to use SSH for signing
git config --global gpg.format ssh

# Set your SSH public key as the signing key (or path to public key file)
git config --global user.signingkey "~/.ssh/id_ed25519.pub"

# Enable signing globally
git config --global commit.gpgsign true
git config --global tag.gpgsign true
```
Make sure to upload this public key to your GitHub account under **Settings -> SSH and GPG keys -> New SSH Key** and set the **Key type** to **Signing Key**.

#### 2. GPG Key Signing
Alternatively, you can use a GPG key:
```bash
# Generate a new GPG key
gpg --full-generate-key

# List your keys to find the Key ID
gpg --list-secret-keys --keyid-format=long

# Set the key in Git (replace <KEY_ID> with your key ID)
git config --global user.signingkey <KEY_ID>
git config --global commit.gpgsign true
git config --global tag.gpgsign true
```
Export and paste your GPG public key (`gpg --armor --export <KEY_ID>`) into GitHub under **Settings -> SSH and GPG keys -> New GPG Key**.

For interactive configuration assistance, you can run the [scripts/setup-git-signing.sh](../../scripts/setup-git-signing.sh) script.

## 🏷️ Labels and Metadata

GitHub labels should help organize work rather than create clutter.

Prefer a small, stable label system covering:

- type
- priority
- status
- area or domain

The canonical catalog lives in [`.github/labels.yml`](../../.github/labels.yml). Apply it (or a downstream-adapted copy) with `scripts/sync-labels.sh`, which is idempotent and safe to re-run whenever the catalog changes:

```bash
scripts/sync-labels.sh --repo <owner>/<repo>
```

| Axis | Values |
| --- | --- |
| `type:` | `bug`, `feature`, `chore`, `docs`, `refactor` |
| `priority:` | `p0-critical`, `p1-high`, `p2-normal`, `p3-low` |
| `status:` | `triage`, `ready`, `in-progress`, `blocked`, `done` |
| `area:` | `docs`, `tooling`, `ci`, `templates` (example set — adapt per project) |

Issue templates (`.github/ISSUE_TEMPLATE/*.md`) reference the `type:` axis by default (`type:bug`, `type:feature`, `type:chore`) — run `sync-labels.sh` before those templates are used, or issue creation will silently drop the label.

A solo project can usually run with `type:` alone; `priority:`/`status:`/`area:` exist to coordinate work across people and can be added later once multiple contributors are involved (see [`docs/guides/APPLICABILITY.md`](../guides/APPLICABILITY.md)).

**Rule: never create an issue or PR unlabeled.** When opening an issue or PR (via `gh issue create`, `gh pr create`, or the GitHub UI), attach at least a `type:` label in the same action — check `gh label list` (or this catalog) first if unsure what exists, rather than creating it bare and labeling as a follow-up step. An unlabeled issue/PR is harder to triage and defeats the point of having a catalog at all.

## 📄 Templates

Repositories should provide templates where they reduce ambiguity.

Recommended templates:

- issue templates
- pull request template
- bug report template
- feature request template

### 📝 Commit Message Linting Template

To enforce the commit message convention (defined in [CONTRIBUTING.md](../../CONTRIBUTING.md#commit-message-standard)) automatically in pull requests or via Git Hooks, a shared configuration template is provided:

- [templates/_shared/commitlint/commitlint.config.mjs](../../templates/_shared/commitlint/commitlint.config.mjs)
- [templates/_shared/commitlint/contribution-title.mjs](../../templates/_shared/commitlint/contribution-title.mjs)

Downstream projects can copy these files into their repository roots and integrate them with tools like Husky or GitHub Actions to block non-compliant commit messages.

## 🤖 Automation Expectations

GitHub should be used as an operational surface, not just a code host.  
That means repository automation should support:

- CI validation
- consistent review flow
- visibility into project health
- structured collaboration

## 🎬 Summary

Well-managed GitHub workflows reduce coordination cost.  
The goal is to make repository collaboration explicit, reviewable, and repeatable across teams and projects.

# 🚀 Init Guide

## Purpose

This document is written for AI coding agents (not humans) that are applying
`Soku-Convention-Boilerplate` to a downstream repository. It defines how to
detect the target repository's stack and which setup steps to execute, in
order. Follow it as a procedure, not as background reading.

## 1️⃣ Detect the Stack

Inspect the target repository root for these marker files. A repository may
match more than one row (for example, a Node.js API with a PostgreSQL schema).

| Marker file(s) found in target repo | Stack | Template source |
| --- | --- | --- |
| `package.json` | JavaScript/TypeScript/Node.js | `templates/javascript-typescript-node/` |
| `pyproject.toml`, `requirements.txt`, or `setup.py` | Python | `templates/python/` |
| `go.mod` | Go | `templates/go/` |
| `pom.xml` (Spring Boot parent/dependencies) | Java/Spring | `templates/java-spring/` |
| `*.sql` migrations referencing `AUTO_INCREMENT` / MySQL dialect, or an existing MySQL connection config | MySQL | `templates/mysql/` |
| `*.sql` migrations using `BIGSERIAL`/PostgreSQL dialect, or an existing PostgreSQL connection config | PostgreSQL | `templates/postgresql/` |
| `.github/workflows/*` referencing `gcloud`/Cloud Run/Cloud Build | GCP | `templates/gcloud/` |
| `.github/workflows/*` referencing CodeBuild/CodePipeline, or `buildspec.yml` | AWS | `templates/aws/` |
| `.github/workflows/*` referencing `azure-pipelines.yml` or Azure DevOps | Azure | `templates/azure/` |
| No marker files (empty/new repository) | Ask the user which stack(s) to bootstrap, or default to the stack named in the user's request | — |

## 2️⃣ Common Setup Checklist

Run these steps regardless of detected stack:

- [ ] Copy `.editorconfig` and `.gitignore` from this boilerplate to the target repo root (do not overwrite if the target already customizes them — merge instead).
- [ ] Copy `.gitmessage` to the target repo root, then instruct the user to run `git config commit.template .gitmessage` (or run it yourself if you have shell access to the target repo).
- [ ] Copy the template directory (or directories) identified in Step 1 into the target repo, preserving their internal path structure (e.g. `templates/python/*` contents go to the target repo root, not into a `templates/` subfolder).
- [ ] Copy `templates/_shared/ci/downstream-ci.yml` to `.github/workflows/ci.yml` in the target repo, then uncomment only the job(s) matching the detected stack(s) and delete the rest.
- [ ] Do **not** copy this boilerplate's own `.github/workflows/ci.yml` (`repository-hygiene` job) — that job checks this boilerplate's own files, not the target repo's.
- [ ] Ask the user which collaboration language to use for commit messages, issues, and pull requests in the target repo (Korean-only, English-only, Japanese-only, or an explicit mix) — do not assume this boilerplate's English-language `.gitmessage` examples apply as-is. Record the decision in the target repo's `CONTRIBUTING.md`. See the Collaboration Language section in `docs/standards/GITHUB_STANDARDS.md` for the underlying rule.

## 3️⃣ Replace Placeholders

Search the copied files for these exact placeholder strings and replace them with real project values:

| Placeholder | Found in | Replace with |
| --- | --- | --- |
| `your-project-name` | `package.json` (`name`), `pyproject.toml` (`[project].name`) | the target repo's actual project/package name |
| `your-org/your-repo` | `go.mod` (`module`) | the target repo's actual module path, e.g. `github.com/<org>/<repo>` |
| `com.example` | `pom.xml` (`groupId`), Java package declarations under `src/main/java/com/example/...` and `src/test/java/com/example/...` | the target org's real Java group ID (rename both the `groupId` and the Java package directories/declarations together) |
| `your-service` | `pom.xml` (`artifactId`, `name`) | the target repo's actual service name |

## 4️⃣ Verify the Copied Skeleton Runs

Each stack template ships with a minimal example and test so the copy can be verified immediately. Run the command for each stack you copied:

- JavaScript/TypeScript/Node.js: `npm ci && npm run typecheck && npm test`
- Python: `pip install -e . pytest && pytest`
- Go: `go test ./...`
- Java/Spring: `mvn test`

If any command fails, fix the copy (usually a missed placeholder or a path mismatch) before considering setup complete.

## 5️⃣ Report Back

Summarize for the user: which stack(s) were detected, which files were copied, which placeholders were replaced (with old → new values), and whether verification (Step 4) passed.

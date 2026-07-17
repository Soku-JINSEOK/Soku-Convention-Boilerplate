# 🧰 Stack Configs

## 🎯 Purpose

This document maps the copyable starter configuration files that live under `templates/`.

The files in this area are meant to be dropped into downstream repositories and adapted, not treated as the only possible way to configure a stack.

## 🛠️ How To Use

Start with the template set for the stack you are using, then replace placeholder names, ports, package IDs, and service names with the real project values.

## 🧱 Shared Baseline

The following file applies across most stacks:

- [`.editorconfig`](../../.editorconfig)

## 🗂️ Directory Layout Convention

Each stack directory under `templates/` follows that language's own idiomatic layout rather than a single forced shape:

- JavaScript/TypeScript/Node.js and Python use `src/` + `test/`/`tests/`.
- Go uses a flat package layout (no `src/`), matching Go convention.
- Java/Spring uses the Maven-standard `src/main/java/...` and `src/test/java/...` layout.

This divergence is intentional. Cross-stack, non-language-specific templates (such as CI starters) live separately under `templates/_shared/` so they are not mistaken for a stack's own files.

## 🟨 JavaScript, TypeScript, Node.js

Template files:

- [`templates/javascript-typescript-node/package.json`](../../templates/javascript-typescript-node/package.json)
- [`templates/javascript-typescript-node/package-lock.json`](../../templates/javascript-typescript-node/package-lock.json)
- [`templates/javascript-typescript-node/tsconfig.json`](../../templates/javascript-typescript-node/tsconfig.json)
- [`templates/javascript-typescript-node/eslint.config.mjs`](../../templates/javascript-typescript-node/eslint.config.mjs)
- [`templates/javascript-typescript-node/prettier.config.cjs`](../../templates/javascript-typescript-node/prettier.config.cjs)
- [`templates/javascript-typescript-node/vitest.config.ts`](../../templates/javascript-typescript-node/vitest.config.ts)
- [`templates/javascript-typescript-node/src/profile.ts`](../../templates/javascript-typescript-node/src/profile.ts)
- [`templates/javascript-typescript-node/test/profile.test.ts`](../../templates/javascript-typescript-node/test/profile.test.ts)

The template includes a minimal TypeScript example and a Vitest test so a copied project can be validated immediately after dependency installation:

```bash
npm ci
npm run typecheck
npm test
```

## 🐍 Python

Template files:

- [`templates/python/pyproject.toml`](../../templates/python/pyproject.toml)
- [`templates/python/requirements-lock.txt`](../../templates/python/requirements-lock.txt)
- [`templates/python/src/user_profile.py`](../../templates/python/src/user_profile.py)
- [`templates/python/tests/test_user_profile.py`](../../templates/python/tests/test_user_profile.py)

The example module is named `user_profile.py`, not `profile.py` — `profile` is a Python standard library module name, and shadowing it breaks `import profile` once the project is installed (`pip install -e .` puts `src/` on `sys.path` directly per the src-layout convention, so top-level module names must not collide with the standard library).

Formatting uses [`pyink`](https://github.com/google/pyink) (Google's Black fork, 80-column) rather than `ruff format` — see the comment in `pyproject.toml`. The template includes a minimal example and a pytest test so a copied project can be validated immediately:

```bash
python -m venv .venv && source .venv/bin/activate
pip install -r requirements-lock.txt -e ".[dev]"
pyink --check .
ruff check .
mypy .
pytest
```

`requirements-lock.txt` pins the `dev` extra's transitive dependencies for reproducible installs (generated with `pip-compile --extra=dev --output-file=requirements-lock.txt pyproject.toml`, from the [`pip-tools`](https://github.com/jazzband/pip-tools) package). Regenerate it with that same command whenever `pyproject.toml`'s `dev` extra changes.

## 🐹 Go

Template files:

- [`templates/go/go.mod`](../../templates/go/go.mod)
- [`templates/go/.golangci.yml`](../../templates/go/.golangci.yml)
- [`templates/go/Makefile`](../../templates/go/Makefile)
- [`templates/go/profile.go`](../../templates/go/profile.go)
- [`templates/go/profile_test.go`](../../templates/go/profile_test.go)

The template includes a minimal example and a Go test so a copied project can be validated immediately:

```bash
go test ./...
```

## ☕ Java, Spring

Template files:

- [`templates/java-spring/pom.xml`](../../templates/java-spring/pom.xml)
- [`templates/java-spring/src/main/resources/application.yml`](../../templates/java-spring/src/main/resources/application.yml)
- [`templates/java-spring/src/main/java/com/example/profile`](../../templates/java-spring/src/main/java/com/example/profile)
- [`templates/java-spring/src/test/java/com/example/profile`](../../templates/java-spring/src/test/java/com/example/profile)

Checkstyle uses the official `google_checks.xml` bundled with `maven-checkstyle-plugin` (set via `configLocation` in `pom.xml`) instead of a hand-written ruleset. Formatting should use [`google-java-format`](https://github.com/google/google-java-format). The template includes a minimal example and a JUnit test so a copied project can be validated immediately:

```bash
mvn test
```

## 🗄️ Databases

Template files:

- [`templates/mysql/schema.sql`](../../templates/mysql/schema.sql)
- [`templates/postgresql/schema.sql`](../../templates/postgresql/schema.sql)

## ☁️ Cloud

Template files:

- [`templates/gcloud/cloudbuild.yaml`](../../templates/gcloud/cloudbuild.yaml)
- [`templates/aws/buildspec.yml`](../../templates/aws/buildspec.yml)
- [`templates/azure/azure-pipelines.yml`](../../templates/azure/azure-pipelines.yml)

## 🔗 Shared, Cross-Stack

Not tied to any single stack:

- [`templates/_shared/ci/downstream-ci.yml`](../../templates/_shared/ci/downstream-ci.yml): starter CI workflow with stack jobs commented out, meant to be copied to a downstream repo's `.github/workflows/ci.yml`
- [`templates/_shared/agents/`](../../templates/_shared/agents/): tool-agnostic domain agent charters for parallel AI agent ownership under the [Multi-Domain Layout](../standards/PROJECT_STRUCTURE.md#multi-domain-layout-alternative)

## 🎬 Summary

These files are starter-quality defaults.  
They are designed to make it easy to bootstrap a new project with familiar conventions, then refine the setup as the project matures.

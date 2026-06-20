# Stack Configs

## Purpose

This document maps the copyable starter configuration files that live under `templates/`.

The files in this area are meant to be dropped into downstream repositories and adapted, not treated as the only possible way to configure a stack.

## How To Use

Start with the template set for the stack you are using, then replace placeholder names, ports, package IDs, and service names with the real project values.

## Shared Baseline

The following file applies across most stacks:

- [`.editorconfig`](./.editorconfig)

## JavaScript, TypeScript, Node.js

Template files:

- [`templates/javascript-typescript-node/package.json`](./templates/javascript-typescript-node/package.json)
- [`templates/javascript-typescript-node/package-lock.json`](./templates/javascript-typescript-node/package-lock.json)
- [`templates/javascript-typescript-node/tsconfig.json`](./templates/javascript-typescript-node/tsconfig.json)
- [`templates/javascript-typescript-node/eslint.config.mjs`](./templates/javascript-typescript-node/eslint.config.mjs)
- [`templates/javascript-typescript-node/prettier.config.cjs`](./templates/javascript-typescript-node/prettier.config.cjs)
- [`templates/javascript-typescript-node/vitest.config.ts`](./templates/javascript-typescript-node/vitest.config.ts)

## Python

Template files:

- [`templates/python/pyproject.toml`](./templates/python/pyproject.toml)

## Go

Template files:

- [`templates/go/go.mod`](./templates/go/go.mod)
- [`templates/go/.golangci.yml`](./templates/go/.golangci.yml)
- [`templates/go/Makefile`](./templates/go/Makefile)

## Java, Spring

Template files:

- [`templates/java-spring/pom.xml`](./templates/java-spring/pom.xml)
- [`templates/java-spring/checkstyle.xml`](./templates/java-spring/checkstyle.xml)
- [`templates/java-spring/src/main/resources/application.yml`](./templates/java-spring/src/main/resources/application.yml)

## Databases

Template files:

- [`templates/mysql/schema.sql`](./templates/mysql/schema.sql)
- [`templates/postgresql/schema.sql`](./templates/postgresql/schema.sql)

## Cloud

Template files:

- [`templates/gcloud/cloudbuild.yaml`](./templates/gcloud/cloudbuild.yaml)
- [`templates/aws/buildspec.yml`](./templates/aws/buildspec.yml)
- [`templates/azure/azure-pipelines.yml`](./templates/azure/azure-pipelines.yml)

## Summary

These files are starter-quality defaults.  
They are designed to make it easy to bootstrap a new project with familiar conventions, then refine the setup as the project matures.

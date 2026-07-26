# npm Trusted Publishing

## Purpose

The `soku` npm package is published from GitHub Actions through npm Trusted
Publishing. The release workflow uses GitHub OIDC instead of a long-lived npm
automation token.

## Trusted Publisher Configuration

Configure the trusted publisher on the npm package settings page with this
exact identity:

- organization or user: `Soku-JINSEOK`
- repository: `Soku-Convention-Boilerplate`
- workflow filename: `release.yml`
- environment: leave empty unless the workflow is updated to use a reviewed
  GitHub environment
- allowed action: publish

The repository and workflow names are case-sensitive. The package
`repository.url` must remain
`https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate`.

## Workflow Contract

The `publish-npm` job must:

- run on a GitHub-hosted runner
- grant `contents: read` and `id-token: write`
- use Node.js 24 and the reviewed npm CLI version
- disable automatic package-manager caching
- run `npm publish --provenance --access public`
- never inject `NPM_TOKEN` or `NODE_AUTH_TOKEN`
- run only for a validated `soku/vMAJOR.MINOR.PATCH` tag in the canonical
  repository

Run `node scripts/verify-npm-publishing.mjs` to check this contract locally.
Repository CI runs the same verifier and its regression tests.

## Migration Gate

Do not create a release tag solely to test infrastructure. Use the next
reviewed CLI release:

1. Configure the trusted publisher on npm before pushing the release tag.
2. Publish the next reviewed CLI version through `release.yml`.
3. Confirm that the GitHub release workflow succeeds without a token.
4. Confirm that npm shows the expected package version and provenance.
5. Only after those checks pass, remove the obsolete GitHub `NPM_TOKEN` secret
   and revoke the corresponding npm automation token.

If the OIDC publication fails, keep the existing credential available while
the publisher identity is corrected. Restoring token-based publication
requires a reviewed workflow change; do not add an implicit fallback path.

## Change Control

Changes to the owner, repository, workflow filename, environment, npm CLI
version, or package repository URL require a pull request that updates:

- the npm trusted publisher configuration
- `.github/workflows/release.yml`
- `scripts/verify-npm-publishing.mjs`
- this guide

Credential retirement is an external, irreversible step and must be recorded
with the successful release evidence.

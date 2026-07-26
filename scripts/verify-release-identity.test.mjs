import assert from 'node:assert/strict';
import test from 'node:test';

import {verifyReleaseIdentity} from './verify-release-identity.mjs';

const identity = {
  schemaVersion: 1,
  boilerplate: {
    version: '1.0.5',
    tag: 'v1.0.5',
    releaseNotes: 'docs/releases/v1.0.5.md',
  },
  cli: {
    version: '0.2.1',
    tag: 'soku/v0.2.1',
    releaseNotes: 'docs/releases/soku-v0.2.1.md',
  },
  npm: {
    package: '@soku-jinseok/soku',
    version: '0.2.1',
  },
  compatibility: {
    boilerplateTag: 'v1.0.5',
    cliTag: 'soku/v0.1.4',
    npmSupportedFromCliTag: 'soku/v0.2.0',
  },
};

const currentIdentity =
  'v1.0.5 soku/v0.2.1 npm install -g @soku-jinseok/soku@0.2.1 soku/v0.1.4';
const files = {
  'README.md': currentIdentity,
  'README.ko.md': currentIdentity,
  'README.ja.md': currentIdentity,
  'soku/README.md': currentIdentity,
  'soku/npm/package.json': JSON.stringify({
    name: '@soku-jinseok/soku',
    version: '0.2.1',
  }),
  '.github/workflows/release.yml': `
    boilerplate-tag:
      default: v1.0.5
    cli-tag:
      default: soku/v0.2.1
  `,
  'docs/guides/USAGE_MANUAL.md': currentIdentity,
  'docs/releases/v1.0.5.md': `
# Boilerplate v1.0.5
Compatible soku: soku/v0.1.4
  `,
  'docs/releases/soku-v0.2.1.md': `
# soku v0.2.1
Boilerplate compatibility: v1.0.5
execution of soku/v0.2.1
  `,
};

const verify = (nextIdentity = identity, nextFiles = files) =>
  verifyReleaseIdentity(nextIdentity, nextFiles);

test('accepts one aligned release identity', () => {
  assert.deepEqual(verify(), []);
});

test('rejects mismatched tag and package versions', () => {
  const drifted = structuredClone(identity);
  drifted.cli.version = '0.2.2';
  const errors = verify(drifted).join('\n');

  assert.match(errors, /CLI tag and version must match/);
  assert.match(errors, /npm and CLI versions must match/);
});

test('rejects npm package metadata drift', () => {
  const drifted = {
    ...files,
    'soku/npm/package.json': '{"name":"wrong","version":"9.9.9"}',
  };
  const errors = verify(identity, drifted).join('\n');

  assert.match(errors, /npm package name has drifted/);
  assert.match(errors, /npm package version has drifted/);
});

test('rejects workflow default drift', () => {
  const drifted = {
    ...files,
    '.github/workflows/release.yml': `
      boilerplate-tag:
        default: v9.9.9
      cli-tag:
        default: soku/v9.9.9
    `,
  };

  assert.match(verify(identity, drifted).join('\n'), /workflow .* drifted/);
});

test('rejects documentation and release-note drift', () => {
  const drifted = {
    ...files,
    'README.md': 'stale',
    'docs/releases/v1.0.5.md': '# Boilerplate v1.0.5',
  };
  const errors = verify(identity, drifted).join('\n');

  assert.match(errors, /README\.md is missing release identity/);
  assert.match(errors, /Compatible soku/);
});

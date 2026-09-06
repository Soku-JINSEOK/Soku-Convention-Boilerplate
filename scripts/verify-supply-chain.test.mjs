import assert from 'node:assert/strict';
import {dirname, resolve} from 'node:path';
import test from 'node:test';
import {fileURLToPath} from 'node:url';

import {
  inspectContent,
  verifyDependabotCoverage,
  verifyRepository,
} from './verify-supply-chain.mjs';

test('accepts immutable action, image, and tool references', () => {
  const content = [
    'uses: actions/checkout@0123456789abcdef0123456789abcdef01234567',
    'image: mysql:8.4.10@sha256:' + 'a'.repeat(64),
    'FROM alpine:3.21.7@sha256:' + 'b'.repeat(64),
    'run: go install example.com/tool@v1.2.3',
    'run: npx --yes yaml-lint@1.7.0 file.yml',
  ].join('\n');

  assert.deepEqual(inspectContent('fixture.yml', content), []);
});

test('rejects floating latest references', () => {
  const findings = inspectContent(
    'fixture.yml',
    'run: go install example.com/tool@latest',
  );

  assert.ok(findings.some(({rule}) => rule === 'floating-latest'));
});

test('rejects mutable action tags', () => {
  const findings = inspectContent(
    'fixture.yml',
    'uses: actions/checkout@v7',
  );

  assert.ok(findings.some(({rule}) => rule === 'action-sha'));
});

test('rejects container images without a digest', () => {
  const findings = inspectContent('fixture.yml', 'image: postgres:16.14');

  assert.ok(findings.some(({rule}) => rule === 'image-digest'));
});

test('enforces immutable Cloud Build builder images', () => {
  const accepted = inspectContent(
    'cloudbuild/validation.yaml',
    'name: node:24.17.0@sha256:' + 'c'.repeat(64),
  );
  const rejected = inspectContent(
    'cloudbuild/validation.yaml',
    'name: node:24.17.0',
  );

  assert.deepEqual(accepted, []);
  assert.ok(rejected.some(({rule}) => rule === 'image-digest'));
});

test('rejects unversioned executable tools', () => {
  const findings = inspectContent(
    'fixture.yml',
    ['run: go install example.com/tool', 'run: npx --yes yaml-lint'].join(
      '\n',
    ),
  );

  assert.ok(findings.some(({rule}) => rule === 'go-tool-version'));
  assert.ok(findings.some(({rule}) => rule === 'npx-tool-version'));
});

test('reports missing dependency update coverage', () => {
  const findings = verifyDependabotCoverage(`
version: 2
updates:
  - package-ecosystem: github-actions
    directory: /
`);

  assert.ok(findings.some(({message}) => message.includes('/soku/npm')));
  assert.ok(findings.some(({message}) => message.includes('/soku/internal/manual/assets/runner')));
  assert.ok(findings.some(({message}) => message.includes('/infra/gcp')));
});

test('current repository satisfies the immutable supply-chain contract', () => {
  const root = resolve(dirname(fileURLToPath(import.meta.url)), '..');
  const result = verifyRepository(root);

  assert.deepEqual(result.findings, []);
});

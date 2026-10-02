import assert from 'node:assert/strict';
import {dirname, resolve} from 'node:path';
import {mkdtempSync, mkdirSync, readFileSync, writeFileSync, rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import test from 'node:test';
import {fileURLToPath} from 'node:url';

import {
  inspectContent,
  renderToolsPolicy,
  verifyPolicyDocumentation,
  verifyPythonSupport,
  verifyRuntimeSupport,
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

test('policy documentation rejects a stale threshold after a source change', () => {
  const tools = new Map([['NPM_AUDIT_LEVEL', 'high']]);
  const rendered = renderToolsPolicy(tools);
  assert.deepEqual(verifyPolicyDocumentation(rendered, tools), []);
  assert.deepEqual(verifyPolicyDocumentation(rendered.replaceAll('\n', '\r\n'), tools), []);
  tools.set('NPM_AUDIT_LEVEL', 'critical');
  assert.equal(verifyPolicyDocumentation(rendered, tools)[0].rule, 'policy-documentation');
  assert.deepEqual(verifyPolicyDocumentation(renderToolsPolicy(tools), tools), []);
});

test('Python support rejects untested minimum, new upper bound, and floating ranges', () => {
  const workflow = "python-version: ['3.11', '3.12', '3.13', '3.14']\n" +
    'python-version: ${{ matrix.python-version }}\nfail-fast: false';
  assert.deepEqual(verifyPythonSupport('requires-python = ">=3.11,<3.15"', workflow), []);
  for (const range of ['>=3.10,<3.15', '>=3.11,<3.16', '>=3.11']) {
    assert.equal(verifyPythonSupport(`requires-python = "${range}"`, workflow)[0].rule, 'python-support');
  }
  assert.equal(verifyPythonSupport('requires-python = ">=3.11,<3.15"', workflow.replace('fail-fast: false', 'fail-fast: true')).length, 1);
});

test('runtime parity catches declaration and matrix drift', () => {
  const root = resolve(dirname(fileURLToPath(import.meta.url)), '..');
  const fixture = mkdtempSync(resolve(tmpdir(), 'soku-runtime-contract-'));
  const paths = [
    'templates/javascript-typescript-node/package.json',
    'templates/javascript-typescript-node/package-lock.json',
    'templates/go/go.mod', 'templates/java-spring/pom.xml',
    '.github/workflows/templates-ci.yml', 'templates/_shared/ci/downstream-ci.yml',
    'scripts/verify.sh', 'infra/gcp/versions.tf', 'infra/gcp/cloud-build-logging/versions.tf',
  ];
  try {
    for (const path of paths) {
      mkdirSync(dirname(resolve(fixture, path)), {recursive: true});
      writeFileSync(resolve(fixture, path), readFileSync(resolve(root, path)));
    }
    assert.deepEqual(verifyRuntimeSupport(fixture), []);
    for (const [path, before, after] of [
      ['templates/javascript-typescript-node/package.json', '>=22.12.0', '>=22.0.0'],
      ['templates/go/go.mod', 'go 1.26', 'go 1.27'],
      ['templates/java-spring/pom.xml', '<java.version>21', '<java.version>25'],
      ['infra/gcp/versions.tf', '>= 1.15.3', '>= 1.8.0'],
    ]) {
      const original = readFileSync(resolve(fixture, path), 'utf8');
      assert.ok(original.includes(before), path);
      writeFileSync(resolve(fixture, path), original.replace(before, after));
      assert.ok(verifyRuntimeSupport(fixture).length > 0, path);
      writeFileSync(resolve(fixture, path), original);
    }
  } finally {
    rmSync(fixture, {recursive: true, force: true});
  }
});

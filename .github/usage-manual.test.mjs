import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import test from 'node:test';

const read = (path) => readFileSync(new URL(`../${path}`, import.meta.url), 'utf8');
const escapeRegExp = (value) => value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');

const manual = read('docs/guides/USAGE_MANUAL.md');
const cliReadme = read('soku/README.md');
const releaseWorkflow = read('.github/workflows/release.yml');
const commandSource = read('soku/internal/cli/command.go');
const catalogIndex = JSON.parse(read('soku/catalog/index-v2.json'));
const coreCatalog = JSON.parse(read('soku/catalog/core-v1.json'));

test('usage manual follows the published release workflow defaults', () => {
  const boilerplate = releaseWorkflow.match(
    /boilerplate-tag:[\s\S]*?default:\s*([^\s]+)/,
  )?.[1];
  const cli = releaseWorkflow.match(/cli-tag:[\s\S]*?default:\s*([^\s]+)/)?.[1];

  assert.equal(boilerplate, 'v1.0.5');
  assert.equal(cli, 'soku/v0.1.4');
  assert.match(manual, new RegExp('boilerplate `' + escapeRegExp(boilerplate) + '`'));
  assert.match(manual, new RegExp('CLI `' + escapeRegExp(cli) + '`'));
  assert.match(cliReadme, /recommended full-verification baseline[\s\S]*`v1\.0\.5`[\s\S]*`soku\/v0\.1\.4`/);
  assert.match(manual, /checksums\.txt/);
  assert.match(manual, /sha256sum --check/);
});

test('usage manual names every built-in profile and stack ID', () => {
  for (const profile of catalogIndex.profiles.map(({id}) => id)) {
    assert.match(manual, new RegExp('`' + escapeRegExp(profile) + '`'));
  }

  for (const stack of coreCatalog.stacks.map(({id}) => id)) {
    assert.match(manual, new RegExp('`' + escapeRegExp(stack) + '`'));
  }
});

test('usage manual covers the implemented public lifecycle commands', () => {
  const commands = [
    ...commandSource.matchAll(/newLifecycleCommand\("([^"]+)"/g),
  ].map((match) => match[1]);

  assert.deepEqual(commands, ['init', 'status', 'diff', 'upgrade']);
  for (const command of commands) {
    assert.match(manual, new RegExp(`soku ${escapeRegExp(command)}(?: |\\n)`));
  }

  for (const flag of ['--dry-run', '--verify', '--yes']) {
    assert.match(manual, new RegExp(escapeRegExp(flag)));
  }
});

test('all requested discovery entrypoints link to the same manual', () => {
  for (const path of [
    'README.md',
    'README.ko.md',
    'README.ja.md',
    'BLUEPRINT.md',
    'docs/guides/APPLICABILITY.md',
    'soku/README.md',
  ]) {
    assert.match(read(path), /USAGE_MANUAL\.md/, path);
  }
});

test('manual delegates detailed rules to authoritative documents', () => {
  for (const target of [
    'INIT_GUIDE.md',
    'SOKU_LIFECYCLE.md',
    'RELEASE_AND_SYNC.md',
    'GITHUB_STANDARDS.md',
    'CLOUD_RUN_CICD.md',
    'VERIFICATION_GUIDE.md',
  ]) {
    assert.match(manual, new RegExp(escapeRegExp(target)), target);
  }
});

import assert from 'node:assert/strict';
import {
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from 'node:fs';
import {tmpdir} from 'node:os';
import {join, resolve} from 'node:path';
import {spawnSync} from 'node:child_process';
import test from 'node:test';
import {
  detectScopes,
  parseFilesInput,
  parseNameStatus,
} from './detect-verification-scope.mjs';

const root = resolve(new URL('..', import.meta.url).pathname);
const detector = resolve(root, 'scripts/detect-verification-scope.mjs');
const config = JSON.parse(
  readFileSync(resolve(root, 'verification/scopes.yml'), 'utf8'),
);
const profiles = JSON.parse(
  readFileSync(resolve(root, 'verification/profiles.yml'), 'utf8'),
);

function git(workspace, ...args) {
  const result = spawnSync('git', args, {cwd: workspace, encoding: 'utf8'});
  assert.equal(result.status, 0, result.stderr);
  return result.stdout.trim();
}

function run(workspace, args, input) {
  return spawnSync(process.execPath, [detector, ...args], {
    cwd: workspace,
    encoding: 'utf8',
    input,
  });
}

function repository() {
  const workspace = mkdtempSync(join(tmpdir(), 'verification-scope-'));
  git(workspace, 'init', '-q');
  git(workspace, 'config', 'user.name', 'Scope Test');
  git(workspace, 'config', 'user.email', 'scope@example.invalid');
  git(workspace, 'config', 'commit.gpgsign', 'false');
  writeFileSync(join(workspace, 'README.md'), 'baseline\n');
  git(workspace, 'add', 'README.md');
  git(workspace, 'commit', '-qm', 'baseline');
  return workspace;
}

test('maps one template path to only its runtime scope', () => {
  const result = detectScopes(
    ['templates/javascript-typescript-node/src/profile.ts'],
    config,
  );
  assert.deepEqual(result.scopes, ['javascript-typescript-node']);
  assert.equal(result.allSelected, false);
});

test('keeps detector scopes aligned with the fast profile', () => {
  assert.deepEqual(profiles.profiles.fast.scopes, config.allScopes);
  assert.equal(profiles.schemaVersion, config.schemaVersion);
});

test('shared, provider, and unknown paths fail closed to every scope', () => {
  for (const path of [
    'verification/tools.env',
    'soku/schema/catalog-core-v1.schema.json',
    'unclassified/new-system.conf',
  ]) {
    const result = detectScopes([path], config);
    assert.deepEqual(result.scopes, config.allScopes);
    assert.equal(result.allSelected, true);
  }
});

test('documentation changes retain always-on checks without runtime scopes', () => {
  const result = detectScopes(['docs/guides/USAGE_MANUAL.md'], config);
  assert.deepEqual(result.scopes, []);
  assert.equal(result.reasons[0].rule, 'repository-documentation');
});

test('parses rename and delete name-status inputs', () => {
  const parsed = parseNameStatus(
    Buffer.from('R100\0old/path.go\0new/path.go\0D\0removed.txt\0'),
  );
  assert.deepEqual(parsed, ['old/path.go', 'new/path.go', 'removed.txt']);
  assert.deepEqual(
    parseFilesInput(
      Buffer.from('R090\told.py\tnew.py\nD\tdeleted.yml\nplain.md\n'),
    ),
    ['old.py', 'new.py', 'deleted.yml', 'plain.md'],
  );
});

test('defaults to staged changes and emits the versioned JSON contract', () => {
  const workspace = repository();
  try {
    writeFileSync(join(workspace, 'README.md'), 'staged\n');
    git(workspace, 'add', 'README.md');
    const result = run(workspace, ['--json']);
    assert.equal(result.status, 0, result.stderr);
    assert.deepEqual(Object.keys(JSON.parse(result.stdout)), [
      'schemaVersion',
      'changedFiles',
      'scopes',
      'reasons',
      'allSelected',
    ]);
    assert.deepEqual(JSON.parse(result.stdout).changedFiles, ['README.md']);
  } finally {
    rmSync(workspace, {recursive: true, force: true});
  }
});

test('supports an explicit base/head commit range', () => {
  const workspace = repository();
  try {
    const base = git(workspace, 'rev-parse', 'HEAD');
    writeFileSync(join(workspace, 'README.md'), 'range\n');
    git(workspace, 'commit', '-qam', 'range');
    const head = git(workspace, 'rev-parse', 'HEAD');
    const result = run(workspace, [
      '--base',
      base,
      '--head',
      head,
      '--json',
    ]);
    assert.equal(result.status, 0, result.stderr);
    assert.deepEqual(JSON.parse(result.stdout).changedFiles, ['README.md']);
  } finally {
    rmSync(workspace, {recursive: true, force: true});
  }
});

test('supports file and stdin input, including rename/delete records', () => {
  const workspace = repository();
  try {
    writeFileSync(
      join(workspace, 'changes.txt'),
      'R100\ttemplates/go/old.go\ttemplates/go/new.go\n' +
        'D\ttemplates/python/src/removed.py\n',
    );
    const fromFile = run(workspace, [
      '--files-from',
      'changes.txt',
      '--json',
    ]);
    assert.equal(fromFile.status, 0, fromFile.stderr);
    assert.deepEqual(JSON.parse(fromFile.stdout).scopes, ['python', 'go']);

    const fromStdin = run(
      workspace,
      ['--files-from', '-', '--json'],
      'templates/java-spring/pom.xml\n',
    );
    assert.equal(fromStdin.status, 0, fromStdin.stderr);
    assert.deepEqual(JSON.parse(fromStdin.stdout).scopes, ['java-spring']);
  } finally {
    rmSync(workspace, {recursive: true, force: true});
  }
});

test('recognizes staged rename/delete changes and an empty staged diff', () => {
  const workspace = repository();
  try {
    const empty = run(workspace, ['--staged', '--json']);
    assert.equal(empty.status, 0, empty.stderr);
    assert.deepEqual(JSON.parse(empty.stdout).changedFiles, []);

    git(workspace, 'mv', 'README.md', 'README-renamed.md');
    const renamed = run(workspace, ['--staged', '--json']);
    assert.equal(renamed.status, 0, renamed.stderr);
    assert.deepEqual(JSON.parse(renamed.stdout).changedFiles, [
      'README.md',
      'README-renamed.md',
    ]);

    git(workspace, 'reset', '--hard', '-q', 'HEAD');
    rmSync(join(workspace, 'README.md'));
    git(workspace, 'add', '-u');
    const deleted = run(workspace, ['--staged', '--json']);
    assert.equal(deleted.status, 0, deleted.stderr);
    assert.deepEqual(JSON.parse(deleted.stdout).changedFiles, ['README.md']);
  } finally {
    rmSync(workspace, {recursive: true, force: true});
  }
});

test('rejects incomplete ranges and conflicting input modes', () => {
  const workspace = repository();
  try {
    for (const args of [
      ['--base', 'a'.repeat(40)],
      ['--staged', '--files-from', '-'],
    ]) {
      const result = run(workspace, args);
      assert.equal(result.status, 2);
    }
  } finally {
    rmSync(workspace, {recursive: true, force: true});
  }
});

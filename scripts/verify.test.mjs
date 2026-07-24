import assert from 'node:assert/strict';
import {resolve} from 'node:path';
import {spawnSync} from 'node:child_process';
import test from 'node:test';

const root = resolve(new URL('..', import.meta.url).pathname);

function run(script, args) {
  return spawnSync('bash', [resolve(root, 'scripts', script), ...args], {
    cwd: root,
    encoding: 'utf8',
  });
}

test('verify.sh --help documents the full profile and exits 0', () => {
  const result = run('verify.sh', ['--help']);
  assert.equal(result.status, 0, result.stderr);
  assert.match(result.stdout, /--profile <name>/);
  assert.match(result.stdout, /\bfull\b/);
});

test('verify.sh requires --profile', () => {
  const result = run('verify.sh', []);
  assert.equal(result.status, 2);
  assert.match(result.stderr, /--profile is required/);
});

test('verify.sh rejects unknown profiles', () => {
  const result = run('verify.sh', ['--profile', 'nonsense']);
  assert.equal(result.status, 2);
  assert.match(result.stderr, /unknown profile 'nonsense'/);
});

for (const profile of ['fast', 'ci-quick', 'hosted-full', 'release', 'deploy']) {
  test(`verify.sh fails loudly for the not-yet-implemented '${profile}' profile`, () => {
    const result = run('verify.sh', ['--profile', profile]);
    assert.equal(result.status, 3);
    assert.match(result.stderr, /not yet implemented/);
  });
}

test('ci-local.sh delegates to verify.sh --profile full', () => {
  const wrapper = run('ci-local.sh', ['--help']);
  const target = run('verify.sh', ['--help']);
  assert.equal(wrapper.status, target.status);
  assert.equal(wrapper.stdout, target.stdout);
});

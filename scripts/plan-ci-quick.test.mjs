import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {resolve} from 'node:path';
import test from 'node:test';
import {planQuickGroups} from './plan-ci-quick.mjs';

const root = resolve(new URL('..', import.meta.url).pathname);
const profiles = JSON.parse(
  readFileSync(resolve(root, 'verification/profiles.yml'), 'utf8'),
);

function plan(scopes) {
  return planQuickGroups({schemaVersion: 1, scopes}, profiles).include;
}

test('always schedules the repository shard', () => {
  assert.deepEqual(plan([]), [
    {
      id: 'always',
      name: 'Always-on repository checks',
      toolchains: ['node', 'go'],
    },
  ]);
});

test('maps selected scopes to independent runtime and service shards', () => {
  assert.deepEqual(
    plan([
      'soku',
      'javascript-typescript-node',
      'mysql',
      'postgresql',
      'gcloud',
      'infra-gcp',
    ]).map(({id}) => id),
    [
      'always',
      'soku',
      'javascript-typescript-node',
      'database',
      'cloud',
      'infra-gcp',
    ],
  );
});

test('coalesces scopes that share one shard', () => {
  const groups = plan(['mysql', 'postgresql', 'gcloud', 'cloud-config']);
  assert.equal(groups.filter(({id}) => id === 'database').length, 1);
  assert.equal(groups.filter(({id}) => id === 'cloud').length, 1);
});

test('fails closed for detector scopes outside the profile', () => {
  assert.throws(() => plan(['unknown']), /unknown ci-quick scopes/);
});

test('rejects invalid group toolchains and duplicate IDs', () => {
  const invalid = structuredClone(profiles);
  invalid.profiles['ci-quick'].groups[0].toolchains.push('ruby');
  assert.throws(
    () => planQuickGroups({schemaVersion: 1, scopes: []}, invalid),
    /invalid ci-quick group/,
  );

  const duplicate = structuredClone(profiles);
  duplicate.profiles['ci-quick'].groups[1].id = 'always';
  assert.throws(
    () => planQuickGroups({schemaVersion: 1, scopes: []}, duplicate),
    /duplicate ci-quick group/,
  );
});

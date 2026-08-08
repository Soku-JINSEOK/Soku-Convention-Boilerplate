import assert from 'node:assert/strict';
import test from 'node:test';
import {verifyPromotion} from './verify-image-promotion.mjs';

const repository = 'Soku-JINSEOK/Soku-Convention-Boilerplate';
const sourceSha = 'a'.repeat(40);
const digest = `sha256:${'b'.repeat(64)}`;
const manifest = {
  schemaVersion: 1,
  repository,
  sourceSha,
  sourceRef: 'refs/heads/main',
  workflowRunId: '1234',
  imageUri: `asia-docker.pkg.dev/project/repository/service@${digest}`,
  digest,
  createdAt: '2026-07-26T00:00:00Z',
};
const run = {
  id: 1234,
  event: 'push',
  head_branch: 'main',
  head_sha: sourceSha,
  conclusion: 'success',
  path: '.github/workflows/validation.yml',
  repository: {full_name: repository},
  head_repository: {full_name: repository},
};

test('accepts a successful canonical main Validation artifact', () => {
  assert.deepEqual(verifyPromotion({manifest, run, repository}), {
    sourceSha,
    imageUri: manifest.imageUri,
    digest,
  });
});

for (const [name, mutate, message] of [
  ['run id', ({manifest: value}) => { value.workflowRunId = '999'; }, /run ID mismatch/],
  ['source sha', ({run: value}) => { value.head_sha = 'c'.repeat(40); }, /source SHA mismatch/],
  ['digest', ({manifest: value}) => { value.digest = `sha256:${'d'.repeat(64)}`; }, /image URI and digest mismatch/],
  ['event', ({run: value}) => { value.event = 'pull_request'; }, /not a push/],
  ['branch', ({run: value}) => { value.head_branch = 'feature'; }, /not on main/],
  ['conclusion', ({run: value}) => { value.conclusion = 'failure'; }, /did not succeed/],
  ['workflow', ({run: value}) => { value.path = '.github/workflows/release.yml'; }, /workflow mismatch/],
  ['repository', ({run: value}) => { value.head_repository.full_name = 'fork/repo'; }, /head repository mismatch/],
]) {
  test(`rejects ${name} mismatch`, () => {
    const values = {
      manifest: structuredClone(manifest),
      run: structuredClone(run),
    };
    mutate(values);
    assert.throws(
      () => verifyPromotion({...values, repository}),
      message,
    );
  });
}

import assert from 'node:assert/strict';
import test from 'node:test';
import {
  buildAuthenticatedEvent,
  validatePullRequestIdentity,
} from './authenticated-pr-metadata.mjs';

function fixture() {
  const repository = 'owner/repository';
  const headRepository = 'contributor/repository';
  const headSha = 'a'.repeat(40);
  const pullRequest = {
    number: 19,
    head: {sha: headSha, repo: {full_name: headRepository}},
    base: {repo: {full_name: repository}},
  };
  return {
    event: {
      repository: {full_name: repository},
      pull_request: structuredClone(pullRequest),
    },
    current: structuredClone(pullRequest),
    files: ['README.md'],
    trusted: {
      repository,
      pullRequestNumber: '19',
      headRepository,
      headSha,
    },
  };
}

test('builds an authenticated policy event', () => {
  const input = fixture();
  assert.deepEqual(
    buildAuthenticatedEvent(input).pull_request.changed_files_list,
    ['README.md'],
  );
});

test('rejects repository, number, head repository, and SHA mismatches', () => {
  for (const mutate of [
    (input) => { input.current.base.repo.full_name = 'other/repository'; },
    (input) => { input.current.number = 20; },
    (input) => { input.current.head.repo.full_name = 'attacker/repository'; },
    (input) => { input.current.head.sha = 'b'.repeat(40); },
  ]) {
    const input = fixture();
    mutate(input);
    assert.throws(() => buildAuthenticatedEvent(input), /does not match/);
  }
});

test('rejects untrusted event values and invalid file metadata', () => {
  const input = fixture();
  input.event.pull_request.head.sha = 'b'.repeat(40);
  assert.ok(validatePullRequestIdentity({
    event: input.event,
    current: input.current,
    ...input.trusted,
  }).length > 0);
  const invalid = fixture();
  invalid.files = [{filename: 'README.md'}];
  assert.throws(() => buildAuthenticatedEvent(invalid), /array of strings/);
});

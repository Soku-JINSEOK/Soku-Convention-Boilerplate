import assert from 'node:assert/strict';
import {mkdtempSync, rmSync, writeFileSync} from 'node:fs';
import {join} from 'node:path';
import {tmpdir} from 'node:os';
import test from 'node:test';

import {
  GitHubApiClient,
  parseArgs,
  readSyncConfig,
  relationNumbers,
} from './github-project-sync.mjs';

test('installed Project Sync configuration is portable and opt-in', () => {
  const directory = mkdtempSync(join(tmpdir(), 'soku-project-sync-'));
  const configPath = join(directory, 'project-sync.yml');
  writeFileSync(configPath, JSON.stringify({
    schemaVersion: 1,
    project: {
      owner: '@me',
      number: 17,
      fields: {
        status: 'Status',
        priority: 'Priority',
        size: 'Size',
        workstream: 'Workstream',
        targetDate: 'Target date',
      },
    },
    backfill: {relationMappings: {}, dependencyTrackingIssue: null},
  }));
  try {
    const config = readSyncConfig(configPath);
    assert.equal(config.repository, undefined);
    assert.equal(config.project.owner, '@me');
    assert.equal(config.project.number, 17);
    assert.deepEqual(config.project.fields, {
      status: 'Status',
      priority: 'Priority',
      size: 'Size',
      workstream: 'Workstream',
      targetDate: 'Target date',
    });
    assert.deepEqual(config.backfill.relationMappings, {});
    assert.equal(config.backfill.dependencyTrackingIssue, null);
  } finally {
    rmSync(directory, {recursive: true, force: true});
  }
});

test('runtime accepts repository and positive Project selectors without API access', () => {
  const parsed = parseArgs([
    '--mode', 'audit',
    '--repo', 'owner/repository',
    '--project-owner', '@me',
    '--project-number', '1',
  ]);
  assert.equal(parsed.mode, 'audit');
  assert.equal(parsed.projectNumber, 1);
  assert.deepEqual(relationNumbers('Related to #7'), [7]);
});

function apiResponse(status, data, headers = {}) {
  return {ok: status >= 200 && status < 300, status, headers,
    text: async () => JSON.stringify(data)};
}

test('serializes concurrent GitHub requests including reads', async () => {
  let active = 0;
  let peak = 0;
  const client = new GitHubApiClient({
    token: 'test-token',
    fetchImpl: async () => {
      active += 1;
      peak = Math.max(peak, active);
      await Promise.resolve();
      active -= 1;
      return apiResponse(200, {ok: true});
    },
  });
  await Promise.all(['/one', '/two', '/three'].map(path => client.rest('GET', path)));
  assert.equal(peak, 1);
});

test('retries secondary-limit GET responses with bounded exponential waits', async () => {
  let attempts = 0;
  const waits = [];
  const client = new GitHubApiClient({
    token: 'test-token',
    wait: async milliseconds => waits.push(milliseconds),
    fetchImpl: async () => {
      attempts += 1;
      return attempts < 3
        ? apiResponse(403, {message: 'You have exceeded a secondary rate limit.'})
        : apiResponse(200, {recovered: true});
    },
  });
  assert.deepEqual(await client.rest('GET', '/items'), {recovered: true});
  assert.deepEqual(waits, [60_000, 120_000]);
  assert.equal(attempts, 3);
});

test('honors Retry-After seconds and HTTP dates before retrying', async () => {
  for (const retryAfter of ['90', 'Thu, 01 Jan 1970 00:01:30 GMT']) {
    let attempts = 0;
    const waits = [];
    const client = new GitHubApiClient({
      token: 'test-token', now: () => 0,
      wait: async milliseconds => waits.push(milliseconds),
      fetchImpl: async () => ++attempts === 1
        ? apiResponse(429, {message: 'rate limited'}, {'retry-after': retryAfter})
        : apiResponse(200, []),
    });
    await client.rest('GET', '/items');
    assert.deepEqual(waits, [90_000]);
  }
});

test('honors the primary reset time and refuses waits outside the budget', async () => {
  for (const reset of ['90', '3600']) {
    let attempts = 0;
    const waits = [];
    const client = new GitHubApiClient({
      token: 'test-token', now: () => 0,
      wait: async milliseconds => waits.push(milliseconds),
      fetchImpl: async () => ++attempts === 1
        ? apiResponse(403, {message: 'API rate limit exceeded'},
          {'x-ratelimit-remaining': '0', 'x-ratelimit-reset': reset})
        : apiResponse(200, []),
    });
    if (reset === '90') {
      await client.rest('GET', '/items');
      assert.deepEqual(waits, [90_000]);
    } else {
      await assert.rejects(() => client.rest('GET', '/items'), /rate limit exceeded/);
      assert.equal(attempts, 1);
      assert.deepEqual(waits, []);
    }
  }
});

test('fails closed after the bounded retry count', async () => {
  let attempts = 0;
  const waits = [];
  const client = new GitHubApiClient({
    token: 'test-token',
    wait: async milliseconds => waits.push(milliseconds),
    fetchImpl: async () => {
      attempts += 1;
      return apiResponse(429, {message: 'rate limited'});
    },
  });
  await assert.rejects(() => client.rest('GET', '/items'), /failed \(429\)/);
  assert.equal(attempts, 3);
  assert.deepEqual(waits, [60_000, 120_000]);
});

test('does not replay mutations or ordinary authorization failures', async () => {
  for (const [method, status, message] of [
    ['POST', 429, 'rate limited'],
    ['PATCH', 403, 'secondary rate limit'],
    ['GET', 401, 'Bad credentials'],
    ['GET', 403, 'Resource not accessible by integration'],
  ]) {
    let attempts = 0;
    const waits = [];
    const client = new GitHubApiClient({
      token: 'test-token',
      wait: async milliseconds => waits.push(milliseconds),
      fetchImpl: async () => {
        attempts += 1;
        return apiResponse(status, {message});
      },
    });
    await assert.rejects(() => client.rest(method, '/items'), /failed/);
    assert.equal(attempts, 1);
    assert.deepEqual(waits, []);
  }
});

test('a failed request does not poison the serialized queue', async () => {
  let attempts = 0;
  const client = new GitHubApiClient({
    token: 'test-token',
    fetchImpl: async () => ++attempts === 1
      ? apiResponse(403, {message: 'Permission denied'})
      : apiResponse(200, {ok: true}),
  });
  await assert.rejects(() => client.rest('GET', '/denied'), /Permission denied/);
  assert.deepEqual(await client.rest('GET', '/allowed'), {ok: true});
});

import assert from 'node:assert/strict';
import {mkdtempSync, rmSync, writeFileSync} from 'node:fs';
import {join} from 'node:path';
import {tmpdir} from 'node:os';
import test from 'node:test';

import {
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

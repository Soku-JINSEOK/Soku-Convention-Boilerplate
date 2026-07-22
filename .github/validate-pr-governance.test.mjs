import assert from 'node:assert/strict';
import {spawnSync} from 'node:child_process';
import {mkdtempSync, writeFileSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join} from 'node:path';
import test from 'node:test';
import {fileURLToPath} from 'node:url';

const validator = fileURLToPath(
  new URL('./validate-pr-governance.mjs', import.meta.url),
);

function validPullRequest() {
  return {
    body: `## 🔗 Common Metadata

- **Issue:** Related to #55
- **Task report:** docs/issues/issue-55-task-report.md
- **Governance profile:** boilerplate

## 🇬🇧 English — Normative Source

### 🧪 Verification

- [x] Node regression tests passed

## 🇰🇷 한국어 요약

Governance verification.

## 🇯🇵 日本語の要約

Governance verification.

## Gitmoji Checklist

- [x] 🔧 Chore

## 🤖 AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex
`,
    labels: [{name: 'type:chore'}, {name: 'area:ci'}],
  };
}

function runValidator(pullRequest) {
  const directory = mkdtempSync(join(tmpdir(), 'pr-governance-'));
  const eventPath = join(directory, 'event.json');
  writeFileSync(eventPath, JSON.stringify({pull_request: pullRequest}));
  return spawnSync(process.execPath, [validator], {
    encoding: 'utf8',
    env: {...process.env, GITHUB_EVENT_PATH: eventPath},
  });
}

test('accepts a complete governed pull request', () => {
  const result = runValidator(validPullRequest());
  assert.equal(result.status, 0, result.stderr);
  assert.match(result.stdout, /governance fields satisfied/);
});

test('rejects a pull request without a canonical issue relationship', () => {
  const pullRequest = validPullRequest();
  pullRequest.body = pullRequest.body.replace('Related to #55', 'Issue 55');
  const result = runValidator(pullRequest);
  assert.equal(result.status, 1);
  assert.match(result.stderr, /must include at least one issue relation/);
});

test('rejects missing canonical labels', () => {
  const pullRequest = validPullRequest();
  pullRequest.labels = [{name: 'type:chore'}];
  const result = runValidator(pullRequest);
  assert.equal(result.status, 1);
  assert.match(result.stderr, /area:\*/);
});

test('rejects unchecked verification evidence', () => {
  const pullRequest = validPullRequest();
  pullRequest.body = pullRequest.body.replace('- [x] Node', '- [ ] Node');
  const result = runValidator(pullRequest);
  assert.equal(result.status, 1);
  assert.match(result.stderr, /checked result line/);
});

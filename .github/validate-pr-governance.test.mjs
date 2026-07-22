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
- **Task report:** \`docs/issues/issue-55-task-report.md\`
- **Governance profile:** \`boilerplate\`

## 🇬🇧 English — Normative Source

### 🎯 Goal

Goal

### 📦 Scope

Scope

### ✅ Acceptance Criteria

Criteria

### 🔒️ Security Boundary

Boundary

### 🧪 Verification

- [x] Node regression tests passed

### ⚠️ Risks and Follow-up

Risks

## 🇰🇷 한국어 요약

### 🎯 목표

목표

### 📦 핵심 범위

범위

### 🧪 검증

검증

### 🔒️ 비파괴 조건

조건

### ⚠️ 잔여 위험과 후속 작업

위험

## 🇯🇵 日本語の要約

### 🎯 目標

目標

### 📦 主な範囲

範囲

### 🧪 検証

検証

### 🔒️ 非破壊条件

条件

### ⚠️ 残存リスクと後続作業

リスク

## Gitmoji Checklist

- [x] 🔧 Chore

## 🤖 AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex
`,
    labels: [{name: 'type:chore'}, {name: 'area:ci'}, {name: 'custom:kept'}],
    assignees: [{login: 'Soku-JINSEOK'}],
    draft: true,
  };
}

function runValidator(pullRequest) {
  const directory = mkdtempSync(join(tmpdir(), 'pr-governance-'));
  const eventPath = join(directory, 'event.json');
  writeFileSync(eventPath, JSON.stringify({pull_request: pullRequest}));
  return spawnSync(process.execPath, [validator], {
    encoding: 'utf8',
    env: {...process.env, CURRENT_PR_EVENT_PATH: eventPath},
  });
}

function rejects(name, mutate, message) {
  test(name, () => {
    const pullRequest = validPullRequest();
    mutate(pullRequest);
    const result = runValidator(pullRequest);
    assert.equal(result.status, 1, result.stdout);
    assert.match(result.stderr, message);
  });
}

test('accepts canonical labels together with a custom label', () => {
  const result = runValidator(validPullRequest());
  assert.equal(result.status, 0, result.stderr);
  assert.match(result.stdout, /governance fields satisfied/);
});

rejects('rejects a non-canonical type label', (pr) => {
  pr.labels[0] = {name: 'type:not-canonical'};
}, /canonical `type:\*`/);
rejects('rejects a non-canonical area label', (pr) => {
  pr.labels[1] = {name: 'area:not-canonical'};
}, /canonical `area:\*`/);
rejects('rejects a missing assignee', (pr) => {
  pr.assignees = [];
}, /assigned to Soku-JINSEOK/);
rejects('rejects a closing relationship on a Draft PR', (pr) => {
  pr.body = pr.body.replace('Related to #55', 'Closes #55');
}, /Draft PRs/);
rejects('rejects multiple relationships', (pr) => {
  pr.body = pr.body.replace('Related to #55', 'Related to #55; Closes #54');
}, /exactly one/);
rejects('rejects an unsupported relationship', (pr) => {
  pr.body = pr.body.replace('Related to #55', 'Refs #55');
}, /exactly one/);
rejects('rejects a relation outside the Issue line', (pr) => {
  pr.body = pr.body
    .replace('Related to #55', 'Issue 55')
    .replace('\nGoal\n\n### 📦 Scope', '\nRelated to #55\n\nGoal\n\n### 📦 Scope');
}, /Issue:` must be exactly one/);
rejects('rejects a task report for a different Issue', (pr) => {
  pr.body = pr.body.replace('issue-55-task-report', 'issue-54-task-report');
}, /must match the linked Issue/);
rejects('rejects a missing task report file', (pr) => {
  pr.body = pr.body
    .replace('Related to #55', 'Related to #999')
    .replace('issue-55-task-report', 'issue-999-task-report');
}, /does not exist/);
rejects('rejects the wrong governance profile', (pr) => {
  pr.body = pr.body.replace('`boilerplate`', '`control-plane`');
}, /Governance profile must be `boilerplate`/);
rejects('rejects placeholders', (pr) => {
  pr.body = pr.body.replace('\nGoal\n\n### 📦 Scope', '\nTODO\n\n### 📦 Scope');
}, /placeholder/);
rejects('rejects a missing heading', (pr) => {
  pr.body = pr.body.replace('## 🇰🇷 한국어 요약', '한국어 요약');
}, /required heading/);
rejects('rejects headings out of order', (pr) => {
  pr.body = pr.body
    .replace('## 🇰🇷 한국어 요약', '## TEMP')
    .replace('## 🇯🇵 日本語の要約', '## 🇰🇷 한국어 요약')
    .replace('## TEMP', '## 🇯🇵 日本語の要約');
}, /not in the required template order/);
rejects('rejects unchecked verification evidence', (pr) => {
  pr.body = pr.body.replace('- [x] Node', '- [ ] Node');
}, /checked result line/);
rejects('rejects an unsupported AI disclosure', (pr) => {
  pr.body = pr.body.replace('OpenAI Codex', 'Undisclosed');
}, /AI Assistance/);

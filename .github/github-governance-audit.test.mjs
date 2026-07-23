import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import test from 'node:test';
import {
  classifyIssue,
  classifyPullRequest,
  createGhReader,
  epochFor,
  hashBody,
  parseArgs,
  renderReport,
  validateTitleForEpoch,
} from '../scripts/github-governance-audit.mjs';

const completeIssueBody = `
## 🇬🇧 English — Normative Source
### Goal
Goal text.
### Scope
Scope text.
### Acceptance Criteria
Acceptance text.
### Security Boundary
Non-destructive constraint.
### Verification
Verification evidence.
### AI Assistance
Codex.
Task report: docs/issues/issue-91-task-report.md
## 🇰🇷 한국어 요약
목표 범위 검증 비파괴 조건.
## 🇯🇵 日本語の要約
目標 範囲 検証 非破壊条件.
`;

const checks = [
  {name: 'Validation Gate', conclusion: 'success'},
  {name: 'PR Metadata Gate', conclusion: 'success'},
];

test('arguments and policy epochs are deterministic', () => {
  assert.deepEqual(
    parseArgs([
      '--repo',
      'owner/repo',
      '--as-of',
      '2026-07-23T01:00:19Z',
      '--output',
      'report.md',
    ]),
    {
      repo: 'owner/repo',
      asOf: '2026-07-23T01:00:19.000Z',
      output: new URL('../report.md', import.meta.url).pathname,
    },
  );
  assert.equal(epochFor('2026-07-14T03:50:00Z', 'Issue', 1), 'legacy');
  assert.equal(epochFor('2026-07-14T03:50:01Z', 'Issue', 1), 'multilingual');
  assert.equal(epochFor('2026-07-18T04:27:29Z', 'PR', 1), 'strengthened');
  assert.equal(epochFor('2026-07-21T20:23:25Z', 'PR', 1), 'strict');
  assert.equal(epochFor('2026-07-01T00:00:00Z', 'Issue', 91), 'current-normalized');
  assert.equal(epochFor('2026-07-01T00:00:00Z', 'PR', 89), 'dependabot-current');
  const historicalLongTitle =
    '📚 docs(history): document a subject that intentionally exceeds the later limit.';
  assert.equal(
    validateTitleForEpoch(historicalLongTitle, 'multilingual').valid,
    true,
  );
  assert.equal(validateTitleForEpoch(historicalLongTitle, 'strengthened').valid, false);
});

test('reader can issue only explicit GET requests', async () => {
  let command = null;
  const read = createGhReader({
    execute: async (args) => {
      command = args;
      return {stdout: '[[{"number":1}],[{"number":2}]]'};
    },
  });
  assert.deepEqual(await read('repos/owner/repo/issues', {paginate: true}), [
    {number: 1},
    {number: 2},
  ]);
  assert.deepEqual(command, [
    'api',
    '--method',
    'GET',
    '--paginate',
    '--slurp',
    'repos/owner/repo/issues',
  ]);
});

test('current normalized Issue distinguishes metadata from immutable evidence', () => {
  const evidence = {taskReportExists: true, linkedPullRequests: [92]};
  const result = classifyIssue({
    number: 91,
    title: '🐛 fix(gcp): restore deployment evidence',
    body: completeIssueBody,
    created_at: '2026-07-01T00:00:00Z',
    state: 'closed',
    state_reason: 'completed',
    labels: [
      {name: 'type:bug'},
      {name: 'priority:p1-high'},
      {name: 'status:done'},
      {name: 'area:ci'},
    ],
    assignees: [{login: 'Soku-JINSEOK'}],
  }, evidence);
  assert.equal(result.classification, 'Compliant');
  assert.equal(result.bodyHash, hashBody(completeIssueBody));

  const missingMetadata = classifyIssue({
    number: 91,
    title: '🐛 fix(gcp): restore deployment evidence',
    body: completeIssueBody,
    created_at: '2026-07-01T00:00:00Z',
    state: 'closed',
    state_reason: 'completed',
    labels: [],
    assignees: [],
  }, evidence);
  assert.equal(missingMetadata.classification, 'Correctable metadata');
  assert.ok(missingMetadata.candidates.length > 0);
});

test('exact Dependabot PR passes only with scoped files and current gates', () => {
  const item = {
    number: 89,
    title: 'build(deps): bump dependency from 1.0.0 to 1.0.1',
    body: 'Dependabot original body',
    created_at: '2026-07-01T00:00:00Z',
    state: 'open',
    merged_at: null,
    labels: [{name: 'type:chore'}, {name: 'area:tooling'}],
    assignees: [{login: 'Soku-JINSEOK'}],
    user: {login: 'dependabot[bot]'},
    head: {ref: 'dependabot/npm_and_yarn/templates/javascript-typescript-node/deps'},
  };
  const options = {
    dependabotConfigurations: [
      {ecosystem: 'npm', directory: '/templates/javascript-typescript-node'},
    ],
  };
  const result = classifyPullRequest(item, {
    files: [{filename: 'templates/javascript-typescript-node/package.json'}],
    checks,
    commits: [{commit: {message: 'build(deps): bump dependency', verification: {verified: true}}}],
  }, options);
  assert.equal(result.classification, 'Compliant');

  const outside = classifyPullRequest(item, {
    files: [{filename: 'README.md'}],
    checks,
    commits: [{commit: {message: 'build(deps): bump dependency', verification: {verified: true}}}],
  }, options);
  assert.equal(outside.classification, 'Blocked/missing evidence');
});

test('rendered report stores hashes and judgments but never raw bodies', () => {
  const secretMarker = 'RAW-BODY-MUST-NOT-LEAK';
  const issue = classifyIssue({
    number: 1,
    title: 'Initial issue',
    body: `${secretMarker}\nGoal Scope Acceptance`,
    created_at: '2026-01-01T00:00:00Z',
    state: 'open',
    state_reason: null,
    labels: [],
    assignees: [],
  });
  const report = renderReport({
    repo: 'owner/repo',
    asOf: '2026-07-23T01:00:19.000Z',
    issues: [issue],
    pullRequests: [],
  });
  assert.doesNotMatch(report, new RegExp(secretMarker));
  assert.match(report, new RegExp(hashBody(`${secretMarker}\nGoal Scope Acceptance`)));
  assert.match(report, /GitHub API contains 0 pull requests/);
  assert.match(report, /No mutation was applied/);
});

test('committed report covers the complete reconciled inventory', () => {
  const report = readFileSync(
    new URL('../docs/audits/github-governance-2026-07-23.md', import.meta.url),
    'utf8',
  );
  const rows = [
    ...report.matchAll(
      /^\| (Issue|PR) #(\d+) \| ([^|]+) \| ([^|]+) \| `([a-f0-9]{64})` \|/gmu,
    ),
  ];
  assert.equal(rows.filter(([, kind]) => kind === 'Issue').length, 33);
  assert.equal(rows.filter(([, kind]) => kind === 'PR').length, 59);
  assert.equal(rows.length, 92);
  const classifications = new Set([
    'Compliant',
    'Correctable metadata',
    'Historical exception',
    'Blocked/missing evidence',
    'Not applicable',
  ]);
  for (const [, , , , classification] of rows) {
    assert.ok(classifications.has(classification.trim()), classification);
  }
  assert.match(
    report,
    /^\| Issue #91 \| current-normalized \| Compliant \|[^\n]+\| 1 comments;/mu,
  );
  assert.match(report, /^\| PR #89 \| dependabot-current \| Compliant \|/mu);
  assert.match(report, /^\| PR #90 \| dependabot-current \| Compliant \|/mu);
  assert.match(report, /the approved plan expected 33 Issues \+ 48 pull requests = 81 items/);
});

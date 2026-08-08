import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import test from 'node:test';

import {
  GitHubApiClient,
  appendIssueRelation,
  auditRepository,
  applyOperations,
  buildIssueStatusLabels,
  computeSizeBucket,
  countScopeItems,
  extractTargetDate,
  inferPriority,
  inferSize,
  inferWorkstream,
  parseArgs,
  readSyncConfig,
} from './github-project-sync.mjs';

const config = readSyncConfig();

test('parses documented modes and target selectors', () => {
  const parsed = parseArgs(['--mode', 'apply', '--repo', 'owner/repo', '--project-owner', '@me', '--project-number', '2', '--issue-number', '42', '--output', 'report.json']);
  assert.equal(parsed.mode, 'apply');
  assert.equal(parsed.repo, 'owner/repo');
  assert.equal(parsed.issueNumber, 42);
  assert.match(parsed.output, /report\.json$/u);
  assert.throws(() => parseArgs(['--issue-number', '1', '--pr-number', '2']), /cannot be combined/);
  assert.throws(() => parseArgs(['--mode', 'mutate']), /audit or apply/);
});

test('appends exactly one relation while preserving the existing body prefix', () => {
  const body = 'Dependabot release notes\n';
  const updated = appendIssueRelation(body, 69);
  assert.ok(updated.startsWith(body));
  assert.match(updated, /Related to #69/);
  assert.equal(appendIssueRelation(updated, 69), updated);
});

test('normalizes Issue status labels without dropping custom labels', () => {
  const result = buildIssueStatusLabels({number: 180, state: 'closed', state_reason: 'completed', labels: [{name: 'type:chore'}, {name: 'custom:kept'}, {name: 'status:in-progress'}]});
  assert.deepEqual(result.labels, ['type:chore', 'custom:kept', 'status:done']);
  const open = buildIssueStatusLabels({number: 1, state: 'open', labels: [{name: 'custom:kept'}, {name: 'status:done'}, {name: 'status:ready'}]});
  assert.deepEqual(open.labels, ['custom:kept', 'status:ready']);
  const notPlanned = buildIssueStatusLabels({number: 2, state: 'closed', state_reason: 'not_planned', labels: [{name: 'status:in-progress'}]});
  assert.equal(notPlanned.plan.warning !== undefined, true);
  assert.deepEqual(notPlanned.labels, ['status:in-progress']);
});

test('derives priority, size, workstream, scope, and only explicit dates', () => {
  const body = '## 📦 Scope\n\n- update the workflow\n- verify the release\n\nPriority: P1 - high\nSize: L\nWorkstream: Delivery\nTarget date: 2026-08-31\n';
  const issue = {title: 'Deploy release metadata', body, labels: []};
  assert.equal(inferPriority(issue), 'P1');
  assert.equal(inferSize(issue, {pullRequests: 0, files: 0, scopeItems: 0}), 'L');
  assert.equal(inferWorkstream(issue), 'Delivery');
  assert.equal(countScopeItems(body), 2);
  assert.equal(extractTargetDate(body), '2026-08-31');
  assert.equal(extractTargetDate('Target date: next month'), null);
  assert.equal(computeSizeBucket({pullRequests: 1, files: 4, scopeItems: 2}), 'S');
  assert.equal(computeSizeBucket({pullRequests: 5, files: 51, scopeItems: 13}), 'XL');
});

test('GitHub REST pagination follows the Link header', async () => {
  const calls = [];
  const client = new GitHubApiClient({token: 'test-token', fetchImpl: async (url) => {
    calls.push(url);
    const second = url.includes('page=2');
    return {ok: true, status: 200, headers: second ? {get: () => ''} : {get: (name) => name === 'link' ? '<https://api.github.com/items?page=2>; rel="next"' : ''}, text: async () => JSON.stringify(second ? [{id: 2}] : [{id: 1}])};
  }});
  assert.deepEqual(await client.paginateRest('/items?per_page=1'), [{id: 1}, {id: 2}]);
  assert.deepEqual(calls, ['https://api.github.com/items?per_page=1', 'https://api.github.com/items?page=2']);
});

class FakeClient {
  constructor({issues, pullRequests, project}) {
    this.issues = issues;
    this.pullRequests = pullRequests;
    this.project = project;
    this.mutations = [];
  }
  async paginateRest(path) {
    if (path.includes('/issues?')) return this.issues;
    if (path.includes('/pulls?')) return this.pullRequests;
    if (path.includes('/files?')) return [{filename: 'a'}, {filename: 'b'}, {filename: 'c'}];
    throw new Error(`Unexpected pagination path: ${path}`);
  }
  async graphql(query) {
    if (query.includes('viewer { login }')) return {viewer: {login: 'Soku-JINSEOK'}};
    if (query.includes('updateProjectV2ItemFieldValue')) {
      this.mutations.push({type: 'project-field'});
      return {updateProjectV2ItemFieldValue: {projectV2Item: {id: 'item-10'}}};
    }
    return {user: {projectV2: {id: this.project.id, number: this.project.number, title: this.project.title, fields: {nodes: this.project.fields}, items: {nodes: this.project.items.map((item) => ({id: item.id, type: item.type, content: item.content, fieldValues: {nodes: item.fieldValues.map((field) => ({__typename: field.type, ...(field.type === 'ProjectV2ItemFieldSingleSelectValue' ? {name: field.value, optionId: field.optionId, field: {id: field.id, name: field.name}} : {text: field.value, field: {id: field.id, name: field.name}})}))}})), pageInfo: {hasNextPage: false, endCursor: null}}}}};
  }
  async rest(method, path, body) {
    this.mutations.push({method, path, body});
    if (method === 'GET' && path.includes('/pulls/')) return this.pullRequests.find((item) => path.endsWith(`/${item.number}`));
    if (method === 'GET' && path.includes('/issues/')) return this.issues.find((item) => path.endsWith(`/${item.number}`));
    const number = Number(path.match(/\/issues\/(\d+)/u)?.[1]);
    const item = this.pullRequests.find((entry) => entry.number === number)
      ?? this.issues.find((entry) => entry.number === number);
    if (item && method === 'PATCH' && body?.body !== undefined) item.body = body.body;
    if (item && method === 'PUT' && Array.isArray(body?.labels)) {
      item.labels = body.labels.map((name) => ({name}));
    }
    return {};
  }
}

function fakeProject() {
  const select = (id, name, values) => ({__typename: 'ProjectV2SingleSelectField', id, name, options: values.map((value) => ({id: `${id}-${value}`, name: value}))});
  return {id: 'project-2', number: 2, title: 'Operations', fields: [select('status', 'Status', ['Inbox', 'Ready', 'In progress', 'Blocked', 'Done']), select('priority', 'Priority', ['P0', 'P1', 'P2', 'P3']), select('size', 'Size', ['XS', 'S', 'M', 'L', 'XL']), select('workstream', 'Workstream', ['Engineering', 'Security', 'Delivery', 'Governance']), {__typename: 'ProjectV2Field', id: 'target', name: 'Target date', dataType: 'DATE'}], items: [{id: 'item-10', type: 'ISSUE', content: {__typename: 'Issue', id: 'issue-10', number: 10, title: 'Deploy release metadata', body: '## 📦 Scope\n\n- workflow\n\nTarget date: 2026-08-31', state: 'OPEN', stateReason: null, repository: {nameWithOwner: 'owner/repo'}, labels: {nodes: [{name: 'priority:p1-high'}, {name: 'status:ready'}]}}, fieldValues: [{type: 'ProjectV2ItemFieldSingleSelectValue', id: 'status', name: 'Status', value: 'Ready', optionId: 'status-Ready'}, {type: 'ProjectV2ItemFieldSingleSelectValue', id: 'priority', name: 'Priority', value: 'P2', optionId: 'priority-P2'}]}]};
}

test('audit plans Project field updates for current-repository Issue items only', async () => {
  const fake = new FakeClient({issues: [{number: 10, title: 'Deploy release metadata', body: '## 📦 Scope\n\n- workflow\n\nTarget date: 2026-08-31', state: 'open', state_reason: null, labels: [{name: 'priority:p1-high'}, {name: 'status:ready'}]}], pullRequests: [{number: 1, body: 'Related to #10', state: 'open', merged_at: null, changed_files: 3, labels: []}], project: fakeProject()});
  const audit = await auditRepository({client: fake, repo: 'owner/repo', projectOwner: '@me', projectNumber: 2, config: {...config, repository: 'owner/repo', backfill: {relationMappings: {}}}});
  const operation = audit.operations.find((item) => item.kind === 'project-fields');
  assert.ok(operation);
  assert.deepEqual(operation.expectedAfter.fields, {Priority: 'P1', Size: 'S', Workstream: 'Delivery', 'Target date': '2026-08-31'});
  assert.equal(audit.operations.some((item) => item.kind === 'pr-relation'), false);
});

test('apply skips only stale targets after a fresh read', async () => {
  const fake = new FakeClient({issues: [{number: 10, title: 'Issue', body: 'Body', state: 'open', state_reason: null, labels: [{name: 'status:ready'}]}], pullRequests: [], project: {...fakeProject(), items: []}});
  const audit = {repository: 'owner/repo', config, dependencyIssueNumber: null, operations: [{target: 'Issue #10', operation: 'normalize issue status labels', beforeHash: 'different-hash', expectedAfter: {labels: ['status:done']}, reason: 'test', conflict: null, kind: 'issue-labels', issueNumber: 10, labels: ['status:done']}], project: {owner: 'Soku-JINSEOK', number: 2, title: 'Operations'}};
  const result = await applyOperations({client: fake, audit, projectOwner: '@me', projectNumber: 2});
  assert.equal(result.appliedOperations[0].result, 'skipped');
  assert.equal(result.appliedOperations[0].conflict, 'stale-read');
  assert.equal(fake.mutations.length, 1);
});

test('applies independent PR relation and status operations without self-conflict', async () => {
  const fake = new FakeClient({
    issues: [],
    pullRequests: [{
      number: 182,
      body: 'Dependabot release notes',
      state: 'closed',
      merged_at: '2026-08-01T00:00:00Z',
      labels: [],
      user: {login: 'dependabot[bot]'},
      head: {ref: 'dependabot/npm_and_yarn/example-182'},
      changed_files: 1,
    }],
    project: {...fakeProject(), items: []},
  });
  const audit = await auditRepository({
    client: fake,
    repo: 'owner/repo',
    projectOwner: '@me',
    projectNumber: 2,
    config,
  });
  audit.operations = audit.operations.filter((item) => item.kind !== 'create-dependency-issue');
  audit.dependencyIssueNumber = 123;
  const result = await applyOperations({
    client: fake,
    audit,
    projectOwner: '@me',
    projectNumber: 2,
  });
  assert.equal(
    result.appliedOperations.filter((item) => item.result === 'applied').length,
    2,
  );
  assert.match(fake.pullRequests[0].body, /Related to #123/u);
  assert.deepEqual(fake.pullRequests[0].labels, [{name: 'status:done'}]);
});

test('does not reuse an initial tracking Issue for unrelated Dependabot PRs', async () => {
  const dependabot = (number) => ({
    number,
    body: 'Dependabot release notes',
    state: 'open',
    merged_at: null,
    labels: [],
    user: {login: 'dependabot[bot]'},
    head: {ref: `dependabot/npm_and_yarn/example-${number}`},
    changed_files: 1,
  });
  const fake = new FakeClient({
    issues: [],
    pullRequests: [dependabot(182), dependabot(143)],
    project: {...fakeProject(), items: []},
  });
  const audit = await auditRepository({
    client: fake,
    repo: 'owner/repo',
    projectOwner: '@me',
    projectNumber: 2,
    config,
  });
  assert.ok(audit.operations.some((item) => item.target === 'PR #182'));
  assert.equal(audit.operations.some((item) => item.target === 'PR #143'), false);
  assert.ok(audit.warnings.some((item) => item.target === 'PR #143'));
});

test('generated configuration disables dependency tracking by default', async () => {
  const dependabot = {
    number: 182,
    body: 'Dependabot release notes',
    state: 'open',
    merged_at: null,
    labels: [],
    user: {login: 'dependabot[bot]'},
    head: {ref: 'dependabot/npm_and_yarn/example-182'},
    changed_files: 1,
  };
  const fake = new FakeClient({
    issues: [],
    pullRequests: [dependabot],
    project: {...fakeProject(), items: []},
  });
  const audit = await auditRepository({
    client: fake,
    repo: 'owner/repo',
    projectOwner: '@me',
    projectNumber: 17,
    config: {
      ...config,
      project: {owner: '@me', number: 17, fields: config.project.fields},
      backfill: {relationMappings: {}, dependencyTrackingIssue: null},
    },
  });
  assert.equal(audit.operations.some((item) => item.kind === 'create-dependency-issue'), false);
  assert.equal(audit.operations.some((item) => item.kind === 'pr-relation'), false);
  assert.ok(audit.warnings.some((item) => item.conflict === 'dependency-tracking-disabled'));
});

test('configuration does not contain raw issue or pull request bodies', () => {
  const source = readFileSync(new URL('../.github/project-sync.yml', import.meta.url), 'utf8');
  assert.doesNotMatch(source, /Dependabot release notes|## 🎯 Goal/);
});

test('credential runbook fixes the least-privilege and rotation contract', () => {
  const runbook = readFileSync(
    new URL('../docs/guides/PROJECT_SYNC_CREDENTIAL_RUNBOOK.md', import.meta.url),
    'utf8',
  );
  const guide = readFileSync(
    new URL('../docs/guides/GITHUB_PROJECT_SYNC.md', import.meta.url),
    'utf8',
  );

  for (const requirement of [
    'Repository | Metadata | Read',
    'Repository | Issues | Read and write',
    'Repository | Pull requests | Read and write',
    'Authenticated user | Projects | Read and write',
    'repository Contents write',
    'Actions administration',
    'workflow edit',
    'organization administration',
    'billing access',
  ]) {
    assert.ok(runbook.includes(requirement), requirement);
  }

  const phases = [
    'Phase 1: prepare and directly audit the replacement',
    'Phase 2: replace the repository secret',
    'Phase 3: post-replacement audit',
    'Phase 4: optional separately approved apply',
    'Phase 5: revoke and close out',
  ];
  let previous = -1;
  for (const phase of phases) {
    const index = runbook.indexOf(phase);
    assert.ok(index > previous, `${phase} must remain ordered`);
    previous = index;
  }

  assert.match(runbook, /Never revoke the old credential before/u);
  assert.match(runbook, /Never retry with a personal or\s+broadly scoped fallback token/u);
  assert.match(runbook, /Abort and rollback matrix/u);
  assert.match(runbook, /Only the UTC rotation date and the allowlisted redacted outcome fields/u);
  assert.doesNotMatch(runbook, /gh[pousr]_[A-Za-z0-9_]{20,}/u);
  assert.doesNotMatch(runbook, /github_pat_[A-Za-z0-9_]{20,}/u);
  assert.match(guide, /PROJECT_SYNC_CREDENTIAL_RUNBOOK\.md/u);
});

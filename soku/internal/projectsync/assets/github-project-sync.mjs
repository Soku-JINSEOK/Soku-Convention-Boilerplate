#!/usr/bin/env node

import {createHash} from 'node:crypto';
import {
  existsSync,
  mkdirSync,
  readFileSync,
  writeFileSync,
} from 'node:fs';
import {dirname, resolve} from 'node:path';
import {execFileSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';

const REPOSITORY_ROOT = fileURLToPath(new URL('../', import.meta.url));
const DEFAULT_CONFIG_PATH = resolve(REPOSITORY_ROOT, '.github/project-sync.yml');
const GITHUB_API = 'https://api.github.com';
const RELATION_PATTERN = /\b(?:Closes|Fixes|Resolves|Related to)\s+#([1-9]\d*)\b/giu;
const ACTIVE_STATUS_LABELS = new Set([
  'status:triage',
  'status:ready',
  'status:in-progress',
  'status:blocked',
]);
const ISSUE_STATUS_LABELS = new Set([...ACTIVE_STATUS_LABELS, 'status:done']);
const STATUS_PRIORITY = [
  ['status:blocked', 'Blocked'],
  ['status:in-progress', 'In progress'],
  ['status:ready', 'Ready'],
  ['status:triage', 'Inbox'],
];
const DEFAULT_PROJECT_FIELDS = Object.freeze({
  status: 'Status',
  priority: 'Priority',
  size: 'Size',
  workstream: 'Workstream',
  targetDate: 'Target date',
});
const SIZE_BUCKETS = Object.freeze([
  {value: 'XS', pullRequests: 0, files: 2, scopeItems: 1},
  {value: 'S', pullRequests: 1, files: 8, scopeItems: 3},
  {value: 'M', pullRequests: 2, files: 20, scopeItems: 6},
  {value: 'L', pullRequests: 4, files: 50, scopeItems: 12},
  {value: 'XL', pullRequests: Infinity, files: Infinity, scopeItems: Infinity},
]);
const DEFAULT_DEPENDENCY_BODY = `## Dependency tracking

Track dependency update pull requests that have been explicitly assigned to this repository's tracking Issue.

## Non-destructive boundary

This Issue records dependency coordination only; it does not authorize source, release, or infrastructure changes.
`;

function usage() {
  return `Usage: scripts/github-project-sync.mjs [options]

Options:
  --mode <audit|apply>       Read-only audit (default) or guarded mutation
  --repo <owner/repo>        Repository to inspect
  --project-owner <owner>    Project owner, or @me for the authenticated user
  --project-number <number>  Project v2 number (defaults to the configuration)
  --issue-number <number>    Limit issue synchronization to one Issue
  --pr-number <number>       Limit PR synchronization to one pull request
  --output <path>            Write the redacted JSON report to this path
  --config <path>            JSON-compatible YAML configuration path
  --help                     Show this help`;
}

function positiveNumber(value, flag) {
  if (!/^[1-9]\d*$/.test(String(value))) {
    throw new Error(`${flag} must be a positive integer.`);
  }
  return Number(value);
}

export function parseArgs(argv) {
  const options = {
    mode: 'audit',
    config: DEFAULT_CONFIG_PATH,
  };
  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    if (argument === '--help') return {help: true};
    if (
      ![
        '--mode',
        '--repo',
        '--project-owner',
        '--project-number',
        '--issue-number',
        '--pr-number',
        '--output',
        '--config',
      ].includes(argument)
    ) {
      throw new Error(`Unknown argument: ${argument}`);
    }
    const value = argv[index + 1];
    if (!value || value.startsWith('--')) {
      throw new Error(`${argument} requires a value.`);
    }
    index += 1;
    if (argument === '--mode') options.mode = value;
    if (argument === '--repo') options.repo = value;
    if (argument === '--project-owner') options.projectOwner = value;
    if (argument === '--project-number') {
      options.projectNumber = positiveNumber(value, argument);
    }
    if (argument === '--issue-number') {
      options.issueNumber = positiveNumber(value, argument);
    }
    if (argument === '--pr-number') {
      options.prNumber = positiveNumber(value, argument);
    }
    if (argument === '--output') options.output = resolve(value);
    if (argument === '--config') options.config = resolve(value);
  }
  if (!['audit', 'apply'].includes(options.mode)) {
    throw new Error('--mode must be audit or apply.');
  }
  if (options.issueNumber && options.prNumber) {
    throw new Error('--issue-number and --pr-number cannot be combined.');
  }
  if (options.repo && !/^[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+$/.test(options.repo)) {
    throw new Error('--repo must use owner/repository syntax.');
  }
  return options;
}

function clone(value) {
  if (value === undefined) return undefined;
  return JSON.parse(JSON.stringify(value));
}

function stableValue(value) {
  if (Array.isArray(value)) return value.map(stableValue);
  if (value && typeof value === 'object') {
    return Object.fromEntries(
      Object.keys(value)
        .sort()
        .map((key) => [key, stableValue(value[key])]),
    );
  }
  return value;
}

export function hashValue(value) {
  return createHash('sha256')
    .update(JSON.stringify(stableValue(value)), 'utf8')
    .digest('hex');
}

export function hashBody(body) {
  return createHash('sha256').update(body ?? '', 'utf8').digest('hex');
}

function labelNames(item) {
  return (item?.labels ?? [])
    .map((label) => (typeof label === 'string' ? label : label?.name))
    .filter(Boolean);
}

function sortedUnique(values) {
  return [...new Set(values)].sort((left, right) => left.localeCompare(right));
}

function sameLabelSet(left, right) {
  return JSON.stringify(sortedUnique(left)) === JSON.stringify(sortedUnique(right));
}

export function relationNumbers(body) {
  RELATION_PATTERN.lastIndex = 0;
  return [...(body ?? '').matchAll(RELATION_PATTERN)].map((match) => Number(match[1]));
}

function hasRelation(body, issueNumber) {
  return relationNumbers(body).includes(issueNumber);
}

export function appendIssueRelation(body, issueNumber) {
  if (hasRelation(body, issueNumber)) return body ?? '';
  const source = body ?? '';
  const separator = source.length === 0
    ? ''
    : source.endsWith('\n')
      ? '\n'
      : '\n\n';
  return `${source}${separator}## 🔗 Issue Relationship\n\nRelated to #${issueNumber}\n`;
}

function issueStateSnapshot(issue) {
  return {
    bodyHash: hashBody(issue?.body),
    labels: labelNames(issue),
    state: issue?.state ?? null,
    stateReason: issue?.state_reason ?? issue?.stateReason ?? null,
  };
}

function pullRequestBodySnapshot(pullRequest) {
  return {
    bodyHash: hashBody(pullRequest?.body),
    state: pullRequest?.state ?? null,
    mergedAt: pullRequest?.merged_at ?? pullRequest?.mergedAt ?? null,
  };
}

function pullRequestLabelSnapshot(pullRequest) {
  return {
    labels: labelNames(pullRequest),
    state: pullRequest?.state ?? null,
    mergedAt: pullRequest?.merged_at ?? pullRequest?.mergedAt ?? null,
  };
}

function projectItemSnapshot(item) {
  const fields = {};
  for (const field of item?.fieldValues ?? []) {
    const name = field.name ?? field.fieldName;
    if (!name) continue;
    fields[name] = field.value ?? null;
  }
  return {itemId: item?.id ?? null, issueNumber: item?.content?.number ?? null, fields};
}

function issueStateHash(issue) {
  return hashValue(issueStateSnapshot(issue));
}

function pullRequestBodyHash(pullRequest) {
  return hashValue(pullRequestBodySnapshot(pullRequest));
}

function pullRequestLabelHash(pullRequest) {
  return hashValue(pullRequestLabelSnapshot(pullRequest));
}

function projectItemHash(item) {
  return hashValue(projectItemSnapshot(item));
}

function normalizeToken() {
  const configured =
    process.env.PROJECT_SYNC_TOKEN ?? process.env.GH_TOKEN ?? process.env.GITHUB_TOKEN;
  if (configured) return configured;
  try {
    return execFileSync('gh', ['auth', 'token'], {encoding: 'utf8'}).trim();
  } catch {
    return '';
  }
}

function headerValue(headers, name) {
  if (!headers) return '';
  if (typeof headers.get === 'function') return headers.get(name) ?? '';
  return headers[name] ?? headers[name.toLowerCase()] ?? '';
}

function nextLink(headers) {
  const link = headerValue(headers, 'link');
  return link.match(/<([^>]+)>;\s*rel="next"/)?.[1] ?? null;
}

export class GitHubApiClient {
  constructor({token = normalizeToken(), fetchImpl = globalThis.fetch, apiBase = GITHUB_API} = {}) {
    if (typeof fetchImpl !== 'function') {
      throw new Error('A fetch implementation is required.');
    }
    if (!token) throw new Error('PROJECT_SYNC_TOKEN, GH_TOKEN, or GITHUB_TOKEN is required.');
    this.token = token;
    this.fetchImpl = fetchImpl;
    this.apiBase = apiBase.replace(/\/$/, '');
  }

  async request(method, pathOrUrl, body) {
    const url = pathOrUrl.startsWith('http')
      ? pathOrUrl
      : `${this.apiBase}${pathOrUrl.startsWith('/') ? pathOrUrl : `/${pathOrUrl}`}`;
    const response = await this.fetchImpl(url, {
      method,
      headers: {
        Accept: 'application/vnd.github+json',
        Authorization: `Bearer ${this.token}`,
        'X-GitHub-Api-Version': '2022-11-28',
        ...(body === undefined ? {} : {'Content-Type': 'application/json'}),
      },
      ...(body === undefined ? {} : {body: JSON.stringify(body)}),
    });
    const raw = typeof response.text === 'function'
      ? await response.text()
      : JSON.stringify(await response.json());
    let data = null;
    if (raw) {
      try {
        data = JSON.parse(raw);
      } catch {
        data = raw;
      }
    }
    if (!response.ok) {
      const message = typeof data === 'object' ? data?.message : data;
      throw new Error(`GitHub ${method} ${pathOrUrl} failed (${response.status}): ${message ?? 'unknown error'}`);
    }
    return {data, headers: response.headers, status: response.status};
  }

  async rest(method, path, body) {
    return (await this.request(method, path, body)).data;
  }

  async paginateRest(path) {
    const values = [];
    let next = path;
    while (next) {
      const response = await this.request('GET', next);
      if (!Array.isArray(response.data)) {
        throw new Error(`Expected a paginated array from ${next}.`);
      }
      values.push(...response.data);
      next = nextLink(response.headers);
    }
    return values;
  }

  async graphql(query, variables = {}) {
    const response = await this.request('POST', '/graphql', {query, variables});
    if (response.data?.errors?.length) {
      throw new Error(
        `GitHub GraphQL query failed: ${response.data.errors.map((error) => error.message).join('; ')}`,
      );
    }
    return response.data?.data;
  }
}

export function readSyncConfig(path = DEFAULT_CONFIG_PATH) {
  if (!existsSync(path)) throw new Error(`Sync configuration does not exist: ${path}`);
  let config;
  try {
    config = JSON.parse(readFileSync(path, 'utf8'));
  } catch (error) {
    throw new Error(`Sync configuration must be JSON-compatible YAML: ${error.message}`);
  }
  if (config.schemaVersion !== 1) {
    throw new Error('Unsupported project sync configuration schema.');
  }
  if (config.repository && !/^[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+$/.test(config.repository)) {
    throw new Error('project-sync.yml repository must use owner/repository syntax.');
  }
  const backfill = config.backfill ?? {};
  const dependencyTrackingIssue = backfill.dependencyTrackingIssue;
  const normalizedBackfill = {
    ...backfill,
    relationMappings: {...(backfill.relationMappings ?? {})},
  };
  if (dependencyTrackingIssue === null || dependencyTrackingIssue === false) {
    normalizedBackfill.dependencyTrackingIssue = dependencyTrackingIssue;
  } else if (dependencyTrackingIssue !== undefined) {
    normalizedBackfill.dependencyTrackingIssue = {...dependencyTrackingIssue};
  }
  return {
    ...config,
    project: {
      ...(config.project ?? {}),
      fields: {
        ...DEFAULT_PROJECT_FIELDS,
        ...(config.project?.fields ?? {}),
      },
    },
    backfill: normalizedBackfill,
  };
}

function configOptions(parsed, config) {
  const repo = parsed.repo ?? config.repository;
  if (!repo) throw new Error('--repo is required when project-sync.yml has no repository.');
  const projectOwner = parsed.projectOwner ?? config.project?.owner ?? '@me';
  const projectNumber = parsed.projectNumber ?? config.project?.number;
  if (!Number.isInteger(projectNumber) || projectNumber < 1) {
    throw new Error('--project-number or a positive project.number is required.');
  }
  return {
    ...parsed,
    repo,
    projectOwner,
    projectNumber,
  };
}

async function listIssues(client, repo) {
  const values = await client.paginateRest(
    `/repos/${repo}/issues?state=all&per_page=100`,
  );
  return values.filter((item) => !item.pull_request);
}

async function listPullRequests(client, repo) {
  return client.paginateRest(`/repos/${repo}/pulls?state=all&per_page=100`);
}

async function fetchIssue(client, repo, number) {
  return client.rest('GET', `/repos/${repo}/issues/${number}`);
}

async function fetchPullRequest(client, repo, number) {
  return client.rest('GET', `/repos/${repo}/pulls/${number}`);
}

async function fetchPullRequestFiles(client, repo, number) {
  const files = await client.paginateRest(
    `/repos/${repo}/pulls/${number}/files?per_page=100`,
  );
  return files.length;
}

const PROJECT_QUERY = `
query($owner: String!, $number: Int!, $after: String) {
  user(login: $owner) {
    projectV2(number: $number) {
      id
      number
      title
      fields(first: 100) {
        nodes {
          __typename
          ... on ProjectV2Field { id name dataType }
          ... on ProjectV2SingleSelectField {
            id
            name
            options { id name }
          }
          ... on ProjectV2IterationField { id name }
        }
      }
      items(first: 100, after: $after) {
        nodes {
          id
          type
          content {
            __typename
            ... on Issue {
              id
              number
              title
              body
              state
              stateReason
              repository { nameWithOwner }
              labels(first: 100) { nodes { name } }
            }
            ... on PullRequest {
              id
              number
              title
              body
              state
              mergedAt
              repository { nameWithOwner }
            }
          }
          fieldValues(first: 100) {
            nodes {
              __typename
              ... on ProjectV2ItemFieldTextValue {
                text
                field { ... on ProjectV2Field { id name } }
              }
              ... on ProjectV2ItemFieldSingleSelectValue {
                name
                optionId
                field { ... on ProjectV2SingleSelectField { id name } }
              }
              ... on ProjectV2ItemFieldDateValue {
                date
                field { ... on ProjectV2Field { id name } }
              }
              ... on ProjectV2ItemFieldNumberValue {
                number
                field { ... on ProjectV2Field { id name } }
              }
              ... on ProjectV2ItemFieldIterationValue {
                title
                iterationId
                field { ... on ProjectV2IterationField { id name } }
              }
            }
          }
        }
        pageInfo { hasNextPage endCursor }
      }
    }
  }
}`;

async function resolveProjectOwner(client, owner) {
  if (owner !== '@me') return owner;
  const data = await client.graphql('query { viewer { login } }');
  if (!data?.viewer?.login) throw new Error('Could not resolve @me to a GitHub login.');
  return data.viewer.login;
}

function projectFieldValue(node) {
  if (!node) return null;
  const field = node.field ?? {};
  const base = {id: field.id, name: field.name, type: node.__typename};
  if (node.__typename === 'ProjectV2ItemFieldSingleSelectValue') {
    return {...base, value: node.name ?? null, optionId: node.optionId ?? null};
  }
  if (node.__typename === 'ProjectV2ItemFieldTextValue') {
    return {...base, value: node.text ?? null};
  }
  if (node.__typename === 'ProjectV2ItemFieldDateValue') {
    return {...base, value: node.date ?? null};
  }
  if (node.__typename === 'ProjectV2ItemFieldNumberValue') {
    return {...base, value: node.number ?? null};
  }
  if (node.__typename === 'ProjectV2ItemFieldIterationValue') {
    return {
      ...base,
      value: node.title ?? null,
      iterationId: node.iterationId ?? null,
    };
  }
  return null;
}

function normalizeProjectItem(item) {
  const content = item.content;
  return {
    id: item.id,
    type: item.type,
    content: content
      ? {
        id: content.id,
        number: content.number,
        title: content.title,
        body: content.body ?? '',
        state: content.state,
        stateReason: content.stateReason ?? null,
        mergedAt: content.mergedAt ?? null,
        repository: content.repository?.nameWithOwner ?? null,
        labels: content.labels?.nodes ?? [],
        __typename: content.__typename,
      }
      : null,
    fieldValues: (item.fieldValues?.nodes ?? [])
      .map(projectFieldValue)
      .filter(Boolean),
  };
}

async function fetchProject(client, owner, number) {
  const resolvedOwner = await resolveProjectOwner(client, owner);
  const items = [];
  let fields = [];
  let project = null;
  let after = null;
  do {
    const data = await client.graphql(PROJECT_QUERY, {
      owner: resolvedOwner,
      number,
      after,
    });
    project = data?.user?.projectV2;
    if (!project) {
      throw new Error(`User-owned Project #${number} was not found for ${resolvedOwner}.`);
    }
    if (fields.length === 0) fields = project.fields?.nodes ?? [];
    items.push(...(project.items?.nodes ?? []).map(normalizeProjectItem));
    const pageInfo = project.items?.pageInfo;
    after = pageInfo?.hasNextPage ? pageInfo.endCursor : null;
  } while (after);
  return {
    id: project.id,
    number: project.number,
    title: project.title,
    owner: resolvedOwner,
    fields,
    items,
  };
}

function normalizedName(value) {
  return String(value ?? '').trim().toLowerCase();
}

function findProjectField(project, name) {
  const wanted = normalizedName(name);
  return project.fields.find((field) => normalizedName(field.name) === wanted) ?? null;
}

function findProjectItem(project, repo, issueNumber) {
  return project.items.find(
    (item) =>
      item.content?.__typename === 'Issue' &&
      item.content.repository === repo &&
      item.content.number === issueNumber,
  ) ?? null;
}

function currentProjectValue(item, fieldName) {
  const wanted = normalizedName(fieldName);
  return item.fieldValues.find((field) => normalizedName(field.name) === wanted) ?? null;
}

function canonicalStatusLabel(labels) {
  const labelSet = new Set(labels);
  return STATUS_PRIORITY.find(([label]) => labelSet.has(label))?.[0] ?? null;
}

export function deriveIssueStatus(issue) {
  const labels = labelNames(issue);
  const state = issue.state;
  const reason = issue.state_reason ?? issue.stateReason ?? null;
  if (state === 'closed') {
    if (reason === 'completed') return {value: 'Done', label: 'status:done'};
    return {
      value: null,
      label: null,
      warning: `Issue #${issue.number} is closed with non-completed reason ${reason ?? 'unknown'}; status is audit-only.`,
    };
  }
  const label = canonicalStatusLabel(labels);
  if (label) {
    return {
      value: STATUS_PRIORITY.find(([candidate]) => candidate === label)?.[1] ?? 'Inbox',
      label,
    };
  }
  return {value: 'Inbox', label: 'status:triage'};
}

export function buildIssueStatusLabels(issue) {
  const plan = deriveIssueStatus(issue);
  if (!plan.label) return {plan, labels: labelNames(issue)};
  const existing = labelNames(issue);
  const labels = existing.filter((label) => !ISSUE_STATUS_LABELS.has(label));
  labels.push(plan.label);
  return {plan, labels};
}

export function inferPriority(issue) {
  const labels = labelNames(issue);
  const label = labels
    .map((candidate) => /^priority:p([0-3])(?:-|$)/u.exec(candidate))
    .find(Boolean);
  if (label) return `P${label[1]}`;
  const body = issue.body ?? '';
  const match = /\bP([0-3])\s*(?:-|:)\s*(?:critical|high|normal|low)\b/iu.exec(body)
    ?? /\bpriority\s*[:=]\s*P([0-3])\b/iu.exec(body);
  return match ? `P${match[1]}` : 'P2';
}

export function extractExplicitSize(body) {
  return /(?:^|\n)\s*(?:[-*]\s*)?(?:size|estimate)\s*[:=]\s*(XS|S|M|L|XL)\b/iu.exec(
    body ?? '',
  )?.[1]?.toUpperCase() ?? null;
}

export function countScopeItems(body) {
  const match = /(?:^|\n)##[^\n]*scope[^\n]*\n([\s\S]*?)(?=\n##|$)/iu.exec(body ?? '');
  if (!match) return 0;
  return match[1]
    .split(/\r?\n/u)
    .filter((line) => /^\s*(?:[-*+]\s+|\d+[.)]\s+)/u.test(line))
    .length;
}

export function computeSizeBucket({pullRequests = 0, files = 0, scopeItems = 0} = {}) {
  return SIZE_BUCKETS.find((bucket) =>
    pullRequests <= bucket.pullRequests &&
    files <= bucket.files &&
    scopeItems <= bucket.scopeItems,
  )?.value ?? 'XL';
}

export function inferSize(issue, metrics = {}) {
  return extractExplicitSize(issue.body) ?? computeSizeBucket(metrics);
}

export function inferWorkstream(issue) {
  const explicit = /(?:^|\n)\s*(?:[-*]\s*)?workstream\s*[:=]\s*([^\n]+)/iu.exec(
    issue.body ?? '',
  )?.[1]?.trim();
  if (explicit) {
    const value = explicit.toLowerCase();
    if (value.includes('security')) return 'Security';
    if (value.includes('delivery')) return 'Delivery';
    if (value.includes('governance')) return 'Governance';
    if (value.includes('engineering')) return 'Engineering';
    return explicit;
  }
  const signals = `${issue.title ?? ''}\n${issue.body ?? ''}`.toLowerCase();
  if (/security/u.test(signals)) return 'Security';
  if (/(?:release|publication|deploy)/u.test(signals)) return 'Delivery';
  if (/(?:governance|audit|policy)/u.test(signals)) return 'Governance';
  return 'Engineering';
}

export function extractTargetDate(body) {
  const candidate = /(?:^|\n)\s*(?:[-*]\s*)?(?:target date|due date)\s*[:=]\s*(\d{4}-\d{2}-\d{2})\b/iu.exec(
    body ?? '',
  )?.[1];
  if (!candidate) return null;
  const date = new Date(`${candidate}T00:00:00Z`);
  return Number.isNaN(date.valueOf()) || date.toISOString().slice(0, 10) !== candidate
    ? null
    : candidate;
}

function relationMetrics(issueNumber, pullRequests, fileCounts) {
  const linked = pullRequests.filter((pullRequest) =>
    relationNumbers(pullRequest.body).includes(issueNumber),
  );
  return {
    pullRequests: linked.length,
    files: linked.reduce(
      (total, pullRequest) => total + (fileCounts.get(pullRequest.number) ?? 0),
      0,
    ),
  };
}

function projectFieldPlan(issue, metrics) {
  const targetDate = extractTargetDate(issue.body);
  return {
    Priority: inferPriority(issue),
    Size: inferSize(issue, metrics),
    Workstream: inferWorkstream(issue),
    ...(targetDate ? {'Target date': targetDate} : {}),
  };
}

function optionFor(field, value) {
  const options = field?.options ?? [];
  const exact = options.find((option) => normalizedName(option.name) === normalizedName(value));
  if (exact) return exact;
  if (/^P[0-3]$/u.test(value)) {
    return options.find((option) =>
      new RegExp(`^${value}(?:\\s|-)`, 'iu').test(option.name),
    ) ?? null;
  }
  return null;
}

function projectValuesEqual(field, current, desired) {
  if (normalizedName(current?.value) === normalizedName(desired)) return true;
  if (field.__typename !== 'ProjectV2SingleSelectField') return false;
  const option = optionFor(field, desired);
  return Boolean(option && normalizedName(current?.value) === normalizedName(option.name));
}

function fieldMutation(field, value) {
  if (field.__typename === 'ProjectV2SingleSelectField') {
    const option = optionFor(field, value);
    if (!option) return {error: `Project field ${field.name} has no option named ${value}.`};
    return {
      fieldId: field.id,
      inputValue: {singleSelectOptionId: option.id},
      value,
    };
  }
  if (field.__typename === 'ProjectV2IterationField') {
    return {error: `Project iteration field ${field.name} cannot be inferred safely.`};
  }
  if (field.dataType?.toUpperCase() === 'DATE' || field.name.toLowerCase().includes('date')) {
    return {fieldId: field.id, inputValue: {date: value}, value};
  }
  return {fieldId: field.id, inputValue: {text: value}, value};
}

function createPublicOperation(operation, result = {}) {
  return {
    target: operation.target,
    operation: operation.operation,
    beforeHash: operation.beforeHash,
    expectedAfter: clone(operation.expectedAfter),
    reason: operation.reason,
    conflict: result.conflict ?? operation.conflict ?? null,
    ...(result.result ? {result: result.result} : {}),
  };
}

function warning(target, reason, conflict = 'audit-warning') {
  return {target, reason, conflict};
}

function dependencyIssueSettings(config) {
  const configured = config.backfill?.dependencyTrackingIssue;
  if (configured === null || configured === false) return null;
  const settings = configured ?? {};
  return {
    title: settings.title ?? 'Dependency updates',
    variable: settings.variable ?? 'DEPENDENCY_TRACKING_ISSUE',
    labels: settings.labels ?? [
      'type:chore',
      'area:tooling',
      'priority:p2-normal',
      'status:in-progress',
    ],
    assignees: settings.assignees ?? [],
  };
}

function dependencyIssueNumberFromEnvironment(settings) {
  const value = process.env[settings.variable];
  return value && /^[1-9]\d*$/.test(value) ? Number(value) : null;
}

function relationMapping(config, number) {
  const value = config.backfill?.relationMappings?.[String(number)];
  if (value === undefined || value === null) return null;
  if (value === 'dependency-tracking') return value;
  return positiveNumber(value, `relationMappings.${number}`);
}

function isDependabotPullRequest(pullRequest) {
  return pullRequest.user?.login === 'dependabot[bot]' &&
    pullRequest.head?.ref?.startsWith('dependabot/');
}

function addDependencyCreateOperation({issues, settings, operations}) {
  const existing = issues.find((issue) => issue.title?.trim() === settings.title);
  if (existing) return existing.number;
  operations.push({
    target: `Issue "${settings.title}"`,
    operation: 'create dependency tracking issue',
    beforeHash: hashValue({exists: false, title: settings.title}),
    expectedAfter: {
      title: settings.title,
      labels: sortedUnique(settings.labels),
      issueNumber: 'allocated-on-apply',
    },
    reason: 'The approved initial backfill requires one monthly aggregation Issue.',
    conflict: null,
    kind: 'create-dependency-issue',
  });
  return 'dependency-tracking';
}

function buildRelationOperations({pullRequests, issues, config, selectedPrNumber}) {
  const operations = [];
  const warnings = [];
  const settings = dependencyIssueSettings(config);
  const configuredDependencyIssueNumber = settings
    ? dependencyIssueNumberFromEnvironment(settings)
    : null;
  let dependencyIssueNumber = configuredDependencyIssueNumber;
  const relevantPullRequests = pullRequests.filter((pullRequest) =>
    !selectedPrNumber || pullRequest.number === selectedPrNumber,
  );
  const initialBackfillNeedsIssue = relevantPullRequests.some((pullRequest) =>
    relationMapping(config, pullRequest.number) === 'dependency-tracking' &&
    relationNumbers(pullRequest.body).length === 0,
  );
  if (initialBackfillNeedsIssue && !dependencyIssueNumber) {
    if (settings) {
      dependencyIssueNumber = addDependencyCreateOperation({issues, settings, operations});
    } else {
      warnings.push(warning(
        'Dependency tracking',
        'Dependency-tracking Issue creation is disabled by configuration.',
        'dependency-tracking-disabled',
      ));
    }
  }
  for (const pullRequest of relevantPullRequests) {
    if (relationNumbers(pullRequest.body).length > 0) continue;
    let target = relationMapping(config, pullRequest.number);
    if (target === null && isDependabotPullRequest(pullRequest)) {
      if (!settings) {
        warnings.push(warning(
          `PR #${pullRequest.number}`,
          'Dependabot PR relations are not configured because dependency-tracking is disabled.',
          'dependency-tracking-disabled',
        ));
        continue;
      }
      target = configuredDependencyIssueNumber;
      if (!target) {
        warnings.push(warning(
          `PR #${pullRequest.number}`,
          `Dependabot PR has no Issue relation and ${settings.variable} is not configured.`,
          'missing-dependency-tracking-issue',
        ));
        continue;
      }
    }
    if (!target) continue;
    const afterBody = target === 'dependency-tracking'
      ? appendIssueRelation(pullRequest.body, 0).replace('#0', '#<allocated-on-apply>')
      : appendIssueRelation(pullRequest.body, target);
    const relation = target === 'dependency-tracking'
      ? 'Related to #<allocated-on-apply>'
      : `Related to #${target}`;
    operations.push({
      target: `PR #${pullRequest.number}`,
      operation: 'append issue relation',
      beforeHash: pullRequestBodyHash(pullRequest),
      expectedAfter: {
        bodyHash: hashBody(afterBody),
        relation,
      },
      reason: target === 'dependency-tracking'
        ? 'Dependabot dependency update belongs to the configured tracking Issue.'
        : `Approved relation to Issue #${target}.`,
      conflict: null,
      kind: 'pr-relation',
      prNumber: pullRequest.number,
      relationIssueNumber: target,
    });
  }
  return {operations, warnings, dependencyIssueNumber};
}

function buildMergedPullRequestStatusOperations({pullRequests, selectedPrNumber}) {
  const operations = [];
  const warnings = [];
  for (const pullRequest of pullRequests) {
    if (selectedPrNumber && pullRequest.number !== selectedPrNumber) continue;
    const labels = labelNames(pullRequest);
    if (pullRequest.merged_at ?? pullRequest.mergedAt) {
      const desired = labels.filter((label) => !ACTIVE_STATUS_LABELS.has(label));
      if (!desired.includes('status:done')) desired.push('status:done');
      if (sameLabelSet(desired, labels)) continue;
      operations.push({
        target: `PR #${pullRequest.number}`,
        operation: 'normalize merged PR status labels',
        beforeHash: pullRequestLabelHash(pullRequest),
        expectedAfter: {labels: desired},
        reason: 'Merged pull requests must retain status:done and no active status label.',
        conflict: null,
        kind: 'pr-labels',
        prNumber: pullRequest.number,
        labels: desired,
      });
    } else if (pullRequest.state === 'closed' && labels.includes('status:done')) {
      warnings.push(warning(
        `PR #${pullRequest.number}`,
        'Closed-unmerged PR claims status:done; human review is required before removal.',
        'closed-without-merge',
      ));
    }
  }
  return {operations, warnings};
}

function buildIssueStatusOperations({issues, selectedIssueNumber}) {
  const operations = [];
  const warnings = [];
  for (const issue of issues) {
    if (selectedIssueNumber && issue.number !== selectedIssueNumber) continue;
    const status = buildIssueStatusLabels(issue);
    if (status.plan.warning) {
      warnings.push(warning(`Issue #${issue.number}`, status.plan.warning, 'close-reason'));
      continue;
    }
    const current = labelNames(issue);
    if (sameLabelSet(status.labels, current)) continue;
    operations.push({
      target: `Issue #${issue.number}`,
      operation: 'normalize issue status labels',
      beforeHash: issueStateHash(issue),
      expectedAfter: {
        bodyHash: hashBody(issue.body),
        labels: status.labels,
      },
      reason: issue.state === 'closed'
        ? 'Completed Issue must use status:done.'
        : 'Open Issue status labels must resolve to exactly one Project status.',
      conflict: null,
      kind: 'issue-labels',
      issueNumber: issue.number,
      labels: status.labels,
    });
  }
  return {operations, warnings};
}

async function collectFileCounts(client, repo, issues, pullRequests) {
  const needed = new Set();
  for (const issue of issues) {
    for (const pullRequest of pullRequests) {
      if (relationNumbers(pullRequest.body).includes(issue.number)) {
        needed.add(pullRequest.number);
      }
    }
  }
  const counts = new Map();
  await Promise.all([...needed].map(async (number) => {
    const pullRequest = pullRequests.find((item) => item.number === number);
    if (typeof pullRequest?.changed_files === 'number') {
      counts.set(number, pullRequest.changed_files);
    } else {
      counts.set(number, await fetchPullRequestFiles(client, repo, number));
    }
  }));
  return counts;
}

function projectOperations({issues, pullRequests, project, config, fileCounts, repo, selectedIssueNumber}) {
  const operations = [];
  const warnings = [];
  const fields = config.project.fields;
  const projectIssues = issues.filter((issue) =>
    !selectedIssueNumber || issue.number === selectedIssueNumber,
  );
  for (const issue of projectIssues) {
    const item = findProjectItem(project, repo, issue.number);
    if (!item) {
      warnings.push(warning(
        `Issue #${issue.number}`,
        `Issue is not currently a Project #${project.number} item; no Project item will be added.`,
        'not-in-project',
      ));
      continue;
    }
    const status = deriveIssueStatus(issue);
    if (status.warning) continue;
    const metrics = relationMetrics(issue.number, pullRequests, fileCounts);
    const desired = projectFieldPlan(issue, {
      ...metrics,
      scopeItems: countScopeItems(issue.body),
    });
    if (status.value) desired[fields.status] = status.value;
    const changes = [];
    const expectedFields = {};
    for (const [configuredName, value] of Object.entries({
      [fields.status]: desired[fields.status],
      [fields.priority]: desired.Priority,
      [fields.size]: desired.Size,
      [fields.workstream]: desired.Workstream,
      ...(desired['Target date'] ? {[fields.targetDate]: desired['Target date']} : {}),
    })) {
      if (value === undefined) continue;
      const field = findProjectField(project, configuredName);
      if (!field) {
        warnings.push(warning(
          `Issue #${issue.number}`,
          `Project field ${configuredName} is missing; it was not changed.`,
          'missing-project-field',
        ));
        continue;
      }
      const current = currentProjectValue(item, configuredName);
      if (projectValuesEqual(field, current, value)) continue;
      const mutation = fieldMutation(field, value);
      if (mutation.error) {
        warnings.push(warning(`Issue #${issue.number}`, mutation.error, 'missing-project-option'));
        continue;
      }
      changes.push({...mutation, fieldName: configuredName});
      expectedFields[configuredName] = value;
    }
    if (changes.length === 0) continue;
    operations.push({
      target: `Issue #${issue.number} / Project #${project.number}`,
      operation: 'synchronize Project fields',
      beforeHash: projectItemHash(item),
      expectedAfter: {fields: expectedFields},
      reason: 'Project fields are derived from current Issue metadata under the fixed synchronization rules.',
      conflict: null,
      kind: 'project-fields',
      issueNumber: issue.number,
      itemId: item.id,
      changes,
    });
  }
  return {operations, warnings};
}

function dependencyIssueNumberForRelation(operation, dependencyIssueNumber) {
  return operation.relationIssueNumber === 'dependency-tracking'
    ? dependencyIssueNumber
    : operation.relationIssueNumber;
}

export async function auditRepository({client, repo, projectOwner = '@me', projectNumber, config, issueNumber, prNumber} = {}) {
  if (!client) throw new Error('auditRepository requires a GitHub API client.');
  const [issues, pullRequests, project] = await Promise.all([
    listIssues(client, repo),
    listPullRequests(client, repo),
    fetchProject(client, projectOwner, projectNumber),
  ]);
  const fileCounts = await collectFileCounts(client, repo, issues, pullRequests);
  const selectedPullRequestNumber = prNumber ?? (issueNumber ? -1 : undefined);
  const selectedIssue = issueNumber ?? (prNumber ? -1 : undefined);
  const relation = buildRelationOperations({
    pullRequests,
    issues,
    config,
    selectedPrNumber: selectedPullRequestNumber,
  });
  const merged = buildMergedPullRequestStatusOperations({
    pullRequests,
    selectedPrNumber: selectedPullRequestNumber,
  });
  const issueStatus = buildIssueStatusOperations({issues, selectedIssueNumber: selectedIssue});
  const projectPlan = projectOperations({
    issues,
    pullRequests,
    project,
    config,
    fileCounts,
    repo,
    selectedIssueNumber: selectedIssue,
  });
  const operations = [
    ...relation.operations,
    ...merged.operations,
    ...issueStatus.operations,
    ...projectPlan.operations,
  ];
  const warnings = [
    ...relation.warnings,
    ...merged.warnings,
    ...issueStatus.warnings,
    ...projectPlan.warnings,
  ];
  return {
    repository: repo,
    project: {owner: project.owner, number: project.number, title: project.title},
    issues,
    pullRequests,
    projectData: project,
    config,
    dependencyIssueNumber: relation.dependencyIssueNumber,
    operations,
    warnings,
  };
}

async function freshProjectItem(client, owner, number, repo, issueNumber) {
  const project = await fetchProject(client, owner, number);
  return {project, item: findProjectItem(project, repo, issueNumber)};
}

async function applyCreateDependencyIssue({client, repo, config, operation}) {
  const settings = dependencyIssueSettings(config);
  const issues = await listIssues(client, repo);
  const existing = issues.find((issue) => issue.title?.trim() === settings.title);
  if (existing) return {result: 'already-exists', issueNumber: existing.number};
  const created = await client.rest('POST', `/repos/${repo}/issues`, {
    title: settings.title,
    body: DEFAULT_DEPENDENCY_BODY,
    labels: settings.labels,
    assignees: settings.assignees,
  });
  if (!created?.number) throw new Error('GitHub did not return the new dependency tracking Issue number.');
  return {result: 'applied', issueNumber: created.number};
}

async function applyPullRequestRelation({client, repo, operation, dependencyIssueNumber}) {
  const issueNumber = dependencyIssueNumberForRelation(operation, dependencyIssueNumber);
  if (!issueNumber) return {result: 'skipped', conflict: 'missing-dependency-tracking-issue'};
  const pullRequest = await fetchPullRequest(client, repo, operation.prNumber);
  if (pullRequestBodyHash(pullRequest) !== operation.beforeHash) {
    return {result: 'skipped', conflict: 'stale-read'};
  }
  if (relationNumbers(pullRequest.body).length > 0) {
    return {result: 'skipped', conflict: 'relation-changed'};
  }
  const body = appendIssueRelation(pullRequest.body, issueNumber);
  await client.rest('PATCH', `/repos/${repo}/issues/${operation.prNumber}`, {body});
  return {result: 'applied'};
}

async function applyLabels({client, repo, operation, number, kind}) {
  const current = kind === 'pr-labels'
    ? await fetchPullRequest(client, repo, number)
    : await fetchIssue(client, repo, number);
  const currentHash = kind === 'pr-labels'
    ? pullRequestLabelHash(current)
    : issueStateHash(current);
  if (currentHash !== operation.beforeHash) return {result: 'skipped', conflict: 'stale-read'};
  await client.rest('PUT', `/repos/${repo}/issues/${number}/labels`, {
    labels: operation.labels,
  });
  return {result: 'applied'};
}

async function updateProjectField(client, projectId, itemId, change) {
  const mutation = `
mutation($projectId: ID!, $itemId: ID!, $fieldId: ID!, $value: ProjectV2FieldValue!) {
  updateProjectV2ItemFieldValue(input: {
    projectId: $projectId
    itemId: $itemId
    fieldId: $fieldId
    value: $value
  }) { projectV2Item { id } }
}`;
  await client.graphql(mutation, {
    projectId,
    itemId,
    fieldId: change.fieldId,
    value: change.inputValue,
  });
}

async function applyProjectFields({client, repo, operation, projectOwner, projectNumber}) {
  const {project, item} = await freshProjectItem(
    client,
    projectOwner,
    projectNumber,
    repo,
    operation.issueNumber,
  );
  if (!item) return {result: 'skipped', conflict: 'not-in-project'};
  if (projectItemHash(item) !== operation.beforeHash) {
    return {result: 'skipped', conflict: 'stale-read'};
  }
  for (const change of operation.changes) {
    await updateProjectField(client, project.id, item.id, change);
  }
  return {result: 'applied'};
}

export async function applyOperations({client, audit, projectOwner = '@me', projectNumber} = {}) {
  let dependencyIssueNumber = audit.dependencyIssueNumber;
  const results = [];
  for (const operation of audit.operations) {
    try {
      if (operation.kind === 'create-dependency-issue') {
        const result = await applyCreateDependencyIssue({
          client,
          repo: audit.repository,
          config: audit.config,
          operation,
        });
        dependencyIssueNumber = result.issueNumber ?? dependencyIssueNumber;
        results.push(createPublicOperation(operation, result));
        continue;
      }
      let result;
      if (operation.kind === 'pr-relation') {
        result = await applyPullRequestRelation({
          client,
          repo: audit.repository,
          operation,
          dependencyIssueNumber,
        });
      } else if (operation.kind === 'pr-labels') {
        result = await applyLabels({
          client,
          repo: audit.repository,
          operation,
          number: operation.prNumber,
          kind: operation.kind,
        });
      } else if (operation.kind === 'issue-labels') {
        result = await applyLabels({
          client,
          repo: audit.repository,
          operation,
          number: operation.issueNumber,
          kind: operation.kind,
        });
      } else if (operation.kind === 'project-fields') {
        result = await applyProjectFields({
          client,
          repo: audit.repository,
          operation,
          projectOwner,
          projectNumber,
        });
      } else {
        result = {result: 'skipped', conflict: 'unknown-operation'};
      }
      results.push(createPublicOperation(operation, result));
    } catch (error) {
      results.push(createPublicOperation(operation, {
        result: 'skipped',
        conflict: `mutation-error:${error.message}`,
      }));
    }
  }
  return {
    ...audit,
    appliedOperations: results,
    dependencyIssueNumber,
  };
}

function reportFor(audit, mode, operations, warnings) {
  const resultCounts = operations.reduce((counts, operation) => {
    const key = operation.result ?? 'planned';
    counts[key] = (counts[key] ?? 0) + 1;
    return counts;
  }, {});
  const warningFailures = warnings.filter((item) => [
    'missing-dependency-tracking-issue',
    'missing-project-field',
    'missing-project-option',
  ].includes(item.conflict));
  const operationFailures = operations
    .filter((operation) => operation.conflict)
    .map(({target, reason, conflict}) => ({target, reason, conflict}));
  const failures = [...warningFailures, ...operationFailures];
  return {
    schemaVersion: 1,
    generatedAt: new Date().toISOString(),
    mode,
    repository: audit.repository,
    project: audit.project,
    privacy: {
      bodiesStored: false,
      bodyHashesStored: true,
      rawMetadataExcluded: true,
    },
    summary: {
      ...resultCounts,
      warnings: warnings.length,
      failures: failures.length,
    },
    operations,
    warnings,
    failures,
  };
}

export async function run({argv = process.argv.slice(2), client} = {}) {
  const parsed = parseArgs(argv);
  if (parsed.help) {
    console.log(usage());
    return 0;
  }
  const config = readSyncConfig(parsed.config);
  const options = configOptions(parsed, config);
  const api = client ?? new GitHubApiClient();
  const audit = await auditRepository({
    client: api,
    repo: options.repo,
    projectOwner: options.projectOwner,
    projectNumber: options.projectNumber,
    config,
    issueNumber: options.issueNumber,
    prNumber: options.prNumber,
  });
  let operations = audit.operations.map((operation) => createPublicOperation(operation));
  let warnings = audit.warnings;
  if (options.mode === 'apply') {
    const applied = await applyOperations({
      client: api,
      audit,
      projectOwner: options.projectOwner,
      projectNumber: options.projectNumber,
    });
    operations = applied.appliedOperations;
    audit.dependencyIssueNumber = applied.dependencyIssueNumber;
  }
  const report = reportFor(audit, options.mode, operations, warnings);
  const output = `${JSON.stringify(report, null, 2)}\n`;
  if (options.output) {
    mkdirSync(dirname(options.output), {recursive: true});
    writeFileSync(options.output, output, {mode: 0o600});
  }
  console.log(output.trimEnd());
  return report.failures.length > 0 ? 1 : 0;
}

if (import.meta.url === `file://${process.argv[1]}`) {
  run().then((code) => {
    process.exitCode = code;
  }).catch((error) => {
    console.error(`github-project-sync: ${error.message}`);
    process.exitCode = 1;
  });
}

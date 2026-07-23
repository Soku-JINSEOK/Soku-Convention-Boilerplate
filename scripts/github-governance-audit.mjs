#!/usr/bin/env node

import {execFile} from 'node:child_process';
import {createHash} from 'node:crypto';
import {
  existsSync,
  mkdirSync,
  readFileSync,
  writeFileSync,
} from 'node:fs';
import {dirname, resolve} from 'node:path';
import {promisify} from 'node:util';
import {fileURLToPath, pathToFileURL} from 'node:url';
import {
  TITLE_CONVENTIONS,
  contributionTitleOptionsForPullRequest,
  validateContributionTitle,
} from '../templates/_shared/commitlint/contribution-title.mjs';
import {
  readDependabotConfigurations,
  validateDependabotFiles,
} from './pull-request-policy.mjs';

const execFileAsync = promisify(execFile);
const REPOSITORY_ROOT = fileURLToPath(new URL('../', import.meta.url));
const DEFAULT_AS_OF = '2026-07-23T01:00:19Z';
const POLICY_EPOCHS = Object.freeze([
  {
    id: 'legacy',
    from: '1970-01-01T00:00:00Z',
    description: 'Initial repository rules',
  },
  {
    id: 'multilingual',
    from: '2026-07-14T03:50:01Z',
    description: 'English/Korean contribution templates',
  },
  {
    id: 'strengthened',
    from: '2026-07-18T04:27:29Z',
    description: 'Strengthened English/Korean/Japanese templates',
  },
  {
    id: 'strict',
    from: '2026-07-21T20:23:25Z',
    description: 'Strict automated governance',
  },
]);
const CLASSIFICATIONS = Object.freeze([
  'Compliant',
  'Correctable metadata',
  'Historical exception',
  'Blocked/missing evidence',
  'Not applicable',
]);
const CURRENT_ISSUES = new Set([91]);
const CURRENT_PULL_REQUESTS = new Set([92]);
const CURRENT_DEPENDABOT_PULL_REQUESTS = new Set([89, 90]);
const SUCCESSFUL_CONCLUSIONS = new Set(['success', 'neutral', 'skipped']);
const ACTIVE_STATUS_LABELS = new Set([
  'status:triage',
  'status:ready',
  'status:in-progress',
  'status:blocked',
]);

function usage() {
  return `Usage: scripts/github-governance-audit.mjs [options]

Read-only GitHub governance audit. All GitHub API calls use GET.

Options:
  --repo <owner/repo>  Repository to inspect (required)
  --as-of <ISO time>   Creation-time inventory cutoff (default: ${DEFAULT_AS_OF})
  --output <path>      Markdown report path (required)
  --help               Show this help`;
}

export function parseArgs(argv) {
  const options = {asOf: DEFAULT_AS_OF};
  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    if (argument === '--help') return {help: true};
    if (!['--repo', '--as-of', '--output'].includes(argument)) {
      throw new Error(`Unknown argument: ${argument}`);
    }
    const value = argv[index + 1];
    if (!value || value.startsWith('--')) {
      throw new Error(`${argument} requires a value.`);
    }
    index += 1;
    if (argument === '--repo') options.repo = value;
    if (argument === '--as-of') options.asOf = value;
    if (argument === '--output') options.output = value;
  }
  if (!/^[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+$/.test(options.repo ?? '')) {
    throw new Error('--repo must use owner/repository syntax.');
  }
  if (!options.output) throw new Error('--output is required.');
  if (Number.isNaN(Date.parse(options.asOf))) {
    throw new Error('--as-of must be an ISO-8601 timestamp.');
  }
  options.asOf = new Date(options.asOf).toISOString();
  options.output = resolve(options.output);
  return options;
}

export function hashBody(body) {
  return createHash('sha256').update(body ?? '', 'utf8').digest('hex');
}

export function epochFor(createdAt, kind, number) {
  if (kind === 'Issue' && CURRENT_ISSUES.has(number)) return 'current-normalized';
  if (kind === 'PR' && CURRENT_PULL_REQUESTS.has(number)) return 'current-normalized';
  if (kind === 'PR' && CURRENT_DEPENDABOT_PULL_REQUESTS.has(number)) {
    return 'dependabot-current';
  }
  let epoch = POLICY_EPOCHS[0].id;
  for (const candidate of POLICY_EPOCHS) {
    if (Date.parse(createdAt) >= Date.parse(candidate.from)) epoch = candidate.id;
  }
  return epoch;
}

function labelNames(item) {
  return (item.labels ?? []).map((label) =>
    typeof label === 'string' ? label : label.name,
  );
}

function assigneeNames(item) {
  return (item.assignees ?? []).map((assignee) => assignee.login);
}

function hasAxis(labels, axis) {
  return labels.some((label) => label.startsWith(`${axis}:`));
}

function hasLanguage(body, language) {
  const patterns = {
    english: /(?:🇬🇧|\bEnglish\b|Defect and expected behavior|Reproduction and scope|Goal and background|## Description|## Verification)/iu,
    korean: /(?:🇰🇷|한국어|목표|검증)/u,
    japanese: /(?:🇯🇵|日本語|目標|検証)/u,
  };
  return patterns[language].test(body);
}

function semanticSignals(body) {
  return {
    goal: /(?:goal|objective|expected behavior|background|defect|목표|目標)/iu.test(body),
    scope: /(?:scope|범위|範囲)/iu.test(body),
    constraints: /(?:constraint|boundary|non-destructive|비파괴|제약|非破壊|制約)/iu.test(
      body,
    ),
    acceptance: /(?:acceptance|definition of done|수용 기준|완료 조건|受け入れ|完了条件)/iu.test(
      body,
    ),
    evidence: /(?:evidence|verification|test|증거|검증|証拠|検証)/iu.test(body),
    ai: /(?:AI Assistance|AI 지원|AI 支援|Codex|Claude|None)/iu.test(body),
  };
}

function relationNumbers(body) {
  return [
    ...body.matchAll(
      /\b(?:Closes|Fixes|Resolves|Related to)\s+#([1-9]\d*)\b/giu,
    ),
  ].map((match) => Number(match[1]));
}

function taskReportNumbers(body) {
  return [
    ...body.matchAll(/docs\/issues\/issue-([1-9]\d*)-task-report\.md/giu),
  ].map((match) => Number(match[1]));
}

export function validateTitleForEpoch(title, epoch, options = {}) {
  if (epoch !== 'multilingual') return validateContributionTitle(title, options);
  const value = title.trim();
  const convention = TITLE_CONVENTIONS.find(([emoji, type]) =>
    value.startsWith(`${emoji} ${type}(`),
  );
  if (!convention) return {valid: false};
  const [emoji, type] = convention;
  const remainder = value.slice(`${emoji} ${type}(`.length);
  const separator = remainder.indexOf('): ');
  if (separator < 1) return {valid: false};
  const scope = remainder.slice(0, separator);
  const subject = remainder.slice(separator + 3);
  return {
    valid:
      /^[a-z0-9]+(?:-[a-z0-9]+)*$/u.test(scope) &&
      Boolean(subject) &&
      subject === subject.trim() &&
      /^[\x20-\x7e]+$/u.test(subject),
  };
}

function issueTitleValid(title, epoch) {
  return validateTitleForEpoch(title, epoch).valid;
}

function expectedTypeLabel(title) {
  const type = /(?:^|\s)(feat|fix|docs|chore|refactor|style|test|ci|build|security|perf)(?:!?\()/u.exec(
    title,
  )?.[1];
  return {
    feat: 'type:feature',
    fix: 'type:bug',
    docs: 'type:docs',
    chore: 'type:chore',
    refactor: 'type:refactor',
    style: 'type:chore',
    test: 'type:chore',
    ci: 'type:chore',
    build: 'type:chore',
    security: 'type:chore',
    perf: 'type:refactor',
  }[type];
}

function expectedPriorityLabel(body) {
  const match = /\bP([0-3])\s*-\s*(critical|high|normal|low)\b/iu.exec(body);
  if (!match) return null;
  return `priority:p${match[1]}-${match[2].toLowerCase()}`;
}

function expectedAreaLabel(body) {
  const match = /##\s*📦\s*Area\s*\n+\s*(standards|templates|automation|ci|docs|security|tooling)\b/iu.exec(
    body,
  );
  return match ? `area:${match[1].toLowerCase()}` : null;
}

function addCandidate(candidates, target, mutation, evidence) {
  candidates.push({target, mutation, evidence, approval: 'Required'});
}

function determineClassification({
  epoch,
  metadataProblems,
  immutableProblems,
  notApplicable = false,
}) {
  if (immutableProblems.length > 0) {
    return ['strict', 'current-normalized', 'dependabot-current'].includes(epoch)
      ? 'Blocked/missing evidence'
      : 'Historical exception';
  }
  if (metadataProblems.length > 0) return 'Correctable metadata';
  if (notApplicable) return 'Not applicable';
  return 'Compliant';
}

function issueEvidenceSummary(evidence) {
  const links = evidence.linkedPullRequests.length > 0
    ? evidence.linkedPullRequests.map((number) => `#${number}`).join(', ')
    : 'none';
  return `${evidence.commentCount} comments; linked PRs ${links}`;
}

export function classifyIssue(item, evidence = {}) {
  const body = item.body ?? '';
  const labels = labelNames(item);
  const assignees = assigneeNames(item);
  const epoch = epochFor(item.created_at, 'Issue', item.number);
  const strict = ['strict', 'current-normalized'].includes(epoch);
  const multilingual = epoch !== 'legacy';
  const strengthened = ['strengthened', 'strict', 'current-normalized'].includes(epoch);
  const signals = semanticSignals(body);
  const metadataProblems = [];
  const immutableProblems = [];
  const candidates = [];

  if (!body.trim()) immutableProblems.push('Issue body is empty.');
  if (multilingual && !issueTitleValid(item.title, epoch)) {
    immutableProblems.push('Title does not match the contribution convention.');
  }
  for (const key of ['goal', 'scope', 'acceptance']) {
    if (!signals[key]) immutableProblems.push(`Body does not identify ${key}.`);
  }
  if (multilingual && (!hasLanguage(body, 'english') || !hasLanguage(body, 'korean'))) {
    immutableProblems.push('Required English/Korean contribution blocks are incomplete.');
  }
  if (strengthened && !hasLanguage(body, 'japanese')) {
    immutableProblems.push('Required Japanese summary is missing.');
  }
  if (strengthened) {
    for (const key of ['constraints', 'evidence', 'ai']) {
      if (!signals[key]) immutableProblems.push(`Body does not record ${key}.`);
    }
  }

  const requiredAxes = strict
    ? ['type', 'priority', 'status', 'area']
    : multilingual
      ? ['type']
      : [];
  for (const axis of requiredAxes) {
    if (!hasAxis(labels, axis)) metadataProblems.push(`Missing ${axis}: label.`);
  }
  if (strict && !assignees.some((name) => name.toLowerCase() === 'soku-jinseok')) {
    metadataProblems.push('Missing Soku-JINSEOK assignee.');
    addCandidate(
      candidates,
      `Issue #${item.number}`,
      'Assign Soku-JINSEOK',
      'Current strict Issue contract',
    );
  }

  const expectedType = expectedTypeLabel(item.title);
  if (requiredAxes.includes('type') && !hasAxis(labels, 'type') && expectedType) {
    addCandidate(
      candidates,
      `Issue #${item.number}`,
      `Add ${expectedType}`,
      'Title type maps unambiguously to the canonical label',
    );
  }
  const inferredMetadata = [
    ['priority', expectedPriorityLabel(body)],
    ['area', expectedAreaLabel(body)],
  ];
  for (const [axis, label] of inferredMetadata) {
    if (strict && !hasAxis(labels, axis) && label) {
      addCandidate(
        candidates,
        `Issue #${item.number}`,
        `Add ${label}`,
        `Issue form records an exact ${axis} value`,
      );
    }
  }
  if (item.state === 'closed' && item.state_reason === 'completed') {
    if (!labels.includes('status:done') && strict) {
      metadataProblems.push('Completed Issue lacks status:done.');
      const removals = labels.filter((label) => ACTIVE_STATUS_LABELS.has(label));
      addCandidate(
        candidates,
        `Issue #${item.number}`,
        `Add status:done${removals.length ? `; remove ${removals.join(', ')}` : ''}`,
        'Closed with completed reason',
      );
    }
  }
  if (item.state === 'closed' && item.state_reason !== 'completed' && labels.includes('status:done')) {
    metadataProblems.push('Uncompleted Issue claims status:done.');
    addCandidate(
      candidates,
      `Issue #${item.number}`,
      'Remove status:done or correct the close reason after human review',
      `Close reason is ${item.state_reason ?? 'not recorded'}`,
    );
  }

  const reportNumbers = taskReportNumbers(body);
  if (
    strict &&
    !reportNumbers.includes(item.number) &&
    !evidence.taskReportExists
  ) {
    immutableProblems.push('No matching task-report evidence exists.');
  }
  if (strict && (evidence.linkedPullRequests ?? []).length === 0) {
    immutableProblems.push('No linked pull-request evidence was found.');
  }
  const findings = [...immutableProblems, ...metadataProblems];
  if (findings.length === 0) findings.push('No defects under the applicable creation-time contract.');
  return {
    number: item.number,
    kind: 'Issue',
    epoch,
    bodyHash: hashBody(body),
    classification: determineClassification({
      epoch,
      metadataProblems,
      immutableProblems,
    }),
    findings,
    evidence: issueEvidenceSummary({
      commentCount: evidence.commentCount ?? 0,
      linkedPullRequests: evidence.linkedPullRequests ?? [],
    }),
    candidates,
  };
}

function checksSummary(evidence) {
  const checks = evidence.checks ?? [];
  const successful = checks.filter((check) =>
    SUCCESSFUL_CONCLUSIONS.has(check.conclusion),
  ).length;
  const verified = (evidence.commits ?? []).filter(
    (commit) => commit.commit?.verification?.verified,
  ).length;
  const mergeSignature = evidence.mergeCommit
    ? evidence.mergeCommit.commit?.verification?.verified
      ? 'verified'
      : 'unverified'
    : 'not applicable';
  const parentCount = evidence.mergeCommit?.parents?.length;
  const mergeTopology = evidence.mergeCommit
    ? parentCount === 1
      ? 'single-parent squash/rebase'
      : `${parentCount ?? 'unknown'}-parent merge`
    : 'not merged';
  return `${evidence.commits?.length ?? 0} commits (${verified} verified); ` +
    `${checks.length} checks (${successful} successful); ` +
    `${evidence.reviews?.length ?? 0} reviews; ` +
    `${(evidence.issueComments?.length ?? 0) + (evidence.reviewComments?.length ?? 0)} comments; ` +
    `${mergeTopology}; signature ${mergeSignature}`;
}

function requiredCheckMissing(checks, pattern) {
  return !checks.some(
    (check) => pattern.test(check.name) && SUCCESSFUL_CONCLUSIONS.has(check.conclusion),
  );
}

export function classifyPullRequest(item, evidence = {}, options = {}) {
  const body = item.body ?? '';
  const labels = labelNames(item);
  const assignees = assigneeNames(item);
  const epoch = epochFor(item.created_at, 'PR', item.number);
  const strict = ['strict', 'current-normalized', 'dependabot-current'].includes(epoch);
  const multilingual = epoch !== 'legacy';
  const strengthened = ['strengthened', 'strict', 'current-normalized'].includes(epoch);
  const metadataProblems = [];
  const immutableProblems = [];
  const candidates = [];
  const botContract = epoch === 'dependabot-current';
  const titleOptions = contributionTitleOptionsForPullRequest(
    item.user?.login ?? '',
    item.head?.ref ?? '',
  );
  const titleResult = validateTitleForEpoch(
    item.title,
    epoch,
    titleOptions,
  );

  if (multilingual && !titleResult.valid) {
    immutableProblems.push('PR title does not match the applicable contribution convention.');
  }
  if (botContract) {
    if (item.user?.login !== 'dependabot[bot]' || !item.head?.ref?.startsWith('dependabot/')) {
      immutableProblems.push('PR does not satisfy the exact Dependabot author/ref identity.');
    }
    const fileErrors = validateDependabotFiles({
      headRef: item.head?.ref ?? '',
      changedFiles: (evidence.files ?? []).map((file) => file.filename),
      configurations: options.dependabotConfigurations ?? [],
    });
    immutableProblems.push(...fileErrors);
  } else {
    const signals = semanticSignals(body);
    const relations = relationNumbers(body);
    const reports = taskReportNumbers(body);
    if (!body.trim()) immutableProblems.push('PR body is empty.');
    if (multilingual && relations.length !== 1) {
      immutableProblems.push('PR does not record exactly one local Issue relation.');
    }
    if (
      strict &&
      item.draft &&
      /\b(?:Closes|Fixes|Resolves)\s+#[1-9]\d*\b/iu.test(body)
    ) {
      immutableProblems.push('Draft PR records a closing Issue relation.');
    }
    if (multilingual && !reports.includes(relations[0])) {
      immutableProblems.push('PR does not link the matching task report.');
    }
    if (multilingual && (!hasLanguage(body, 'english') || !hasLanguage(body, 'korean'))) {
      immutableProblems.push('Required English/Korean contribution blocks are incomplete.');
    }
    if (strengthened && !hasLanguage(body, 'japanese')) {
      immutableProblems.push('Required Japanese summary is missing.');
    }
    if (strengthened) {
      for (const key of ['goal', 'scope', 'acceptance', 'constraints', 'evidence', 'ai']) {
        if (!signals[key]) immutableProblems.push(`PR body does not record ${key}.`);
      }
      if (!/Governance profile/iu.test(body)) {
        immutableProblems.push('PR body does not record a governance profile.');
      }
    }
  }

  const requiredType = botContract ? 'type:chore' : null;
  const requiredArea = botContract ? 'area:tooling' : null;
  if (strict && !hasAxis(labels, 'type')) metadataProblems.push('Missing type: label.');
  if (strict && !hasAxis(labels, 'area')) metadataProblems.push('Missing area: label.');
  if (requiredType && !labels.includes(requiredType)) {
    metadataProblems.push(`Dependabot PR lacks ${requiredType}.`);
  }
  if (requiredArea && !labels.includes(requiredArea)) {
    metadataProblems.push(`Dependabot PR lacks ${requiredArea}.`);
  }
  if (strict && !assignees.some((name) => name.toLowerCase() === 'soku-jinseok')) {
    metadataProblems.push('Missing Soku-JINSEOK assignee.');
    addCandidate(
      candidates,
      `PR #${item.number}`,
      'Assign Soku-JINSEOK',
      'Applicable strict PR contract',
    );
  }
  if (strict && !hasAxis(labels, 'type')) {
    const inferred = expectedTypeLabel(item.title);
    if (inferred) {
      addCandidate(
        candidates,
        `PR #${item.number}`,
        `Add ${inferred}`,
        'Title type maps unambiguously to the canonical label',
      );
    }
  }
  if (botContract && !labels.includes('area:tooling')) {
    addCandidate(
      candidates,
      `PR #${item.number}`,
      'Add area:tooling',
      'Exact Dependabot contract',
    );
  }

  const checks = evidence.checks ?? [];
  if (strict && requiredCheckMissing(checks, /^Validation Gate$/u)) {
    immutableProblems.push('Successful Validation Gate evidence is missing.');
  }
  if (strict && requiredCheckMissing(checks, /^PR Metadata Gate$/u)) {
    immutableProblems.push('Successful PR Metadata Gate evidence is missing.');
  }
  const commits = evidence.commits ?? [];
  if (multilingual && commits.length === 0) {
    immutableProblems.push('Commit evidence is missing.');
  }
  if (multilingual && commits.some((commit) => {
    const title = commit.commit?.message?.split(/\r?\n/u)[0] ?? '';
    const commitOptions = botContract
      ? contributionTitleOptionsForPullRequest(
        item.user?.login ?? '',
        item.head?.ref ?? '',
      )
      : {};
    return !validateTitleForEpoch(title, epoch, commitOptions).valid;
  })) {
    immutableProblems.push('One or more commit titles violate the applicable convention.');
  }
  if (strict && item.merged_at && !evidence.mergeCommit) {
    immutableProblems.push('Merge commit/signature evidence could not be read.');
  }
  if (strict && item.merged_at && !labels.includes('status:done')) {
    metadataProblems.push('Merged PR lacks status:done.');
    const removals = labels.filter((label) => ACTIVE_STATUS_LABELS.has(label));
    addCandidate(
      candidates,
      `PR #${item.number}`,
      `Add status:done${removals.length ? `; remove ${removals.join(', ')}` : ''}`,
      'PR is merged',
    );
  }
  if (!item.merged_at && item.state === 'closed' && labels.includes('status:done')) {
    metadataProblems.push('Closed-unmerged PR claims status:done.');
    addCandidate(
      candidates,
      `PR #${item.number}`,
      'Remove status:done',
      'PR was closed without merge',
    );
  }

  const notApplicable =
    epoch === 'legacy' && item.state === 'closed' && !item.merged_at &&
    immutableProblems.length === 0 && metadataProblems.length === 0;
  const findings = [...immutableProblems, ...metadataProblems];
  if (findings.length === 0) {
    findings.push(
      notApplicable
        ? 'Closed without merge before formal contribution governance.'
        : 'No defects under the applicable creation-time contract.',
    );
  }
  return {
    number: item.number,
    kind: 'PR',
    epoch,
    bodyHash: hashBody(body),
    classification: determineClassification({
      epoch,
      metadataProblems,
      immutableProblems,
      notApplicable,
    }),
    findings,
    evidence: checksSummary(evidence),
    candidates,
  };
}

function flattenPaginated(value) {
  if (!Array.isArray(value)) return value;
  return value.flatMap((entry) => (Array.isArray(entry) ? entry : [entry]));
}

export function createGhReader({execute} = {}) {
  const run = execute ?? (async (args) => execFileAsync('gh', args, {
    encoding: 'utf8',
    maxBuffer: 64 * 1024 * 1024,
  }));
  return async function read(endpoint, {paginate = false, optional = false} = {}) {
    const args = ['api', '--method', 'GET'];
    if (paginate) args.push('--paginate', '--slurp');
    args.push(endpoint);
    try {
      const result = await run(args);
      const stdout = typeof result === 'string' ? result : result.stdout;
      return flattenPaginated(JSON.parse(stdout || 'null'));
    } catch (error) {
      const diagnostic = `${error.stderr ?? ''}\n${error.message ?? ''}`;
      if (optional && /(?:HTTP 404|HTTP 409|HTTP 422)/u.test(diagnostic)) return null;
      throw new Error(`GET ${endpoint} failed: ${diagnostic.trim()}`, {cause: error});
    }
  };
}

async function mapLimit(items, limit, operation) {
  const results = new Array(items.length);
  let nextIndex = 0;
  async function worker() {
    while (nextIndex < items.length) {
      const index = nextIndex;
      nextIndex += 1;
      results[index] = await operation(items[index], index);
    }
  }
  await Promise.all(Array.from({length: Math.min(limit, items.length)}, worker));
  return results;
}

function linkedPullRequestsFromComments(comments) {
  const numbers = new Set();
  for (const comment of comments) {
    for (const match of (comment.body ?? '').matchAll(
      /(?:pull\/|pull request\s+#?|PR\s+#)([1-9]\d*)/giu,
    )) {
      numbers.add(Number(match[1]));
    }
  }
  return [...numbers].sort((left, right) => left - right);
}

function linkedPullRequestsFromText(body) {
  const numbers = new Set();
  for (const match of (body ?? '').matchAll(
    /(?:pull\/|pull request\s+#?|PR\s+#)([1-9]\d*)/giu,
  )) {
    numbers.add(Number(match[1]));
  }
  return [...numbers].sort((left, right) => left - right);
}

async function readPullRequestEvidence(read, repo, pullRequest) {
  const base = `repos/${repo}`;
  const number = pullRequest.number;
  const [commits, checksResponse, reviews, issueComments, reviewComments, files] =
    await Promise.all([
      read(`${base}/pulls/${number}/commits?per_page=100`, {paginate: true}),
      read(`${base}/commits/${pullRequest.head.sha}/check-runs?per_page=100`, {
        paginate: true,
        optional: true,
      }),
      read(`${base}/pulls/${number}/reviews?per_page=100`, {paginate: true}),
      read(`${base}/issues/${number}/comments?per_page=100`, {paginate: true}),
      read(`${base}/pulls/${number}/comments?per_page=100`, {paginate: true}),
      read(`${base}/pulls/${number}/files?per_page=100`, {paginate: true}),
    ]);
  const mergeCommit = pullRequest.merged_at && pullRequest.merge_commit_sha
    ? await read(`${base}/commits/${pullRequest.merge_commit_sha}`, {optional: true})
    : null;
  return {
    commits: commits ?? [],
    checks: Array.isArray(checksResponse)
      ? checksResponse.flatMap((page) => page.check_runs ?? [])
      : checksResponse?.check_runs ?? [],
    reviews: reviews ?? [],
    issueComments: issueComments ?? [],
    reviewComments: reviewComments ?? [],
    files: files ?? [],
    mergeCommit,
  };
}

function escapeTable(value) {
  return String(value).replaceAll('|', '\\|').replaceAll('\n', ' ');
}

function renderInventoryTable(items) {
  const lines = [
    '| Item | Epoch | Classification | Body SHA-256 | Evidence | Findings |',
    '| --- | --- | --- | --- | --- | --- |',
  ];
  for (const item of items) {
    lines.push(
      `| ${item.kind} #${item.number} | ${item.epoch} | ${item.classification} | ` +
      `\`${item.bodyHash}\` | ${escapeTable(item.evidence)} | ` +
      `${escapeTable(item.findings.join(' '))} |`,
    );
  }
  return lines.join('\n');
}

function renderMutationManifest(candidates) {
  if (candidates.length === 0) {
    return 'No unambiguous metadata mutation candidate was identified.';
  }
  const unique = [...new Map(candidates.map((candidate) => [
    `${candidate.target}\0${candidate.mutation}`,
    candidate,
  ])).values()];
  const lines = [
    '| Target | Proposed mutation | Evidence | Approval |',
    '| --- | --- | --- | --- |',
  ];
  for (const candidate of unique) {
    lines.push(
      `| ${escapeTable(candidate.target)} | ${escapeTable(candidate.mutation)} | ` +
      `${escapeTable(candidate.evidence)} | ${candidate.approval} |`,
    );
  }
  return lines.join('\n');
}

export function renderReport({repo, asOf, issues, pullRequests}) {
  const items = [...issues, ...pullRequests].sort((left, right) => {
    if (left.kind !== right.kind) return left.kind.localeCompare(right.kind);
    return left.number - right.number;
  });
  const counts = Object.fromEntries(CLASSIFICATIONS.map((name) => [name, 0]));
  for (const item of items) counts[item.classification] += 1;
  const candidates = items.flatMap((item) => item.candidates);
  const policyCommits = {
    multilingual: 'baa32118e473551b76bcd047db01faf3bbfbbf68',
    strengthened: '79f24250f9f27f9106b542152081261ed0d18ccd',
    strict: '30ab5b39766b26cef09726d048097c4ef0a857d8',
  };
  const epochs = POLICY_EPOCHS.map(({id, from, description}) => {
    const commit = policyCommits[id] ? ` (commit \`${policyCommits[id]}\`)` : '';
    return `- \`${id}\` from ${from}${commit}: ${description}.`;
  }).join('\n');
  const summary = CLASSIFICATIONS.map(
    (name) => `| ${name} | ${counts[name]} |`,
  ).join('\n');

  return `# GitHub Governance History Audit — 2026-07-23

## Scope and safety boundary

- Repository: \`${repo}\`
- Creation-time inventory cutoff: \`${asOf}\`
- Audited inventory: ${issues.length} Issues + ${pullRequests.length} pull requests = ${items.length} items
- Planned snapshot reconciliation: the approved plan expected 33 Issues + 48 pull requests = 81 items. The GitHub API contains ${pullRequests.length} pull requests before the same cutoff, so this report audits all ${items.length} items rather than silently omitting history.
- Collection method: \`gh api --method GET\`; the audit performs no GitHub mutation.
- Privacy boundary: Issue, PR, and comment bodies are processed only in memory. This report stores each Issue/PR body SHA-256, evidence counts, and judgments, never raw body content.
- Temporal boundary: inventory membership is fixed by creation time; the audit uses a fresh read of current metadata and evidence. Current normalization rules are intentionally used for Issue #91 and PR #92, and the merged Dependabot exception is used for open PRs #89 and #90.

## Policy epochs

${epochs}
- \`current-normalized\`: current contract explicitly requested for Issue #91 and PR #92.
- \`dependabot-current\`: exact current bot contract explicitly requested for PRs #89 and #90.

Current rules are not otherwise applied retroactively. Immutable historical body, commit, check, review, or merge gaps are classified as historical exceptions before strict governance and as blocked/missing evidence once strict governance applied.

This personal repository permits zero required approvals when no independent reviewer exists. Review and conversation counts remain evidence, but a zero count alone is not a defect. A single-parent merged commit is reported conservatively as squash/rebase topology because REST evidence alone cannot distinguish those two methods.

## Result summary

| Classification | Count |
| --- | ---: |
${summary}
| **Total** | **${items.length}** |

## Complete inventory

${renderInventoryTable(items)}

## Follow-up mutation manifest

This is a review-only manifest. No mutation was applied. Only unambiguous label, assignee, status, or close-reason candidates are listed; body, commit, check, review, merge, and signature history must remain unchanged. Every row requires separate approval and a fresh pre-mutation read.

${renderMutationManifest(candidates)}

## Approved mutation procedure

If a later approval authorizes any row, read the target again, confirm that its body hash still equals the value above, apply only the approved metadata mutation, and rerun this audit. Record before/after classification counts and prove that all body hashes remain unchanged.
`;
}

export async function auditRepository(options, dependencies = {}) {
  const read = dependencies.read ?? createGhReader();
  const base = `repos/${options.repo}`;
  const cutoff = Date.parse(options.asOf);
  const [issueFeed, pullRequestFeed] = await Promise.all([
    read(`${base}/issues?state=all&per_page=100&sort=created&direction=asc`, {
      paginate: true,
    }),
    read(`${base}/pulls?state=all&per_page=100&sort=created&direction=asc`, {
      paginate: true,
    }),
  ]);
  const issues = issueFeed.filter(
    (item) => !item.pull_request && Date.parse(item.created_at) <= cutoff,
  );
  const pullRequests = pullRequestFeed.filter(
    (item) => Date.parse(item.created_at) <= cutoff,
  );
  const dependabotConfigurations = readDependabotConfigurations(
    readFileSync(resolve(REPOSITORY_ROOT, '.github/dependabot.yml'), 'utf8'),
  );
  const pullRequestsByIssue = new Map();
  for (const pullRequest of pullRequests) {
    for (const issueNumber of relationNumbers(pullRequest.body ?? '')) {
      const linked = pullRequestsByIssue.get(issueNumber) ?? [];
      linked.push(pullRequest.number);
      pullRequestsByIssue.set(issueNumber, linked);
    }
  }

  const issueResults = await mapLimit(issues, 6, async (issue, index) => {
    const comments = await read(
      `${base}/issues/${issue.number}/comments?per_page=100`,
      {paginate: true},
    );
    if ((index + 1) % 10 === 0 || index + 1 === issues.length) {
      process.stderr.write(`Audited Issues: ${index + 1}/${issues.length}\n`);
    }
    return classifyIssue(issue, {
      commentCount: comments.length,
      linkedPullRequests: [...new Set([
        ...linkedPullRequestsFromText(issue.body),
        ...linkedPullRequestsFromComments(comments),
        ...(pullRequestsByIssue.get(issue.number) ?? []),
      ])].sort((left, right) => left - right),
      taskReportExists: existsSync(
        resolve(REPOSITORY_ROOT, `docs/issues/issue-${issue.number}-task-report.md`),
      ),
    });
  });
  const pullRequestResults = await mapLimit(
    pullRequests,
    6,
    async (pullRequest, index) => {
      const evidence = await readPullRequestEvidence(read, options.repo, pullRequest);
      if ((index + 1) % 10 === 0 || index + 1 === pullRequests.length) {
        process.stderr.write(
          `Audited pull requests: ${index + 1}/${pullRequests.length}\n`,
        );
      }
      return classifyPullRequest(pullRequest, evidence, {
        dependabotConfigurations,
      });
    },
  );
  return {issues: issueResults, pullRequests: pullRequestResults};
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  if (options.help) {
    process.stdout.write(`${usage()}\n`);
    return;
  }
  const result = await auditRepository(options);
  const report = renderReport({
    repo: options.repo,
    asOf: options.asOf,
    ...result,
  });
  mkdirSync(dirname(options.output), {recursive: true});
  writeFileSync(options.output, report, 'utf8');
  process.stdout.write(
    `Wrote ${result.issues.length + result.pullRequests.length} read-only judgments to ${options.output}\n`,
  );
}

const invokedPath = process.argv[1] ? pathToFileURL(resolve(process.argv[1])).href : '';
if (import.meta.url === invokedPath) {
  main().catch((error) => {
    process.stderr.write(`${error.message}\n`);
    process.exitCode = 1;
  });
}

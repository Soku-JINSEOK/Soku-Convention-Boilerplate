#!/usr/bin/env node
import {existsSync, readFileSync} from 'node:fs';
import {fileURLToPath} from 'node:url';
import {
  contributionTitleOptionsForPullRequest,
  isDependabotPullRequest,
  validateContributionTitle,
} from './contribution-title.mjs';

const CLOSING_PATTERN = /(?:closes|fixes|resolves)\s+(?:#[1-9]\d*|[\w.-]+\/[\w.-]+#[1-9]\d*|https:\/\/github\.com\/[\w.-]+\/[\w.-]+\/issues\/[1-9]\d*)/i;
const RELATION_PATTERN = /\b(?:Closes|Related to)\s+#[1-9]\d*\b/g;
const PLACEHOLDER_PATTERN = /<actual tool or none>|(?:closes|related to)\s+#\s*(?:\n|$)|issue-<n>|<!--\s*replace|\b(?:tbd|todo|no-issue)\b/iu;
const REQUIRED_BODY_SECTIONS = [
  '## 🔗 Common Metadata',
  '## 🇬🇧 English — Normative Source',
  '### 🎯 Goal',
  '### 📦 Scope',
  '### ✅ Acceptance Criteria',
  '### 🔒️ Security Boundary',
  '### 🧪 Verification',
  '### ⚠️ Risks and Follow-up',
  '## 🇰🇷 한국어 요약',
  '### 🎯 목표',
  '### 📦 핵심 범위',
  '### 🧪 검증',
  '### 🔒️ 비파괴 조건',
  '### ⚠️ 잔여 위험과 후속 작업',
  '## 🇯🇵 日本語の要約',
  '### 🎯 目標',
  '### 📦 主な範囲',
  '### 🧪 検証',
  '### 🔒️ 非破壊条件',
  '### ⚠️ 残存リスクと後続作業',
  '## Gitmoji Checklist',
  '## 🤖 AI Assistance',
];
const DEPENDABOT_ECOSYSTEMS = new Map([
  ['github_actions', 'github-actions'],
  ['go_modules', 'gomod'],
  ['npm_and_yarn', 'npm'],
  ['pip', 'pip'],
]);

function followsRequiredSectionOrder(body) {
  let previousIndex = -1;
  for (const section of REQUIRED_BODY_SECTIONS) {
    const index = body.indexOf(section);
    if (index === -1 || index <= previousIndex) return false;
    previousIndex = index;
  }
  return true;
}

function extractCommonMetadata(body) {
  const start = body.indexOf('## 🔗 Common Metadata');
  const end = body.indexOf('## 🇬🇧 English — Normative Source');
  return start >= 0 && end > start ? body.slice(start, end) : '';
}

function unquote(value) {
  return value.replace(/^['"]|['"]$/g, '');
}

export function readCanonicalLabels(source) {
  return new Set(
    [...source.matchAll(/^\s*- name:\s*["']?([^"'\n]+)["']?\s*$/gm)].map(
      ([, name]) => name.trim(),
    ),
  );
}

export function readExpectedProfile(source) {
  return /^- \*\*Governance profile:\*\* `([^`\r\n]+)`\s*$/m.exec(source)?.[1] ?? '';
}

export function readDependabotConfigurations(source) {
  const configurations = [];
  let current = null;
  for (const line of source.split(/\r?\n/)) {
    const ecosystem = /^\s*- package-ecosystem:\s*(\S+)\s*$/.exec(line);
    if (ecosystem) {
      current = {ecosystem: unquote(ecosystem[1]), directory: ''};
      configurations.push(current);
      continue;
    }
    const directory = /^\s+directory:\s*(\S+)\s*$/.exec(line);
    if (current && directory) current.directory = unquote(directory[1]);
  }
  return configurations.filter(({ecosystem, directory}) => ecosystem && directory);
}

function relativeToDirectory(path, directory) {
  const normalizedDirectory = directory.replace(/^\/+|\/+$/g, '');
  if (!normalizedDirectory) return path;
  const prefix = `${normalizedDirectory}/`;
  return path.startsWith(prefix) ? path.slice(prefix.length) : null;
}

function matchesEcosystemPath(path, configuration) {
  const relative = relativeToDirectory(path, configuration.directory);
  if (relative === null) return false;
  switch (configuration.ecosystem) {
    case 'github-actions':
      return /^\.github\/workflows\/.+\.ya?ml$/.test(relative);
    case 'gomod':
      return /^(?:go\.mod|go\.sum)$/.test(relative);
    case 'npm':
      return /^(?:package\.json|package-lock\.json|npm-shrinkwrap\.json|yarn\.lock|pnpm-lock\.yaml)$/.test(relative);
    case 'pip':
      return /^(?:requirements(?:[-_.][^/]*)?\.txt|pyproject\.toml|poetry\.lock|Pipfile(?:\.lock)?|setup\.py|setup\.cfg)$/.test(relative);
    default:
      return false;
  }
}

export function validateDependabotFiles({headRef, changedFiles, configurations}) {
  const branchEcosystem = DEPENDABOT_ECOSYSTEMS.get(headRef.split('/')[1]);
  if (!branchEcosystem) {
    return ['Dependabot head ref does not identify a supported configured ecosystem.'];
  }
  const candidates = configurations.filter(
    ({ecosystem}) => ecosystem === branchEcosystem,
  );
  if (candidates.length === 0) {
    return [`Dependabot ecosystem \`${branchEcosystem}\` is not configured.`];
  }
  if (changedFiles.length === 0) {
    return ['Dependabot policy requires the complete changed-file list.'];
  }
  const outside = changedFiles.filter(
    (path) => !candidates.some((configuration) => matchesEcosystemPath(path, configuration)),
  );
  return outside.length === 0
    ? []
    : [`Dependabot changed files outside its configured manifest/lock/workflow scope: ${outside.join(', ')}`];
}

export function validatePullRequest({
  title = '',
  body = '',
  labels = [],
  assignees = [],
  isDraft = false,
  author = '',
  headRef = '',
  changedFiles = [],
  dependabotConfigurations = [],
  canonicalLabels = new Set(),
  expectedProfile = '',
  taskReportExists = () => false,
}) {
  const errors = [];
  const dependabot = isDependabotPullRequest(author, headRef);
  const titleResult = validateContributionTitle(
    title,
    contributionTitleOptionsForPullRequest(author, headRef),
  );
  if (!titleResult.valid) errors.push(titleResult.message);

  const canonicalTypes = new Set(
    [...canonicalLabels].filter((label) => label.startsWith('type:')),
  );
  const canonicalAreas = new Set(
    [...canonicalLabels].filter((label) => label.startsWith('area:')),
  );
  if (!labels.some((label) => canonicalTypes.has(label))) {
    errors.push('PR requires a canonical type:* label from .github/labels.yml.');
  }
  if (!labels.some((label) => canonicalAreas.has(label))) {
    errors.push('PR requires a canonical area:* label from .github/labels.yml.');
  }
  if (
    !assignees.some(
      (assignee) =>
        typeof assignee === 'string' &&
        assignee.toLowerCase() === 'soku-jinseok',
    )
  ) {
    errors.push('PR must be assigned to Soku-JINSEOK.');
  }

  if (dependabot) {
    if (!labels.includes('type:chore')) {
      errors.push('Dependabot PR requires the `type:chore` label.');
    }
    if (!labels.includes('area:tooling')) {
      errors.push('Dependabot PR requires the `area:tooling` label.');
    }
    errors.push(
      ...validateDependabotFiles({
        headRef,
        changedFiles,
        configurations: dependabotConfigurations,
      }),
    );
    return errors;
  }

  const metadata = extractCommonMetadata(body);
  const issueMatch = /^- \*\*Issue:\*\* (Closes|Related to) #([1-9]\d*)\s*$/m.exec(
    metadata,
  );
  const relations = body.match(RELATION_PATTERN) ?? [];
  if (!issueMatch || relations.length !== 1) {
    errors.push(
      'Common Metadata Issue must be exactly one Closes #N or Related to #N relation.',
    );
  }
  if (isDraft && CLOSING_PATTERN.test(body)) {
    errors.push('Draft PRs must use Related to #N, not a closing relation.');
  }

  const taskReportMatch =
    /^- \*\*Task report:\*\* `([^`\s]+)`\s*$/m.exec(metadata);
  if (!taskReportMatch) {
    errors.push('Common Metadata must include a non-empty Task report line.');
  } else if (issueMatch) {
    const expectedTaskReport = `docs/issues/issue-${issueMatch[2]}-task-report.md`;
    if (taskReportMatch[1] !== expectedTaskReport) {
      errors.push(`Task report must match the linked Issue: ${expectedTaskReport}.`);
    } else if (!taskReportExists(expectedTaskReport)) {
      errors.push(`Task report does not exist: ${expectedTaskReport}.`);
    }
  }

  const profileMatch =
    /^- \*\*Governance profile:\*\* `([^`\r\n]+)`\s*$/m.exec(metadata);
  if (!profileMatch) {
    errors.push('Common Metadata must include a non-empty Governance profile line.');
  } else if (!expectedProfile || profileMatch[1] !== expectedProfile) {
    errors.push(`Governance profile must be ${expectedProfile || 'repository-defined'}.`);
  }

  if (PLACEHOLDER_PATTERN.test(body)) {
    errors.push('PR body contains an unfilled contribution placeholder.');
  }
  if (!followsRequiredSectionOrder(body)) {
    errors.push('PR body must follow the repository template heading order.');
  }
  if (!/AI Assistance[\s\S]*(?:Codex|Claude Code|Antigravity|None)/i.test(body)) {
    errors.push('PR body must record actual AI assistance or None.');
  }
  const verification = /### 🧪 Verification([\s\S]*?)(?=\n### |\n## |$)/.exec(
    body,
  );
  if (!verification || !/^- \[[xX]\] .+/m.test(verification[1])) {
    errors.push('English Verification must include a checked result.');
  }
  return errors;
}

export function runPullRequestPolicy({eventPath, repositoryRoot} = {}) {
  const resolvedEventPath =
    eventPath ?? process.env.CURRENT_PR_EVENT_PATH ?? process.env.GITHUB_EVENT_PATH;
  if (!resolvedEventPath) {
    console.error('CURRENT_PR_EVENT_PATH or GITHUB_EVENT_PATH is required.');
    return 1;
  }
  const root =
    repositoryRoot ?? fileURLToPath(new URL('../', import.meta.url));
  const event = JSON.parse(readFileSync(resolvedEventPath, 'utf8'));
  const pullRequest = event.pullRequest ?? event.pull_request;
  if (!pullRequest) {
    console.error('This validator must run on pull request events.');
    return 1;
  }
  const canonicalLabels = readCanonicalLabels(
    readFileSync(`${root}.github/labels.yml`, 'utf8'),
  );
  const expectedProfile = readExpectedProfile(
    readFileSync(`${root}.github/PULL_REQUEST_TEMPLATE.md`, 'utf8'),
  );
  const dependabotConfigurations = readDependabotConfigurations(
    readFileSync(`${root}.github/dependabot.yml`, 'utf8'),
  );
  const errors = validatePullRequest({
    title: pullRequest.title ?? '',
    body: pullRequest.body ?? '',
    labels: (pullRequest.labels ?? []).map((label) => label?.name),
    assignees: (pullRequest.assignees ?? []).map((assignee) => assignee?.login),
    isDraft: Boolean(pullRequest.draft),
    author: pullRequest.user?.login ?? '',
    headRef: pullRequest.head?.ref ?? '',
    changedFiles: pullRequest.changed_files_list ?? [],
    dependabotConfigurations,
    canonicalLabels,
    expectedProfile,
    taskReportExists: (path) => existsSync(`${root}${path}`),
  });
  if (errors.length) {
    console.error(errors.map((error) => `- ${error}`).join('\n'));
    return 1;
  }
  console.log(
    isDependabotPullRequest(
      pullRequest.user?.login ?? '',
      pullRequest.head?.ref ?? '',
    )
      ? 'Dependabot pull request policy passed.'
      : 'Pull request policy passed.',
  );
  return 0;
}

if (import.meta.url === `file://${process.argv[1]}`) {
  process.exitCode = runPullRequestPolicy();
}

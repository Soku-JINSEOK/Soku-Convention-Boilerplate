import {existsSync, readFileSync} from 'node:fs';
import {fileURLToPath} from 'node:url';

const eventPath = process.env.CURRENT_PR_EVENT_PATH ?? process.env.GITHUB_EVENT_PATH;
if (!eventPath) {
  console.error('CURRENT_PR_EVENT_PATH or GITHUB_EVENT_PATH is required.');
  process.exit(1);
}

const event = JSON.parse(readFileSync(eventPath, 'utf8'));
const pullRequest = event.pullRequest ?? event.pull_request;
if (!pullRequest) {
  console.error('This validator must run on pull request events.');
  process.exit(1);
}

function fail(message) {
  console.error(`✗ ${message}`);
  process.exit(1);
}

const repositoryRoot = fileURLToPath(new URL('../', import.meta.url));
const labelsSource = readFileSync(
  new URL('./labels.yml', import.meta.url),
  'utf8',
);
const catalogLabels = new Set(
  [...labelsSource.matchAll(/^\s*- name:\s*["']?([^"'\n]+)["']?\s*$/gm)].map(
    ([, name]) => name.trim(),
  ),
);
const canonicalTypes = new Set(
  [...catalogLabels].filter((label) => label.startsWith('type:')),
);
const canonicalAreas = new Set(
  [...catalogLabels].filter((label) => label.startsWith('area:')),
);
if (canonicalTypes.size === 0 || canonicalAreas.size === 0) {
  fail('Canonical label catalog must define at least one type and area label.');
}

const templateSource = readFileSync(
  new URL('./PULL_REQUEST_TEMPLATE.md', import.meta.url),
  'utf8',
);
const expectedProfile =
  /^- \*\*Governance profile:\*\* `([^`\r\n]+)`\s*$/m.exec(
    templateSource,
  )?.[1];
if (!expectedProfile) {
  fail('Repository PR template must define its governance profile.');
}

const body = pullRequest.body ?? '';
const labels = new Set(
  (pullRequest.labels ?? []).map((label) => label?.name).filter(Boolean),
);
if (![...labels].some((label) => canonicalTypes.has(label))) {
  fail('PR must include a canonical `type:*` label from `.github/labels.yml`.');
}
if (![...labels].some((label) => canonicalAreas.has(label))) {
  fail('PR must include a canonical `area:*` label from `.github/labels.yml`.');
}
if (
  !(pullRequest.assignees ?? []).some(
    (assignee) => assignee?.login?.toLowerCase() === 'soku-jinseok',
  )
) {
  fail('PR must be assigned to Soku-JINSEOK.');
}

const requiredHeadings = [
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
const headingIndexes = requiredHeadings.map((heading) => body.indexOf(heading));
if (headingIndexes.some((index) => index < 0)) {
  fail('PR body must include every required heading.');
}
for (let index = 1; index < headingIndexes.length; index += 1) {
  if (headingIndexes[index] <= headingIndexes[index - 1]) {
    fail('PR body headings are not in the required template order.');
  }
}

const placeholderPatterns = [
  /<Closes #N for final work \| Related to #N for partial work>/i,
  /<docs\/issues\/issue-<n>-task-report\.md>/i,
  /<profile name or None>/i,
  /<scope>/i,
  /<actual tool or None>/i,
  /<issue number>/i,
  /\b(?:TBD|TODO|No-Issue)\b/i,
];
for (const pattern of placeholderPatterns) {
  if (pattern.test(body)) fail(`Template placeholder still present: ${pattern}`);
}

const commonStart = body.indexOf('## 🔗 Common Metadata');
const englishStart = body.indexOf('## 🇬🇧 English — Normative Source');
const commonMetadata = body.slice(commonStart, englishStart);
const issueMatch =
  /^- \*\*Issue:\*\* (Closes|Related to) #([1-9]\d*)\s*$/m.exec(
    commonMetadata,
  );
const relations = body.match(/\b(?:Closes|Related to)\s+#[1-9]\d*\b/g) ?? [];
if (!issueMatch || relations.length !== 1) {
  fail('Common Metadata `Issue:` must be exactly one `Closes #N` or `Related to #N`.');
}
if (pullRequest.draft && issueMatch[1] === 'Closes') {
  fail('Draft PRs must use `Related to #N`, not a closing relation.');
}

const taskReportMatch =
  /^- \*\*Task report:\*\* `([^`\s]+)`\s*$/m.exec(commonMetadata);
if (!taskReportMatch) {
  fail('Common Metadata must include a non-empty `Task report:` line.');
}
const expectedTaskReport = `docs/issues/issue-${issueMatch[2]}-task-report.md`;
if (taskReportMatch[1] !== expectedTaskReport) {
  fail(`Task report must match the linked Issue: \`${expectedTaskReport}\`.`);
}
if (!existsSync(`${repositoryRoot}${expectedTaskReport}`)) {
  fail(`Task report does not exist: \`${expectedTaskReport}\`.`);
}

const profileMatch =
  /^- \*\*Governance profile:\*\* `([^`\r\n]+)`\s*$/m.exec(commonMetadata);
if (!profileMatch) {
  fail('Common Metadata must include a non-empty `Governance profile:` line.');
}
if (profileMatch[1] !== expectedProfile) {
  fail(`Governance profile must be \`${expectedProfile}\`.`);
}

const verification = /### 🧪 Verification([\s\S]*?)(?=\n### |\n## |$)/.exec(
  body,
);
if (!verification || !/^- \[[xX]\] .+/m.test(verification[1])) {
  fail('English Verification must include at least one checked result line.');
}

const aiStart = body.indexOf('## 🤖 AI Assistance');
const aiSection = body.slice(aiStart);
if (
  !/^- \*\*Planning\/implementation\/drafting:\*\* (?:OpenAI Codex|Claude Code|Antigravity|None)(?:\s|$)/m.test(
    aiSection,
  )
) {
  fail('AI Assistance must record the actual supported tool or None.');
}

console.log(`✓ PR issue reference: ${issueMatch[1]} #${issueMatch[2]}`);
console.log(`✓ PR governance profile: ${profileMatch[1]}`);
console.log('✓ PR governance fields satisfied.');

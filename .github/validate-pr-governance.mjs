import {readFileSync} from 'node:fs';

const eventPath = process.env.GITHUB_EVENT_PATH;
if (!eventPath) {
  console.error('GITHUB_EVENT_PATH is required.');
  process.exit(1);
}

const rawEvent = readFileSync(eventPath, 'utf8');
const event = JSON.parse(rawEvent);
const pullRequest = event.pullRequest ? event.pullRequest : event.pull_request;
if (!pullRequest) {
  console.error('This validator must run on pull request events.');
  process.exit(1);
}

function fail(message) {
  console.error(`✗ ${message}`);
  process.exit(1);
}

const body = pullRequest.body ?? '';
const labels = new Set(
  (pullRequest.labels || []).map((label) => label?.name).filter(Boolean),
);

const issueLinks = body.match(/\b(?:Closes|Related to)\s+#\d+\b/g) || [];
if (issueLinks.length === 0) {
  fail(
    'PR body must include at least one issue relation: `Closes #N` or `Related to #N`.',
  );
}

if (/\bNo-Issue\b/i.test(body)) {
  fail('Do not use `No-Issue`; every PR must reference an issue in metadata.');
}

if (!Array.from(labels).some((label) => label.startsWith('type:'))) {
  fail('PR must include at least one `type:*` label.');
}
if (!Array.from(labels).some((label) => label.startsWith('area:'))) {
  fail('PR must include at least one `area:*` label.');
}

const requiredHeadings = [
  '## 🔗 Common Metadata',
  '## 🇬🇧 English — Normative Source',
  '## 🇰🇷 한국어 요약',
  '## 🇯🇵 日本語の要約',
  '## Gitmoji Checklist',
  '## 🤖 AI Assistance',
];
const headingIndexes = requiredHeadings.map((heading) => body.indexOf(heading));
if (headingIndexes.some((index) => index < 0)) {
  fail('PR body must include all required headings in the project template order.');
}
for (let i = 1; i < headingIndexes.length; i += 1) {
  if (headingIndexes[i] <= headingIndexes[i - 1]) {
    fail('PR body headings are not in the required template order.');
  }
}

const placeholderPatterns = [
  /<Closes #N for final work \| Related to #N for partial work>/,
  /<docs\/issues\/issue-<n>-task-report\.md>/,
  /<profile name or None>/,
  /<scope>/,
  /<actual tool or None>/,
  /<issue number>/,
  /No-Issue/,
];

for (const pattern of placeholderPatterns) {
  if (pattern.test(body)) {
    fail(`Template placeholder still present: ${pattern}`);
  }
}

function extractSectionBody(headingText) {
  const start = body.indexOf(headingText);
  if (start < 0) return '';
  const nextHeadings = requiredHeadings
    .map((heading) => body.indexOf(heading, start + headingText.length))
    .filter((index) => index > start);
  const end = nextHeadings.length ? Math.min(...nextHeadings) : body.length;
  return body.slice(start, end);
}

const commonMetadata = extractSectionBody('## 🔗 Common Metadata');
if (!/- \*\*Issue:\*\*/.test(commonMetadata)) {
  fail('PR metadata section must include an `Issue:` line.');
}

const englishSection = extractSectionBody('## 🇬🇧 English — Normative Source');
const verificationSection = /### 🧪 Verification([\s\S]*?)(?=\n### |\n## |\n$)/.exec(englishSection);
if (!verificationSection || !/\- \[[xX]\]/.test(verificationSection[1])) {
  fail('English Verification section must include at least one checked result line.');
}

const aiSection = extractSectionBody('## 🤖 AI Assistance');
const aiLineMatch = /- \*\*Planning\/implementation\/drafting:\*\*\s*(.+)/.exec(
  aiSection,
);
if (!aiLineMatch) {
  fail('AI Assistance section must include Planning/implementation/drafting entry.');
}
if (/</.test(aiLineMatch[1])) {
  fail('AI Assistance entry still contains placeholder tokens.');
}

console.log(`✓ PR issue references: ${issueLinks.join(', ')}`);
console.log('✓ PR governance fields satisfied.');

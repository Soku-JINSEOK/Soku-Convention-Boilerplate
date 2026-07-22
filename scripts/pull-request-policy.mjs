#!/usr/bin/env node
import {existsSync, readFileSync} from "node:fs";
import {fileURLToPath} from "node:url";
import {validateContributionTitle} from "./contribution-title.mjs";

const CLOSING_PATTERN = /(?:closes|fixes|resolves)\s+(?:#[1-9]\d*|[\w.-]+\/[\w.-]+#[1-9]\d*|https:\/\/github\.com\/[\w.-]+\/[\w.-]+\/issues\/[1-9]\d*)/i;
const RELATION_PATTERN = /\b(?:Closes|Related to)\s+#[1-9]\d*\b/g;
const PLACEHOLDER_PATTERN = /<actual tool or none>|(?:closes|related to)\s+#\s*(?:\n|$)|issue-<n>|<!--\s*replace|\b(?:tbd|todo|no-issue)\b/iu;
const REQUIRED_BODY_SECTIONS = [
  "## 🔗 Common Metadata",
  "## 🇬🇧 English — Normative Source",
  "### 🎯 Goal",
  "### 📦 Scope",
  "### ✅ Acceptance Criteria",
  "### 🔒️ Security Boundary",
  "### 🧪 Verification",
  "### ⚠️ Risks and Follow-up",
  "## 🇰🇷 한국어 요약",
  "### 🎯 목표",
  "### 📦 핵심 범위",
  "### 🧪 검증",
  "### 🔒️ 비파괴 조건",
  "### ⚠️ 잔여 위험과 후속 작업",
  "## 🇯🇵 日本語の要約",
  "### 🎯 目標",
  "### 📦 主な範囲",
  "### 🧪 検証",
  "### 🔒️ 非破壊条件",
  "### ⚠️ 残存リスクと後続作業",
  "## Gitmoji Checklist",
  "## 🤖 AI Assistance",
];

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
  const start = body.indexOf("## 🔗 Common Metadata");
  const end = body.indexOf("## 🇬🇧 English — Normative Source");
  return start >= 0 && end > start ? body.slice(start, end) : "";
}

export function readCanonicalLabels(source) {
  return new Set(
    [...source.matchAll(/^\s*- name:\s*["']?([^"'\n]+)["']?\s*$/gm)].map(
      ([, name]) => name.trim(),
    ),
  );
}

export function readExpectedProfile(source) {
  return /^- \*\*Governance profile:\*\* `([^`\r\n]+)`\s*$/m.exec(source)?.[1] ?? "";
}

export function validatePullRequest({
  title = "",
  body = "",
  labels = [],
  assignees = [],
  isDraft = false,
  canonicalLabels = new Set(),
  expectedProfile = "",
  taskReportExists = () => false,
}) {
  const errors = [];
  const titleResult = validateContributionTitle(title);
  if (!titleResult.valid) errors.push(titleResult.message);

  const canonicalTypes = new Set(
    [...canonicalLabels].filter((label) => label.startsWith("type:")),
  );
  const canonicalAreas = new Set(
    [...canonicalLabels].filter((label) => label.startsWith("area:")),
  );
  if (!labels.some((label) => canonicalTypes.has(label))) {
    errors.push("PR requires a canonical type:* label from .github/labels.yml.");
  }
  if (!labels.some((label) => canonicalAreas.has(label))) {
    errors.push("PR requires a canonical area:* label from .github/labels.yml.");
  }
  if (
    !assignees.some(
      (assignee) =>
        typeof assignee === "string" &&
        assignee.toLowerCase() === "soku-jinseok",
    )
  ) {
    errors.push("PR must be assigned to Soku-JINSEOK.");
  }

  const metadata = extractCommonMetadata(body);
  const issueMatch = /^- \*\*Issue:\*\* (Closes|Related to) #([1-9]\d*)\s*$/m.exec(
    metadata,
  );
  const relations = body.match(RELATION_PATTERN) ?? [];
  if (!issueMatch || relations.length !== 1) {
    errors.push(
      "Common Metadata Issue must be exactly one Closes #N or Related to #N relation.",
    );
  }
  if (isDraft && CLOSING_PATTERN.test(body)) {
    errors.push("Draft PRs must use Related to #N, not a closing relation.");
  }

  const taskReportMatch =
    /^- \*\*Task report:\*\* `([^`\s]+)`\s*$/m.exec(metadata);
  if (!taskReportMatch) {
    errors.push("Common Metadata must include a non-empty Task report line.");
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
    errors.push("Common Metadata must include a non-empty Governance profile line.");
  } else if (!expectedProfile || profileMatch[1] !== expectedProfile) {
    errors.push(`Governance profile must be ${expectedProfile || "repository-defined"}.`);
  }

  if (PLACEHOLDER_PATTERN.test(body)) {
    errors.push("PR body contains an unfilled contribution placeholder.");
  }
  if (!followsRequiredSectionOrder(body)) {
    errors.push("PR body must follow the repository template heading order.");
  }
  if (!/AI Assistance[\s\S]*(?:Codex|Claude Code|Antigravity|None)/i.test(body)) {
    errors.push("PR body must record actual AI assistance or None.");
  }
  const verification = /### 🧪 Verification([\s\S]*?)(?=\n### |\n## |$)/.exec(
    body,
  );
  if (!verification || !/^- \[[xX]\] .+/m.test(verification[1])) {
    errors.push("English Verification must include a checked result.");
  }
  return errors;
}

const eventPath = process.env.CURRENT_PR_EVENT_PATH ?? process.env.GITHUB_EVENT_PATH;

if (import.meta.url === `file://${process.argv[1]}` && eventPath) {
  const event = JSON.parse(readFileSync(eventPath, "utf8"));
  const pullRequest = event.pull_request ?? {};
  const root = fileURLToPath(new URL("../", import.meta.url));
  const canonicalLabels = readCanonicalLabels(
    readFileSync(new URL("../.github/labels.yml", import.meta.url), "utf8"),
  );
  const expectedProfile = readExpectedProfile(
    readFileSync(
      new URL("../.github/PULL_REQUEST_TEMPLATE.md", import.meta.url),
      "utf8",
    ),
  );
  const errors = validatePullRequest({
    title: pullRequest.title ?? "",
    body: pullRequest.body ?? "",
    labels: (pullRequest.labels ?? []).map((label) => label?.name),
    assignees: (pullRequest.assignees ?? []).map((assignee) => assignee?.login),
    isDraft: Boolean(pullRequest.draft),
    canonicalLabels,
    expectedProfile,
    taskReportExists: (path) => existsSync(`${root}${path}`),
  });
  if (errors.length) {
    console.error(errors.map((error) => `- ${error}`).join("\n"));
    process.exitCode = 1;
  } else {
    console.log("Pull request policy passed.");
  }
}

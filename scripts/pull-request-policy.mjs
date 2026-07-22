#!/usr/bin/env node
import fs from "node:fs";
import {validateContributionTitle} from "./contribution-title.mjs";

const ISSUE_PATTERN = /(?:closes|fixes|resolves|related to|refs)\s+(?:#[1-9]\d*|[\w.-]+\/[\w.-]+#[1-9]\d*|https:\/\/github\.com\/[\w.-]+\/[\w.-]+\/issues\/[1-9]\d*)/i;
const CLOSING_PATTERN = /(?:closes|fixes|resolves)\s+(?:#[1-9]\d*|[\w.-]+\/[\w.-]+#[1-9]\d*|https:\/\/github\.com\/[\w.-]+\/[\w.-]+\/issues\/[1-9]\d*)/i;
const PLACEHOLDER_PATTERN = /<actual tool or none>|(?:closes|fixes|resolves|related to|refs)\s+#\s*(?:\n|$)|issue-<n>|<!--\s*replace/iu;
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

export function validatePullRequest({
  title = "",
  body = "",
  labels = [],
  files = [],
  assignees = [],
  isDraft = false,
}) {
  const errors = [];
  const titleResult = validateContributionTitle(title);
  if (!titleResult.valid) errors.push(titleResult.message);
  if (!ISSUE_PATTERN.test(body)) errors.push("PR body must link an Issue with Closes, Fixes, Resolves, Related to, or Refs.");
  if (isDraft && CLOSING_PATTERN.test(body)) {
    errors.push("Draft PRs must use only Related to or Refs, not closing keywords.");
  }
  if (!labels.some((label) => label.startsWith("type:"))) errors.push("PR requires a type:* label.");
  if (!labels.some((label) => label.startsWith("area:"))) errors.push("PR requires an area:* label.");
  if (!assignees.some((assignee) => typeof assignee === "string" && assignee.toLowerCase() === "soku-jinseok")) errors.push("PR must be assigned to Soku-JINSEOK.");
  if (PLACEHOLDER_PATTERN.test(body)) errors.push("PR body contains an unfilled contribution placeholder.");
  if (!followsRequiredSectionOrder(body)) errors.push("PR body must follow the PR #2 section order from the repository template.");
  if (!/AI Assistance[\s\S]*(?:Codex|Claude Code|Antigravity|None)/i.test(body)) errors.push("PR body must record actual AI assistance or None.");
  if (!/Verification[\s\S]*(?:\[x\]|Result:|passed|failed)/i.test(body)) errors.push("PR body must record an actual verification result.");
  const needsReport = labels.some((label) => ["size:l", "size:xl", "area:security"].includes(label.toLowerCase())) ||
    files.some((file) => /^(registry\/.*(?:schema|projects)|registry\/schema\/|pipelines\/|.*delivery.*)/.test(file));
  if (needsReport && !/docs\/issues\/issue-[1-9]\d*-task-report\.md/.test(body)) errors.push("This risk class requires an approved task report link.");
  return errors;
}

const eventPath = process.env.CURRENT_PR_EVENT_PATH ?? process.env.GITHUB_EVENT_PATH;

if (import.meta.url === `file://${process.argv[1]}` && eventPath) {
  const event = JSON.parse(fs.readFileSync(eventPath, "utf8"));
  const pullRequest = event.pull_request ?? {};
  const changedFiles = process.env.CHANGED_FILES_PATH ? fs.readFileSync(process.env.CHANGED_FILES_PATH, "utf8") : (process.env.CHANGED_FILES ?? "");
  const files = changedFiles.split("\n").filter(Boolean);
  const errors = validatePullRequest({
    title: pullRequest.title ?? "",
    body: pullRequest.body ?? "",
    labels: (pullRequest.labels ?? []).map((label) => label.name),
    files,
    assignees: (pullRequest.assignees ?? []).map((assignee) => assignee.login),
    isDraft: Boolean(pullRequest.draft),
  });
  if (errors.length) {
    console.error(errors.map((error) => `- ${error}`).join("\n"));
    process.exitCode = 1;
  } else {
    console.log("Pull request policy passed.");
  }
}

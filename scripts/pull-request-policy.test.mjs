import assert from "node:assert/strict";
import fs from "node:fs";
import test from "node:test";
import {
  readCanonicalLabels,
  readExpectedProfile,
  validatePullRequest,
} from "./pull-request-policy.mjs";

const validBody = `## 🔗 Common Metadata

- **Issue:** Related to #21
- **Task report:** \`docs/issues/issue-21-task-report.md\`
- **Governance profile:** \`control-plane\`

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

- [x] npm test — passed

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

- [x] ✨ Feature

## 🤖 AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex`;

const validTitle = "✨ feat(governance): enforce repository contract";
const canonicalLabels = new Set([
  "type:feature",
  "type:chore",
  "area:registry",
  "area:ci",
]);

function validPullRequest(overrides = {}) {
  return {
    title: validTitle,
    body: validBody,
    labels: ["type:feature", "area:registry", "custom:kept"],
    assignees: ["Soku-JINSEOK"],
    canonicalLabels,
    expectedProfile: "control-plane",
    taskReportExists: (path) => path === "docs/issues/issue-21-task-report.md",
    ...overrides,
  };
}

function errors(overrides = {}) {
  return validatePullRequest(validPullRequest(overrides));
}

test("accepts canonical and custom labels with complete metadata", () => {
  assert.deepEqual(errors(), []);
});

test("reads repository label and profile sources", () => {
  assert.deepEqual(
    [...readCanonicalLabels('- name: "type:feature"\n- name: area:ci\n')],
    ["type:feature", "area:ci"],
  );
  assert.equal(
    readExpectedProfile("- **Governance profile:** `boilerplate`\n"),
    "boilerplate",
  );
});

test("rejects labels that merely imitate canonical axes", () => {
  assert.match(errors({labels: ["type:invented", "area:registry"]}).join(" "), /canonical type/);
  assert.match(errors({labels: ["type:feature", "area:invented"]}).join(" "), /canonical area/);
});

test("rejects missing assignment", () => {
  assert.match(errors({assignees: []}).join(" "), /assigned to Soku-JINSEOK/);
});

test("rejects closing relationships for Draft pull requests", () => {
  assert.match(
    errors({body: validBody.replace("Related to #21", "Closes #21"), isDraft: true}).join(" "),
    /Draft PRs/,
  );
});

test("rejects missing, multiple, and unsupported Issue relationships", () => {
  assert.match(errors({body: validBody.replace("Related to #21", "Refs #21")}).join(" "), /exactly one/);
  assert.match(
    errors({body: validBody.replace("Related to #21", "Related to #21; Closes #22")}).join(" "),
    /exactly one/,
  );
});

test("binds the task report to the Issue and requires the file", () => {
  assert.match(
    errors({body: validBody.replace("issue-21-task-report", "issue-22-task-report")}).join(" "),
    /must match/,
  );
  assert.match(errors({taskReportExists: () => false}).join(" "), /does not exist/);
});

test("requires the repository governance profile", () => {
  assert.match(
    errors({body: validBody.replace("`control-plane`", "`boilerplate`")}).join(" "),
    /Governance profile must be control-plane/,
  );
});

test("rejects heading, placeholder, verification, and AI errors", () => {
  assert.match(errors({body: validBody.replace("## 🇰🇷", "## missing 🇰🇷")}).join(" "), /heading order/);
  assert.match(errors({body: validBody.replace("Goal\n", "TODO\n")}).join(" "), /placeholder/);
  assert.match(errors({body: validBody.replace("- [x] npm", "- [ ] npm")}).join(" "), /checked result/);
  assert.match(errors({body: validBody.replace("OpenAI Codex", "Undisclosed")}).join(" "), /AI assistance/i);
});

test("workflow reruns on metadata changes and reads the current PR", () => {
  const workflow = fs.readFileSync(
    new URL("../.github/workflows/pull-request-policy.yml", import.meta.url),
    "utf8",
  );
  assert.match(workflow, /assigned, unassigned/);
  assert.match(workflow, /ready_for_review, converted_to_draft/);
  assert.match(workflow, /gh api "repos\/\$\{GITHUB_REPOSITORY\}\/pulls\/\$\{PR_NUMBER\}"/);
  assert.match(workflow, /CURRENT_PR_EVENT_PATH:\s*\/tmp\/current-pr-event\.json/);
});

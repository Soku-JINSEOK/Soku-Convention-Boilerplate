import assert from "node:assert/strict";
import fs from "node:fs";
import test from "node:test";
import {validatePullRequest} from "./pull-request-policy.mjs";

const validBody = `## 🔗 Common Metadata
Closes #12
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
Codex`;
const validTitle = "✨ feat(dashboard): add portfolio metrics";

test("accepts a linked, labeled, verified pull request", () => {
  assert.deepEqual(
    validatePullRequest({
      title: validTitle,
      body: validBody,
      labels: ["type:feature", "area:dashboard"],
      files: ["apps/dashboard/a.ts"],
      assignees: ["Soku-JINSEOK"],
    }),
    [],
  );
});

test("accepts a non-closing issue relationship", () => {
  const body = validBody.replace("Closes #12", "Related to #12");
  assert.deepEqual(
    validatePullRequest({
      title: validTitle,
      body,
      labels: ["type:feature", "area:dashboard"],
      files: ["apps/dashboard/a.ts"],
      assignees: ["Soku-JINSEOK"],
    }),
    [],
  );
});

test("rejects closing keywords on draft pull requests", () => {
  assert.match(
    validatePullRequest({
      title: validTitle,
      body: validBody,
      labels: ["type:feature", "area:dashboard"],
      files: ["apps/dashboard/a.ts"],
      assignees: ["Soku-JINSEOK"],
      isDraft: true,
    }).join(" "),
    /Draft PRs must use only Related to or Refs/,
  );
});

test("accepts non-closing references on draft pull requests", () => {
  const body = validBody.replace("Closes #12", "Related to #12");
  assert.deepEqual(
    validatePullRequest({
      title: validTitle,
      body,
      labels: ["type:feature", "area:dashboard"],
      files: ["apps/dashboard/a.ts"],
      assignees: ["Soku-JINSEOK"],
      isDraft: true,
    }),
    [],
  );
});

test("accepts local and qualified Refs relationships", () => {
  for (const reference of ["Refs #12", "Refs Soku-JINSEOK/CutVi#101"]) {
    const body = validBody.replace("Closes #12", reference);
    assert.deepEqual(
      validatePullRequest({
        title: validTitle,
        body,
        labels: ["type:feature", "area:dashboard"],
        files: ["apps/dashboard/a.ts"],
        assignees: ["Soku-JINSEOK"],
      }),
      [],
    );
  }
});

test("rejects placeholders, missing labels, and a malformed issue reference", () => {
  const errors = validatePullRequest({
    title: "bad",
    body: "Related to #\n<actual tool or None>",
    labels: [],
    files: [],
  });
  assert.equal(errors.length, 9);
});

test("rejects a pull request without the required assignee", () => {
  assert.match(
    validatePullRequest({
      title: validTitle,
      body: validBody,
      labels: ["type:feature", "area:dashboard"],
      files: ["apps/dashboard/a.ts"],
    }).join(" "),
    /assigned to Soku-JINSEOK/,
  );
});

test("handles malformed assignee entries safely", () => {
  assert.equal(
    validatePullRequest({
      title: validTitle,
      body: validBody,
      labels: ["type:feature", "area:dashboard"],
      files: ["apps/dashboard/a.ts"],
      assignees: [null, 42],
    }).filter((message) => message.includes("assigned to Soku-JINSEOK"))[0],
    "PR must be assigned to Soku-JINSEOK.",
  );
});

test("reruns policy when assignment or draft state changes", () => {
  const workflow = fs.readFileSync(
    new URL("../.github/workflows/pull-request-policy.yml", import.meta.url),
    "utf8",
  );
  assert.match(workflow, /types:\s*\[[^\]]*assigned[^\]]*unassigned[^\]]*\]/);
  assert.match(
    workflow,
    /types:\s*\[[^\]]*ready_for_review[^\]]*converted_to_draft[^\]]*\]/,
  );
});

test("validates the current pull request metadata instead of a stale event", () => {
  const workflow = fs.readFileSync(
    new URL("../.github/workflows/pull-request-policy.yml", import.meta.url),
    "utf8",
  );
  assert.match(
    workflow,
    /gh api "repos\/\$\{GITHUB_REPOSITORY\}\/pulls\/\$\{PR_NUMBER\}"/,
  );
  assert.match(
    workflow,
    /CURRENT_PR_EVENT_PATH:\s*\/tmp\/current-pr-event\.json/,
  );
});

test("rejects missing or reordered PR #2 sections", () => {
  const body = validBody.replace(
    "## 🇰🇷 한국어 요약",
    "## 🇯🇵 日本語の要約",
  );
  assert.match(
    validatePullRequest({
      title: validTitle,
      body,
      labels: ["type:feature", "area:dashboard"],
      files: [],
      assignees: ["Soku-JINSEOK"],
    }).join(" "),
    /PR #2 section order/,
  );
});

test("requires a task report for schema, delivery, security, and large work", () => {
  assert.match(
    validatePullRequest({
      title: validTitle,
      body: validBody,
      labels: ["type:chore", "area:registry"],
      files: ["registry/schema/x.json"],
      assignees: ["Soku-JINSEOK"],
    }).join(" "),
    /task report/,
  );
  assert.deepEqual(
    validatePullRequest({
      title: validTitle,
      body: `${validBody}\ndocs/issues/issue-12-task-report.md`,
      labels: ["type:chore", "area:registry"],
      files: ["registry/schema/x.json"],
      assignees: ["Soku-JINSEOK"],
    }),
    [],
  );
});

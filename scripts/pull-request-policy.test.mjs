import assert from "node:assert/strict";
import fs from "node:fs";
import test from "node:test";
import {
  readCanonicalLabels,
  readDependabotConfigurations,
  readExpectedProfile,
  validateDependabotFiles,
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
  "area:tooling",
]);
const dependabotSource = `version: 2
updates:
  - package-ecosystem: github-actions
    directory: /
  - package-ecosystem: gomod
    directory: /soku
  - package-ecosystem: npm
    directory: /templates/javascript-typescript-node
  - package-ecosystem: pip
    directory: /templates/python
`;
const dependabotConfigurations = readDependabotConfigurations(dependabotSource);

function validPullRequest(overrides = {}) {
  return {
    title: validTitle,
    body: validBody,
    labels: ["type:feature", "area:registry", "custom:kept"],
    assignees: ["Soku-JINSEOK"],
    canonicalLabels,
    expectedProfile: "control-plane",
    taskReportExists: (path) => path === "docs/issues/issue-21-task-report.md",
    dependabotConfigurations,
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

test("reads configured Dependabot ecosystems and directories", () => {
  assert.deepEqual(dependabotConfigurations, [
    {ecosystem: "github-actions", directory: "/"},
    {ecosystem: "gomod", directory: "/soku"},
    {ecosystem: "npm", directory: "/templates/javascript-typescript-node"},
    {ecosystem: "pip", directory: "/templates/python"},
  ]);
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

function dependabotPullRequest(overrides = {}) {
  return validPullRequest({
    title: "build(deps): bump package group to the latest minor versions with extended release notes",
    body: "Dependabot-generated release notes without the human template.",
    labels: ["type:chore", "area:tooling"],
    assignees: ["Soku-JINSEOK"],
    author: "dependabot[bot]",
    headRef: "dependabot/npm_and_yarn/templates/javascript-typescript-node/group",
    changedFiles: [
      "templates/javascript-typescript-node/package.json",
      "templates/javascript-typescript-node/package-lock.json",
    ],
    ...overrides,
  });
}

test("accepts configured Dependabot npm, pip, gomod, and Actions paths", () => {
  const scenarios = [
    [
      "dependabot/npm_and_yarn/templates/javascript-typescript-node/group",
      ["templates/javascript-typescript-node/package-lock.json"],
    ],
    [
      "dependabot/pip/templates/python/group",
      ["templates/python/requirements-lock.txt", "templates/python/pyproject.toml"],
    ],
    ["dependabot/go_modules/soku/group", ["soku/go.mod", "soku/go.sum"]],
    [
      "dependabot/github_actions/actions/group",
      [".github/workflows/validation.yml"],
    ],
  ];
  for (const [headRef, changedFiles] of scenarios) {
    assert.deepEqual(
      validatePullRequest(dependabotPullRequest({headRef, changedFiles})),
      [],
    );
  }
});

test("rejects Dependabot impersonation and incorrect head refs as human PRs", () => {
  for (const overrides of [
    {author: "dependabot"},
    {author: "dependabot-user"},
    {headRef: "automation/npm_and_yarn/group"},
  ]) {
    const result = validatePullRequest(dependabotPullRequest(overrides));
    assert.match(result.join(" "), /Gitmoji|Common Metadata/);
  }
});

test("rejects Dependabot files outside its configured ecosystem scope", () => {
  const result = validatePullRequest(
    dependabotPullRequest({changedFiles: ["scripts/pull-request-policy.mjs"]}),
  );
  assert.match(result.join(" "), /outside its configured/);
  assert.deepEqual(
    validateDependabotFiles({
      headRef: "dependabot/unsupported/example",
      changedFiles: ["file.lock"],
      configurations: dependabotConfigurations,
    }),
    ["Dependabot head ref does not identify a supported configured ecosystem."],
  );
});

test("keeps Dependabot labels and assignment mandatory", () => {
  assert.match(
    validatePullRequest(
      dependabotPullRequest({labels: ["type:chore", "area:ci"]}),
    ).join(" "),
    /area:tooling/,
  );
  assert.match(
    validatePullRequest(dependabotPullRequest({assignees: []})).join(" "),
    /assigned to Soku-JINSEOK/,
  );
});

test("workflow reruns on metadata changes and reads the current PR", () => {
  const workflow = fs.readFileSync(
    new URL("../.github/workflows/pull-request-policy.yml", import.meta.url),
    "utf8",
  );
  assert.match(workflow, /workflow_call/);
  assert.doesNotMatch(workflow, /^\s{2}pull_request:/m);
  assert.match(workflow, /gh api "repos\/\$\{GITHUB_REPOSITORY\}\/pulls\/\$\{PR_NUMBER\}"/);
  assert.match(workflow, /files\?per_page=100/);
  assert.match(workflow, /CURRENT_PR_EVENT_PATH:\s*\/tmp\/current-pr-event\.json/);
});

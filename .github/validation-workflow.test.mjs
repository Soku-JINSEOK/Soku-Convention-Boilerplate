import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import test from 'node:test';

const workflow = readFileSync(
  new URL('./workflows/validation.yml', import.meta.url),
  'utf8',
);
const releaseWorkflow = readFileSync(
  new URL('./workflows/release.yml', import.meta.url),
  'utf8',
);

test('separates full validation from current PR metadata validation', () => {
  assert.match(workflow, /'Validation Gate' \|\| 'Full Validation Not Required'/);
  assert.match(workflow, /name: PR Metadata Gate/);
  assert.match(workflow, /CURRENT_PR_EVENT_PATH:\s*\/tmp\/current-pr-event\.json/);
  assert.match(
    workflow,
    /gh api "repos\/\$\{GITHUB_REPOSITORY\}\/pulls\/\$\{PR_NUMBER\}"/,
  );
});

test('runs full validation only for code-bearing pull request events', () => {
  for (const action of ['opened', 'synchronize', 'reopened']) {
    assert.match(workflow, new RegExp(`github\\.event\\.action == '${action}'`));
  }
  assert.match(workflow, /github\.event\.changes\.base != null/);
  assert.match(workflow, /'Validation Gate' \|\| 'Full Validation Not Required'/);
});

test('keeps full and metadata cancellation domains independent', () => {
  assert.match(workflow, /group: validation-full-repository-/);
  assert.match(workflow, /group: validation-full-templates-/);
  assert.match(workflow, /group: validation-full-security-/);
  assert.match(workflow, /group: validation-metadata-titles-/);
  assert.match(workflow, /group: validation-metadata-governance-/);
  assert.doesNotMatch(workflow, /^concurrency:/m);
});

test('does not subscribe to closed pull request events', () => {
  const trigger = /pull_request:\n\s+types: \[([^\]]+)\]/.exec(workflow);
  assert.ok(trigger, 'pull_request event list must be explicit');
  assert.doesNotMatch(trigger[1], /closed/);
  for (const action of [
    'edited',
    'labeled',
    'unlabeled',
    'assigned',
    'unassigned',
    'ready_for_review',
    'converted_to_draft',
  ]) {
    assert.match(trigger[1], new RegExp(`\\b${action}\\b`));
  }
});

test('release preflight can call validation without enabling delivery', () => {
  assert.match(releaseWorkflow, /boilerplate-tag:[\s\S]*default: v1\.0\.2/);
  assert.match(releaseWorkflow, /cli-tag:[\s\S]*default: soku\/v0\.1\.3/);
  assert.match(
    releaseWorkflow,
    /permissions:\n\s+contents: read\n\s+pull-requests: read/,
  );
  assert.match(
    releaseWorkflow,
    /github\.event_name == 'push' &&\n\s+github\.repository == 'Soku-JINSEOK\/Soku-Convention-Boilerplate'/,
  );
  assert.doesNotMatch(
    releaseWorkflow,
    /github\.event_name == 'workflow_dispatch'[^\n]*deliver/,
  );
});

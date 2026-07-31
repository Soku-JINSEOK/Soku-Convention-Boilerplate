import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import test from 'node:test';

const workflow = readFileSync(
  new URL('./workflows/validation.yml', import.meta.url),
  'utf8',
);
const repositoryWorkflow = readFileSync(
  new URL('./workflows/ci.yml', import.meta.url),
  'utf8',
);
const quickWorkflow = readFileSync(
  new URL('./workflows/ci-quick.yml', import.meta.url),
  'utf8',
);
const hostedFullWorkflow = readFileSync(
  new URL('./workflows/full-validation.yml', import.meta.url),
  'utf8',
);
const releaseWorkflow = readFileSync(
  new URL('./workflows/release.yml', import.meta.url),
  'utf8',
);
const releaseIdentity = JSON.parse(
  readFileSync(new URL('../release-identity.json', import.meta.url), 'utf8'),
);
const contributionWorkflow = readFileSync(
  new URL('./workflows/contribution-title-check.yml', import.meta.url),
  'utf8',
);
const policyWorkflow = readFileSync(
  new URL('./workflows/pull-request-policy.yml', import.meta.url),
  'utf8',
);

test('separates full validation from current PR metadata validation', () => {
  assert.match(workflow, /ci-quick-gate:\n\s+name: CI Quick Gate/);
  assert.match(workflow, /uses: \.\/\.github\/workflows\/ci-quick\.yml/);
  assert.match(workflow, /validation-gate:[\s\S]*name: Validation Gate/);
  assert.match(workflow, /name: Full Validation Not Required/);
  assert.match(workflow, /name: PR Metadata Gate/);
  assert.match(workflow, /uses: \.\/\.github\/workflows\/contribution-title-check\.yml/);
  assert.match(workflow, /uses: \.\/\.github\/workflows\/pull-request-policy\.yml/);
});

test('runs CI Quick in parallel without replacing the full gate', () => {
  assert.match(quickWorkflow, /^\s{2}workflow_call:/m);
  assert.match(quickWorkflow, /plan:\n\s+name: Plan changed-scope shards/);
  assert.match(quickWorkflow, /matrix: \$\{\{ fromJSON\(needs\.plan\.outputs\.matrix\) \}\}/);
  assert.match(quickWorkflow, /node scripts\/plan-ci-quick\.mjs/);
  assert.match(quickWorkflow, /--group '\$\{\{ matrix\.id \}\}'/);
  assert.match(quickWorkflow, /if: contains\(matrix\.toolchains, 'node'\)/);
  assert.match(quickWorkflow, /if: contains\(matrix\.toolchains, 'go'\)/);
  assert.match(quickWorkflow, /if: contains\(matrix\.toolchains, 'python'\)/);
  assert.match(quickWorkflow, /if: contains\(matrix\.toolchains, 'java'\)/);
  assert.match(quickWorkflow, /fetch-depth: 0/);
  assert.match(
    quickWorkflow,
    /\.\/scripts\/verify\.sh --profile ci-quick/,
  );
  assert.match(workflow, /name: CI Quick Gate/);
  assert.match(workflow, /name: Validation Gate/);
  assert.match(workflow, /group: validation-quick-/);
  assert.match(workflow, /name: CI Quick Not Required/);
  assert.match(workflow, /NOT_REQUIRED_RESULT:/);
  assert.match(workflow, /QUICK_RESULT" = success/);
  assert.match(workflow, /QUICK_RESULT" = skipped/);
  assert.match(workflow, /exit 1/);
  assert.match(workflow, /head-sha: \$\{\{ github\.event\.pull_request\.head\.sha \|\| github\.sha \}\}/);
});

test('provides a valid Quick base for manual reusable validation', () => {
  assert.match(
    workflow,
    /base-sha: \$\{\{ github\.event\.pull_request\.base\.sha \|\| github\.event\.before \|\| github\.sha \}\}/,
  );
});

test('runs full validation only for code-bearing pull request events', () => {
  for (const action of ['opened', 'synchronize', 'reopened']) {
    assert.match(workflow, new RegExp(`github\\.event\\.action == '${action}'`));
  }
  assert.match(workflow, /github\.event\.changes\.base != null/);
  assert.match(workflow, /REPOSITORY_RESULT:/);
});

test('metadata-only events preserve the required Validation Gate context', () => {
  assert.match(workflow, /validation-gate:\n\s+name: Validation Gate/);
  assert.match(workflow, /full-validation-not-required:/);
  assert.match(workflow, /Metadata-only event preserves the existing Validation Gate/);
  assert.match(workflow, /validation-metadata-not-required-/);
});

test('keeps full and metadata cancellation domains independent', () => {
  assert.match(workflow, /group: validation-full-repository-/);
  assert.match(workflow, /group: validation-full-templates-/);
  assert.match(workflow, /group: validation-full-security-/);
  assert.match(workflow, /group: validation-metadata-titles-/);
  assert.match(workflow, /group: validation-metadata-governance-/);
  assert.match(workflow, /group: validation-full-gate-/);
  assert.doesNotMatch(workflow, /^concurrency:/m);
});

test('only Validation directly subscribes to pull request and main push events', () => {
  assert.match(workflow, /^\s{2}pull_request:/m);
  assert.match(workflow, /^\s{2}push:/m);
  for (const component of [contributionWorkflow, policyWorkflow]) {
    assert.match(component, /^\s{2}workflow_call:/m);
    assert.doesNotMatch(component, /^\s{2}pull_request:/m);
    assert.doesNotMatch(component, /^\s{2}push:/m);
  }
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

test('isolates token-backed provider conformance from fork pull requests', () => {
  assert.match(
    repositoryWorkflow,
    /name: Run hermetic lifecycle conformance gate\n\s+shell: bash\n\s+run:/,
  );
  assert.match(
    repositoryWorkflow,
    /name: Fetch pinned external provider conformance fixture\n\s+if: >-\n\s+github\.event_name != 'pull_request' \|\|\n\s+github\.event\.pull_request\.head\.repo\.full_name == github\.repository\n\s+env:\n\s+GITHUB_TOKEN:/,
  );
  assert.doesNotMatch(
    `${workflow}\n${repositoryWorkflow}`,
    /^\s{2}pull_request_target:/m,
  );

  const repository = 'Soku-JINSEOK/Soku-Convention-Boilerplate';
  const shouldRunNetworkConformance = ({
    eventName,
    headRepository = '',
  }) => eventName !== 'pull_request' || headRepository === repository;
  const cases = [
    {
      name: 'fork pull request',
      eventName: 'pull_request',
      headRepository: 'external-contributor/Soku-Convention-Boilerplate',
      expected: false,
    },
    {
      name: 'same-repository pull request',
      eventName: 'pull_request',
      headRepository: repository,
      expected: true,
    },
    {name: 'main push', eventName: 'push', expected: true},
    {name: 'scheduled run', eventName: 'schedule', expected: true},
    {name: 'manual run', eventName: 'workflow_dispatch', expected: true},
  ];

  for (const context of cases) {
    assert.equal(
      shouldRunNetworkConformance(context),
      context.expected,
      context.name,
    );
  }
});

test('release preflight can call validation without enabling delivery', () => {
  assert.match(
    releaseWorkflow,
    new RegExp(
      `boilerplate-tag:[\\s\\S]*default: ${releaseIdentity.boilerplate.tag.replaceAll('.', '\\.')}`,
    ),
  );
  assert.match(
    releaseWorkflow,
    new RegExp(
      `cli-tag:[\\s\\S]*default: ${releaseIdentity.cli.tag.replaceAll('.', '\\.')}`,
    ),
  );
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

test('Hosted Full supports exact-source reusable, manual, and daily runs', () => {
  assert.match(hostedFullWorkflow, /^\s{2}workflow_call:/m);
  assert.match(hostedFullWorkflow, /^\s{2}workflow_dispatch:/m);
  assert.match(hostedFullWorkflow, /^\s{2}schedule:/m);
  assert.match(hostedFullWorkflow, /cron: '41 2 \* \* \*'/);
  assert.match(hostedFullWorkflow, /source-sha:[\s\S]*required: true/);
  assert.match(hostedFullWorkflow, /name: Hosted Full Gate/);
  for (const dependency of ['repository', 'templates', 'security']) {
    assert.match(
      hostedFullWorkflow,
      new RegExp(
        `${dependency}:[\\s\\S]*source-sha: \\$\\{\\{ inputs\\.source-sha \\|\\| github\\.sha \\}\\}`,
      ),
    );
  }
  for (const result of [
    'REPOSITORY_RESULT',
    'TEMPLATES_RESULT',
    'SECURITY_RESULT',
  ]) {
    assert.match(hostedFullWorkflow, new RegExp(`test "\\$${result}" = success`));
  }
});

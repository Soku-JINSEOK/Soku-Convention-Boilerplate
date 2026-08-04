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
const releaseWorkflow = readFileSync(
  new URL('./workflows/release.yml', import.meta.url),
  'utf8',
);
const deployWorkflow = readFileSync(
  new URL('./workflows/deploy-gcp.yml', import.meta.url),
  'utf8',
);
const releaseIdentity = JSON.parse(
  readFileSync(new URL('../release-identity.json', import.meta.url), 'utf8'),
);
const securityWorkflow = readFileSync(
  new URL('./workflows/security.yml', import.meta.url),
  'utf8',
);
const policyWorkflow = readFileSync(
  new URL('./workflows/pull-request-policy.yml', import.meta.url),
  'utf8',
);
const projectSyncWorkflow = readFileSync(
  new URL('./workflows/project-sync.yml', import.meta.url),
  'utf8',
);

test('keeps general validation manual and reusable', () => {
  assert.match(workflow, /^\s{2}workflow_call:/m);
  assert.match(workflow, /^\s{2}workflow_dispatch:/m);
  assert.doesNotMatch(workflow, /^\s{2}pull_request:/m);
  assert.doesNotMatch(workflow, /^\s{2}push:/m);
  assert.match(workflow, /ci-quick-gate:\n\s+name: CI Quick Gate/);
  assert.match(workflow, /uses: \.\/\.github\/workflows\/ci-quick\.yml/);
  assert.match(workflow, /validation-gate:[\s\S]*name: Validation Gate/);
  assert.doesNotMatch(workflow, /PR Metadata Gate/);
});

test('runs CI Quick in parallel without replacing the full gate', () => {
  assert.match(quickWorkflow, /^\s{2}workflow_call:/m);
  assert.match(quickWorkflow, /plan:\n\s+name: Plan changed-scope shards/);
  assert.match(quickWorkflow, /matrix: \$\{\{ fromJSON\(needs\.plan\.outputs\.matrix\) \}\}/);
  assert.match(quickWorkflow, /node scripts\/plan-ci-quick\.mjs/);
  assert.match(quickWorkflow, /--group '\$\{\{ matrix\.id \}\}'/);
  assert.match(quickWorkflow, /if: contains\(matrix\.toolchains, 'node'\)/);
  assert.match(quickWorkflow, /if: contains\(matrix\.toolchains, 'go'\)/);
  assert.match(
    quickWorkflow,
    /cache-dependency-path: \|\n\s+soku\/go\.sum\n\s+templates\/go\/go\.mod/,
  );
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

test('retains the full validation result contract', () => {
  assert.match(workflow, /REPOSITORY_RESULT:/);
});

test('metadata-only events preserve the required Validation Gate context', () => {
  assert.match(workflow, /validation-gate:\n\s+name: Validation Gate/);
  assert.match(workflow, /full-validation-not-required:/);
  assert.match(workflow, /Metadata-only event preserves the existing Validation Gate/);
  assert.match(workflow, /validation-metadata-not-required-/);
});

test('keeps full validation cancellation domains independent', () => {
  assert.match(workflow, /group: validation-full-repository-/);
  assert.match(workflow, /group: validation-full-templates-/);
  assert.match(workflow, /group: validation-full-gate-/);
  assert.doesNotMatch(workflow, /^concurrency:/m);
});

test('keeps event-driven policy, Security, and Project synchronization explicit', () => {
  assert.match(policyWorkflow, /^\s{2}pull_request:/m);
  assert.doesNotMatch(policyWorkflow, /^\s{2}push:/m);
  assert.match(policyWorkflow, /policy:\n\s+name: PR Metadata Gate/);
  assert.match(policyWorkflow, /validation-gate:\n\s+name: Validation Gate/);
  assert.match(policyWorkflow, /POLICY_RESULT: \$\{\{ needs\.policy\.result \}\}/);
  assert.match(securityWorkflow, /^\s{2}pull_request:/m);
  assert.match(securityWorkflow, /^\s{2}push:/m);
  assert.match(securityWorkflow, /^\s{2}schedule:/m);
  assert.match(projectSyncWorkflow, /^\s{2}issues:/m);
  assert.match(projectSyncWorkflow, /^\s{2}pull_request:/m);
  assert.match(projectSyncWorkflow, /^\s{2}schedule:/m);
  assert.match(projectSyncWorkflow, /^\s{2}workflow_dispatch:/m);
  assert.match(projectSyncWorkflow, /GH_TOKEN: \$\{\{ secrets\.PROJECT_SYNC_TOKEN \}\}/);
  assert.match(projectSyncWorkflow, /contents: read/);
  assert.match(projectSyncWorkflow, /! -f scripts\/github-project-sync\.mjs/);
  assert.match(projectSyncWorkflow, /pre-merge event skipped/);
  assert.doesNotMatch(projectSyncWorkflow, /contents: write/);
  assert.doesNotMatch(projectSyncWorkflow, /pull_request_target/);

  const securityPullRequestTrigger =
    /pull_request:\n\s+types:\n((?:\s+- .+\n)+)/.exec(securityWorkflow);
  assert.ok(securityPullRequestTrigger, 'Security PR event list must be explicit');
  assert.match(securityPullRequestTrigger[1], /\bready_for_review\b/);
  assert.doesNotMatch(securityPullRequestTrigger[1], /\bclosed\b/);
});

test('executes policy and history controls from the trusted base checkout', () => {
  assert.match(policyWorkflow, /name: Checkout trusted policy source/);
  assert.match(policyWorkflow, /path: trusted-policy/);
  assert.match(policyWorkflow, /ref: \$\{\{ github\.event\.pull_request\.base\.sha \}\}/);
  assert.match(policyWorkflow, /path: pr-head/);
  assert.match(
    policyWorkflow,
    /import \{runPullRequestPolicy\} from '\.\/trusted-policy\/scripts\/pull-request-policy\.mjs'/,
  );
  assert.match(policyWorkflow, /event head SHA/);
  assert.match(policyWorkflow, /API head SHA/);
  assert.match(policyWorkflow, /repositoryRoot: `\$\{process\.env\.PR_HEAD_WORKSPACE\}\//);
  assert.match(policyWorkflow, /PR_HEAD_WORKSPACE:/);
  assert.doesNotMatch(policyWorkflow, /pr-head\/scripts\//);

  assert.match(securityWorkflow, /name: Checkout trusted security policy/);
  assert.match(securityWorkflow, /path: trusted-security/);
  assert.match(securityWorkflow, /path: repository/);
  assert.match(
    securityWorkflow,
    /HISTORICAL_BASELINE_COMMIT: 2be9df2421e5661ea5d978fa7832a0ae32936e9d/,
  );
  assert.match(securityWorkflow, /historical \.gitleaks\.toml raw-byte hash mismatch/);
  assert.match(securityWorkflow, /--config \/trusted\/\.gitleaks\.toml/);
  assert.match(securityWorkflow, /--log-opts HEAD/);
  assert.doesNotMatch(securityWorkflow, /repository\/scripts\//);
});

test('preserves release tag and manual deployment trigger exceptions', () => {
  assert.match(releaseWorkflow, /^\s{2}push:\n\s+tags:/m);
  assert.match(releaseWorkflow, /^\s{2}workflow_dispatch:/m);
  assert.match(deployWorkflow, /^\s{2}workflow_dispatch:/m);
  assert.doesNotMatch(deployWorkflow, /^\s{2}(?:pull_request|push|schedule):/m);
});

test('does not subscribe to closed pull request events', () => {
  const trigger = /pull_request:\n\s+types:\n((?:\s+- .+\n)+)/.exec(policyWorkflow);
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

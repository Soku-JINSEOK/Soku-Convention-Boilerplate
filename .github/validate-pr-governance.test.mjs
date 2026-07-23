import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import test from 'node:test';

import {
  readDependabotConfigurations,
  validatePullRequest,
} from '../scripts/pull-request-policy.mjs';

test('GitHub validator is a thin adapter over the shared policy module', () => {
  const wrapper = readFileSync(
    new URL('./validate-pr-governance.mjs', import.meta.url),
    'utf8',
  );
  assert.match(wrapper, /runPullRequestPolicy/);
  assert.match(wrapper, /\.\.\/scripts\/pull-request-policy\.mjs/);
  assert.doesNotMatch(wrapper, /requiredHeadings|canonicalTypes|placeholderPatterns/);
});

test('shared adapter policy accepts only a file-scoped Dependabot PR', () => {
  const configurations = readDependabotConfigurations(`updates:
  - package-ecosystem: gomod
    directory: /soku
`);
  const base = {
    title: 'build(deps): bump Go module dependencies with detailed release notes',
    body: 'Dependabot release notes',
    labels: ['type:chore', 'area:tooling'],
    assignees: ['Soku-JINSEOK'],
    author: 'dependabot[bot]',
    headRef: 'dependabot/go_modules/soku/group',
    canonicalLabels: new Set(['type:chore', 'area:tooling']),
    dependabotConfigurations: configurations,
  };
  assert.deepEqual(
    validatePullRequest({...base, changedFiles: ['soku/go.mod', 'soku/go.sum']}),
    [],
  );
  assert.match(
    validatePullRequest({...base, changedFiles: ['README.md']}).join(' '),
    /outside its configured/,
  );
});

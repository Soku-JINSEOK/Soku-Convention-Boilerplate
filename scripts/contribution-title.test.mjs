import assert from 'node:assert/strict';
import test from 'node:test';

import * as repositoryModule from './contribution-title.mjs';
import * as templateModule from '../templates/_shared/commitlint/contribution-title.mjs';

const cases = [
  ['✨ feat(auth): add login flow', {}],
  ['♻️ refactor(api): simplify routing', {}],
  ['fix(deps): bump dependency', {allowConventionalWithoutGitmoji: true}],
  ['add login flow', {}],
  ['✨ feat(auth): 로그인 흐름 추가', {}],
  ['✨ feat(auth): add login flow.', {}],
];

test('repository and downstream contribution-title modules stay behaviorally identical', () => {
  assert.deepEqual(repositoryModule.TITLE_CONVENTIONS, templateModule.TITLE_CONVENTIONS);
  for (const [title, options] of cases) {
    assert.deepEqual(
      repositoryModule.validateContributionTitle(title, options),
      templateModule.validateContributionTitle(title, options),
    );
  }
});

test('Dependabot title options require both the exact author and head prefix', () => {
  assert.deepEqual(
    repositoryModule.contributionTitleOptionsForPullRequest(
      'dependabot[bot]',
      'dependabot/npm_and_yarn/example',
    ),
    {allowConventionalWithoutGitmoji: true, maxLength: null},
  );
  for (const [author, headRef] of [
    ['dependabot', 'dependabot/npm_and_yarn/example'],
    ['dependabot[bot]', 'automation/npm_and_yarn/example'],
  ]) {
    assert.deepEqual(
      repositoryModule.contributionTitleOptionsForPullRequest(author, headRef),
      {},
    );
  }
});

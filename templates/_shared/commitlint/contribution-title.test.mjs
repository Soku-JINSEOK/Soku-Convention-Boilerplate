import test from 'node:test';
import assert from 'node:assert/strict';

import {
  contributionTitleOptionsForAuthor,
  validateContributionTitle,
} from './contribution-title.mjs';

test('accepts a standard feat title', () => {
  const result = validateContributionTitle('✨ feat(auth): add login flow');
  assert.equal(result.valid, true);
});

test('accepts a standard fix title', () => {
  const result = validateContributionTitle('🐛 fix(api): handle empty response body');
  assert.equal(result.valid, true);
});

test('accepts bot-style conventional title when option enabled', () => {
  const result = validateContributionTitle('fix(deps): bump dependency', {
    allowConventionalWithoutGitmoji: true,
  });
  assert.equal(result.valid, true);
});

test('rejects bot-style conventional title when option is disabled', () => {
  const result = validateContributionTitle('fix(deps): bump dependency');
  assert.equal(result.valid, false);
});

test('accepts a breaking change title', () => {
  const result = validateContributionTitle('💥 feat!(api): remove legacy endpoint');
  assert.equal(result.valid, true);
});

test('rejects a title without a supported gitmoji/type prefix', () => {
  const result = validateContributionTitle('add login flow');
  assert.equal(result.valid, false);
});

test('rejects a title without a scope in parentheses', () => {
  const result = validateContributionTitle('✨ feat: add login flow');
  assert.equal(result.valid, false);
});

test('rejects a title with a non-kebab-case scope', () => {
  const result = validateContributionTitle('✨ feat(Auth Module): add login flow');
  assert.equal(result.valid, false);
});

test('rejects a title with an empty subject', () => {
  const result = validateContributionTitle('✨ feat(auth): ');
  assert.equal(result.valid, false);
});

test('rejects a title with a non-ASCII subject', () => {
  const result = validateContributionTitle('✨ feat(auth): 로그인 흐름 추가');
  assert.equal(result.valid, false);
});

test('rejects a title whose subject ends with a period', () => {
  const result = validateContributionTitle('✨ feat(auth): add login flow.');
  assert.equal(result.valid, false);
});

test('rejects a title longer than 72 characters', () => {
  const longSubject = 'a very long subject that keeps going past the character budget';
  const result = validateContributionTitle(`✨ feat(auth): ${longSubject}`);
  assert.equal(result.valid, false);
  assert.match(result.message, /72 characters/);
});

test('defaults to 72 when maxLength is omitted', () => {
  const longTitle = '✨ feat(auth): a very long subject that keeps going past the character budget';
  const result = validateContributionTitle(longTitle, {});
  assert.equal(result.valid, false);
});

test('supports numeric maxLength override', () => {
  const longTitle = '✨ feat(auth): this subject is just long enough for the new budget';
  const long = validateContributionTitle(longTitle, {maxLength: 100});
  assert.equal(long.valid, true);

  const short = validateContributionTitle(longTitle, {maxLength: 20});
  assert.equal(short.valid, false);
});

test('applies maxLength null only for dependabot[bot]', () => {
  const longSubject =
    'a very long subject that keeps going past the character budget even with extra detail and more than 72 characters';

  const bot = validateContributionTitle(`✨ feat(auth): ${longSubject}`,
    contributionTitleOptionsForAuthor('dependabot[bot]'),
  );
  assert.equal(bot.valid, true);

  const user = validateContributionTitle(`✨ feat(auth): ${longSubject}`,
    contributionTitleOptionsForAuthor('dependabot'),
  );
  assert.equal(user.valid, false);

  const unknown = validateContributionTitle(`✨ feat(auth): ${longSubject}`,
    contributionTitleOptionsForAuthor('human'),
  );
  assert.equal(unknown.valid, false);

  const bypass = contributionTitleOptionsForAuthor('dependabot[bot]');
  assert.equal(bypass.maxLength, null);
  assert.equal(contributionTitleOptionsForAuthor('dependabot').maxLength, undefined);
});

test('skips length validation when maxLength is null', () => {
  const longSubject =
    'a very long subject that keeps going past the character budget even with extra detail and more than 72 characters';
  const result = validateContributionTitle(`✨ feat(auth): ${longSubject}`, {
    maxLength: null,
  });
  assert.equal(result.valid, true);
});

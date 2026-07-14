import test from 'node:test';
import assert from 'node:assert/strict';

import {validateContributionTitle} from './contribution-title.mjs';

test('accepts a standard feat title', () => {
  const result = validateContributionTitle('✨ feat(auth): add login flow');
  assert.equal(result.valid, true);
});

test('accepts a standard fix title', () => {
  const result = validateContributionTitle('🐛 fix(api): handle empty response body');
  assert.equal(result.valid, true);
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

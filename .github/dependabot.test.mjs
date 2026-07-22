import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {test} from 'node:test';

const source = readFileSync(new URL('dependabot.yml', import.meta.url), 'utf8');

test('uses the Dependabot schema value for ignored major updates', () => {
  const validValues = source.match(/^\s+- version-update:semver-major$/gm) ?? [];

  assert.equal(validValues.length, 4);
  assert.doesNotMatch(source, /^\s+- major$/m);
});

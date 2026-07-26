import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {test} from 'node:test';

import {
  parseDependabotUpdates,
  REQUIRED_UPDATE_TARGETS,
} from '../scripts/verify-supply-chain.mjs';

const source = readFileSync(new URL('dependabot.yml', import.meta.url), 'utf8');

test('uses valid major-ignore values with complete update coverage', () => {
  const validValues = source.match(/^\s+- version-update:semver-major$/gm) ?? [];
  const updates = parseDependabotUpdates(source);
  const configured = new Set(
    updates.map(({ecosystem, directory}) => `${ecosystem}:${directory}`),
  );

  assert.equal(validValues.length, updates.length);
  assert.doesNotMatch(source, /^\s+- major$/m);
  for (const [ecosystem, directory] of REQUIRED_UPDATE_TARGETS) {
    assert.ok(configured.has(`${ecosystem}:${directory}`));
  }
});

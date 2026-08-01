import assert from 'node:assert/strict';
import {readFileSync, readdirSync} from 'node:fs';
import {join} from 'node:path';
import {test} from 'node:test';
import {fileURLToPath} from 'node:url';

import {
  parseDependabotUpdates,
  REQUIRED_UPDATE_TARGETS,
} from '../scripts/verify-supply-chain.mjs';

const source = readFileSync(new URL('dependabot.yml', import.meta.url), 'utf8');
const repositoryRoot = fileURLToPath(new URL('../', import.meta.url));

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

test('binds every Docker update target to a local manifest', () => {
  const updates = parseDependabotUpdates(source);
  for (const {ecosystem, directory} of updates) {
    if (ecosystem !== 'docker') continue;
    const target = join(repositoryRoot, directory.replace(/^\/+/, ''));
    const manifests = readdirSync(target).filter((name) => (
      name === 'Dockerfile' || /\.ya?ml$/u.test(name)
    ));
    assert.ok(
      manifests.length > 0,
      `Dependabot Docker target has no manifest: ${directory}`,
    );
  }
});

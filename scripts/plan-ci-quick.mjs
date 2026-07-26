#!/usr/bin/env node

import {readFileSync} from 'node:fs';
import {fileURLToPath} from 'node:url';
import {resolve} from 'node:path';

const allowedRunners = new Set([
  'always',
  'soku-fast',
  'templates',
  'database-schema',
  'infrastructure',
]);
const allowedToolchains = new Set(['node', 'go', 'python', 'java', 'docker']);

function validateProfiles(profiles) {
  const quick = profiles?.profiles?.['ci-quick'];
  if (
    profiles?.schemaVersion !== 1 ||
    profiles?.format !== 'json-compatible-yaml' ||
    !Array.isArray(quick?.scopes) ||
    !Array.isArray(quick?.groups)
  ) {
    throw new Error('verification/profiles.yml has no supported ci-quick groups');
  }

  const knownScopes = new Set(quick.scopes);
  const ids = new Set();
  const groupForScope = new Map();
  for (const group of quick.groups) {
    if (
      !group?.id ||
      !group?.name ||
      !allowedRunners.has(group.runner) ||
      !Array.isArray(group.scopes) ||
      !Array.isArray(group.toolchains) ||
      group.scopes.some((scope) => !knownScopes.has(scope)) ||
      group.toolchains.some((toolchain) => !allowedToolchains.has(toolchain))
    ) {
      throw new Error(`invalid ci-quick group: ${group?.id ?? '(missing id)'}`);
    }
    if (ids.has(group.id)) {
      throw new Error(`duplicate ci-quick group: ${group.id}`);
    }
    ids.add(group.id);
    for (const scope of group.scopes) {
      if (groupForScope.has(scope)) {
        throw new Error(
          `ci-quick scope '${scope}' is assigned to more than one group`,
        );
      }
      groupForScope.set(scope, group.id);
    }
  }
  if (!quick.groups.some((group) => group.always === true)) {
    throw new Error('ci-quick requires at least one always-on group');
  }
  const unmapped = quick.scopes.filter((scope) => !groupForScope.has(scope));
  if (unmapped.length > 0) {
    throw new Error(`ci-quick scopes have no group: ${unmapped.join(', ')}`);
  }
  return quick;
}

export function planQuickGroups(scopeResult, profiles) {
  const quick = validateProfiles(profiles);
  if (
    scopeResult?.schemaVersion !== 1 ||
    !Array.isArray(scopeResult.scopes)
  ) {
    throw new Error('scope detector result has an unsupported schema');
  }

  const knownScopes = new Set(quick.scopes);
  const unexpected = scopeResult.scopes.filter(
    (scope) => !knownScopes.has(scope),
  );
  if (unexpected.length > 0) {
    throw new Error(`unknown ci-quick scopes: ${unexpected.join(', ')}`);
  }

  const selectedScopes = new Set(scopeResult.scopes);
  const include = quick.groups
    .filter(
      (group) =>
        group.always === true ||
        group.scopes.some((scope) => selectedScopes.has(scope)),
    )
    .map(({id, name, toolchains}) => ({id, name, toolchains}));

  return {include};
}

function parseArguments(args) {
  const options = {};
  for (let index = 0; index < args.length; index += 1) {
    const argument = args[index];
    if (argument === '--scope-json' || argument === '--profiles') {
      const value = args[index + 1];
      if (!value) throw new Error(`missing value for ${argument}`);
      options[argument.slice(2)] = value;
      index += 1;
    } else {
      throw new Error(`unknown argument: ${argument}`);
    }
  }
  if (!options['scope-json'] || !options.profiles) {
    throw new Error('--scope-json and --profiles are required');
  }
  return options;
}

function main() {
  try {
    const options = parseArguments(process.argv.slice(2));
    const scopeResult = JSON.parse(
      readFileSync(resolve(options['scope-json']), 'utf8'),
    );
    const profiles = JSON.parse(
      readFileSync(resolve(options.profiles), 'utf8'),
    );
    process.stdout.write(
      `${JSON.stringify(planQuickGroups(scopeResult, profiles))}\n`,
    );
  } catch (error) {
    process.stderr.write(`ci-quick planner: ${error.message}\n`);
    process.exitCode = 2;
  }
}

if (
  process.argv[1] &&
  resolve(process.argv[1]) === fileURLToPath(import.meta.url)
) {
  main();
}

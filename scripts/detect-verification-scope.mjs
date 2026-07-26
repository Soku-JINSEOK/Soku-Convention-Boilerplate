#!/usr/bin/env node

import {readFileSync} from 'node:fs';
import {dirname, resolve} from 'node:path';
import {spawnSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';

const scriptRoot = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const defaultConfigPath = resolve(scriptRoot, 'verification/scopes.yml');

function usage() {
  process.stdout.write(`Usage: scripts/detect-verification-scope.mjs [options]

Changed-file input (choose one):
  --staged                  Inspect staged changes (default).
  --base <sha> --head <sha> Inspect a commit range.
  --files-from <path|->     Read paths or name-status records from a file/stdin.

Output and location:
  --json                    Emit the versioned JSON result.
  --workspace <path>        Git workspace. Defaults to the current directory.
  --help                    Show this help.

Without --json, selected scopes are printed one per line.
`);
}

function parseArguments(argv) {
  const options = {
    mode: 'staged',
    base: '',
    head: '',
    filesFrom: '',
    json: false,
    workspace: process.cwd(),
  };
  let explicitInput = false;

  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    const value = () => {
      const next = argv[index + 1];
      if (!next) throw new Error(`Missing value for ${argument}`);
      index += 1;
      return next;
    };

    switch (argument) {
      case '--staged':
        if (explicitInput) throw new Error('Changed-file input modes are mutually exclusive');
        explicitInput = true;
        options.mode = 'staged';
        break;
      case '--base':
        if (options.filesFrom) throw new Error('Changed-file input modes are mutually exclusive');
        explicitInput = true;
        options.mode = 'range';
        options.base = value();
        break;
      case '--head':
        if (options.filesFrom) throw new Error('Changed-file input modes are mutually exclusive');
        explicitInput = true;
        options.mode = 'range';
        options.head = value();
        break;
      case '--files-from':
        if (options.mode === 'range' || explicitInput) {
          throw new Error('Changed-file input modes are mutually exclusive');
        }
        explicitInput = true;
        options.mode = 'files';
        options.filesFrom = value();
        break;
      case '--workspace':
        options.workspace = value();
        break;
      case '--json':
        options.json = true;
        break;
      case '--help':
        options.help = true;
        break;
      default:
        throw new Error(`Unknown argument: ${argument}`);
    }
  }

  if (options.mode === 'range' && (!options.base || !options.head)) {
    throw new Error('--base and --head must be provided together');
  }
  for (const [name, value] of [
    ['--base', options.base],
    ['--head', options.head],
  ]) {
    if (value && !/^[0-9a-f]{7,64}$/i.test(value)) {
      throw new Error(`${name} must be a hexadecimal commit SHA`);
    }
  }
  if (options.mode === 'files' && !options.filesFrom) {
    throw new Error('--files-from requires a path or -');
  }
  return options;
}

function runGit(workspace, args) {
  const result = spawnSync('git', args, {
    cwd: workspace,
    encoding: null,
    maxBuffer: 16 * 1024 * 1024,
  });
  if (result.status !== 0) {
    const message = result.stderr?.toString('utf8').trim() || 'git command failed';
    throw new Error(message);
  }
  return result.stdout;
}

function isStatus(value) {
  return /^(?:[ACDMRTUXB]|R\d{1,3}|C\d{1,3})$/.test(value);
}

function addChangedPath(paths, value) {
  const normalized = value.replaceAll('\\', '/').replace(/^\.\//, '');
  if (normalized && !paths.includes(normalized)) paths.push(normalized);
}

export function parseNameStatus(buffer) {
  const paths = [];
  const values = buffer.toString('utf8').split('\0');

  for (let index = 0; index < values.length; ) {
    const status = values[index++];
    if (!status) continue;
    if (!isStatus(status)) {
      addChangedPath(paths, status);
      continue;
    }
    const firstPath = values[index++] ?? '';
    addChangedPath(paths, firstPath);
    if (/^[RC]/.test(status)) {
      const secondPath = values[index++] ?? '';
      addChangedPath(paths, secondPath);
    }
  }
  return paths;
}

export function parseFilesInput(buffer) {
  if (buffer.includes(0)) return parseNameStatus(buffer);

  const paths = [];
  for (const rawLine of buffer.toString('utf8').split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line) continue;
    const fields = line.split('\t');
    if (isStatus(fields[0])) {
      for (const path of fields.slice(1)) addChangedPath(paths, path);
    } else {
      addChangedPath(paths, line);
    }
  }
  return paths;
}

function changedFiles(options) {
  if (options.mode === 'staged') {
    return parseNameStatus(
      runGit(options.workspace, [
        'diff',
        '--cached',
        '--name-status',
        '-z',
        '--find-renames',
        '--',
      ]),
    );
  }
  if (options.mode === 'range') {
    return parseNameStatus(
      runGit(options.workspace, [
        'diff',
        '--name-status',
        '-z',
        '--find-renames',
        options.base,
        options.head,
        '--',
      ]),
    );
  }
  const input =
    options.filesFrom === '-'
      ? readFileSync(0)
      : readFileSync(resolve(options.workspace, options.filesFrom));
  return parseFilesInput(input);
}

function globExpression(pattern) {
  let expression = '^';
  for (let index = 0; index < pattern.length; index += 1) {
    const character = pattern[index];
    if (character === '*' && pattern[index + 1] === '*') {
      if (pattern[index + 2] === '/') {
        expression += '(?:.*/)?';
        index += 2;
      } else {
        expression += '.*';
        index += 1;
      }
    } else if (character === '*') {
      expression += '[^/]*';
    } else if (character === '?') {
      expression += '[^/]';
    } else {
      expression += character.replace(/[|\\{}()[\]^$+?.]/g, '\\$&');
    }
  }
  return new RegExp(`${expression}$`);
}

function loadConfig(path = defaultConfigPath) {
  const config = JSON.parse(readFileSync(path, 'utf8'));
  if (
    config.schemaVersion !== 1 ||
    config.format !== 'json-compatible-yaml' ||
    !Array.isArray(config.allScopes) ||
    !Array.isArray(config.rules) ||
    !config.default
  ) {
    throw new Error('verification/scopes.yml does not match schema version 1');
  }
  const knownScopes = new Set(config.allScopes);
  for (const rule of config.rules) {
    if (!rule.id || !Array.isArray(rule.paths) || !rule.reason) {
      throw new Error('verification/scopes.yml contains an invalid rule');
    }
    if (
      rule.scopes !== 'all' &&
      (!Array.isArray(rule.scopes) ||
        rule.scopes.some((scope) => !knownScopes.has(scope)))
    ) {
      throw new Error(`verification/scopes.yml rule '${rule.id}' has invalid scopes`);
    }
  }
  return config;
}

export function detectScopes(files, config = loadConfig()) {
  const selected = new Set();
  const reasons = [];
  let allSelected = false;

  for (const path of files) {
    const safePath =
      !path.startsWith('/') &&
      !path.split('/').includes('..') &&
      !path.includes('\0');
    const rule = safePath
      ? config.rules.find(({paths}) =>
          paths.some((pattern) => globExpression(pattern).test(path)),
        )
      : undefined;
    const effective = rule ?? {
      id: 'unknown-path',
      scopes: config.default.scopes,
      reason: config.default.reason,
    };
    const scopes =
      effective.scopes === 'all' ? [...config.allScopes] : effective.scopes;
    if (effective.scopes === 'all') allSelected = true;
    for (const scope of scopes) selected.add(scope);
    reasons.push({
      path,
      rule: effective.id,
      scopes,
      reason: effective.reason,
    });
  }

  return {
    schemaVersion: 1,
    changedFiles: [...files],
    scopes: config.allScopes.filter((scope) => selected.has(scope)),
    reasons,
    allSelected,
  };
}

function main() {
  let options;
  try {
    options = parseArguments(process.argv.slice(2));
    if (options.help) {
      usage();
      return;
    }
    const result = detectScopes(changedFiles(options));
    if (options.json) {
      process.stdout.write(`${JSON.stringify(result, null, 2)}\n`);
    } else if (result.scopes.length > 0) {
      process.stdout.write(`${result.scopes.join('\n')}\n`);
    }
  } catch (error) {
    process.stderr.write(`scope detector: ${error.message}\n`);
    process.exitCode = 2;
  }
}

if (
  process.argv[1] &&
  resolve(process.argv[1]) === fileURLToPath(import.meta.url)
) {
  main();
}

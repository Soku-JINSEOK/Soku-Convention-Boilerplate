#!/usr/bin/env node

import fs from 'node:fs';
import path from 'node:path';
import {fileURLToPath} from 'node:url';

const scriptPath = fileURLToPath(import.meta.url);
const requiredStaticFiles = [
  'README.md',
  'README.ko.md',
  'README.ja.md',
  'soku/README.md',
  'soku/npm/package.json',
  '.github/workflows/release.yml',
  'docs/guides/USAGE_MANUAL.md',
];

const escapeRegExp = (value) =>
  value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');

function requireVersion(value, field, errors) {
  if (typeof value !== 'string' || !/^\d+\.\d+\.\d+$/.test(value)) {
    errors.push(`${field} must be a stable MAJOR.MINOR.PATCH version.`);
  }
}

function requireTag(value, pattern, field, errors) {
  if (typeof value !== 'string' || !pattern.test(value)) {
    errors.push(`${field} has an invalid release tag.`);
  }
}

function requireSnippet(files, filePath, snippet, errors) {
  if (!files[filePath]?.includes(snippet)) {
    errors.push(`${filePath} is missing release identity: ${snippet}`);
  }
}

function verifySignerRotationLog(content, activeFingerprint, errors) {
  if (typeof content !== 'string') return;

  const rows = content
    .split(/\r?\n/)
    .filter((line) => line.startsWith('|'))
    .map((line) => line.split('|').slice(1, -1).map((cell) => cell.trim()))
    .filter(
      (cells) =>
        cells[0] !== 'Previous fingerprint' &&
        !cells.every((cell) => /^-+$/.test(cell)),
    );
  if (rows.length === 0) {
    errors.push('signer rotation log must contain at least one history row.');
    return;
  }

  const normalizedFingerprint = (value) =>
    value === 'none' ? value : value.replace(/^`|`$/g, '');
  let previousNewFingerprint;
  const observedFingerprints = new Set();

  rows.forEach((cells, index) => {
    if (cells.length !== 5 || cells.some((cell) => cell.length === 0)) {
      errors.push(
        `signer rotation history row ${index + 1} must contain five non-empty fields.`,
      );
      return;
    }

    const previousFingerprint = normalizedFingerprint(cells[0]);
    const newFingerprint = normalizedFingerprint(cells[1]);
    if (
      (index === 0 && previousFingerprint !== 'none') ||
      (index > 0 && previousFingerprint !== previousNewFingerprint)
    ) {
      errors.push(
        `signer rotation history row ${index + 1} breaks fingerprint continuity.`,
      );
    }
    if (!/^[A-F0-9]{40}$/.test(newFingerprint)) {
      errors.push(
        `signer rotation history row ${index + 1} has an invalid new fingerprint.`,
      );
    }
    if (observedFingerprints.has(newFingerprint)) {
      errors.push(
        `signer rotation history row ${index + 1} reuses a fingerprint.`,
      );
    }
    observedFingerprints.add(newFingerprint);
    previousNewFingerprint = newFingerprint;
  });

  if (previousNewFingerprint !== activeFingerprint) {
    errors.push(
      'the final signer rotation fingerprint must match the active fingerprint.',
    );
  }
}

function workflowDefault(workflow, input) {
  return workflow.match(
    new RegExp(`${escapeRegExp(input)}:[\\s\\S]*?default:\\s*([^\\s]+)`),
  )?.[1];
}

export function verifyReleaseIdentity(identity, files) {
  const errors = [];
  const boilerplate = identity?.boilerplate ?? {};
  const cli = identity?.cli ?? {};
  const npm = identity?.npm ?? {};
  const compatibility = identity?.compatibility ?? {};
  const signing = identity?.signing ?? {};

  if (identity?.schemaVersion !== 1) {
    errors.push('schemaVersion must be 1.');
  }
  requireVersion(boilerplate.version, 'boilerplate.version', errors);
  requireVersion(cli.version, 'cli.version', errors);
  requireVersion(npm.version, 'npm.version', errors);
  requireTag(
    boilerplate.tag,
    /^v\d+\.\d+\.\d+$/,
    'boilerplate.tag',
    errors,
  );
  requireTag(cli.tag, /^soku\/v\d+\.\d+\.\d+$/, 'cli.tag', errors);
  requireTag(
    compatibility.boilerplateTag,
    /^v\d+\.\d+\.\d+$/,
    'compatibility.boilerplateTag',
    errors,
  );
  requireTag(
    compatibility.cliTag,
    /^soku\/v\d+\.\d+\.\d+$/,
    'compatibility.cliTag',
    errors,
  );
  requireTag(
    compatibility.npmSupportedFromCliTag,
    /^soku\/v\d+\.\d+\.\d+$/,
    'compatibility.npmSupportedFromCliTag',
    errors,
  );
  if (
    typeof signing.activeFingerprint !== 'string' ||
    !/^[A-F0-9]{40}$/.test(signing.activeFingerprint)
  ) {
    errors.push(
      'signing.activeFingerprint must be an uppercase 40-character GPG fingerprint.',
    );
  }
  if (signing.rotationLog !== 'docs/releases/SIGNER_ROTATIONS.md') {
    errors.push(
      'signing.rotationLog must reference docs/releases/SIGNER_ROTATIONS.md.',
    );
  }

  if (boilerplate.tag !== `v${boilerplate.version}`) {
    errors.push('boilerplate tag and version must match.');
  }
  if (cli.tag !== `soku/v${cli.version}`) {
    errors.push('CLI tag and version must match.');
  }
  if (npm.version !== cli.version) {
    errors.push('npm and CLI versions must match for coordinated publication.');
  }
  if (compatibility.boilerplateTag !== boilerplate.tag) {
    errors.push('compatibility boilerplate tag must match the current release.');
  }

  let packageMetadata;
  try {
    packageMetadata = JSON.parse(files['soku/npm/package.json']);
  } catch {
    errors.push('soku/npm/package.json must contain valid JSON.');
  }
  if (packageMetadata?.name !== npm.package) {
    errors.push('npm package name has drifted from release-identity.json.');
  }
  if (packageMetadata?.version !== npm.version) {
    errors.push('npm package version has drifted from release-identity.json.');
  }

  const workflow = files['.github/workflows/release.yml'] ?? '';
  if (workflowDefault(workflow, 'boilerplate-tag') !== boilerplate.tag) {
    errors.push('Release workflow boilerplate default has drifted.');
  }
  if (workflowDefault(workflow, 'cli-tag') !== cli.tag) {
    errors.push('Release workflow CLI default has drifted.');
  }

  for (const filePath of ['README.md', 'README.ko.md', 'README.ja.md']) {
    requireSnippet(
      files,
      filePath,
      `npm install -g ${npm.package}@${npm.version}`,
      errors,
    );
    requireSnippet(files, filePath, boilerplate.tag, errors);
    requireSnippet(files, filePath, cli.tag, errors);
  }

  for (const snippet of [
    cli.tag,
    `${npm.package}@${npm.version}`,
    compatibility.cliTag,
  ]) {
    requireSnippet(files, 'soku/README.md', snippet, errors);
  }

  for (const snippet of [
    boilerplate.tag,
    cli.tag,
    `${npm.package}@${npm.version}`,
    compatibility.cliTag,
  ]) {
    requireSnippet(files, 'docs/guides/USAGE_MANUAL.md', snippet, errors);
  }

  requireSnippet(
    files,
    boilerplate.releaseNotes,
    `# Boilerplate ${boilerplate.tag}`,
    errors,
  );
  requireSnippet(
    files,
    boilerplate.releaseNotes,
    `Compatible soku: ${compatibility.cliTag}`,
    errors,
  );
  requireSnippet(
    files,
    cli.releaseNotes,
    `# soku v${cli.version}`,
    errors,
  );
  requireSnippet(
    files,
    cli.releaseNotes,
    `Boilerplate compatibility: ${compatibility.boilerplateTag}`,
    errors,
  );
  requireSnippet(files, cli.releaseNotes, cli.tag, errors);
  requireSnippet(
    files,
    signing.rotationLog,
    `Current active fingerprint: \`${signing.activeFingerprint}\``,
    errors,
  );
  verifySignerRotationLog(
    files[signing.rotationLog],
    signing.activeFingerprint,
    errors,
  );

  return errors;
}

export function readReleaseIdentity(repositoryRoot) {
  const manifestPath = path.join(repositoryRoot, 'release-identity.json');
  const identity = JSON.parse(fs.readFileSync(manifestPath, 'utf8'));
  const filePaths = [
    ...requiredStaticFiles,
    identity.boilerplate.releaseNotes,
    identity.cli.releaseNotes,
    ...(typeof identity.signing?.rotationLog === 'string'
      ? [identity.signing.rotationLog]
      : []),
  ];
  const files = Object.fromEntries(
    filePaths.map((filePath) => [
      filePath,
      fs.readFileSync(path.join(repositoryRoot, filePath), 'utf8'),
    ]),
  );
  return {identity, files};
}

if (process.argv[1] && path.resolve(process.argv[1]) === scriptPath) {
  const repositoryRoot = process.argv[2]
    ? path.resolve(process.argv[2])
    : path.resolve(path.dirname(scriptPath), '..');
  const {identity, files} = readReleaseIdentity(repositoryRoot);
  const errors = verifyReleaseIdentity(identity, files);

  if (errors.length > 0) {
    for (const error of errors) console.error(`- ${error}`);
    process.exit(1);
  }

  console.log(
    `Release identity verification passed: ${identity.boilerplate.tag}, ` +
      `${identity.cli.tag}, ${identity.npm.package}@${identity.npm.version}.`,
  );
}

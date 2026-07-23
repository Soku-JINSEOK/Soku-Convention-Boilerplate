#!/usr/bin/env node
import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';

import {
  isNpmPublishReady,
  releaseTagFromVersion,
} from '../lib/launcher.mjs';

function usage() {
  console.error(
    'Usage: node scripts/prepare-package.mjs --version <0.0.0> [--repo-root <path>]',
  );
  process.exit(2);
}

const args = process.argv.slice(2);
let version = '';
let repoRoot = process.cwd();
for (let i = 0; i < args.length; i++) {
  if (args[i] === '--version') {
    const next = args[i + 1];
    if (!next || next.startsWith('--')) {
      usage();
    }
    version = next;
    i++;
    continue;
  }
  if (args[i] === '--repo-root') {
    const next = args[i + 1];
    if (!next || next.startsWith('--')) {
      usage();
    }
    repoRoot = next;
    i++;
    continue;
  }
  usage();
}

if (!version) {
  usage();
}
const npmPackagePath = path.join(repoRoot, 'soku', 'npm', 'package.json');
const notesPath = path.join(repoRoot, 'docs', 'releases', `soku-v${version}.md`);

if (!isNpmPublishReady(version)) {
  throw new Error(`npm publication is not enabled for soku/v${version} yet`);
}

if (!fs.existsSync(npmPackagePath)) {
  throw new Error(`missing npm package manifest: ${npmPackagePath}`);
}

const packageJson = JSON.parse(fs.readFileSync(npmPackagePath, 'utf8'));
if (packageJson.version !== version) {
  throw new Error(`package version mismatch: ${packageJson.version} != ${version}`);
}

if (!fs.existsSync(notesPath)) {
  throw new Error(`missing release notes: ${notesPath}`);
}

const packageTag = releaseTagFromVersion(packageJson.version);
if (!packageTag.startsWith('soku/v')) {
  throw new Error(`invalid release tag derivation for package version ${packageJson.version}`);
}

if (!packageJson.soku?.githubRepository) {
  throw new Error('missing soku.githubRepository in package.json');
}

console.log(`Prepared npm package ${packageJson.name}@${packageJson.version} for ${packageTag}.`);

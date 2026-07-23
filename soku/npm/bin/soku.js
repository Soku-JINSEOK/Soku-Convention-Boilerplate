#!/usr/bin/env node
import fs from 'node:fs';
import {spawnSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';
import {dirname, resolve} from 'node:path';
import process from 'node:process';

import {isNpmPublishReady, releaseTagFromVersion, resolveBinary, resolveTarget} from '../lib/launcher.mjs';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const packagePath = resolve(__dirname, '../package.json');
const packageJson = JSON.parse(fs.readFileSync(packagePath, 'utf8'));

function getRepository(cliPackage) {
  return process.env.SOKU_GITHUB_REPOSITORY || cliPackage?.soku?.githubRepository || 'Soku-JINSEOK/Soku-Convention-Boilerplate';
}

async function main() {
  const version = packageJson.version;
  const targetTag = releaseTagFromVersion(version);
  if (!isNpmPublishReady(version)) {
    throw new Error(
      `package is intentionally not published before ${packageJson.soku.minNpmVersion}; version is ${version}`,
    );
  }

  const repository = getRepository(packageJson);
  const target = resolveTarget(process.platform, process.arch);
  const binaryPath = await resolveBinary({
    version,
    repository,
    target,
    targetTag,
  });
  const child = spawnSync(binaryPath, process.argv.slice(2), {
    stdio: 'inherit',
    env: {
      ...process.env,
      SOKU_LAUNCHER: 'npm',
    },
  });
  if (child.status === null) {
    throw new Error(child.error?.message ?? 'failed to execute native soku binary');
  }
  process.exit(child.status);
}

main().catch((error) => {
  console.error(`[soku] ${error?.message ?? error}`);
  process.exit(1);
});

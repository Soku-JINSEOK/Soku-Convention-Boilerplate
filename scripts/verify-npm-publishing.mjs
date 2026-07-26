#!/usr/bin/env node
import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import {fileURLToPath, pathToFileURL} from 'node:url';

const EXPECTED_NPM_CLI = '12.0.1';
const EXPECTED_REPOSITORY =
  'https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate.git';

export function verifyNpmPublishing(workflow, packageJson) {
  const errors = [];
  const marker = '\n  publish-npm:\n';
  const markerIndex = workflow.indexOf(marker);
  const publishJob = markerIndex >= 0 ? workflow.slice(markerIndex) : '';

  if (!publishJob) {
    return ['release workflow is missing the publish-npm job'];
  }

  if (!/\n\s+contents:\s+read\s*$/m.test(publishJob)) {
    errors.push('publish-npm must grant contents: read');
  }
  if (!/\n\s+id-token:\s+write\s*$/m.test(publishJob)) {
    errors.push('publish-npm must grant id-token: write');
  }
  if (!/node-version:\s+["']?24["']?\s*$/m.test(publishJob)) {
    errors.push('publish-npm must use Node.js 24');
  }
  if (!/package-manager-cache:\s+false\s*$/m.test(publishJob)) {
    errors.push('publish-npm must disable automatic package-manager caching');
  }
  if (
    !publishJob.includes(`npm install --global npm@${EXPECTED_NPM_CLI}`)
  ) {
    errors.push(`publish-npm must install npm@${EXPECTED_NPM_CLI}`);
  }
  if (publishJob.includes('registry-url:')) {
    errors.push('publish-npm must not create token-oriented registry authentication');
  }
  if (workflow.includes('NPM_TOKEN') || workflow.includes('NODE_AUTH_TOKEN')) {
    errors.push('release workflow must not inject an npm publication token');
  }
  if (!/npm publish\s+--provenance\s+--access public/.test(publishJob)) {
    errors.push('publish-npm must publish publicly with provenance');
  }

  const repositoryUrl =
    typeof packageJson.repository === 'string'
      ? packageJson.repository
      : packageJson.repository?.url;
  if (repositoryUrl !== EXPECTED_REPOSITORY) {
    errors.push(`npm repository URL must exactly match ${EXPECTED_REPOSITORY}`);
  }

  return errors;
}

function main() {
  const repoRoot = path.resolve(
    path.dirname(fileURLToPath(import.meta.url)),
    '..',
  );
  const workflowPath = path.join(repoRoot, '.github', 'workflows', 'release.yml');
  const packagePath = path.join(repoRoot, 'soku', 'npm', 'package.json');
  const workflow = fs.readFileSync(workflowPath, 'utf8');
  const packageJson = JSON.parse(fs.readFileSync(packagePath, 'utf8'));
  const errors = verifyNpmPublishing(workflow, packageJson);

  if (errors.length > 0) {
    for (const error of errors) {
      console.error(`ERROR: ${error}`);
    }
    process.exit(1);
  }

  console.log(
    `Verified npm Trusted Publishing contract with npm@${EXPECTED_NPM_CLI}.`,
  );
}

if (
  process.argv[1] &&
  import.meta.url === pathToFileURL(path.resolve(process.argv[1])).href
) {
  main();
}

#!/usr/bin/env node

import {appendFileSync, readFileSync} from 'node:fs';
import {fileURLToPath} from 'node:url';
import {resolve} from 'node:path';

export function verifyPromotion({
  manifest,
  run,
  repository,
  workflowPath = '.github/workflows/validation.yml',
}) {
  const errors = [];
  if (manifest?.schemaVersion !== 1) errors.push('unsupported manifest schema');
  if (manifest?.repository !== repository) errors.push('manifest repository mismatch');
  if (manifest?.sourceRef !== 'refs/heads/main') errors.push('manifest sourceRef is not canonical main');
  if (!/^[0-9a-f]{40}$/.test(manifest?.sourceSha ?? '')) errors.push('manifest sourceSha is invalid');
  if (!/^[0-9]+$/.test(String(manifest?.workflowRunId ?? ''))) errors.push('manifest workflowRunId is invalid');
  if (!/^sha256:[0-9a-f]{64}$/.test(manifest?.digest ?? '')) errors.push('manifest digest is invalid');
  if (
    typeof manifest?.imageUri !== 'string' ||
    !manifest.imageUri.endsWith(`@${manifest.digest}`)
  ) {
    errors.push('manifest image URI and digest mismatch');
  }
  if (Number(manifest?.workflowRunId) !== Number(run?.id)) errors.push('run ID mismatch');
  if (run?.event !== 'push') errors.push('source run is not a push');
  if (run?.head_branch !== 'main') errors.push('source run is not on main');
  if (run?.head_sha !== manifest?.sourceSha) errors.push('source SHA mismatch');
  if (run?.conclusion !== 'success') errors.push('source run did not succeed');
  if (run?.path !== workflowPath) errors.push('source workflow mismatch');
  if (run?.repository?.full_name !== repository) errors.push('run repository mismatch');
  if (run?.head_repository?.full_name !== repository) errors.push('run head repository mismatch');
  if (errors.length > 0) throw new Error(errors.join('; '));
  return {
    sourceSha: manifest.sourceSha,
    imageUri: manifest.imageUri,
    digest: manifest.digest,
  };
}

function parseArguments(args) {
  const options = {};
  for (let index = 0; index < args.length; index += 1) {
    const name = args[index];
    if (['--manifest', '--run-json', '--repository', '--github-output'].includes(name)) {
      const value = args[index + 1];
      if (!value) throw new Error(`missing value for ${name}`);
      options[name.slice(2)] = value;
      index += 1;
    } else {
      throw new Error(`unknown argument: ${name}`);
    }
  }
  for (const required of ['manifest', 'run-json', 'repository']) {
    if (!options[required]) throw new Error(`--${required} is required`);
  }
  return options;
}

function main() {
  try {
    const options = parseArguments(process.argv.slice(2));
    const verified = verifyPromotion({
      manifest: JSON.parse(readFileSync(resolve(options.manifest), 'utf8')),
      run: JSON.parse(readFileSync(resolve(options['run-json']), 'utf8')),
      repository: options.repository,
    });
    const output = [
      `source_sha=${verified.sourceSha}`,
      `image_uri=${verified.imageUri}`,
      `digest=${verified.digest}`,
    ].join('\n');
    if (options['github-output']) {
      appendFileSync(resolve(options['github-output']), `${output}\n`);
    } else {
      process.stdout.write(`${output}\n`);
    }
  } catch (error) {
    process.stderr.write(`image promotion: ${error.message}\n`);
    process.exitCode = 2;
  }
}

if (
  process.argv[1] &&
  resolve(process.argv[1]) === fileURLToPath(import.meta.url)
) {
  main();
}

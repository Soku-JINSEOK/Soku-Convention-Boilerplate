#!/usr/bin/env node

import { spawnSync } from 'node:child_process';
import { existsSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, isAbsolute, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const SHA40 = /^[0-9a-f]{40}$/;
const IMAGE_DIGEST = /^[a-z0-9./_-]+:[^@\s]+@sha256:[0-9a-f]{64}$/;
const APPROVED_BUILDER =
  'node:24.17.0-bookworm@sha256:733e1c06ada118ed9f6133a31aa1290be6929664026fb28821500437c61f2c6f';
const REQUIRED_SAMPLE_FIELDS = [
  'provider',
  'repository',
  'sourceSha',
  'baseSha',
  'pullRequest',
  'naturalEventId',
  'buildId',
  'triggerId',
  'triggerVersion',
  'attempt',
  'validationProfile',
  'startedAt',
  'completedAt',
  'status',
  'conclusion',
  'commandContractVersion',
  'evidenceReference',
  'duplicate',
  'cancellationReason',
];

const FORBIDDEN_CONFIG_PATTERNS = [
  /\bdocker\s+push\b/i,
  /\bgcloud\s+run\s+deploy\b/i,
  /\b(?:npm|pnpm)\s+publish\b/i,
  /\bartifactregistry\b/i,
  /\bsecretEnv\b/i,
  /\bavailableSecrets\b/i,
  /credentials\.json/i,
  /service-account-key/i,
  /BEGIN\s+(?:RSA|EC|OPENSSH)?\s*PRIVATE\s+KEY/i,
  /\bworkflow_dispatch\b/i,
  /branch\s+protection/i,
  /required[-\s]?check.*(?:change|transition|replace)/i,
  /actions.*(?:disable|replace|retire)/i,
];

function fail(message) {
  throw new Error(message);
}

function isObject(value) {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function requireValue(condition, message) {
  if (!condition) fail(message);
}

function requireExact(actual, expected, label) {
  requireValue(actual === expected, label + ' must equal ' + JSON.stringify(expected));
}

function requireInteger(actual, expected, label) {
  requireValue(Number.isInteger(actual) && actual === expected, label + ' must equal ' + expected);
}

function requireBoolean(actual, expected, label) {
  requireValue(typeof actual === 'boolean' && actual === expected, label + ' must be ' + expected);
}

function readJson(filePath, label) {
  requireValue(existsSync(filePath), label + ' does not exist: ' + filePath);
  try {
    return JSON.parse(readFileSync(filePath, 'utf8'));
  } catch (error) {
    fail(label + ' must be valid JSON-compatible YAML/JSON: ' + error.message);
  }
}

function validateSafeRelativePath(value, label) {
  requireValue(typeof value === 'string' && value.length > 0, label + ' is required');
  requireValue(!isAbsolute(value), label + ' must be relative');
  requireValue(!value.split('/').includes('..'), label + ' must not traverse');
}

export function validateShadowConfig(config) {
  requireValue(isObject(config), 'shadow config must be an object');
  requireValue(Array.isArray(config.steps) && config.steps.length === 1, 'shadow config must have exactly one step');

  const step = config.steps[0];
  requireValue(isObject(step), 'shadow step must be an object');
  requireExact(step.id, 'gcp-shadow-quick', 'shadow step id');
  requireExact(step.entrypoint, 'bash', 'shadow step entrypoint');
  requireValue(typeof step.name === 'string' && IMAGE_DIGEST.test(step.name), 'builder must use an immutable sha256 digest');
  requireExact(step.name, APPROVED_BUILDER, 'builder identity');
  requireValue(!/@sha256:.*latest/i.test(step.name) && !/:latest@/i.test(step.name), 'latest builder reference is prohibited');
  requireValue(Array.isArray(step.args) && step.args.length === 2, 'shadow step args must contain bash mode and one script');
  requireExact(step.args[0], '-ceu', 'shadow step shell mode');
  requireValue(typeof step.args[1] === 'string', 'shadow step command must be a string');

  const command = step.args[1];
  requireValue(command.includes('node scripts/verify-cloud-build-shadow.mjs'), 'shadow step must invoke the shadow validator');
  requireValue(!command.includes('scripts/verify.sh'), 'validation commands must not be duplicated in Cloud Build YAML');
  requireExact((command.match(/--profile\s+ci-quick/g) || []).length, 1, 'ci-quick profile invocation count');
  requireValue(!/--profile\s+full\b/.test(command), 'Quick and Full profiles must not be invoked together');
  for (const token of [
    '\${COMMIT_SHA}',
    '\${_BASE_SHA}',
    '\${REPO_FULL_NAME}',
    '\${_PR_NUMBER}',
    '\${BUILD_ID}',
    '\${TRIGGER_NAME}',
  ]) {
    requireValue(command.includes(token), 'Cloud Build source identity substitution missing: ' + token);
  }
  for (const flag of [
    '--source-sha',
    '--base-sha',
    '--repository',
    '--pull-request',
    '--build-id',
    '--trigger',
    '--attempt',
    '--event pull_request',
    '--execute',
  ]) {
    requireValue(command.includes(flag), 'shadow runtime binding missing: ' + flag);
  }
  for (const binding of [
    '--repository "\${REPO_FULL_NAME}"',
    '--pull-request "\${_PR_NUMBER}"',
    '--base-sha "\${_BASE_SHA}"',
    '--source-sha "\${COMMIT_SHA}"',
    '--build-id "\${BUILD_ID}"',
    '--trigger "\${TRIGGER_NAME}"',
    '--attempt "\${BUILD_ID}"',
  ]) {
    requireValue(command.includes(binding), 'shadow runtime binding is not exact: ' + binding);
  }

  const serialized = JSON.stringify(config) + '\n' + command;
  for (const pattern of FORBIDDEN_CONFIG_PATTERNS) {
    requireValue(!pattern.test(serialized), 'forbidden deployment, credential, or control-plane content detected: ' + pattern);
  }

  requireExact(config.serviceAccount, '\${_CLOUD_BUILD_SERVICE_ACCOUNT}', 'serviceAccount');
  requireExact(config.timeout, '900s', 'build timeout');
  requireValue(isObject(config.options), 'build options are required');
  requireExact(config.options.logging, 'CLOUD_LOGGING_ONLY', 'build logging mode');
  requireValue(!Object.hasOwn(config, 'images'), 'artifact image publication is prohibited');
  requireValue(!Object.hasOwn(config, 'availableSecrets'), 'secret configuration is prohibited');
  requireValue(!Object.hasOwn(config, 'secrets'), 'secret configuration is prohibited');
  requireValue(!Object.hasOwn(config, 'env'), 'ambient environment injection is prohibited');
  requireValue(isObject(config.substitutions), 'explicit substitution contract is required');
  const substitutionKeys = Object.keys(config.substitutions).sort();
  requireValue(
    JSON.stringify(substitutionKeys) === JSON.stringify(['_BASE_SHA', '_CLOUD_BUILD_SERVICE_ACCOUNT']),
    'unexpected substitutions are not allowed',
  );
  requireExact(config.substitutions._BASE_SHA, '', '_BASE_SHA default');
  requireExact(config.substitutions._CLOUD_BUILD_SERVICE_ACCOUNT, '', '_CLOUD_BUILD_SERVICE_ACCOUNT default');

  return {
    builder: step.name,
    profile: 'ci-quick',
    timeout: config.timeout,
    format: 'json-compatible-yaml',
  };
}

export function validateProviderContract(contract) {
  requireValue(isObject(contract), 'provider contract must be an object');
  requireInteger(contract.schemaVersion, 1, 'schemaVersion');
  requireExact(contract.contractVersion, 'option-b-v1', 'contractVersion');
  requireExact(contract.provider, 'gcp-cloud-build-shadow', 'provider');
  requireExact(contract.validationProfile, 'ci-quick', 'validationProfile');
  requireExact(contract.validationEntrypoint, 'scripts/verify.sh', 'validationEntrypoint');
  requireInteger(contract.historicalActionsNaturalSamples, 3, 'historicalActionsNaturalSamples');
  requireInteger(contract.currentGcpNaturalSamples, 0, 'currentGcpNaturalSamples');
  requireInteger(contract.currentGcpSyntheticSamples, 0, 'currentGcpSyntheticSamples');
  requireInteger(contract.pilotNaturalTarget, 3, 'pilotNaturalTarget');
  requireInteger(contract.transitionEvaluationNaturalTarget, 10, 'transitionEvaluationNaturalTarget');
  requireBoolean(contract.providerSampleMergeAllowed, false, 'providerSampleMergeAllowed');

  requireValue(
    JSON.stringify(contract.eligibleNaturalSampleFields) === JSON.stringify(REQUIRED_SAMPLE_FIELDS),
    'eligibleNaturalSampleFields must preserve the complete evidence schema',
  );

  requireValue(isObject(contract.sourceShaBinding), 'sourceShaBinding is required');
  requireBoolean(contract.sourceShaBinding.required, true, 'sourceShaBinding.required');
  requireExact(contract.sourceShaBinding.cloudBuildSubstitution, 'COMMIT_SHA', 'sourceShaBinding.cloudBuildSubstitution');
  requireExact(contract.sourceShaBinding.argument, '--source-sha', 'sourceShaBinding.argument');
  requireExact(contract.sourceShaBinding.baseShaSubstitution, '_BASE_SHA', 'sourceShaBinding.baseShaSubstitution');
  requireExact(contract.sourceShaBinding.format, '40-lower-hex', 'sourceShaBinding.format');
  requireBoolean(contract.sourceShaBinding.mustMatchBuildSource, true, 'sourceShaBinding.mustMatchBuildSource');

  requireValue(isObject(contract.failureCancellationPolicy), 'failureCancellationPolicy is required');
  requireExact(contract.failureCancellationPolicy.denominator, 'included', 'failure denominator');
  requireExact(contract.failureCancellationPolicy.successfulNumerator, 'success-only', 'success numerator');
  requireExact(contract.failureCancellationPolicy.failedBuilds, 'included-and-classified', 'failed build handling');
  requireExact(contract.failureCancellationPolicy.cancelledBuilds, 'included-and-classified', 'cancelled build handling');
  requireExact(contract.failureCancellationPolicy.supersededHeads, 'included-and-classified', 'superseded head handling');

  requireValue(isObject(contract.duplicateAttemptPolicy), 'duplicateAttemptPolicy is required');
  requireExact(contract.duplicateAttemptPolicy.sameBuildCallbacks, 'count-once', 'same-build callback handling');
  requireExact(contract.duplicateAttemptPolicy.sameShaRerun, 'new-attempt-not-new-natural-sample', 'rerun handling');
  requireExact(contract.duplicateAttemptPolicy.duplicateNaturalSample, 'excluded', 'duplicate sample handling');

  requireValue(isObject(contract.syntheticPolicy), 'syntheticPolicy is required');
  requireBoolean(contract.syntheticPolicy.manualBuildCountsAsNatural, false, 'manualBuildCountsAsNatural');
  requireBoolean(contract.syntheticPolicy.syntheticBuildCountsAsNatural, false, 'syntheticBuildCountsAsNatural');
  requireExact(contract.syntheticPolicy.syntheticSamplesSeparateCounter, 'currentGcpSyntheticSamples', 'synthetic counter');
  requireExact(contract.syntheticPolicy.emptyCommitForSampleCollection, 'prohibited', 'empty commit sample policy');

  requireValue(isObject(contract.builderPolicy), 'builderPolicy is required');
  requireBoolean(contract.builderPolicy.immutableDigestRequired, true, 'immutableDigestRequired');
  requireBoolean(contract.builderPolicy.mutableTagOnlyAllowed, false, 'mutableTagOnlyAllowed');
  requireBoolean(contract.builderPolicy.rootPrivilegeAllowed, false, 'rootPrivilegeAllowed');
  requireExact(contract.builderPolicy.externalDownloads, 'must-be-pinned-or-locked', 'externalDownloads');
  requireBoolean(contract.builderPolicy.credentialMaterialInBuild, false, 'credentialMaterialInBuild');

  requireValue(isObject(contract.executionBoundary), 'executionBoundary is required');
  for (const key of [
    'requiredCheckTransition',
    'existingActionsDisabled',
    'deliveryAuthority',
    'publicationAuthority',
    'deploymentAuthority',
    'githubWriteCallback',
    'liveExecutionInThisCheckpoint',
  ]) {
    requireBoolean(contract.executionBoundary[key], false, 'executionBoundary.' + key);
  }

  requireValue(isObject(contract.evidenceContract), 'evidenceContract is required');
  for (const key of [
    'buildIdRequired',
    'triggerIdentityRequired',
    'attemptRequired',
    'validationProfileRequired',
    'timestampsRequired',
    'conclusionRequired',
    'artifactOrLogReferenceRequired',
    'actionsHistoricalSamplesPreserved',
  ]) {
    requireBoolean(contract.evidenceContract[key], true, 'evidenceContract.' + key);
  }
  requireExact(contract.evidenceContract.providerSpecificContext, 'gcp-cloud-build-shadow', 'evidence provider context');
  requireBoolean(contract.evidenceContract.providerSampleMergeAllowed, false, 'evidence provider merge');

  return {
    provider: contract.provider,
    profile: contract.validationProfile,
    actionsSamples: contract.historicalActionsNaturalSamples,
    gcpSamples: contract.currentGcpNaturalSamples,
    syntheticSamples: contract.currentGcpSyntheticSamples,
  };
}

export function validateOptionB(config, contract) {
  const configResult = validateShadowConfig(config);
  const contractResult = validateProviderContract(contract);
  requireExact(configResult.profile, contractResult.profile, 'config/contract profile');
  return { config: configResult, contract: contractResult };
}

export function validateRuntimeIdentity(metadata, contract) {
  requireValue(isObject(metadata), 'runtime identity must be an object');
  requireExact(metadata.profile, contract.validationProfile, 'runtime profile');
  requireValue(typeof metadata.repository === 'string' && /^[^/\s]+\/[^/\s]+$/.test(metadata.repository), 'repository identity is invalid');
  requireValue(typeof metadata.pullRequest === 'string' && /^[1-9][0-9]*$/.test(metadata.pullRequest), 'pull request identity is invalid');
  requireValue(typeof metadata.baseSha === 'string' && SHA40.test(metadata.baseSha), 'base SHA must be 40 lowercase hexadecimal characters');
  requireValue(typeof metadata.sourceSha === 'string' && SHA40.test(metadata.sourceSha), 'source SHA must be 40 lowercase hexadecimal characters');
  requireValue(typeof metadata.buildId === 'string' && metadata.buildId.length > 0, 'build ID is required');
  requireValue(typeof metadata.trigger === 'string' && metadata.trigger.length > 0, 'trigger identity is required');
  requireValue(typeof metadata.attempt === 'string' && metadata.attempt.length > 0, 'attempt identity is required');
  requireExact(metadata.event, 'pull_request', 'event');
  requireValue(!metadata.synthetic, 'synthetic builds cannot be natural samples');
  if (process.env.COMMIT_SHA) {
    requireExact(metadata.sourceSha, process.env.COMMIT_SHA, 'source SHA/build substitution');
  }
  return metadata;
}

export function validateShadowFiles({
  root = process.cwd(),
  configPath = 'cloudbuild/shadow-validation.yaml',
  contractPath = 'verification/provider-shadow-contract.json',
} = {}) {
  validateSafeRelativePath(configPath, 'configPath');
  validateSafeRelativePath(contractPath, 'contractPath');
  const resolvedRoot = resolve(root);
  const resolvedConfig = join(resolvedRoot, configPath);
  const resolvedContract = join(resolvedRoot, contractPath);
  const config = readJson(resolvedConfig, 'shadow config');
  const contract = readJson(resolvedContract, 'provider contract');
  const result = validateOptionB(config, contract);
  return { ...result, root: resolvedRoot, configPath: resolvedConfig, contractPath: resolvedContract };
}

function runNodeScript(scriptPath, args, root) {
  const result = spawnSync(process.execPath, [scriptPath, ...args], {
    cwd: root,
    encoding: 'utf8',
    env: process.env,
  });
  if (result.status !== 0) {
    fail(
      'shared verification planning failed: ' +
        scriptPath +
        '\nstdout:\n' +
        result.stdout +
        '\nstderr:\n' +
        result.stderr,
    );
  }
  return result.stdout;
}

export function runSharedVerification({ root, baseSha, sourceSha }) {
  const scopeScript = join(root, 'scripts', 'detect-verification-scope.mjs');
  const planScript = join(root, 'scripts', 'plan-ci-quick.mjs');
  const verifyScript = join(root, 'scripts', 'verify.sh');
  requireValue(existsSync(scopeScript), 'detect-verification-scope.mjs is missing');
  requireValue(existsSync(planScript), 'plan-ci-quick.mjs is missing');
  requireValue(existsSync(verifyScript), 'scripts/verify.sh is missing');

  const tempRoot = mkdtempSync(join(tmpdir(), 'boilerplate-208-shadow-'));
  try {
    const scopeJson = runNodeScript(
      scopeScript,
      ['--base', baseSha, '--head', sourceSha, '--json', '--workspace', root],
      root,
    );
    const scopePath = join(tempRoot, 'scope.json');
    writeFileSync(scopePath, scopeJson, 'utf8');
    const planJson = runNodeScript(
      planScript,
      ['--scope-json', scopePath, '--profiles', join(root, 'verification', 'profiles.yml')],
      root,
    );
    let plan;
    try {
      plan = JSON.parse(planJson);
    } catch (error) {
      fail('Quick profile planner did not return JSON: ' + error.message);
    }
    requireValue(Array.isArray(plan.include) && plan.include.length > 0, 'Quick profile selected no verification groups');
    const groups = plan.include.map((entry) => entry && entry.id).filter(Boolean);
    requireValue(groups.length === plan.include.length, 'Quick profile returned an invalid group');

    for (const group of groups) {
      const result = spawnSync(
        'bash',
        [verifyScript, '--profile', 'ci-quick', '--base', baseSha, '--head', sourceSha, '--group', group],
        { cwd: root, stdio: 'inherit', env: process.env },
      );
      requireValue(result.status === 0, 'shared verification failed for group: ' + group);
    }
    return { groups };
  } finally {
    rmSync(tempRoot, { recursive: true, force: true });
  }
}

function parseArgs(argv) {
  const valueOptions = new Set([
    '--config',
    '--contract',
    '--profile',
    '--repository',
    '--pull-request',
    '--base-sha',
    '--source-sha',
    '--build-id',
    '--trigger',
    '--attempt',
    '--event',
  ]);
  const result = { execute: false };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === '--execute') {
      requireValue(!result.execute, 'duplicate --execute');
      result.execute = true;
      continue;
    }
    requireValue(valueOptions.has(arg), 'unexpected argument: ' + arg);
    const key = arg.slice(2).replaceAll('-', '');
    requireValue(!Object.hasOwn(result, key), 'duplicate argument: ' + arg);
    requireValue(index + 1 < argv.length && !argv[index + 1].startsWith('--'), 'missing value for ' + arg);
    result[key] = argv[index + 1];
    index += 1;
  }
  return result;
}

export function main(argv = process.argv.slice(2)) {
  const options = parseArgs(argv);
  const root = process.cwd();
  const files = validateShadowFiles({
    root,
    configPath: options.config || 'cloudbuild/shadow-validation.yaml',
    contractPath: options.contract || 'verification/provider-shadow-contract.json',
  });
  requireExact(options.profile || files.contract.profile, files.contract.profile, 'requested profile');

  const runtimeKeys = [
    'repository',
    'pullrequest',
    'basesha',
    'sourcesha',
    'buildid',
    'trigger',
    'attempt',
    'event',
  ];
  const hasRuntime = runtimeKeys.some((key) => Object.hasOwn(options, key));
  if (hasRuntime !== options.execute) {
    fail('runtime identity arguments must be supplied together with --execute');
  }
  if (options.execute) {
    const metadata = validateRuntimeIdentity(
      {
        profile: options.profile || files.contract.profile,
        repository: options.repository,
        pullRequest: options.pullrequest,
        baseSha: options.basesha,
        sourceSha: options.sourcesha,
        buildId: options.buildid,
        trigger: options.trigger,
        attempt: options.attempt,
        event: options.event,
        synthetic: false,
      },
      files.contract,
    );
    const result = runSharedVerification({ root, baseSha: metadata.baseSha, sourceSha: metadata.sourceSha });
    console.log(
      JSON.stringify({
        LOCAL_CONTRACT_STATUS: 'PASS',
        LIVE_GCP_EXECUTION_STATUS: 'NOT_RUN',
        OPERATIONAL_EVIDENCE: 'MISSING',
        provider: files.contract.provider,
        profile: files.contract.profile,
        groups: result.groups,
        sourceSha: metadata.sourceSha,
      }),
    );
    return 0;
  }
  console.log(
    JSON.stringify({
      LOCAL_CONTRACT_STATUS: 'PASS',
      LIVE_GCP_EXECUTION_STATUS: 'NOT_RUN',
      OPERATIONAL_EVIDENCE: 'MISSING',
      provider: files.contract.provider,
      profile: files.contract.profile,
    }),
  );
  return 0;
}

const invokedPath = process.argv[1] ? resolve(process.argv[1]) : '';
if (invokedPath === fileURLToPath(import.meta.url)) {
  try {
    process.exitCode = main();
  } catch (error) {
    console.error('CLOUD_BUILD_SHADOW_VALIDATION_FAILED: ' + error.message);
    process.exitCode = 1;
  }
}

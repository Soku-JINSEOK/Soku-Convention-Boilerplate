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
export const UNATTESTED_CANDIDATE = 'UNATTESTED_CANDIDATE';

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
  'attestationStatus',
  'serverIssuedBuildTriggerId',
  'sourceProvenance',
];

const EXPECTED_STEP_ENV = [
  'CB_COMMIT_SHA=${COMMIT_SHA}',
  'CB_BUILD_ID=${BUILD_ID}',
  'CB_REPO_FULL_NAME=${REPO_FULL_NAME}',
  'CB_PR_NUMBER=${_PR_NUMBER}',
  'CB_HEAD_BRANCH=${_HEAD_BRANCH}',
  'CB_BASE_BRANCH=${_BASE_BRANCH}',
  'CB_HEAD_REPO_URL=${_HEAD_REPO_URL}',
  'CB_BASE_SHA=${_BASE_SHA}',
  'CB_TRIGGER_NAME=${TRIGGER_NAME}',
];

export const CANONICAL_SHADOW_SCRIPT = [
  'set -euo pipefail',
  'require_non_empty() {',
  '  local name="$1"',
  '  local value="$2"',
  '  if [[ -z "$value" || "$value" =~ [[:space:]] ]]; then',
  '    printf "%s must be non-empty and contain no whitespace\\n" "$name" >&2',
  '    exit 2',
  '  fi',
  '}',
  'require_sha() {',
  '  local name="$1"',
  '  local value="$2"',
  '  if [[ ! "$value" =~ ^[0-9a-f]{40}$ ]]; then',
  '    printf "%s must be 40 lowercase hexadecimal characters\\n" "$name" >&2',
  '    exit 2',
  '  fi',
  '}',
  'require_pr() {',
  '  local value="$1"',
  '  if [[ ! "$value" =~ ^[1-9][0-9]*$ ]]; then',
  '    printf "PR_NUMBER must be a positive integer\\n" >&2',
  '    exit 2',
  '  fi',
  '}',
  'require_non_empty "COMMIT_SHA" "$${CB_COMMIT_SHA}"',
  'require_non_empty "BUILD_ID" "$${CB_BUILD_ID}"',
  'require_non_empty "REPO_FULL_NAME" "$${CB_REPO_FULL_NAME}"',
  'require_non_empty "PR_NUMBER" "$${CB_PR_NUMBER}"',
  'require_non_empty "HEAD_BRANCH" "$${CB_HEAD_BRANCH}"',
  'require_non_empty "BASE_BRANCH" "$${CB_BASE_BRANCH}"',
  'require_non_empty "HEAD_REPO_URL" "$${CB_HEAD_REPO_URL}"',
  'require_non_empty "BASE_SHA" "$${CB_BASE_SHA}"',
  'require_non_empty "TRIGGER_NAME" "$${CB_TRIGGER_NAME}"',
  'require_sha "COMMIT_SHA" "$${CB_COMMIT_SHA}"',
  'require_sha "BASE_SHA" "$${CB_BASE_SHA}"',
  'require_pr "$${CB_PR_NUMBER}"',
  'node scripts/verify-cloud-build-shadow.mjs \\',
  '  --config cloudbuild/shadow-validation.yaml \\',
  '  --contract verification/provider-shadow-contract.json \\',
  '  --profile ci-quick \\',
  '  --repository "$${CB_REPO_FULL_NAME}" \\',
  '  --pull-request "$${CB_PR_NUMBER}" \\',
  '  --head-branch "$${CB_HEAD_BRANCH}" \\',
  '  --base-branch "$${CB_BASE_BRANCH}" \\',
  '  --head-repo-url "$${CB_HEAD_REPO_URL}" \\',
  '  --base-sha "$${CB_BASE_SHA}" \\',
  '  --source-sha "$${CB_COMMIT_SHA}" \\',
  '  --build-id "$${CB_BUILD_ID}" \\',
  '  --trigger "$${CB_TRIGGER_NAME}" \\',
  '  --attempt "$${CB_BUILD_ID}" \\',
  '  --event pull_request \\',
  '  --execute',
].join('\n');

const CONFIG_KEYS = ['steps', 'serviceAccount', 'timeout', 'options', 'substitutions'];
const STEP_KEYS = ['id', 'name', 'entrypoint', 'args', 'env'];
const OPTIONS_KEYS = ['logging'];
const SUBSTITUTION_KEYS = ['_BASE_SHA', '_CLOUD_BUILD_SERVICE_ACCOUNT'];
const RUNTIME_IDENTITY_KEYS = [
  'profile',
  'repository',
  'pullRequest',
  'headBranch',
  'baseBranch',
  'headRepoUrl',
  'baseSha',
  'sourceSha',
  'buildId',
  'trigger',
  'attempt',
  'event',
];

const CONTRACT_KEYS = [
  'schemaVersion',
  'contractVersion',
  'provider',
  'validationProfile',
  'validationEntrypoint',
  'historicalActionsNaturalSamples',
  'currentGcpNaturalSamples',
  'currentGcpSyntheticSamples',
  'pilotNaturalTarget',
  'transitionEvaluationNaturalTarget',
  'providerSampleMergeAllowed',
  'eligibleNaturalSampleFields',
  'sourceShaBinding',
  'failureCancellationPolicy',
  'duplicateAttemptPolicy',
  'syntheticPolicy',
  'classificationPolicy',
  'attestationContract',
  'builderPolicy',
  'executionBoundary',
  'evidenceContract',
  'ownerGates',
];

const SOURCE_BINDING_KEYS = [
  'required',
  'cloudBuildSubstitution',
  'baseShaSubstitution',
  'headBranchSubstitution',
  'baseBranchSubstitution',
  'headRepoUrlSubstitution',
  'argument',
  'format',
  'mustMatchBuildSource',
];
const FAILURE_POLICY_KEYS = [
  'denominator',
  'successfulNumerator',
  'failedBuilds',
  'cancelledBuilds',
  'supersededHeads',
];
const DUPLICATE_POLICY_KEYS = ['sameBuildCallbacks', 'sameShaRerun', 'duplicateNaturalSample'];
const SYNTHETIC_POLICY_KEYS = [
  'manualBuildCountsAsNatural',
  'syntheticBuildCountsAsNatural',
  'syntheticSamplesSeparateCounter',
  'emptyCommitForSampleCollection',
];
const CLASSIFICATION_POLICY_KEYS = [
  'buildConfigSelfDeclaredNaturalAllowed',
  'buildConfigOutputClassification',
  'naturalClassificationAuthority',
  'naturalClassificationImplementation',
  'manualBuildCountsAsNatural',
  'unverifiedCandidateCountsAsNatural',
  'sourceGeneratedResultCanAttest',
];
const ATTESTATION_KEYS = [
  'serverIssuedBuildIdRequired',
  'serverIssuedBuildTriggerIdRequired',
  'expectedTriggerResourceIdentityRequired',
  'resolvedRepositoryFromSourceProvenanceRequired',
  'resolvedCommitShaFromSourceProvenanceRequired',
  'expectedRepositoryRequired',
  'expectedPullRequestTriggerRequired',
  'buildStatusConclusionRequired',
  'approvedValidationContractRequired',
  'duplicateBuildAttemptCheckRequired',
  'manualBuildRejected',
  'userSubstitutionsInsufficient',
  'sourceGeneratedResultCannotAttest',
];
const BUILDER_POLICY_KEYS = [
  'immutableDigestRequired',
  'mutableTagOnlyAllowed',
  'rootPrivilegeAllowed',
  'externalDownloads',
  'credentialMaterialInBuild',
];
const EXECUTION_BOUNDARY_KEYS = [
  'requiredCheckTransition',
  'existingActionsDisabled',
  'deliveryAuthority',
  'publicationAuthority',
  'deploymentAuthority',
  'githubWriteCallback',
  'liveExecutionInThisCheckpoint',
];
const EVIDENCE_KEYS = [
  'buildIdRequired',
  'triggerIdentityRequired',
  'attemptRequired',
  'validationProfileRequired',
  'timestampsRequired',
  'conclusionRequired',
  'artifactOrLogReferenceRequired',
  'actionsHistoricalSamplesPreserved',
  'providerSpecificContext',
  'providerSampleMergeAllowed',
  'attestationRequired',
  'operationalEvidenceStatus',
];
const OWNER_GATES = [
  'trusted post-build attestor implementation',
  'live Cloud Build execution',
  'Cloud Build trigger creation',
  'GCP IAM or Secret Manager changes',
  'GitHub required-check transition',
  'delivery or publication',
];

const FORBIDDEN_CONFIG_PATTERNS = [
  /\b(?:docker|podman)\s+push\b/i,
  /\b(?:gcloud\s+run\s+deploy|gcloud\s+functions\s+deploy|kubectl\s+apply)\b/i,
  /\b(?:npm|pnpm)\s+publish\b/i,
  /\b(?:gcloud\s+artifacts|artifactregistry|storage\.googleapis\.com)\b/i,
  /\b(?:artifacts|images|availableSecrets|secretEnv)\b/i,
  /credentials?\.json/i,
  /service-account-key/i,
  /BEGIN\s+(?:RSA|EC|OPENSSH)?\s*PRIVATE\s+KEY/i,
  /\b(?:workflow_dispatch|branch\s+protection|required[-\s]?check)\b/i,
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

function sortedKeys(value) {
  return Object.keys(value).sort();
}

function requireExactKeys(value, expectedKeys, label) {
  requireValue(isObject(value), label + ' must be an object');
  const actual = sortedKeys(value);
  const expected = [...expectedKeys].sort();
  requireValue(JSON.stringify(actual) === JSON.stringify(expected), label + ' has unknown or missing fields');
}

function requireExactArray(actual, expected, label) {
  requireValue(JSON.stringify(actual) === JSON.stringify(expected), label + ' must match the approved structure');
}

class StrictJsonParser {
  constructor(text, label) {
    this.text = text;
    this.label = label;
    this.index = 0;
  }

  error(message) {
    fail(this.label + ' is malformed JSON at offset ' + this.index + ': ' + message);
  }

  skipWhitespace() {
    while (this.index < this.text.length && /[\u0009\u000a\u000d\u0020]/.test(this.text[this.index])) this.index += 1;
  }

  consume(expected) {
    if (this.text[this.index] !== expected) this.error('expected ' + JSON.stringify(expected));
    this.index += 1;
  }

  parse() {
    this.skipWhitespace();
    const value = this.parseValue();
    this.skipWhitespace();
    if (this.index !== this.text.length) this.error('trailing data');
    return value;
  }

  parseValue() {
    this.skipWhitespace();
    const character = this.text[this.index];
    if (character === '{') return this.parseObject();
    if (character === '[') return this.parseArray();
    if (character === '"') return this.parseString();
    if (character === '-' || /[0-9]/.test(character || '')) return this.parseNumber();
    for (const [literal, value] of [
      ['true', true],
      ['false', false],
      ['null', null],
    ]) {
      if (this.text.startsWith(literal, this.index)) {
        this.index += literal.length;
        return value;
      }
    }
    this.error('unexpected token');
  }

  parseObject() {
    this.consume('{');
    this.skipWhitespace();
    const result = {};
    const keys = new Set();
    if (this.text[this.index] === '}') {
      this.index += 1;
      return result;
    }
    while (true) {
      this.skipWhitespace();
      if (this.text[this.index] !== '"') this.error('object keys must be strings');
      const key = this.parseString();
      if (keys.has(key)) this.error('duplicate object key ' + JSON.stringify(key));
      keys.add(key);
      this.skipWhitespace();
      this.consume(':');
      result[key] = this.parseValue();
      this.skipWhitespace();
      if (this.text[this.index] === '}') {
        this.index += 1;
        return result;
      }
      this.consume(',');
    }
  }

  parseArray() {
    this.consume('[');
    this.skipWhitespace();
    const result = [];
    if (this.text[this.index] === ']') {
      this.index += 1;
      return result;
    }
    while (true) {
      result.push(this.parseValue());
      this.skipWhitespace();
      if (this.text[this.index] === ']') {
        this.index += 1;
        return result;
      }
      this.consume(',');
    }
  }

  parseString() {
    const start = this.index;
    this.consume('"');
    let escaped = false;
    while (this.index < this.text.length) {
      const character = this.text[this.index];
      if (escaped) {
        if (character === 'u') {
          const hex = this.text.slice(this.index + 1, this.index + 5);
          if (!/^[0-9a-fA-F]{4}$/.test(hex)) this.error('invalid unicode escape');
          this.index += 5;
          escaped = false;
          continue;
        }
        if (!/["\\/bfnrt]/.test(character)) this.error('invalid string escape');
        this.index += 1;
        escaped = false;
        continue;
      }
      if (character === '\\') {
        escaped = true;
        this.index += 1;
        continue;
      }
      if (character === '"') {
        this.index += 1;
        try {
          return JSON.parse(this.text.slice(start, this.index));
        } catch (error) {
          this.error('invalid string: ' + error.message);
        }
      }
      if (character < ' ') this.error('control character in string');
      this.index += 1;
    }
    this.error('unterminated string');
  }

  parseNumber() {
    const match = this.text.slice(this.index).match(/^-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?/);
    if (!match) this.error('invalid number');
    this.index += match[0].length;
    const value = Number(match[0]);
    if (!Number.isFinite(value)) this.error('number is not finite');
    return value;
  }
}

export function parseStrictJson(text, label = 'JSON') {
  requireValue(typeof text === 'string', label + ' input must be text');
  return new StrictJsonParser(text, label).parse();
}

function readJson(filePath, label) {
  requireValue(existsSync(filePath), label + ' does not exist');
  try {
    return parseStrictJson(readFileSync(filePath, 'utf8'), label);
  } catch (error) {
    fail(error.message);
  }
}

function validateSafeRelativePath(value, label) {
  requireValue(typeof value === 'string' && value.length > 0, label + ' is required');
  requireValue(!isAbsolute(value), label + ' must be relative');
  requireValue(!value.split('/').includes('..'), label + ' must not traverse');
}

function validateNoWhitespaceString(value, label) {
  requireValue(typeof value === 'string' && value.length > 0 && !/[\s]/.test(value), label + ' must be non-empty and contain no whitespace');
}

function validateContractObjectKeys(contract) {
  requireExactKeys(contract, CONTRACT_KEYS, 'provider contract');
  requireExactKeys(contract.sourceShaBinding, SOURCE_BINDING_KEYS, 'sourceShaBinding');
  requireExactKeys(contract.failureCancellationPolicy, FAILURE_POLICY_KEYS, 'failureCancellationPolicy');
  requireExactKeys(contract.duplicateAttemptPolicy, DUPLICATE_POLICY_KEYS, 'duplicateAttemptPolicy');
  requireExactKeys(contract.syntheticPolicy, SYNTHETIC_POLICY_KEYS, 'syntheticPolicy');
  requireExactKeys(contract.classificationPolicy, CLASSIFICATION_POLICY_KEYS, 'classificationPolicy');
  requireExactKeys(contract.attestationContract, ATTESTATION_KEYS, 'attestationContract');
  requireExactKeys(contract.builderPolicy, BUILDER_POLICY_KEYS, 'builderPolicy');
  requireExactKeys(contract.executionBoundary, EXECUTION_BOUNDARY_KEYS, 'executionBoundary');
  requireExactKeys(contract.evidenceContract, EVIDENCE_KEYS, 'evidenceContract');
  requireExactArray(contract.ownerGates, OWNER_GATES, 'ownerGates');
}

export function validateShadowConfig(config) {
  requireExactKeys(config, CONFIG_KEYS, 'shadow config');
  requireValue(Array.isArray(config.steps) && config.steps.length === 1, 'shadow config must have exactly one step');

  const step = config.steps[0];
  requireExactKeys(step, STEP_KEYS, 'shadow step');
  requireExact(step.id, 'gcp-shadow-quick', 'shadow step id');
  requireExact(step.entrypoint, 'bash', 'shadow step entrypoint');
  requireValue(typeof step.name === 'string' && IMAGE_DIGEST.test(step.name), 'builder must use an immutable sha256 digest');
  requireExact(step.name, APPROVED_BUILDER, 'builder identity');
  requireValue(!/@sha256:.*latest/i.test(step.name) && !/:latest@/i.test(step.name), 'latest builder reference is prohibited');
  requireExactArray(step.args, ['-ceu', CANONICAL_SHADOW_SCRIPT], 'shadow step args');
  requireExactArray(step.env, EXPECTED_STEP_ENV, 'shadow step env');

  const serialized = JSON.stringify(config) + '\n' + CANONICAL_SHADOW_SCRIPT;
  for (const pattern of FORBIDDEN_CONFIG_PATTERNS) {
    requireValue(!pattern.test(serialized), 'forbidden deployment, credential, artifact, or control-plane content detected');
  }
  requireExact(config.serviceAccount, '${_CLOUD_BUILD_SERVICE_ACCOUNT}', 'serviceAccount');
  requireExact(config.timeout, '900s', 'build timeout');
  requireExactKeys(config.options, OPTIONS_KEYS, 'build options');
  requireExact(config.options.logging, 'CLOUD_LOGGING_ONLY', 'build logging mode');
  requireExactKeys(config.substitutions, SUBSTITUTION_KEYS, 'substitutions');
  requireExact(config.substitutions._BASE_SHA, '', '_BASE_SHA default');
  requireExact(config.substitutions._CLOUD_BUILD_SERVICE_ACCOUNT, '', '_CLOUD_BUILD_SERVICE_ACCOUNT default');

  return {
    builder: step.name,
    profile: 'ci-quick',
    timeout: config.timeout,
    format: 'json-compatible-yaml',
    classification: UNATTESTED_CANDIDATE,
  };
}

export function validateProviderContract(contract) {
  requireValue(isObject(contract), 'provider contract must be an object');
  validateContractObjectKeys(contract);
  requireInteger(contract.schemaVersion, 2, 'schemaVersion');
  requireExact(contract.contractVersion, 'option-b-v2', 'contractVersion');
  requireExact(contract.provider, 'gcp-cloud-build-shadow', 'provider');
  requireExact(contract.validationProfile, 'ci-quick', 'validationProfile');
  requireExact(contract.validationEntrypoint, 'scripts/verify.sh', 'validationEntrypoint');
  requireInteger(contract.historicalActionsNaturalSamples, 3, 'historicalActionsNaturalSamples');
  requireInteger(contract.currentGcpNaturalSamples, 0, 'currentGcpNaturalSamples');
  requireInteger(contract.currentGcpSyntheticSamples, 0, 'currentGcpSyntheticSamples');
  requireInteger(contract.pilotNaturalTarget, 3, 'pilotNaturalTarget');
  requireInteger(contract.transitionEvaluationNaturalTarget, 10, 'transitionEvaluationNaturalTarget');
  requireBoolean(contract.providerSampleMergeAllowed, false, 'providerSampleMergeAllowed');
  requireExactArray(contract.eligibleNaturalSampleFields, REQUIRED_SAMPLE_FIELDS, 'eligibleNaturalSampleFields');

  requireBoolean(contract.sourceShaBinding.required, true, 'sourceShaBinding.required');
  requireExact(contract.sourceShaBinding.cloudBuildSubstitution, 'COMMIT_SHA', 'sourceShaBinding.cloudBuildSubstitution');
  requireExact(contract.sourceShaBinding.baseShaSubstitution, '_BASE_SHA', 'sourceShaBinding.baseShaSubstitution');
  requireExact(contract.sourceShaBinding.headBranchSubstitution, '_HEAD_BRANCH', 'sourceShaBinding.headBranchSubstitution');
  requireExact(contract.sourceShaBinding.baseBranchSubstitution, '_BASE_BRANCH', 'sourceShaBinding.baseBranchSubstitution');
  requireExact(contract.sourceShaBinding.headRepoUrlSubstitution, '_HEAD_REPO_URL', 'sourceShaBinding.headRepoUrlSubstitution');
  requireExact(contract.sourceShaBinding.argument, '--source-sha', 'sourceShaBinding.argument');
  requireExact(contract.sourceShaBinding.format, '40-lower-hex', 'sourceShaBinding.format');
  requireBoolean(contract.sourceShaBinding.mustMatchBuildSource, true, 'sourceShaBinding.mustMatchBuildSource');

  requireExact(contract.failureCancellationPolicy.denominator, 'included', 'failure denominator');
  requireExact(contract.failureCancellationPolicy.successfulNumerator, 'success-only', 'success numerator');
  requireExact(contract.failureCancellationPolicy.failedBuilds, 'included-and-classified', 'failed build handling');
  requireExact(contract.failureCancellationPolicy.cancelledBuilds, 'included-and-classified', 'cancelled build handling');
  requireExact(contract.failureCancellationPolicy.supersededHeads, 'included-and-classified', 'superseded head handling');

  requireExact(contract.duplicateAttemptPolicy.sameBuildCallbacks, 'count-once', 'same-build callback handling');
  requireExact(contract.duplicateAttemptPolicy.sameShaRerun, 'new-attempt-not-new-natural-sample', 'rerun handling');
  requireExact(contract.duplicateAttemptPolicy.duplicateNaturalSample, 'excluded', 'duplicate sample handling');

  requireBoolean(contract.syntheticPolicy.manualBuildCountsAsNatural, false, 'manualBuildCountsAsNatural');
  requireBoolean(contract.syntheticPolicy.syntheticBuildCountsAsNatural, false, 'syntheticBuildCountsAsNatural');
  requireExact(contract.syntheticPolicy.syntheticSamplesSeparateCounter, 'currentGcpSyntheticSamples', 'synthetic counter');
  requireExact(contract.syntheticPolicy.emptyCommitForSampleCollection, 'prohibited', 'empty commit sample policy');

  requireBoolean(contract.classificationPolicy.buildConfigSelfDeclaredNaturalAllowed, false, 'buildConfigSelfDeclaredNaturalAllowed');
  requireExact(contract.classificationPolicy.buildConfigOutputClassification, UNATTESTED_CANDIDATE, 'buildConfigOutputClassification');
  requireExact(contract.classificationPolicy.naturalClassificationAuthority, 'trusted-post-build-attestor', 'naturalClassificationAuthority');
  requireExact(contract.classificationPolicy.naturalClassificationImplementation, 'not-implemented', 'naturalClassificationImplementation');
  requireBoolean(contract.classificationPolicy.manualBuildCountsAsNatural, false, 'classification manualBuildCountsAsNatural');
  requireBoolean(contract.classificationPolicy.unverifiedCandidateCountsAsNatural, false, 'unverifiedCandidateCountsAsNatural');
  requireBoolean(contract.classificationPolicy.sourceGeneratedResultCanAttest, false, 'sourceGeneratedResultCanAttest');

  for (const key of ATTESTATION_KEYS) requireBoolean(contract.attestationContract[key], true, 'attestationContract.' + key);

  requireBoolean(contract.builderPolicy.immutableDigestRequired, true, 'immutableDigestRequired');
  requireBoolean(contract.builderPolicy.mutableTagOnlyAllowed, false, 'mutableTagOnlyAllowed');
  requireBoolean(contract.builderPolicy.rootPrivilegeAllowed, false, 'rootPrivilegeAllowed');
  requireExact(contract.builderPolicy.externalDownloads, 'must-be-pinned-or-locked', 'externalDownloads');
  requireBoolean(contract.builderPolicy.credentialMaterialInBuild, false, 'credentialMaterialInBuild');

  for (const key of EXECUTION_BOUNDARY_KEYS) requireBoolean(contract.executionBoundary[key], false, 'executionBoundary.' + key);

  for (const key of [
    'buildIdRequired',
    'triggerIdentityRequired',
    'attemptRequired',
    'validationProfileRequired',
    'timestampsRequired',
    'conclusionRequired',
    'artifactOrLogReferenceRequired',
    'actionsHistoricalSamplesPreserved',
    'attestationRequired',
  ]) requireBoolean(contract.evidenceContract[key], true, 'evidenceContract.' + key);
  requireExact(contract.evidenceContract.providerSpecificContext, 'gcp-cloud-build-shadow', 'evidence provider context');
  requireBoolean(contract.evidenceContract.providerSampleMergeAllowed, false, 'evidence provider merge');
  requireExact(contract.evidenceContract.operationalEvidenceStatus, 'MISSING', 'operational evidence status');

  return {
    provider: contract.provider,
    profile: contract.validationProfile,
    actionsSamples: contract.historicalActionsNaturalSamples,
    gcpSamples: contract.currentGcpNaturalSamples,
    syntheticSamples: contract.currentGcpSyntheticSamples,
    classification: UNATTESTED_CANDIDATE,
  };
}

export function validateOptionB(config, contract) {
  const configResult = validateShadowConfig(config);
  const contractResult = validateProviderContract(contract);
  requireExact(configResult.profile, contractResult.profile, 'config/contract profile');
  return { config: configResult, contract: contractResult };
}

function requireRuntimeString(value, label) {
  validateNoWhitespaceString(value, label);
}

export function validateRuntimeIdentity(metadata, contract) {
  requireExactKeys(metadata, RUNTIME_IDENTITY_KEYS, 'runtime identity');
  requireExact(metadata.profile, contract.validationProfile, 'runtime profile');
  requireValue(typeof metadata.repository === 'string' && /^[^/\s]+\/[^/\s]+$/.test(metadata.repository), 'repository identity is invalid');
  requireValue(typeof metadata.pullRequest === 'string' && /^[1-9][0-9]*$/.test(metadata.pullRequest), 'pull request identity is invalid');
  requireRuntimeString(metadata.headBranch, 'head branch');
  requireRuntimeString(metadata.baseBranch, 'base branch');
  requireValue(
    typeof metadata.headRepoUrl === 'string' && /^https:\/\/github\.com\/[^/\s]+\/[^/\s]+(?:\.git)?$/.test(metadata.headRepoUrl),
    'head repository URL is invalid',
  );
  requireValue(typeof metadata.baseSha === 'string' && SHA40.test(metadata.baseSha), 'base SHA must be 40 lowercase hexadecimal characters');
  requireValue(typeof metadata.sourceSha === 'string' && SHA40.test(metadata.sourceSha), 'source SHA must be 40 lowercase hexadecimal characters');
  requireRuntimeString(metadata.buildId, 'build ID');
  requireRuntimeString(metadata.trigger, 'trigger identity');
  requireRuntimeString(metadata.attempt, 'attempt identity');
  requireExact(metadata.event, 'pull_request', 'event');
  if (Object.hasOwn(process.env, 'COMMIT_SHA')) {
    requireValue(typeof process.env.COMMIT_SHA === 'string' && SHA40.test(process.env.COMMIT_SHA), 'COMMIT_SHA environment value is invalid');
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
    fail('shared verification planning failed: ' + scriptPath + '\nstdout:\n' + result.stdout + '\nstderr:\n' + result.stderr);
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
    const scopeJson = runNodeScript(scopeScript, ['--base', baseSha, '--head', sourceSha, '--json', '--workspace', root], root);
    const scopePath = join(tempRoot, 'scope.json');
    writeFileSync(scopePath, scopeJson, 'utf8');
    const planJson = runNodeScript(planScript, ['--scope-json', scopePath, '--profiles', join(root, 'verification', 'profiles.yml')], root);
    let plan;
    try {
      plan = parseStrictJson(planJson, 'Quick profile planner output');
    } catch (error) {
      fail('Quick profile planner did not return JSON: ' + error.message);
    }
    requireValue(Array.isArray(plan.include) && plan.include.length > 0, 'Quick profile selected no verification groups');
    const groups = plan.include.map((entry) => entry && entry.id).filter(Boolean);
    requireValue(groups.length === plan.include.length, 'Quick profile returned an invalid group');

    for (const group of groups) {
      const result = spawnSync('bash', [verifyScript, '--profile', 'ci-quick', '--base', baseSha, '--head', sourceSha, '--group', group], {
        cwd: root,
        stdio: 'inherit',
        env: process.env,
      });
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
    '--head-branch',
    '--base-branch',
    '--head-repo-url',
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
    'headbranch',
    'basebranch',
    'headrepourl',
    'basesha',
    'sourcesha',
    'buildid',
    'trigger',
    'attempt',
    'event',
  ];
  const hasRuntime = runtimeKeys.some((key) => Object.hasOwn(options, key));
  if (hasRuntime !== options.execute) fail('runtime identity arguments must be supplied together with --execute');
  if (options.execute) {
    const metadata = validateRuntimeIdentity(
      {
        profile: options.profile || files.contract.profile,
        repository: options.repository,
        pullRequest: options.pullrequest,
        headBranch: options.headbranch,
        baseBranch: options.basebranch,
        headRepoUrl: options.headrepourl,
        baseSha: options.basesha,
        sourceSha: options.sourcesha,
        buildId: options.buildid,
        trigger: options.trigger,
        attempt: options.attempt,
        event: options.event,
      },
      files.contract,
    );
    const result = runSharedVerification({ root, baseSha: metadata.baseSha, sourceSha: metadata.sourceSha });
    console.log(
      JSON.stringify({
        LOCAL_CONTRACT_STATUS: 'PASS',
        LIVE_GCP_EXECUTION_STATUS: 'NOT_RUN',
        OPERATIONAL_EVIDENCE: 'MISSING',
        BUILD_CONFIG_OUTPUT_CLASSIFICATION: UNATTESTED_CANDIDATE,
        NATURAL_CLASSIFICATION_IMPLEMENTATION: 'NOT_IMPLEMENTED',
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
      BUILD_CONFIG_OUTPUT_CLASSIFICATION: UNATTESTED_CANDIDATE,
      NATURAL_CLASSIFICATION_IMPLEMENTATION: 'NOT_IMPLEMENTED',
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

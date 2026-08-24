import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { test } from 'node:test';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import {
  CANONICAL_SHADOW_SCRIPT,
  MAX_JSON_INPUT_BYTES,
  MAX_JSON_NESTING_DEPTH,
  UNATTESTED_CANDIDATE,
  parseStrictJson,
  scanJsonResourceLimits,
  validateOptionB,
  validateProviderContract,
  validateRuntimeIdentity,
  validateShadowConfig,
  validateShadowFiles,
} from './verify-cloud-build-shadow.mjs';

const root = process.cwd();
const validConfig = parseStrictJson(readFileSync(join(root, 'cloudbuild', 'shadow-validation.yaml'), 'utf8'), 'test config');
const validContract = parseStrictJson(readFileSync(join(root, 'verification', 'provider-shadow-contract.json'), 'utf8'), 'test contract');

function clone(value) {
  return structuredClone(value);
}

function expectFailure(mutator, pattern = /./) {
  const config = clone(validConfig);
  const contract = clone(validContract);
  mutator(config, contract);
  assert.throws(() => validateOptionB(config, contract), pattern);
}

function validRuntime(overrides = {}) {
  return {
    profile: 'ci-quick',
    repository: 'Soku-JINSEOK/Soku-Convention-Boilerplate',
    pullRequest: '208',
    headBranch: 'feature/shadow',
    baseBranch: 'main',
    headRepoUrl: 'https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate.git',
    baseSha: '76c739557d3919eb965d4de2792df1ee1ed2665f',
    sourceSha: '99ae819e5641b1fc7a585aa723c67883eef968f6',
    buildId: 'build-208',
    trigger: 'projects/p/locations/global/triggers/shadow',
    attempt: 'build-208',
    event: 'pull_request',
    ...overrides,
  };
}

function runGuard(overrides = {}) {
  const script = CANONICAL_SHADOW_SCRIPT.replaceAll('$${', '${');
  const environment = {
    ...process.env,
    CB_COMMIT_SHA: '99ae819e5641b1fc7a585aa723c67883eef968f6',
    CB_BUILD_ID: 'build-208',
    CB_REPO_FULL_NAME: 'Soku-JINSEOK/Soku-Convention-Boilerplate',
    CB_PR_NUMBER: '208',
    CB_HEAD_BRANCH: 'feature/shadow',
    CB_BASE_BRANCH: 'main',
    CB_HEAD_REPO_URL: 'https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate.git',
    CB_BASE_SHA: '76c739557d3919eb965d4de2792df1ee1ed2665f',
    CB_TRIGGER_NAME: 'shadow-trigger',
    ...overrides,
  };
  return spawnSync('bash', ['-ceu', script], { cwd: root, env: environment, encoding: 'utf8' });
}

function nestedJson(depth) {
  return '['.repeat(depth) + 'null' + ']'.repeat(depth);
}

function mixedNestedJson(depth) {
  let value = 'null';
  for (let index = depth - 1; index >= 0; index -= 1) {
    value = index % 2 === 0 ? '{"value":' + value + '}' : '[' + value + ']';
  }
  return value;
}

function withJsonFixture(configText, contractText, callback) {
  const fixtureRoot = mkdtempSync(join(tmpdir(), 'boilerplate-208-json-boundary-'));
  try {
    writeFileSync(join(fixtureRoot, 'config.json'), configText, 'utf8');
    writeFileSync(join(fixtureRoot, 'contract.json'), contractText, 'utf8');
    return callback(fixtureRoot);
  } finally {
    rmSync(fixtureRoot, { recursive: true, force: true });
  }
}

function runCliFixture(configText, contractText) {
  return withJsonFixture(configText, contractText, (fixtureRoot) =>
    spawnSync(
      process.execPath,
      [join(root, 'scripts', 'verify-cloud-build-shadow.mjs'), '--config', 'config.json', '--contract', 'contract.json'],
      { cwd: fixtureRoot, env: process.env, encoding: 'utf8' },
    ),
  );
}

function assertControlledFailure(result, expectedMessage) {
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, expectedMessage);
  assert.doesNotMatch(result.stderr, /RangeError|Maximum call stack size exceeded|\n\s+at /);
}

// JSON resource-boundary regression matrix.
test('parser resource policy is explicit and exact', () => {
  assert.deepEqual(validContract.parserResourcePolicy, {
    maximumInputBytes: MAX_JSON_INPUT_BYTES,
    maximumNestingDepth: MAX_JSON_NESTING_DEPTH,
    enforcementPhase: 'pre-parse',
    depthAlgorithm: 'iterative',
    overLimitBehavior: 'fail-closed',
    stackTraceExposure: false,
    appliesTo: ['shadow config', 'provider contract'],
  });
});

test('depth zero and shallow JSON pass', () => {
  assert.deepEqual(scanJsonResourceLimits('null', 'depth zero'), { byteLength: 4, maximumDepth: 0 });
  assert.deepEqual(parseStrictJson('{"value":[1,2,3]}', 'shallow'), { value: [1, 2, 3] });
});

test('depth 63 and 64 pass', () => {
  assert.equal(scanJsonResourceLimits(nestedJson(63), 'depth 63').maximumDepth, 63);
  assert.equal(scanJsonResourceLimits(nestedJson(64), 'depth 64').maximumDepth, 64);
  assert.doesNotThrow(() => parseStrictJson(nestedJson(64), 'depth 64'));
});

test('depth 65, 1000, and 5000 fail before recursive parsing', () => {
  for (const depth of [65, 1000, 5000]) {
    assert.throws(() => parseStrictJson(nestedJson(depth), 'depth ' + depth), /maximum nesting depth of 64/);
  }
});

test('mixed object and array depth 64 passes and depth 65 fails', () => {
  assert.doesNotThrow(() => parseStrictJson(mixedNestedJson(64), 'mixed depth 64'));
  assert.throws(() => parseStrictJson(mixedNestedJson(65), 'mixed depth 65'), /maximum nesting depth of 64/);
});

test('brackets inside strings do not increase depth', () => {
  const value = '{'.repeat(5000) + '['.repeat(5000) + ']'.repeat(5000) + '}'.repeat(5000);
  assert.equal(scanJsonResourceLimits(JSON.stringify(value), 'string brackets').maximumDepth, 0);
  assert.equal(parseStrictJson(JSON.stringify(value), 'string brackets'), value);
});

test('escaped quotes, backslashes, and unicode escape text remain string content', () => {
  const value = String.raw`escaped quote: \" and slash: \\ and unicode text: \\u005b`;
  const encoded = JSON.stringify(value);
  assert.equal(scanJsonResourceLimits(encoded, 'escaped string').maximumDepth, 0);
  assert.equal(parseStrictJson(encoded, 'escaped string'), value);
});

test('malformed delimiter and string structures fail closed', () => {
  assert.throws(() => parseStrictJson(']', 'underflow'), /unmatched closing delimiter/);
  assert.throws(() => parseStrictJson('{]', 'mismatch'), /mismatched closing delimiter/);
  assert.throws(() => parseStrictJson('{', 'unclosed object'), /unclosed delimiter/);
  assert.throws(() => parseStrictJson('[', 'unclosed array'), /unclosed delimiter/);
  assert.throws(() => parseStrictJson('"unterminated', 'unterminated string'), /unterminated string/);
  assert.throws(() => parseStrictJson('{"value":1} trailing', 'trailing data'), /trailing data/);
});

test('input size below and exactly one MiB passes', () => {
  const below = JSON.stringify('a'.repeat(1024));
  assert.ok(Buffer.byteLength(below, 'utf8') < MAX_JSON_INPUT_BYTES);
  assert.doesNotThrow(() => parseStrictJson(below, 'below size'));

  const prefix = '{"value":"';
  const suffix = '"}';
  const exact = prefix + 'a'.repeat(MAX_JSON_INPUT_BYTES - Buffer.byteLength(prefix + suffix, 'utf8')) + suffix;
  assert.equal(Buffer.byteLength(exact, 'utf8'), MAX_JSON_INPUT_BYTES);
  assert.doesNotThrow(() => parseStrictJson(exact, 'exact size'));
});

test('input size one byte over the limit fails before depth or duplicate checks', () => {
  const prefix = '{"value":"';
  const suffix = '"}';
  const oversized = prefix + 'a'.repeat(MAX_JSON_INPUT_BYTES + 1 - Buffer.byteLength(prefix + suffix, 'utf8')) + suffix;
  assert.throws(() => parseStrictJson(oversized, 'oversized'), /exceeds maximum size of 1048576 bytes/);
});

test('multibyte UTF-8 size is measured in bytes rather than characters', () => {
  const multibyte = JSON.stringify('é'.repeat(Math.floor(MAX_JSON_INPUT_BYTES / 2)));
  assert.ok(multibyte.length < MAX_JSON_INPUT_BYTES);
  assert.ok(Buffer.byteLength(multibyte, 'utf8') > MAX_JSON_INPUT_BYTES);
  assert.throws(() => parseStrictJson(multibyte, 'multibyte'), /exceeds maximum size of 1048576 bytes/);
});

test('config file resource limits run before duplicate and semantic validation', () => {
  const validContractText = JSON.stringify(validContract);
  withJsonFixture(nestedJson(5000), validContractText, (fixtureRoot) => {
    assert.throws(
      () => validateShadowFiles({ root: fixtureRoot, configPath: 'config.json', contractPath: 'contract.json' }),
      /shadow config exceeds maximum nesting depth of 64/,
    );
  });
});

test('provider contract file resource limits run before semantic validation', () => {
  const validConfigText = readFileSync(join(root, 'cloudbuild', 'shadow-validation.yaml'), 'utf8');
  withJsonFixture(validConfigText, nestedJson(5000), (fixtureRoot) => {
    assert.throws(
      () => validateShadowFiles({ root: fixtureRoot, configPath: 'config.json', contractPath: 'contract.json' }),
      /provider contract exceeds maximum nesting depth of 64/,
    );
  });
});

test('CLI excessive depth exits cleanly without stack trace', () => {
  const result = runCliFixture(nestedJson(5000), JSON.stringify(validContract));
  assertControlledFailure(result, /shadow config exceeds maximum nesting depth of 64/);
});

test('provider contract parser rejects unknown resource policy fields', () => {
  const contract = clone(validContract);
  contract.parserResourcePolicy.futureUnknown = true;
  assert.throws(() => validateProviderContract(contract), /parserResourcePolicy has unknown or missing fields/);
});

// Existing Option B regression matrix.
test('valid Option B contract passes', () => {
  const result = validateShadowFiles({ root });
  assert.equal(result.config.profile, 'ci-quick');
  assert.equal(result.contract.actionsSamples, 3);
  assert.equal(result.contract.gcpSamples, 0);
  assert.equal(result.config.classification, UNATTESTED_CANDIDATE);
});

test('runtime identity accepts the validated contract summary used by the CLI', () => {
  const result = validateShadowFiles({ root });
  assert.doesNotThrow(() => validateRuntimeIdentity(validRuntime(), result.contract));
});

test('mutable builder reference fails closed', () => {
  expectFailure((config) => {
    config.steps[0].name = 'node:24.17.0-bookworm';
  }, /immutable sha256 digest/);
});

test('invalid builder digest fails closed', () => {
  expectFailure((config) => {
    config.steps[0].name = 'node:24.17.0-bookworm@sha256:bad';
  }, /immutable sha256 digest/);
});

test('missing source SHA binding fails closed', () => {
  expectFailure((config) => {
    config.steps[0].env[0] = 'CB_COMMIT_SHA=';
  }, /shadow step env must match/);
});

test('source SHA bound to a different substitution fails closed', () => {
  expectFailure((config) => {
    config.steps[0].env[0] = 'CB_COMMIT_SHA=${BUILD_ID}';
  }, /shadow step env must match/);
});

test('invalid provider fails closed', () => {
  expectFailure((_config, contract) => {
    contract.provider = 'github-actions';
  }, /provider must equal/);
});

test('Actions and GCP sample merge enabled fails closed', () => {
  expectFailure((_config, contract) => {
    contract.providerSampleMergeAllowed = true;
  }, /providerSampleMergeAllowed must be false/);
});

test('pilot target other than three fails closed', () => {
  expectFailure((_config, contract) => {
    contract.pilotNaturalTarget = 4;
  }, /pilotNaturalTarget must equal 3/);
});

test('transition target other than ten fails closed', () => {
  expectFailure((_config, contract) => {
    contract.transitionEvaluationNaturalTarget = 11;
  }, /transitionEvaluationNaturalTarget must equal 10/);
});

test('synthetic samples cannot count as natural', () => {
  expectFailure((_config, contract) => {
    contract.syntheticPolicy.syntheticBuildCountsAsNatural = true;
  }, /syntheticBuildCountsAsNatural must be false/);
});

test('deployment command fails closed', () => {
  expectFailure((config) => {
    config.steps[0].args[1] += '\ngcloud run deploy shadow';
  }, /shadow step args must match|forbidden/);
});

test('registry push or publish command fails closed', () => {
  expectFailure((config) => {
    config.steps[0].args[1] += '\ndocker push example.invalid/image';
  }, /shadow step args must match|forbidden/);
});

test('inline secret or key material fails closed', () => {
  expectFailure((config) => {
    config.steps[0].args[1] += '\ncat credentials.json';
  }, /shadow step args must match|forbidden/);
});

test('missing timeout fails closed', () => {
  expectFailure((config) => {
    delete config.timeout;
  }, /shadow config has unknown or missing fields/);
});

test('Quick and Full duplicate execution fails closed', () => {
  expectFailure((config) => {
    config.steps[0].args[1] += '\n--profile full';
  }, /shadow step args must match|forbidden/);
});

test('missing existing validation entrypoint fails closed', () => {
  expectFailure((_config, contract) => {
    contract.validationEntrypoint = 'scripts/not-verify.sh';
  }, /validationEntrypoint must equal/);
});

test('required-check transition remains rejected by contract', () => {
  expectFailure((_config, contract) => {
    contract.executionBoundary.requiredCheckTransition = true;
  }, /executionBoundary.requiredCheckTransition must be false/);
});

test('missing source binding contract fails closed', () => {
  const contract = clone(validContract);
  delete contract.sourceShaBinding;
  assert.throws(() => validateProviderContract(contract), /provider contract has unknown or missing fields/);
});

test('config without validator execution fails closed', () => {
  const config = clone(validConfig);
  config.steps[0].args[1] = config.steps[0].args[1].replace('node scripts/verify-cloud-build-shadow.mjs', 'node other-validator.mjs');
  assert.throws(() => validateShadowConfig(config), /shadow step args must match/);
});

// Strict fail-closed correction regression matrix.
test('build config cannot self-declare natural', () => {
  assert.equal(Object.hasOwn(validConfig, 'natural'), false);
  assert.equal(Object.hasOwn(validConfig.steps[0], 'synthetic'), false);
  assert.equal(validContract.classificationPolicy.buildConfigSelfDeclaredNaturalAllowed, false);
  assert.equal(validContract.classificationPolicy.buildConfigOutputClassification, UNATTESTED_CANDIDATE);
});

test('synthetic false field is rejected', () => {
  expectFailure((config) => {
    config.synthetic = false;
  }, /shadow config has unknown or missing fields/);
});

test('natural classification field is rejected', () => {
  expectFailure((config) => {
    config.classification = 'natural';
  }, /shadow config has unknown or missing fields/);
});

test('unattested candidate is the only local classification', () => {
  const result = validateOptionB(clone(validConfig), clone(validContract));
  assert.equal(result.config.classification, UNATTESTED_CANDIDATE);
  assert.equal(result.contract.classification, UNATTESTED_CANDIDATE);
  assert.equal(validContract.classificationPolicy.naturalClassificationImplementation, 'not-implemented');
});

test('runtime identity rejects self-attestation fields', () => {
  assert.throws(() => validateRuntimeIdentity({ ...validRuntime(), synthetic: false }, validContract), /runtime identity has unknown or missing fields/);
  assert.throws(() => validateRuntimeIdentity({ ...validRuntime(), classification: 'natural' }, validContract), /runtime identity has unknown or missing fields/);
});

test('runtime empty COMMIT_SHA is rejected before validator execution', () => {
  const result = runGuard({ CB_COMMIT_SHA: '' });
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /COMMIT_SHA must be non-empty/);
});

test('runtime whitespace-only COMMIT_SHA is rejected', () => {
  const result = runGuard({ CB_COMMIT_SHA: '   ' });
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /COMMIT_SHA must be non-empty/);
});

test('runtime empty PR number is rejected', () => {
  const result = runGuard({ CB_PR_NUMBER: '' });
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /PR_NUMBER must be non-empty/);
});

test('runtime missing head/base/repository identity is rejected', () => {
  for (const key of ['CB_HEAD_BRANCH', 'CB_BASE_BRANCH', 'CB_HEAD_REPO_URL']) {
    const result = runGuard({ [key]: '' });
    assert.notEqual(result.status, 0, key);
  }
});

test('runtime guard removed fails static validation', () => {
  expectFailure((config) => {
    config.steps[0].args[1] = config.steps[0].args[1].replace('require_non_empty "BASE_SHA" "$${CB_BASE_SHA}"\n', '');
  }, /shadow step args must match/);
});

test('runtime guard partially weakened fails static validation', () => {
  expectFailure((config) => {
    config.steps[0].args[1] = config.steps[0].args[1].replace('|| "$value" =~ [[:space:]]', '');
  }, /shadow step args must match/);
});

test('runtime guard cannot be made ineffective by accepting a default', () => {
  expectFailure((config) => {
    config.steps[0].args[1] = config.steps[0].args[1].replace('local value="$2"', 'local value="${2:-default}"');
  }, /shadow step args must match/);
});

test('runtime invalid SHA and PR formats are rejected', () => {
  assert.throws(() => validateRuntimeIdentity(validRuntime({ sourceSha: 'not-a-sha' }), validContract), /source SHA/);
  assert.throws(() => validateRuntimeIdentity(validRuntime({ pullRequest: '0' }), validContract), /pull request/);
});

test('ALLOW_LOOSE cannot bypass runtime guard', () => {
  const result = runGuard({ CB_COMMIT_SHA: '', ALLOW_LOOSE: 'true' });
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /COMMIT_SHA must be non-empty/);
});

test('artifacts field is rejected', () => {
  expectFailure((config) => {
    config.artifacts = {};
  }, /shadow config has unknown or missing fields/);
});

test('images field is rejected', () => {
  expectFailure((config) => {
    config.images = ['example.invalid/image'];
  }, /shadow config has unknown or missing fields/);
});

test('availableSecrets field is rejected', () => {
  expectFailure((config) => {
    config.availableSecrets = {};
  }, /shadow config has unknown or missing fields/);
});

test('secretEnv and credential environment are rejected', () => {
  expectFailure((config) => {
    config.steps[0].env = [...config.steps[0].env, 'SERVICE_ACCOUNT_KEY_JSON=x'];
  }, /shadow step env must match/);
});

test('nested docker push is rejected structurally', () => {
  expectFailure((config) => {
    config.steps[0].args[1] = 'bash -c "d${\"ocker\"} push image"';
  }, /shadow step args must match/);
});

test('quoted or split registry push is rejected structurally', () => {
  expectFailure((config) => {
    config.steps[0].args[1] = '"docker" "push" image';
  }, /shadow step args must match/);
});

test('nested shell is rejected structurally', () => {
  expectFailure((config) => {
    config.steps[0].args[1] = 'sh -c "node scripts/verify-cloud-build-shadow.mjs"';
  }, /shadow step args must match/);
});

test('unknown top-level config field is rejected', () => {
  expectFailure((config) => {
    config.unexpected = true;
  }, /shadow config has unknown or missing fields/);
});

test('unknown step field is rejected', () => {
  expectFailure((config) => {
    config.steps[0].unexpected = true;
  }, /shadow step has unknown or missing fields/);
});

test('duplicate root JSON key is rejected', () => {
  assert.throws(() => parseStrictJson('{"steps":[],"steps":[]}', 'duplicate root'), /duplicate object key/);
});

test('duplicate nested JSON key is rejected', () => {
  assert.throws(() => parseStrictJson('{"steps":[{"id":1,"id":2}]}', 'duplicate nested'), /duplicate object key/);
});

test('duplicate contract key is rejected', () => {
  assert.throws(() => parseStrictJson('{"provider":"gcp-cloud-build-shadow","provider":"github-actions"}', 'provider contract'), /duplicate object key/);
});

test('escaped JSON key parsing works without false duplicate detection', () => {
  assert.deepEqual(parseStrictJson('{"a\\u0062":1,"text":"{x:y,z}"}', 'escaped key'), { ab: 1, text: '{x:y,z}' });
});

test('malformed JSON and trailing garbage fail closed', () => {
  assert.throws(() => parseStrictJson('{"a":1,}', 'malformed'), /malformed JSON/);
  assert.throws(() => parseStrictJson('{"a":1} trailing', 'trailing'), /trailing data/);
});

test('contract unknown field fails closed', () => {
  const contract = clone(validContract);
  contract.futureUnknown = true;
  assert.throws(() => validateProviderContract(contract), /provider contract has unknown or missing fields/);
});

test('nested contract unknown field fails closed', () => {
  const contract = clone(validContract);
  contract.attestationContract.futureUnknown = true;
  assert.throws(() => validateProviderContract(contract), /attestationContract has unknown or missing fields/);
});

test('manual substitution remains unverified and cannot increase natural count', () => {
  assert.equal(validContract.classificationPolicy.manualBuildCountsAsNatural, false);
  assert.equal(validContract.classificationPolicy.unverifiedCandidateCountsAsNatural, false);
  assert.equal(validContract.currentGcpNaturalSamples, 0);
  assert.equal(validContract.currentGcpSyntheticSamples, 0);
});

test('source-generated result cannot attest itself', () => {
  assert.equal(validContract.classificationPolicy.sourceGeneratedResultCanAttest, false);
  assert.equal(validContract.attestationContract.sourceGeneratedResultCannotAttest, true);
});

test('provider contract requires attestation and operational evidence remains missing', () => {
  assert.equal(validContract.evidenceContract.attestationRequired, true);
  assert.equal(validContract.evidenceContract.operationalEvidenceStatus, 'MISSING');
  assert.equal(validContract.executionBoundary.liveExecutionInThisCheckpoint, false);
});

test('natural sample fields include server-issued and provenance evidence', () => {
  assert.ok(validContract.eligibleNaturalSampleFields.includes('serverIssuedBuildTriggerId'));
  assert.ok(validContract.eligibleNaturalSampleFields.includes('sourceProvenance'));
  assert.ok(validContract.eligibleNaturalSampleFields.includes('attestationStatus'));
});

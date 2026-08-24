import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import { join } from 'node:path';
import {
  validateOptionB,
  validateProviderContract,
  validateShadowConfig,
  validateShadowFiles,
} from './verify-cloud-build-shadow.mjs';

const root = process.cwd();
const validConfig = JSON.parse(readFileSync(join(root, 'cloudbuild', 'shadow-validation.yaml'), 'utf8'));
const validContract = JSON.parse(readFileSync(join(root, 'verification', 'provider-shadow-contract.json'), 'utf8'));

function clone(value) {
  return structuredClone(value);
}

function expectFailure(mutator, pattern) {
  const config = clone(validConfig);
  const contract = clone(validContract);
  mutator(config, contract);
  assert.throws(() => validateOptionB(config, contract), pattern);
}

test('valid Option B contract passes', () => {
  const result = validateShadowFiles({ root });
  assert.equal(result.config.profile, 'ci-quick');
  assert.equal(result.contract.actionsSamples, 3);
  assert.equal(result.contract.gcpSamples, 0);
});

test('mutable builder reference fails closed', () => {
  expectFailure((config) => {
    config.steps[0].name = 'node:24.17.0-bookworm';
  }, /immutable sha256 digest/);
});

test('missing source SHA binding fails closed', () => {
  expectFailure((config) => {
    config.steps[0].args[1] = config.steps[0].args[1].replace('\${COMMIT_SHA}', 'missing-source');
  }, /source identity substitution missing/);
});

test('source SHA bound to a different substitution fails closed', () => {
  expectFailure((config) => {
    config.steps[0].args[1] = config.steps[0].args[1].replace(
      '--source-sha "\${COMMIT_SHA}"',
      '--source-sha "\${BUILD_ID}"',
    );
  }, /source identity substitution missing|shadow runtime binding is not exact/);
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
  }, /forbidden deployment/);
});

test('registry push or publish command fails closed', () => {
  expectFailure((config) => {
    config.steps[0].args[1] += '\ndocker push example.invalid/image';
  }, /forbidden deployment/);
});

test('inline secret or key material fails closed', () => {
  expectFailure((config) => {
    config.steps[0].args[1] += '\ncat credentials.json';
  }, /forbidden deployment/);
});

test('missing timeout fails closed', () => {
  expectFailure((config) => {
    delete config.timeout;
  }, /build timeout must equal/);
});

test('Quick and Full duplicate execution fails closed', () => {
  expectFailure((config) => {
    config.steps[0].args[1] += '\n--profile full';
  }, /Quick and Full profiles/);
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
  assert.throws(() => validateProviderContract(contract), /sourceShaBinding is required/);
});

test('config without validator execution fails closed', () => {
  const config = clone(validConfig);
  config.steps[0].args[1] = config.steps[0].args[1].replace(
    'node scripts/verify-cloud-build-shadow.mjs',
    'node other-validator.mjs',
  );
  assert.throws(() => validateShadowConfig(config), /must invoke the shadow validator/);
});

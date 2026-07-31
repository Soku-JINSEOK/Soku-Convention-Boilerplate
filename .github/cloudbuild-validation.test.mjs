import assert from 'node:assert/strict';
import {chmodSync, mkdtempSync, readFileSync, writeFileSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join, resolve} from 'node:path';
import {spawnSync} from 'node:child_process';
import test from 'node:test';

const root = resolve(new URL('..', import.meta.url).pathname);

function executable(path, body) {
  writeFileSync(path, `#!/usr/bin/env bash\nset -euo pipefail\n${body}`);
  chmodSync(path, 0o755);
}

function runBootstrap(args, env = {}) {
  return spawnSync('bash', [join(root, 'scripts/gcp-bootstrap.sh'), ...args], {
    cwd: root,
    encoding: 'utf8',
    env: {...process.env, ...env},
  });
}

test('Cloud Build performs validation only with immutable builders', () => {
  const config = readFileSync(
    join(root, 'cloudbuild/validation.yaml'),
    'utf8',
  );
  const imageReferences = [
    ...config.matchAll(/^\s+name:\s+(\S+)$/gm),
  ].map((match) => match[1]);

  assert.equal(imageReferences.length, 3);
  for (const image of imageReferences) {
    assert.match(
      image,
      /^[\w./-]+:[\w.-]+@sha256:[0-9a-f]{64}$/,
      `builder is not pinned by version and digest: ${image}`,
    );
  }
  assert.match(imageReferences[0], /^node:24\./);
  assert.match(imageReferences[1], /^hashicorp\/terraform:1\.15\.3@sha256:/);
  assert.match(imageReferences[2], /^docker:29\.6\.1-cli@sha256:/);
  assert.match(config, /node --test[\s\S]*deploy-gcp\.test\.mjs/);
  assert.match(config, /node --test[\s\S]*cloudbuild-validation\.test\.mjs/);
  assert.match(config, /terraform fmt -check -recursive/);
  assert.match(
    config,
    /for root in infra\/gcp infra\/gcp\/cloud-build-logging/,
  );
  assert.match(config, /init -backend=false -input=false -lockfile=readonly/);
  assert.match(config, /terraform -chdir="\$root" validate/);
  assert.match(config, /terraform -chdir=infra\/gcp test/);
  assert.match(config, /build[\s\S]*--platform[\s\S]*linux\/amd64/);
  assert.match(config, /^timeout: 900s$/m);

  assert.doesNotMatch(config, /^\s*images:/m);
  assert.doesNotMatch(config, /\bdocker\s+push\b|\bpush\s+--/);
  assert.doesNotMatch(config, /^\s+-\s+push\s*$/m);
  assert.doesNotMatch(config, /\bgcloud\s+run\s+deploy\b/);
  assert.doesNotMatch(config, /secretManager|availableSecrets|secretEnv/i);
  assert.doesNotMatch(config, /service-account-key|credentials\.json/i);
  assert.doesNotMatch(config, /artifactregistry\.(?:writer|repositories\.upload)/i);
});

test('Cloud Build uses the trigger service account and Cloud Logging only', () => {
  const config = readFileSync(
    join(root, 'cloudbuild/validation.yaml'),
    'utf8',
  );

  assert.match(
    config,
    /^serviceAccount: \$\{_CLOUD_BUILD_SERVICE_ACCOUNT\}$/m,
  );
  assert.match(config, /^\s+logging: CLOUD_LOGGING_ONLY$/m);
});

test('Terraform creates two Tokyo first-generation GitHub triggers', () => {
  const main = readFileSync(join(root, 'infra/gcp/main.tf'), 'utf8');
  const variables = readFileSync(join(root, 'infra/gcp/variables.tf'), 'utf8');

  assert.match(
    variables,
    /variable "enable_cloud_build_validation"[\s\S]*default\s+= false/,
  );
  assert.match(main, /"cloudbuild\.googleapis\.com"/);
  assert.match(
    main,
    /resource "google_service_account" "cloud_build_validation"[\s\S]*count\s+= var\.enable_cloud_build_validation \? 1 : 0/,
  );
  assert.match(
    main,
    /resource "google_project_service" "cloud_build"[\s\S]*service\s+= "cloudbuild\.googleapis\.com"[\s\S]*disable_on_destroy\s+= false/,
  );
  assert.match(main, /cloud_build_validation_account_id[\s\S]*cb-[\s\S]*-ci/);
  assert.match(
    main,
    /resource "google_project_iam_member" "cloud_build_validation_log_writer"[\s\S]*roles\/logging\.logWriter/,
  );
  const validationIam = main.match(
    /resource "google_project_iam_member" "cloud_build_validation_log_writer" \{([\s\S]*?)\n\}/,
  )?.[1] ?? '';
  assert.match(validationIam, /roles\/logging\.logWriter/);
  assert.match(
    validationIam,
    /google_service_account\.cloud_build_validation\[0\]\.email/,
  );
  assert.doesNotMatch(validationIam, /roles\/(?:run|artifactregistry|secretmanager)\./);

  const triggers = [
    ...main.matchAll(/resource "google_cloudbuild_trigger" "([^"]+)"/g),
  ];
  assert.equal(triggers.length, 2);
  assert.match(main, /name\s+= "soku-convention-boilerplate-pr"/);
  assert.match(main, /name\s+= "soku-convention-boilerplate-main"/);
  assert.equal(
    (main.match(/count\s+= var\.enable_cloud_build_validation \? 1 : 0/g) ??
      []).length,
    5,
  );
  for (const triggerName of ['pull_request', 'main']) {
    assert.match(
      main,
      new RegExp(
        `resource "google_cloudbuild_trigger" "${triggerName}" \\{[\\s\\S]*?location\\s+= "asia-northeast1"`,
      ),
    );
  }
  assert.equal(
    (main.match(/filename\s+= "cloudbuild\/validation\.yaml"/g) ?? []).length,
    2,
  );
  assert.equal(
    (main.match(/include_build_logs\s+= "INCLUDE_BUILD_LOGS_WITH_STATUS"/g) ??
      []).length,
    2,
  );
  assert.equal((main.match(/branch\s+= "\^main\$"/g) ?? []).length, 2);
  assert.match(
    main,
    /comment_control\s+= "COMMENTS_ENABLED"/,
  );
  const mainTrigger = main.match(
    /resource "google_cloudbuild_trigger" "main" \{([\s\S]*?)\n\}/,
  )?.[1] ?? '';
  for (const includedFile of [
    '.github/cloudbuild-validation.test.mjs',
    '.github/deploy-gcp.test.mjs',
    'cloudbuild/**',
    'infra/gcp/**',
    'scripts/gcp-bootstrap.sh',
    'templates/gcloud/**',
  ]) {
    assert.match(mainTrigger, new RegExp(includedFile.replaceAll('*', '\\*')));
  }
  assert.match(main, /owner\s+= var\.github_org/);
  assert.match(main, /name\s+= var\.github_repo/);
  assert.doesNotMatch(main, /google_cloudbuildv2_|repository_event_config/);
});

test('bootstrap previews Cloud Build resources only when explicitly enabled', () => {
  const temp = mkdtempSync(join(tmpdir(), 'cloud-build-bootstrap-'));
  const bin = join(temp, 'bin');
  spawnSync('mkdir', ['-p', bin]);
  const invoked = join(temp, 'invoked');
  for (const command of ['gcloud', 'docker', 'terraform', 'gh', 'jq']) {
    executable(join(bin, command), `echo ${command} >> '${invoked}'`);
  }
  const env = {PATH: `${bin}:${process.env.PATH}`};

  const disabled = runBootstrap(
    ['--project-id', 'valid-project-123'],
    env,
  );
  assert.equal(disabled.status, 0, disabled.stderr);
  assert.match(disabled.stdout, /Cloud Build validation: disabled/);
  assert.match(disabled.stdout, /State prefix: cloud-run/);
  assert.doesNotMatch(disabled.stdout, /google_cloudbuild_trigger/);

  const enabled = runBootstrap([
    '--project-id',
    'valid-project-123',
    '--enable-cloud-build-validation',
  ], env);
  assert.equal(enabled.status, 0, enabled.stderr);
  assert.match(enabled.stdout, /Cloud Build validation: enabled/);
  assert.match(enabled.stdout, /State prefix: cloud-build-validation/);
  assert.match(enabled.stdout, /cloudbuild\.googleapis\.com/);
  assert.match(enabled.stdout, /google_service_account\.cloud_build_validation/);
  assert.match(enabled.stdout, /google_cloudbuild_trigger\.pull_request/);
  assert.match(enabled.stdout, /google_cloudbuild_trigger\.main/);
  assert.match(enabled.stdout, /enable_cloud_build_validation=true/);
  assert.doesNotMatch(enabled.stdout, /docker (?:build|push)|deploy_runtime=true/);
  assert.equal(spawnSync('test', ['!', '-e', invoked]).status, 0);
});

test('bootstrap previews low-cost controls without exposing billing identity', () => {
  const result = runBootstrap([
    '--project-id',
    'valid-project-123',
    '--max-instances',
    '1',
    '--enable-budget-alerts',
    '--monthly-budget-amount',
    '1500',
  ], {
    TF_VAR_billing_account_id: 'ABCDEF-123456-ABCDEF',
  });

  assert.equal(result.status, 0, result.stderr);
  assert.match(result.stdout, /Max instances: 1/);
  assert.match(result.stdout, /Budget alerts: enabled/);
  assert.match(result.stdout, /Artifact cleanup: dry-run/);
  assert.doesNotMatch(result.stdout, /ABCDEF-123456-ABCDEF/);
  assert.match(result.stdout, /state-lifecycle\.json/);
});

test('bootstrap stops before image work when first-generation trigger creation fails', () => {
  const temp = mkdtempSync(join(tmpdir(), 'cloud-build-connection-'));
  const bin = join(temp, 'bin');
  spawnSync('mkdir', ['-p', bin]);
  const log = join(temp, 'commands.log');

  executable(join(bin, 'gcloud'), `
echo "gcloud $*" >> '${log}'
if [[ "$*" == "storage buckets get-iam-policy"* ]]; then echo '{"bindings":[]}'; fi`);
  executable(join(bin, 'terraform'), `
echo "terraform $*" >> '${log}'
if [[ "$*" == *"apply"* && "$*" == *"google_cloudbuild_trigger.pull_request"* ]]; then
  echo "first-generation GitHub App connection unavailable" >&2
  exit 42
fi`);
  executable(join(bin, 'docker'), `echo "docker $*" >> '${log}'`);
  executable(join(bin, 'jq'), 'exit 1');
  executable(join(bin, 'gh'), `
echo "gh $*" >> '${log}'
if [[ "$*" == "repo view"* ]]; then echo owner/repository;
elif [[ "$*" == "api repos/owner/repository --jq .id" ]]; then echo 123456;
elif [[ "$*" == "api repos/owner/repository --jq .owner.id" ]]; then echo 7890; fi`);

  const result = runBootstrap([
    '--project-id',
    'app-project-123',
    '--enable-cloud-build-validation',
    '--apply',
    '--confirm-project-id',
    'app-project-123',
  ], {PATH: `${bin}:${process.env.PATH}`});

  assert.equal(result.status, 42);
  assert.match(result.stderr, /first-generation GitHub App connection unavailable/);
  const commands = readFileSync(log, 'utf8');
  assert.match(commands, /google_cloudbuild_trigger\.pull_request/);
  assert.match(commands, /enable_cloud_build_validation=true/);
  assert.doesNotMatch(commands, /^docker /m);
});

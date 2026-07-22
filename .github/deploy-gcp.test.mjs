import assert from 'node:assert/strict';
import {chmodSync, mkdtempSync, readFileSync, writeFileSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join, resolve} from 'node:path';
import {spawnSync} from 'node:child_process';
import test from 'node:test';

const root = resolve(new URL('..', import.meta.url).pathname);
const digest = `sha256:${'a'.repeat(64)}`;
const repository = 'asia-docker.pkg.dev/project/artifacts/service';

function executable(path, body) {
  writeFileSync(path, `#!/usr/bin/env bash\nset -euo pipefail\n${body}`);
  chmodSync(path, 0o755);
}

function run(script, args, env = {}) {
  return spawnSync('bash', [join(root, 'scripts', script), ...args], {
    cwd: root,
    encoding: 'utf8',
    env: {...process.env, ...env},
  });
}

function planArgs(output) {
  return [
    '--environment', 'dev', '--project-id', 'project', '--region', 'asia',
    '--service-name', 'service', '--artifact-repository', 'artifacts',
    '--output-dir', output, '--skip-local-checks', '--skip-infra', '--push-image',
  ];
}

test('pushed plans use a digest URI and fail when the digest is unavailable', () => {
  for (const available of [true, false]) {
    const temp = mkdtempSync(join(tmpdir(), 'cd-plan-'));
    const bin = join(temp, 'bin');
    spawnSync('mkdir', ['-p', bin]);
    executable(join(bin, 'gcloud'), ':');
    executable(join(bin, 'docker'), `
case "$1" in
  build|push) exit 0 ;;
  inspect)
    if [[ "$*" == *"{{.Id}}"* ]]; then echo image-id;
    ${available ? `else echo '${repository}@${digest}';` : 'else :;'} fi ;;
esac`);
    const result = run('cd-plan.sh', planArgs(join(temp, 'out')), {
      PATH: `${bin}:${process.env.PATH}`,
      GITHUB_SHA: '1234567890abcdef1234567890abcdef12345678',
    });
    assert.equal(result.status, available ? 0 : 11, result.stderr);
    if (available) {
      const plan = readFileSync(join(temp, 'out/dev/1234567890ab/cd-plan.env'), 'utf8');
      assert.match(plan, new RegExp(`CD_PLAN_IMAGE_URI=${repository}@${digest}`));
      assert.match(plan, /CD_PLAN_IMAGE_TAG_URI=.*:1234567890ab/);
    }
  }
});

test('rollback-only planning does not invoke Docker or Terraform', () => {
  const temp = mkdtempSync(join(tmpdir(), 'cd-rollback-plan-'));
  const bin = join(temp, 'bin');
  spawnSync('mkdir', ['-p', bin]);
  for (const command of ['docker', 'terraform']) {
    executable(join(bin, command), `echo invoked >> '${join(temp, 'invoked')}'`);
  }
  const result = run('cd-plan.sh', [
    '--environment', 'prod', '--project-id', 'project', '--region', 'asia',
    '--service-name', 'service', '--artifact-repository', 'artifacts',
    '--output-dir', join(temp, 'out'), '--rollback-only',
  ], {PATH: `${bin}:${process.env.PATH}`, GITHUB_SHA: 'abcdefabcdefabcdefabcdefabcdefabcdefabcd'});
  assert.equal(result.status, 0, result.stderr);
  assert.equal(spawnSync('test', ['!', '-e', join(temp, 'invoked')]).status, 0);
});

test('failed deploy health check restores the exact pre-deploy revision', () => {
  const temp = mkdtempSync(join(tmpdir(), 'cd-deploy-'));
  const bin = join(temp, 'bin');
  spawnSync('mkdir', ['-p', bin]);
  const log = join(temp, 'gcloud.log');
  const count = join(temp, 'describe-count');
  executable(join(bin, 'gcloud'), `
echo "$*" >> '${log}'
if [[ "$*" == *"print-identity-token"* ]]; then echo test-identity-token;
elif [[ "$*" == *"status.url"* ]]; then echo https://service.example;
elif [[ "$*" == *"latestReadyRevisionName"* ]]; then
  n=0; [[ -f '${count}' ]] && n=$(< '${count}'); n=$((n + 1)); echo "$n" > '${count}'
  if ((n == 1)); then echo service-pre; else echo service-new; fi
elif [[ "$*" == *"revisions list"* ]]; then echo service-older; fi`);
  const curlCount = join(temp, 'curl-count');
  executable(join(bin, 'curl'), `n=0; [[ -f '${curlCount}' ]] && n=$(< '${curlCount}'); n=$((n + 1)); echo "$n" > '${curlCount}'; ((n > 1))`);
  const plan = join(temp, 'plan.env');
  writeFileSync(plan, `CD_PLAN_ENVIRONMENT=dev\nCD_PLAN_COMMIT_SHA=abc\nCD_PLAN_IMAGE_TAG_URI=${repository}:abc\nCD_PLAN_IMAGE_URI=${repository}@${digest}\nCD_PLAN_PROJECT_ID=project\nCD_PLAN_REGION=asia\nCD_PLAN_SERVICE_NAME=service\n`);
  const result = run('cd-deploy.sh', ['--plan-file', plan, '--health-attempts', '1', '--health-delay', '0', '--confirm'], {
    PATH: `${bin}:${process.env.PATH}`,
    CD_DEPLOY_EVIDENCE_DIR: temp,
    GITHUB_RUN_ID: 'test-run',
    GITHUB_RUN_ATTEMPT: '1',
  });
  assert.equal(result.status, 1);
  assert.match(readFileSync(log, 'utf8'), /auth print-identity-token --audiences=https:\/\/service\.example/);
  assert.match(readFileSync(log, 'utf8'), /--to-revisions=service-pre=100/);
  assert.doesNotMatch(readFileSync(log, 'utf8'), /--to-revisions=service-older=100/);
  const evidence = JSON.parse(readFileSync(join(temp, 'deploy-dev-test-run-1.json')));
  assert.equal(evidence.final_status, 'rolled-back');
  assert.equal(evidence.rollback_target, 'service-pre');
});

test('manual rollback evidence distinguishes success, missing target, and failed health', () => {
  for (const scenario of ['success', 'missing', 'unhealthy']) {
    const temp = mkdtempSync(join(tmpdir(), `cd-manual-${scenario}-`));
    const bin = join(temp, 'bin');
    spawnSync('mkdir', ['-p', bin]);
    executable(join(bin, 'gcloud'), `
if [[ "$*" == *"print-identity-token"* ]]; then echo test-identity-token;
elif [[ "$*" == *"status.url"* ]]; then echo https://service.example;
elif [[ "$*" == *"latestReadyRevisionName"* ]]; then echo service-current;
elif [[ "$*" == *"revisions list"* ]]; then ${scenario === 'missing' ? ':' : 'echo service-target;'} fi`);
    executable(join(bin, 'curl'), scenario === 'unhealthy' ? 'exit 1' : 'exit 0');
    const plan = join(temp, 'plan.env');
    writeFileSync(plan, `CD_PLAN_ENVIRONMENT=prod\nCD_PLAN_COMMIT_SHA=abc\nCD_PLAN_PROJECT_ID=project\nCD_PLAN_REGION=asia\nCD_PLAN_SERVICE_NAME=service\n`);
    const args = ['--plan-file', plan, '--rollback-only', '--health-attempts', '1', '--health-delay', '0', '--confirm'];
    if (scenario !== 'missing') args.push('--rollback-revision', 'service-target');
    const result = run('cd-deploy.sh', args, {
      PATH: `${bin}:${process.env.PATH}`,
      CD_DEPLOY_EVIDENCE_DIR: temp,
      GITHUB_RUN_ID: 'test-run',
      GITHUB_RUN_ATTEMPT: '1',
    });
    assert.equal(result.status, scenario === 'success' ? 0 : scenario === 'missing' ? 4 : 9);
    const evidence = JSON.parse(readFileSync(join(temp, 'deploy-prod-test-run-1.json')));
    assert.equal(evidence.final_status, scenario === 'success' ? 'manual-rollback' : 'manual-rollback-failed');
    assert.equal(evidence.error, scenario === 'missing' ? 'missing-revision' : scenario === 'unhealthy' ? 'healthcheck-failed' : '');
  }
});

test('bootstrap dry-run is inert and apply requires matching confirmation', () => {
  const temp = mkdtempSync(join(tmpdir(), 'gcp-bootstrap-'));
  const bin = join(temp, 'bin');
  spawnSync('mkdir', ['-p', bin]);
  const log = join(temp, 'invoked');
  for (const command of ['gcloud', 'docker', 'terraform', 'gh']) {
    executable(join(bin, command), `echo ${command} >> '${log}'`);
  }
  const env = {PATH: `${bin}:${process.env.PATH}`};
  const dryRun = run('gcp-bootstrap.sh', ['--project-id', 'valid-project-123'], env);
  assert.equal(dryRun.status, 0, dryRun.stderr);
  assert.match(dryRun.stdout, /Mode: dry-run/);
  assert.match(dryRun.stdout, /gs:\/\/valid-project-123-tfstate/);
  assert.equal(spawnSync('test', ['!', '-e', log]).status, 0);

  const environmentRun = run('gcp-bootstrap.sh', [], {
    ...env,
    GCP_PROJECT_ID: 'environment-project-123',
  });
  assert.equal(environmentRun.status, 0, environmentRun.stderr);
  assert.match(environmentRun.stdout, /Project: environment-project-123/);
  assert.equal(spawnSync('test', ['!', '-e', log]).status, 0);

  const overrideRun = run('gcp-bootstrap.sh', [
    '--project-id', 'argument-project-123',
  ], {...env, GCP_PROJECT_ID: 'environment-project-123'});
  assert.equal(overrideRun.status, 0, overrideRun.stderr);
  assert.match(overrideRun.stdout, /Project: argument-project-123/);

  for (const args of [
    ['--project-id', 'valid-project-123', '--apply'],
    ['--project-id', 'valid-project-123', '--apply', '--confirm-project-id', 'wrong-project'],
  ]) {
    const result = run('gcp-bootstrap.sh', args, env);
    assert.equal(result.status, 2);
    assert.equal(spawnSync('test', ['!', '-e', log]).status, 0);
  }
});

test('bootstrap apply runs both Terraform stages and upserts repository variables', () => {
  const temp = mkdtempSync(join(tmpdir(), 'gcp-bootstrap-apply-'));
  const bin = join(temp, 'bin');
  spawnSync('mkdir', ['-p', bin]);
  const log = join(temp, 'commands.log');
  executable(join(bin, 'gcloud'), `
echo "gcloud $*" >> '${log}'
if [[ "$*" == "storage buckets describe"* ]]; then exit 1; fi
if [[ "$*" == "artifacts docker images describe"* ]]; then
  echo 'asia-northeast1-docker.pkg.dev/app-project-123/cloud-run/soku-convention-boilerplate@sha256:${'b'.repeat(64)}'
fi`);
  executable(join(bin, 'docker'), `echo "docker $*" >> '${log}'`);
  executable(join(bin, 'terraform'), `
echo "terraform $*" >> '${log}'
if [[ "$*" == *"output -raw wif_provider_name"* ]]; then
  echo projects/123/locations/global/workloadIdentityPools/github-actions/providers/gha
elif [[ "$*" == *"output -raw deployer_service_account_email"* ]]; then
  echo deployer@app-project-123.iam.gserviceaccount.com
fi`);
  executable(join(bin, 'gh'), `
echo "gh $*" >> '${log}'
if [[ "$*" == "repo view"* ]]; then echo owner/repository;
elif [[ "$*" == "api repos/owner/repository --jq .id" ]]; then echo 123456;
elif [[ "$*" == "api repos/owner/repository --jq .owner.id" ]]; then echo 7890; fi`);

  const result = run('gcp-bootstrap.sh', [
    '--apply', '--confirm-project-id', 'app-project-123',
  ], {PATH: `${bin}:${process.env.PATH}`, GCP_PROJECT_ID: 'app-project-123'});
  assert.equal(result.status, 0, result.stderr);
  const commands = readFileSync(log, 'utf8');
  assert.match(commands, /gcloud storage buckets create gs:\/\/app-project-123-tfstate/);
  assert.match(commands, /gcloud storage buckets update gs:\/\/app-project-123-tfstate --uniform-bucket-level-access --public-access-prevention --versioning/);
  assert.match(commands, /gh api repos\/owner\/repository --jq \.id/);
  assert.match(commands, /github_repository_id=123456/);
  assert.match(commands, /github_repository_owner_id=7890/);
  assert.match(commands, /terraform .* apply .*deploy_runtime=false/);
  assert.match(commands, /docker push asia-northeast1-docker\.pkg\.dev\/app-project-123\/cloud-run\/soku-convention-boilerplate:bootstrap/);
  assert.match(commands, /terraform .* apply .*deploy_runtime=true .*image_uri=.*@sha256:b{64}/);
  for (const variable of [
    'GCP_PROJECT_ID', 'GCP_REGION', 'GCP_SERVICE_NAME',
    'GCP_ARTIFACT_REPOSITORY', 'GCP_WIF_PROVIDER', 'GCP_WIF_SERVICE_ACCOUNT',
  ]) assert.match(commands, new RegExp(`gh variable set ${variable} `));
  assert.equal((commands.match(/gh variable set/g) ?? []).length, 6);
});

test('deployment workflow is manual and check cannot authenticate or deploy', () => {
  const workflow = readFileSync(join(root, '.github/workflows/deploy-gcp.yml'), 'utf8');
  for (const sha of [
    '7c6bc770dae815cd3e89ee6cdf493a5fab2cc093',
    'aa5489c8933f4cc7a4f7d45035b3b1440c9c10db',
    '043fb46d1a93c77aae656e7c1c64a875d1fc6a0a',
  ]) assert.match(workflow, new RegExp(sha));
  for (const action of workflow.matchAll(/^\s+uses:\s+([^\s#]+)/gm)) {
    assert.match(action[1], /@[0-9a-f]{40}$/, `action is not pinned: ${action[1]}`);
  }
  assert.doesNotMatch(workflow, /^\s*push:/m);
  assert.match(workflow, /operation:[\s\S]*default: check[\s\S]*options: \[check, deploy, rollback\]/);
  assert.match(workflow, /environment:[\s\S]*default: dev[\s\S]*options: \[dev\]/);
  assert.doesNotMatch(workflow, /options: \[dev, staging, prod\]/);
  const checkJob = workflow.match(/  check:\n([\s\S]*?)\n  deploy:/)?.[1] ?? '';
  assert.doesNotMatch(checkJob, /google-github-actions|gcloud auth|docker push|terraform plan|cd-deploy\.sh --/);
  assert.match(workflow, /if: \$\{\{ inputs\.operation == 'deploy' \}\}/);
  assert.match(workflow, /if: \$\{\{ inputs\.operation == 'rollback' \}\}/);
  assert.match(workflow, /if: \$\{\{ always\(\) \}\}[\s\S]*retention-days: 30/);
});

test('Terraform separates foundation from digest-pinned runtime', () => {
  const variables = readFileSync(join(root, 'infra/gcp/variables.tf'), 'utf8');
  const main = readFileSync(join(root, 'infra/gcp/main.tf'), 'utf8');
  const versions = readFileSync(join(root, 'infra/gcp/versions.tf'), 'utf8');
  assert.match(variables, /variable "deploy_runtime"[\s\S]*default\s+= false/);
  assert.match(variables, /variable "image_uri"[\s\S]*default\s+= null[\s\S]*@sha256:/);
  assert.match(main, /resource "google_cloud_run_service"[\s\S]*count\s+= var\.deploy_runtime \? 1 : 0/);
  assert.match(main, /account_id\s+= "\$\{substr\(var\.service_name, 0, 20\)\}-runtime"/);
  assert.match(main, /account_id\s+= "\$\{substr\(var\.service_name, 0, 15\)\}-gh-deployer"/);
  assert.match(main, /resource "google_service_account_iam_member" "deployer_runtime_user"/);
  assert.match(main, /resource "google_service_account_iam_member" "deployer_self_token_creator"/);
  assert.doesNotMatch(main, /resource "google_project_iam_member"[\s\S]{0,300}roles\/iam\.serviceAccountTokenCreator/);
  assert.match(main, /resource "google_cloud_run_service_iam_member" "deployer_invoker"/);
  assert.doesNotMatch(main, /resource "google_project_iam_member" "deployer_artifact_registry_writer"/);
  assert.match(main, /assertion\.repository_id/);
  assert.match(main, /assertion\.repository_owner_id/);
  assert.match(main, /assertion\.ref == \\"refs\/heads\/main\\"/);
  assert.match(main, /assertion\.workflow_ref/);
  assert.doesNotMatch(main, /assertion\.job_workflow_ref/);
  assert.match(variables, /variable "github_repository_id"/);
  assert.match(variables, /variable "github_repository_owner_id"/);
  assert.doesNotMatch(main, /allowed_audiences/);
  assert.equal((main.match(/display_name\s+= "github-\$\{substr\(var\.service_name, 0, 20\)\}"/g) ?? []).length, 2);
  assert.match(variables, /"run\.googleapis\.com"/);
  assert.doesNotMatch(variables, /"cloudrun\.googleapis\.com"/);
  assert.match(versions, /backend "gcs" \{\}/);
});

test('container builds target Cloud Run amd64 and expose a health endpoint', () => {
  const bootstrap = readFileSync(join(root, 'scripts/gcp-bootstrap.sh'), 'utf8');
  const plan = readFileSync(join(root, 'scripts/cd-plan.sh'), 'utf8');
  const dockerfile = readFileSync(join(root, 'templates/gcloud/Dockerfile'), 'utf8');
  assert.match(bootstrap, /docker build --platform linux\/amd64/);
  assert.match(bootstrap, /FOUNDATION_TARGETS=\(/);
  assert.match(bootstrap, /-target=google_iam_workload_identity_pool_provider\.github/);
  assert.match(bootstrap, /apply "\$\{COMMON_VARS\[@\]\}" "\$\{FOUNDATION_TARGETS\[@\]\}" -var="deploy_runtime=false"/);
  assert.match(plan, /docker build --platform linux\/amd64/);
  assert.match(dockerfile, /\/app\/health/);
  assert.match(dockerfile, /apk add --no-cache busybox-extras/);
  assert.match(dockerfile, /CMD \["httpd", "-f", "-p", "8080"/);
});

#!/usr/bin/env node

import {
  readFileSync,
  readdirSync,
  statSync,
  writeFileSync,
} from 'node:fs';
import {dirname, join, relative, resolve} from 'node:path';
import {fileURLToPath} from 'node:url';

const ACTION_SHA = /^[0-9a-f]{40}$/;
const IMAGE_PIN =
  /^[a-z0-9][a-z0-9./_-]*:[a-z0-9][a-z0-9._-]*@sha256:[0-9a-f]{64}$/i;
const EXACT_VERSION = /^v?\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)?$/;
const VERSION_VARIABLE = /^\$\{[A-Z][A-Z0-9_]*\}$/;

const PROTECTED_PATHS = [
  '.devcontainer',
  '.github/workflows',
  'cloudbuild',
  'docker-compose.verify.yml',
  'docs/releases/SIGNER_ROTATIONS.md',
  'release-identity.json',
  'scripts/create-release-tag.sh',
  'scripts/verify-release-tag.sh',
  'scripts/verify-release-tag_test.sh',
  'soku/scripts',
  'templates/_shared/ci',
  'templates/gcloud/Dockerfile',
  'verification/commands',
  'verification/tools.env',
];

export const REQUIRED_UPDATE_TARGETS = [
  ['github-actions', '/'],
  ['gomod', '/soku'],
  ['gomod', '/templates/go'],
  ['npm', '/soku/npm'],
  ['npm', '/soku/internal/manual/assets/runner'],
  ['npm', '/templates/javascript-typescript-node'],
  ['pip', '/templates/python'],
  ['maven', '/templates/java-spring'],
  ['docker', '/.devcontainer'],
  ['docker', '/templates/gcloud'],
  ['terraform', '/infra/gcp'],
  ['terraform', '/infra/gcp/cloud-build-logging'],
];

function collectFiles(path) {
  if (!statSync(path).isDirectory()) {
    return [path];
  }

  return readdirSync(path, {withFileTypes: true})
    .flatMap((entry) => {
      const child = join(path, entry.name);
      return entry.isDirectory() ? collectFiles(child) : [child];
    })
    .sort();
}

function versionReferenceIsExact(reference) {
  return EXACT_VERSION.test(reference) || VERSION_VARIABLE.test(reference);
}

function imageDefault(value) {
  const match = value.match(/^\$\{[A-Z][A-Z0-9_]*:-(.+)\}$/);
  return match ? match[1] : value;
}

function finding(path, line, rule, message) {
  return {path, line, rule, message};
}

export function inspectContent(path, content) {
  const findings = [];

  for (const [index, line] of content.split(/\r?\n/).entries()) {
    const lineNumber = index + 1;

    if (/@latest\b/.test(line)) {
      findings.push(
        finding(
          path,
          lineNumber,
          'floating-latest',
          'Replace the latest selector with a reviewed exact version.',
        ),
      );
    }

    const action = line.match(/uses:\s*([^#\s]+)@([^#\s]+)/);
    if (action && !ACTION_SHA.test(action[2])) {
      findings.push(
        finding(
          path,
          lineNumber,
          'action-sha',
          `Pin ${action[1]} to a full 40-character commit SHA.`,
        ),
      );
    }

    const cloudBuildImage =
      path.startsWith('cloudbuild/') &&
      line.match(/^\s*(?:#\s*)?name:\s*([^\s#]+)/);
    const image =
      line.match(/^\s*(?:#\s*)?image:\s*([^\s#]+)/) ??
      cloudBuildImage;
    if (image && !IMAGE_PIN.test(imageDefault(image[1]))) {
      findings.push(
        finding(
          path,
          lineNumber,
          'image-digest',
          'Pin the image to an exact tag and sha256 manifest digest.',
        ),
      );
    }

    const dockerFrom =
      /(^|\/)Dockerfile(?:\.[^/]*)?$/.test(path) &&
      line.match(/^\s*FROM\s+([^\s#]+)/i);
    if (
      dockerFrom &&
      dockerFrom[1] !== 'scratch' &&
      !IMAGE_PIN.test(dockerFrom[1])
    ) {
      findings.push(
        finding(
          path,
          lineNumber,
          'image-digest',
          'Pin the base image to an exact tag and sha256 manifest digest.',
        ),
      );
    }

    for (const tool of line.matchAll(
      /\bgo (?:install|run)\s+["']?([^"'\\\s]+)/g,
    )) {
      const at = tool[1].lastIndexOf('@');
      const reference = at > 0 ? tool[1].slice(at + 1) : '';
      const localPackage =
        tool[1].startsWith('./') || tool[1].startsWith('../');
      if (!localPackage && !versionReferenceIsExact(reference)) {
        findings.push(
          finding(
            path,
            lineNumber,
            'go-tool-version',
            `Pin ${tool[1]} to an exact version or tools.env variable.`,
          ),
        );
      }
    }

    const npx = line.match(/\bnpx\s+--yes\s+["']?([^"'\\\s]+)/);
    if (npx) {
      const at = npx[1].lastIndexOf('@');
      const reference = at > 0 ? npx[1].slice(at + 1) : '';
      if (!versionReferenceIsExact(reference)) {
        findings.push(
          finding(
            path,
            lineNumber,
            'npx-tool-version',
            `Pin ${npx[1]} to an exact version or tools.env variable.`,
          ),
        );
      }
    }
  }

  return findings;
}

export function parseToolsEnv(content) {
  return new Map(
    content
      .split(/\r?\n/)
      .map((line) => line.match(/^([A-Z][A-Z0-9_]*)="([^"]+)"$/))
      .filter(Boolean)
      .map((match) => [match[1], match[2]]),
  );
}

export function parseDependabotUpdates(content) {
  const updates = [];
  let current;

  for (const line of content.split(/\r?\n/)) {
    const ecosystem = line.match(
      /^\s*-\s+package-ecosystem:\s*["']?([^"'\s]+)["']?\s*$/,
    );
    if (ecosystem) {
      current = {ecosystem: ecosystem[1]};
      updates.push(current);
      continue;
    }

    const directoryMatch = line.match(
      /^\s+directory:\s*["']?([^"'\s]+)["']?\s*$/,
    );
    if (current && directoryMatch) {
      current.directory = directoryMatch[1];
    }
  }

  return updates;
}

export function verifyDependabotCoverage(content) {
  const configured = new Set(
    parseDependabotUpdates(content).map(
      ({ecosystem, directory}) => `${ecosystem}:${directory}`,
    ),
  );

  return REQUIRED_UPDATE_TARGETS.filter(
    ([ecosystem, directory]) => !configured.has(`${ecosystem}:${directory}`),
  ).map(([ecosystem, directory]) =>
    finding(
      '.github/dependabot.yml',
      1,
      'update-coverage',
      `Add ${ecosystem} update coverage for ${directory}.`,
    ),
  );
}

export function renderToolsPolicy(tools) {
  return [
    '<!-- tools-env:start -->',
    '| Policy key | Reviewed value |',
    '| --- | --- |',
    ...[...tools].map(([key, value]) => `| \`${key}\` | \`${value}\` |`),
    '<!-- tools-env:end -->',
  ].join('\n');
}

export function verifyPolicyDocumentation(content, tools) {
  return content.replaceAll('\r\n', '\n').includes(renderToolsPolicy(tools)) ? [] : [finding(
    'verification/CLASSIFICATION.md', 1, 'policy-documentation',
    'Regenerate shared policy values with node scripts/verify-supply-chain.mjs --write-docs.',
  )];
}

export function verifyPythonSupport(project, workflow) {
  const range = /^requires-python = ">=3\.(\d+),<3\.(\d+)"$/m.exec(project.replaceAll('\r\n', '\n'));
  const matrix = /python-version: \[([^\]]+)\]/.exec(workflow);
  const versions = matrix ? [...matrix[1].matchAll(/['"](3\.\d+)['"]/g)].map(m => m[1]) : [];
  const minimum = Number(range?.[1]);
  const upper = Number(range?.[2]);
  const expected = range && upper > minimum && upper - minimum < 20
    ? Array.from({length: upper - minimum}, (_, i) => `3.${minimum + i}`) : [];
  return expected.length > 0 && JSON.stringify(versions) === JSON.stringify(expected) &&
    workflow.includes('python-version: ${{ matrix.python-version }}') &&
    workflow.includes('fail-fast: false') ? [] : [finding(
      'templates/python/pyproject.toml', 1, 'python-support',
      'Every declared Python minor must run in the Full matrix, with fail-fast disabled.',
    )];
}

export function verifyRuntimeSupport(root) {
  const read = path => readFileSync(resolve(root, path), 'utf8');
  const findings = [];
  const nodeRange = JSON.parse(read('templates/javascript-typescript-node/package.json')).engines.node;
  const lockRange = JSON.parse(read('templates/javascript-typescript-node/package-lock.json')).packages[''].engines.node;
  const clauses = [...nodeRange.matchAll(/>=([0-9]+(?:\.[0-9]+){0,2}) <([0-9]+)/g)];
  const expectedNode = clauses.map(match => match[1]);
  if (clauses.map(match => match[0]).join(' || ') !== nodeRange || clauses.length === 0 ||
      clauses.some(match => Number(match[2]) !== Number(match[1].split('.')[0]) + 1) || nodeRange !== lockRange) {
    findings.push(finding('templates/javascript-typescript-node/package.json', 1, 'runtime-support', 'Declare tested Node release lines and keep the lockfile root in sync.'));
  }
  const goVersion = /^go (\d+\.\d+)$/m.exec(read('templates/go/go.mod').replaceAll('\r\n', '\n'))?.[1];
  const javaVersion = /<java.version>(\d+)<\/java.version>/.exec(read('templates/java-spring/pom.xml'))?.[1];
  for (const path of ['.github/workflows/templates-ci.yml', 'templates/_shared/ci/downstream-ci.yml']) {
    const workflow = read(path);
    const matrix = /node-version: \[([^\]]+)\]/.exec(workflow);
    const actualNode = matrix ? [...matrix[1].matchAll(/['"]([0-9.]+)['"]/g)].map(match => match[1]) : [];
    if (JSON.stringify(actualNode) !== JSON.stringify(expectedNode) || !workflow.includes('node-version: ${{ matrix.node-version }}')) {
      findings.push(finding(path, 1, 'runtime-support', 'Full Node matrix must test the minimum of each declared release line.'));
    }
    for (const [key, version] of [['go-version', goVersion], ['java-version', javaVersion]]) {
      if (!version || !workflow.includes(`${key}: '${version}'`)) {
        findings.push(finding(path, 1, 'runtime-support', `${key} must match the template build declaration.`));
      }
    }
  }
  const terraformVersion = /hashicorp\/terraform:(\d+\.\d+\.\d+)@sha256:/.exec(read('scripts/verify.sh'))?.[1];
  const [major, minor] = (terraformVersion ?? '').split('.').map(Number);
  for (const path of ['infra/gcp/versions.tf', 'infra/gcp/cloud-build-logging/versions.tf']) {
    const range = /required_version = "([^"]+)"/.exec(read(path))?.[1];
    if (!terraformVersion || range !== `>= ${terraformVersion}, < ${major}.${minor + 1}.0`) {
      findings.push(finding(path, 1, 'runtime-support', 'Terraform support must start at the verification image version and remain in its tested minor line.'));
    }
  }
  return findings;
}

function parityChecks(root, tools) {
  const checks = [
    ['GOIMPORTS_VERSION', '.github/workflows/ci.yml', 'goimports@'],
    ['GOIMPORTS_VERSION', '.github/workflows/templates-ci.yml', 'goimports@'],
    ['GOIMPORTS_VERSION', 'templates/_shared/ci/downstream-ci.yml', 'goimports@'],
    ['GOLANGCI_LINT_VERSION', '.github/workflows/ci.yml', 'golangci-lint@'],
    [
      'GOLANGCI_LINT_VERSION',
      '.github/workflows/templates-ci.yml',
      'golangci-lint@',
    ],
    ['MARKDOWNLINT_CLI2_VERSION', '.github/workflows/ci.yml', 'markdownlint-cli2@'],
    ['YAML_LINT_VERSION', '.github/workflows/ci.yml', 'yaml-lint@'],
    ['ACTIONLINT_VERSION', '.github/workflows/ci.yml', 'actionlint@'],
    ['GITLEAKS_VERSION', '.github/workflows/security.yml', 'gitleaks:'],
    ['GITLEAKS_VERSION', 'templates/_shared/ci/downstream-ci-security.yml', 'gitleaks/v8@'],
    ['PIP_AUDIT_VERSION', '.github/workflows/security.yml', 'pip-audit=='],
    ['PIP_AUDIT_VERSION', 'templates/_shared/ci/downstream-ci-security.yml', 'pip-audit=='],
    ['GOVULNCHECK_VERSION', '.github/workflows/security.yml', 'govulncheck@'],
    ['GOVULNCHECK_VERSION', 'templates/_shared/ci/downstream-ci-security.yml', 'govulncheck@'],
    ['OSV_SCANNER_VERSION', '.github/workflows/security.yml', 'osv-scanner:'],
    ['NPM_AUDIT_LEVEL', '.github/workflows/security.yml', 'audit-level='],
    ['NPM_AUDIT_LEVEL', 'templates/_shared/ci/downstream-ci-security.yml', 'audit-level='],
    ['MYSQL_IMAGE', '.github/workflows/templates-ci.template.yml', 'image: '],
    ['MYSQL_IMAGE', '.github/workflows/templates-ci.yml', 'image: '],
    ['MYSQL_IMAGE', 'docker-compose.verify.yml', ':-'],
    ['POSTGRES_IMAGE', '.github/workflows/templates-ci.template.yml', 'image: '],
    ['POSTGRES_IMAGE', '.github/workflows/templates-ci.yml', 'image: '],
    ['POSTGRES_IMAGE', 'docker-compose.verify.yml', ':-'],
  ];
  const findings = [];

  for (const [variable, path, prefix] of checks) {
    const value = tools.get(variable);
    const content = readFileSync(resolve(root, path), 'utf8');
    if (!value || !content.includes(`${prefix}${value}`)) {
      findings.push(
        finding(
          path,
          1,
          'tools-env-parity',
          `${variable} must match the reviewed value in verification/tools.env.`,
        ),
      );
    }
  }

  return findings;
}

export function verifyRepository(root) {
  const findings = [];
  const protectedFiles = PROTECTED_PATHS.flatMap((path) =>
    collectFiles(resolve(root, path)),
  );

  for (const path of protectedFiles) {
    findings.push(
      ...inspectContent(
        relative(root, path),
        readFileSync(path, 'utf8'),
      ),
    );
  }

  const tools = parseToolsEnv(
    readFileSync(resolve(root, 'verification/tools.env'), 'utf8'),
  );
  findings.push(...parityChecks(root, tools));
  findings.push(...verifyRuntimeSupport(root));
  for (const path of ['.github/workflows/templates-ci.yml', 'templates/_shared/ci/downstream-ci.yml']) {
    findings.push(...verifyPythonSupport(
      readFileSync(resolve(root, 'templates/python/pyproject.toml'), 'utf8'),
      readFileSync(resolve(root, path), 'utf8'),
    ));
  }
  findings.push(...verifyPolicyDocumentation(
    readFileSync(resolve(root, 'verification/CLASSIFICATION.md'), 'utf8'), tools,
  ));

  const dependabot = readFileSync(
    resolve(root, '.github/dependabot.yml'),
    'utf8',
  );
  findings.push(...verifyDependabotCoverage(dependabot));

  return {
    findings,
    protectedFileCount: protectedFiles.length,
    updateTargetCount: REQUIRED_UPDATE_TARGETS.length,
  };
}

function main() {
  const root = resolve(dirname(fileURLToPath(import.meta.url)), '..');
  if (process.argv.includes('--write-docs')) {
    const path = resolve(root, 'verification/CLASSIFICATION.md');
    const content = readFileSync(path, 'utf8');
    const tools = parseToolsEnv(readFileSync(resolve(root, 'verification/tools.env'), 'utf8'));
    const marker = /<!-- tools-env:start -->[\s\S]*?<!-- tools-env:end -->/g;
    if ([...content.matchAll(marker)].length !== 1) {
      throw new Error('Expected exactly one tools-env documentation block.');
    }
    writeFileSync(path, content.replace(marker, () => renderToolsPolicy(tools)));
  }
  const result = verifyRepository(root);

  if (result.findings.length > 0) {
    for (const item of result.findings) {
      console.error(
        `${item.path}:${item.line} [${item.rule}] ${item.message}`,
      );
    }
    process.exitCode = 1;
    return;
  }

  console.log(
    `Supply-chain verification passed: ${result.protectedFileCount} ` +
      `protected files, ${result.updateTargetCount} update targets.`,
  );
}

if (
  process.argv[1] &&
  resolve(process.argv[1]) === fileURLToPath(import.meta.url)
) {
  main();
}

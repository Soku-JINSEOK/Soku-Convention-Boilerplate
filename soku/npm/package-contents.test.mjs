import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import test from 'node:test';

const packageRoot = path.dirname(fileURLToPath(import.meta.url));
const repositoryRoot = path.resolve(packageRoot, '..', '..');
const npmCommand = process.platform === 'win32' ? 'npm.cmd' : 'npm';
const expectedFiles = [
  'LICENSE',
  'README.md',
  'bin/soku.js',
  'lib/launcher.mjs',
  'package.json',
];

function pack(args) {
  const output = execFileSync(
    npmCommand,
    ['pack', '--json', '--ignore-scripts', ...args],
    {
      cwd: packageRoot,
      encoding: 'utf8',
    },
  );
  const result = JSON.parse(output);

  assert.equal(result.length, 1);
  return result[0];
}

function inventory(result) {
  return result.files.map(({ path: filePath }) => filePath).sort();
}

test('package license matches the repository license', () => {
  const packageJson = JSON.parse(
    fs.readFileSync(path.join(packageRoot, 'package.json'), 'utf8'),
  );
  const packageLicense = fs.readFileSync(
    path.join(packageRoot, 'LICENSE'),
    'utf8',
  );
  const repositoryLicense = fs.readFileSync(
    path.join(repositoryRoot, 'LICENSE'),
    'utf8',
  );

  assert.equal(packageJson.license, 'MIT');
  assert.equal(packageLicense, repositoryLicense);
});

test('npm pack dry-run exposes only the public package contract', () => {
  assert.deepEqual(inventory(pack(['--dry-run'])), expectedFiles);
});

test('packed tarball exposes only the public package contract', () => {
  const outputDirectory = fs.mkdtempSync(
    path.join(os.tmpdir(), 'soku-npm-pack-'),
  );

  try {
    const result = pack(['--pack-destination', outputDirectory]);

    assert.deepEqual(inventory(result), expectedFiles);
    assert.equal(
      fs.existsSync(path.join(outputDirectory, result.filename)),
      true,
    );
  } finally {
    fs.rmSync(outputDirectory, { recursive: true, force: true });
  }
});

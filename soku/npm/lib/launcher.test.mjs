import assert from 'node:assert/strict';
import test from 'node:test';

import {
  releaseTagFromVersion,
  artifactName,
  checksumLineFromText,
  resolveTarget,
} from './launcher.mjs';

test('releaseTagFromVersion validates and prefixes', () => {
  assert.equal(releaseTagFromVersion('0.2.0'), 'soku/v0.2.0');
  assert.equal(releaseTagFromVersion('0.12.3'), 'soku/v0.12.3');
  assert.throws(() => releaseTagFromVersion('v0.2.0'), /invalid CLI version/);
});

test('resolveTarget maps supported platforms', () => {
  assert.deepEqual(resolveTarget('linux', 'x64'), {os: 'linux', arch: 'amd64', executable: 'soku'});
  assert.deepEqual(resolveTarget('linux', 'arm64'), {os: 'linux', arch: 'arm64', executable: 'soku'});
  assert.deepEqual(resolveTarget('darwin', 'arm64'), {os: 'darwin', arch: 'arm64', executable: 'soku'});
  assert.deepEqual(resolveTarget('win32', 'x64'), {os: 'windows', arch: 'amd64', executable: 'soku.exe'});
});

test('artifactName follows platform naming scheme', () => {
  const linuxAmd64 = {os: 'linux', arch: 'amd64', executable: 'soku'};
  const windows = {os: 'windows', arch: 'amd64', executable: 'soku.exe'};
  assert.equal(artifactName('0.2.0', linuxAmd64), 'soku_v0.2.0_linux_amd64.tar.gz');
  assert.equal(artifactName('0.2.0', windows), 'soku_v0.2.0_windows_amd64.zip');
});

test('checksumLineFromText returns matching hash', () => {
  const raw = `abcd  soku_v0.2.0_linux_amd64.tar.gz
efgh  soku_v0.2.0_windows_amd64.zip`;
  assert.equal(checksumLineFromText(raw, 'soku_v0.2.0_windows_amd64.zip'), 'efgh');
  assert.equal(checksumLineFromText(raw, 'missing'), '');
});

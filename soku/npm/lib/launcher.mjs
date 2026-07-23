import {createHash} from 'node:crypto';
import {createReadStream, createWriteStream, existsSync} from 'node:fs';
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import process from 'node:process';
import {pipeline} from 'node:stream/promises';
import {spawnSync} from 'node:child_process';
import {tmpdir} from 'node:os';

const cliReleasePrefix = 'soku/v';
const releaseDownloads = 'https://github.com';
const cliProgram = 'soku';
const checksumsFile = 'checksums.txt';
const checksumAlgorithm = 'sha256';

const minimumNpmVersion = '0.2.0';

function splitVersion(version) {
  return version.split('.').map((segment) => Number.parseInt(segment, 10));
}

function compareVersions(left, right) {
  const leftParts = splitVersion(left);
  const rightParts = splitVersion(right);
  const maxLength = Math.max(leftParts.length, rightParts.length);

  for (let i = 0; i < maxLength; i++) {
    const leftValue = Number.isNaN(leftParts[i]) ? 0 : leftParts[i];
    const rightValue = Number.isNaN(rightParts[i]) ? 0 : rightParts[i];
    if (leftValue !== rightValue) {
      return leftValue - rightValue;
    }
  }
  return 0;
}

export function releaseTagFromVersion(version) {
  if (!/^\d+\.\d+\.\d+(?:[-+].*)?$/.test(version)) {
    throw new Error(`invalid CLI version: ${version}`);
  }
  return `${cliReleasePrefix}${version}`;
}

export function releaseTagFromPackageJson(cliPackage) {
  return releaseTagFromVersion(cliPackage.version);
}

export function isNpmPublishReady(version) {
  return compareVersions(version, minimumNpmVersion) >= 0;
}

export function resolveTarget(platform = process.platform, architecture = process.arch) {
  const normalizedPlatform = platform;
  const normalizedArch = architecture;

  if (normalizedPlatform === 'linux') {
    if (normalizedArch === 'x64') {
      return {os: 'linux', arch: 'amd64', executable: cliProgram};
    }
    if (normalizedArch === 'arm64') {
      return {os: 'linux', arch: 'arm64', executable: cliProgram};
    }
  }
  if (normalizedPlatform === 'darwin') {
    if (normalizedArch === 'x64') {
      return {os: 'darwin', arch: 'amd64', executable: cliProgram};
    }
    if (normalizedArch === 'arm64') {
      return {os: 'darwin', arch: 'arm64', executable: cliProgram};
    }
  }
  if (normalizedPlatform === 'win32' && normalizedArch === 'x64') {
    return {os: 'windows', arch: 'amd64', executable: `${cliProgram}.exe`};
  }

  throw new Error(`unsupported platform: ${normalizedPlatform}/${normalizedArch}`);
}

export function artifactName(version, target = resolveTarget()) {
  const extension = target.os === 'windows' ? 'zip' : 'tar.gz';
  return `${cliProgram}_v${version}_${target.os}_${target.arch}.${extension}`;
}

export function checksumLineFromText(rawChecksums, targetAsset) {
  const lines = rawChecksums.split('\n');
  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed) {
      continue;
    }
    const [hash, filename] = trimmed.split(/\s+/);
    if (filename === targetAsset) {
      return hash;
    }
  }
  return '';
}

export function resolveReleaseAssetUrl(repository, tag, asset) {
  return `${releaseDownloads}/${repository}/releases/download/${tag}/${asset}`;
}

function readStreamToFile(response, destination) {
  if (!response.ok || !response.body) {
    throw new Error(`failed to download asset: ${response.status} ${response.statusText}`);
  }
  return pipeline(response.body, createWriteStream(destination));
}

export async function ensureAssetDownloaded(url, destination) {
  const response = await fetch(url);
  await readStreamToFile(response, destination);
}

export function cacheRootFor(version, target, repository) {
  const safeRepo = repository.replaceAll('/', '_');
  return path.join(os.homedir(), '.cache', 'soku', safeRepo, version, target.os, target.arch);
}

async function extractArchive(assetPath, destination, target) {
  if (target.os === 'windows') {
    const powershell = process.env.SystemRoot
      ? path.join(process.env.SystemRoot, 'System32', 'WindowsPowerShell', 'v1.0', 'powershell.exe')
      : 'powershell';
    const command = `Expand-Archive -Path '${assetPath}' -DestinationPath '${destination}' -Force`;
    const result = spawnSync(powershell, ['-NoProfile', '-Command', command], {stdio: 'inherit'});
    if (result.status !== 0) {
      throw new Error('failed to extract windows zip');
    }
    return;
  }

  const result = spawnSync('tar', ['-xzf', assetPath, '-C', destination], {stdio: 'inherit'});
  if (result.status !== 0) {
    throw new Error('failed to extract archive');
  }
}

export async function resolveBinary({
  version,
  repository,
  target = resolveTarget(),
  targetTag = releaseTagFromVersion(version),
}) {
  const asset = artifactName(version, target);
  const root = cacheRootFor(targetTag, target, repository);
  const binaryPath = path.join(root, target.executable);

  if (existsSync(binaryPath)) {
    return binaryPath;
  }

  const checksumsUrl = resolveReleaseAssetUrl(repository, targetTag, checksumsFile);
  const response = await fetch(checksumsUrl);
  if (!response.ok) {
    throw new Error(`failed to fetch checksums: ${response.status} ${response.statusText}`);
  }
  const expected = checksumLineFromText(await response.text(), asset);
  if (!expected) {
    throw new Error(`checksum entry not found: ${asset}`);
  }

  const assetUrl = resolveReleaseAssetUrl(repository, targetTag, asset);
  const temporaryDirectory = await fs.mkdtemp(path.join(tmpdir(), 'soku-cli-'));
  const downloadedAsset = path.join(temporaryDirectory, asset);
  const extractedDirectory = path.join(temporaryDirectory, 'payload');

  await ensureAssetDownloaded(assetUrl, downloadedAsset);
  const actual = await fileSha256(downloadedAsset);
  if (actual !== expected) {
    throw new Error(`checksum mismatch for ${asset}`);
  }

  await fs.mkdir(extractedDirectory, {recursive: true});
  await extractArchive(downloadedAsset, extractedDirectory, target);
  await fs.mkdir(root, {recursive: true});
  const extractedBinary = path.join(extractedDirectory, target.executable);
  if (!existsSync(extractedBinary)) {
    throw new Error(`expected executable not found after extraction: ${target.executable}`);
  }
  await fs.copyFile(extractedBinary, binaryPath);
  if (process.platform !== 'win32') {
    await fs.chmod(binaryPath, 0o755);
  }
  return binaryPath;
}

function fileSha256(filePath) {
  const hash = createHash(checksumAlgorithm);
  return pipeline(createReadStream(filePath), hash).then(() => hash.digest('hex'));
}

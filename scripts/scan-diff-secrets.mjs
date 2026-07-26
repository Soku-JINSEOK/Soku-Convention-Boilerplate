#!/usr/bin/env node

import {readFileSync} from 'node:fs';
import {resolve} from 'node:path';
import {fileURLToPath} from 'node:url';

const rules = [
  ['private-key', /-----BEGIN (?:RSA |EC |OPENSSH |DSA )?PRIVATE KEY-----/],
  ['github-token', /\bgh[pousr]_[A-Za-z0-9]{36,}\b/],
  ['npm-token', /\bnpm_[A-Za-z0-9]{36,}\b/],
  ['aws-access-key', /\bAKIA[0-9A-Z]{16}\b/],
  ['google-api-key', /\bAIza[0-9A-Za-z_-]{35}\b/],
  ['slack-token', /\bxox[baprs]-[0-9A-Za-z-]{20,}\b/],
];

export function scanDiff(content) {
  const findings = [];
  let path = '';
  let addedLine = 0;

  for (const line of content.split(/\r?\n/)) {
    if (line.startsWith('+++ b/')) {
      path = line.slice(6);
      continue;
    }
    const hunk = /^@@ -[^+]*\+(\d+)/.exec(line);
    if (hunk) {
      addedLine = Number.parseInt(hunk[1], 10);
      continue;
    }
    if (line.startsWith('+') && !line.startsWith('+++')) {
      const value = line.slice(1);
      for (const [rule, pattern] of rules) {
        if (pattern.test(value)) findings.push({path, line: addedLine, rule});
      }
      addedLine += 1;
    } else if (!line.startsWith('-') && !line.startsWith('\\')) {
      addedLine += 1;
    }
  }
  return findings;
}

function main() {
  const index = process.argv.indexOf('--diff-file');
  const path = index >= 0 ? process.argv[index + 1] : '';
  if (!path) {
    process.stderr.write('Usage: scripts/scan-diff-secrets.mjs --diff-file <path>\n');
    process.exitCode = 2;
    return;
  }
  const findings = scanDiff(readFileSync(path, 'utf8'));
  if (findings.length > 0) {
    for (const finding of findings) {
      process.stderr.write(
        `${finding.path || '<diff>'}:${finding.line} ` +
          `[${finding.rule}] possible secret in added line (value redacted)\n`,
      );
    }
    process.exitCode = 1;
    return;
  }
  process.stdout.write('Changed-line secret scan passed.\n');
}

if (
  process.argv[1] &&
  resolve(process.argv[1]) === fileURLToPath(import.meta.url)
) {
  main();
}

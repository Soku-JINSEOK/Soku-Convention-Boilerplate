import assert from 'node:assert/strict';
import test from 'node:test';
import {scanDiff} from './scan-diff-secrets.mjs';

test('reports high-confidence secrets without retaining the value', () => {
  const token = `gh${'p'}_${'A'.repeat(36)}`;
  const findings = scanDiff(
    `diff --git a/example b/example\n+++ b/example\n@@ -0,0 +1 @@\n+${token}\n`,
  );
  assert.deepEqual(findings, [
    {path: 'example', line: 1, rule: 'github-token'},
  ]);
  assert.doesNotMatch(JSON.stringify(findings), new RegExp(token));
});

test('ignores removed values, context, and secret references', () => {
  const token = `npm_${'B'.repeat(36)}`;
  const findings = scanDiff(
    `diff --git a/example b/example\n+++ b/example\n@@ -1,2 +1,2 @@\n-${token}\n context\n+\${{ secrets.NPM_TOKEN }}\n`,
  );
  assert.deepEqual(findings, []);
});

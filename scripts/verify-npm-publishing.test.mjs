import assert from 'node:assert/strict';
import test from 'node:test';

import {verifyNpmPublishing} from './verify-npm-publishing.mjs';

const validWorkflow = `
name: Release
jobs:
  publish-npm:
    permissions:
      contents: read
      id-token: write
    steps:
      - uses: actions/setup-node@reviewed
        with:
          node-version: "24"
          package-manager-cache: false
      - run: npm install --global npm@12.0.1
      - run: >-
          cd soku/npm &&
          npm publish --provenance --access public
`;

const validPackage = {
  repository: {
    type: 'git',
    url: 'https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate.git',
  },
};

test('accepts the reviewed Trusted Publishing contract', () => {
  assert.deepEqual(verifyNpmPublishing(validWorkflow, validPackage), []);
});

test('rejects token fallback and token-oriented registry configuration', () => {
  const workflow = `${validWorkflow}
      - run: echo "\${NPM_TOKEN}"
        env:
          NODE_AUTH_TOKEN: \${{ secrets.NPM_TOKEN }}
      - with:
          registry-url: https://registry.npmjs.org
`;
  const errors = verifyNpmPublishing(workflow, validPackage);

  assert.ok(
    errors.includes('release workflow must not inject an npm publication token'),
  );
  assert.ok(
    errors.includes(
      'publish-npm must not create token-oriented registry authentication',
    ),
  );
});

test('rejects a repository URL that cannot match the npm trust policy', () => {
  const errors = verifyNpmPublishing(validWorkflow, {
    repository: 'https://github.com/example/fork.git',
  });

  assert.ok(errors.some((error) => error.startsWith('npm repository URL')));
});

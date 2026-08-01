#!/usr/bin/env node

import {readFileSync, writeFileSync} from 'node:fs';

export function validatePullRequestIdentity({
  event,
  current,
  repository,
  pullRequestNumber,
  headRepository,
  headSha,
}) {
  const errors = [];
  const eventPullRequest = event.pull_request;
  const expectedNumber = Number(pullRequestNumber);
  const checks = [
    ['event repository', event.repository?.full_name, repository],
    ['API base repository', current.base?.repo?.full_name, repository],
    ['event pull request number', eventPullRequest?.number, expectedNumber],
    ['API pull request number', current.number, expectedNumber],
    ['event head repository', eventPullRequest?.head?.repo?.full_name, headRepository],
    ['API head repository', current.head?.repo?.full_name, headRepository],
    ['event head SHA', eventPullRequest?.head?.sha, headSha],
    ['API head SHA', current.head?.sha, headSha],
  ];
  for (const [label, actual, expected] of checks) {
    if (!expected || actual !== expected) {
      errors.push(`${label} does not match the trusted event value`);
    }
  }
  return errors;
}

export function buildAuthenticatedEvent({event, current, files, trusted}) {
  const errors = validatePullRequestIdentity({
    event,
    current,
    ...trusted,
  });
  if (errors.length > 0) {
    throw new Error(errors.join('\n'));
  }
  if (!Array.isArray(files) || files.some((file) => typeof file !== 'string')) {
    throw new Error('changed files must be an array of strings');
  }
  return {
    pull_request: {...current, changed_files_list: files},
  };
}

function parseArguments(argv) {
  const values = {};
  for (let index = 0; index < argv.length; index += 2) {
    const key = argv[index]?.replace(/^--/, '');
    const value = argv[index + 1];
    if (!key || value === undefined) throw new Error('invalid arguments');
    values[key] = value;
  }
  return values;
}

export function main(argv = process.argv.slice(2), env = process.env) {
  try {
    const args = parseArguments(argv);
    const authenticated = buildAuthenticatedEvent({
      event: JSON.parse(readFileSync(args.event, 'utf8')),
      current: JSON.parse(readFileSync(args.current, 'utf8')),
      files: JSON.parse(readFileSync(args.files, 'utf8')),
      trusted: {
        repository: env.GITHUB_REPOSITORY,
        pullRequestNumber: env.PR_NUMBER,
        headRepository: env.PR_HEAD_REPOSITORY,
        headSha: env.PR_HEAD_SHA,
      },
    });
    writeFileSync(args.output, `${JSON.stringify(authenticated)}\n`, {
      mode: 0o600,
    });
    return 0;
  } catch (error) {
    console.error(error.message);
    return 1;
  }
}

if (import.meta.url === `file://${process.argv[1]}`) {
  process.exitCode = main();
}

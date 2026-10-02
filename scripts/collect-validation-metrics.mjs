#!/usr/bin/env node

import {execFileSync} from 'node:child_process';
import {resolve} from 'node:path';
import {fileURLToPath} from 'node:url';

const GROUPS = {
  quick: {gate: 'CI Quick Gate', prefixes: ['Quick validation / ']},
  full: {
    gate: 'Validation Gate',
    prefixes: ['Full repository validation / ', 'Full runtime-template validation / ', 'Security validation / '],
  },
};

function summarize(jobs, definition) {
  const selected = jobs.filter(job => job.name === definition.gate ||
    definition.prefixes.some(prefix => job.name.startsWith(prefix)));
  const gate = selected.find(job => job.name === definition.gate);
  const failures = new Set(['failure', 'timed_out', 'startup_failure', 'action_required']);
  const failed = selected.some(job => failures.has(job.conclusion));
  const missingGroup = definition.prefixes.some(prefix => !selected.some(job => job.name.startsWith(prefix)));
  const incomplete = !gate || missingGroup || selected.some(job => job.status !== 'completed' || job.conclusion === 'cancelled');
  const outcome = failed ? 'fail' : !incomplete && gate.conclusion === 'success' ? 'pass' : 'incomplete';
  const intervals = selected.filter(job => job.conclusion !== 'skipped').map(job => {
    const start = Date.parse(job.started_at);
    const end = Date.parse(job.completed_at);
    return Number.isFinite(start) && Number.isFinite(end) && end >= start ? [start, end] : null;
  });
  const timed = intervals.filter(Boolean);
  const timingComplete = intervals.length > 0 && timed.length === intervals.length;
  return {
    outcome,
    jobCount: selected.length,
    executedJobCount: intervals.length,
    timingComplete,
    runnerSeconds: timingComplete ? timed.reduce((sum, [start, end]) => sum + (end - start) / 1000, 0) : null,
    criticalSeconds: timingComplete ? (Math.max(...timed.map(([, end]) => end)) - Math.min(...timed.map(([start]) => start))) / 1000 : null,
    jobs: selected.map(({id, name, status, conclusion, started_at, completed_at}) =>
      ({id, name, status, conclusion, started_at, completed_at})),
  };
}

export function summarizeValidation(run, jobs) {
  if (run.name !== 'Validation' || run.status !== 'completed') {
    throw new Error('Metrics require a completed Validation run.');
  }
  if (new Set(jobs.map(job => job.id)).size !== jobs.length ||
      jobs.some(job => job.run_id !== run.id || job.run_attempt !== run.run_attempt)) {
    throw new Error('Jobs must belong to one run attempt without duplicates.');
  }
  const quick = summarize(jobs, GROUPS.quick);
  const full = summarize(jobs, GROUPS.full);
  const complete = quick.outcome !== 'incomplete' && full.outcome !== 'incomplete';
  return {
    schemaVersion: 1,
    run: {id: run.id, attempt: run.run_attempt, url: run.html_url, event: run.event,
      headSha: run.head_sha, createdAt: run.created_at, conclusion: run.conclusion,
      pullRequests: run.pull_requests ?? []},
    quick,
    full,
    result: `quick-${quick.outcome}/full-${full.outcome}`,
    possibleMiss: complete ? quick.outcome === 'pass' && full.outcome === 'fail' : null,
    criticalRatio: complete && full.criticalSeconds > 0 && quick.criticalSeconds !== null
      ? quick.criticalSeconds / full.criticalSeconds : null,
    runnerReduction: complete && full.runnerSeconds > 0 && quick.runnerSeconds !== null
      ? 1 - quick.runnerSeconds / full.runnerSeconds : null,
    // Jobs API does not expose cache hits or prove changed-scope relevance.
    cacheHitRate: null,
    relevantJobCount: null,
    unclassifiedJobCount: jobs.length - quick.jobCount - full.jobCount,
    qualifiesForGateTransition: false,
  };
}

export function collectValidation(repository, runID, attempt, api) {
  if (!/^[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+$/.test(repository) ||
      !/^[1-9][0-9]*$/.test(String(runID)) || !/^[1-9][0-9]*$/.test(String(attempt))) {
    throw new Error('Expected owner/repository, numeric run ID, and numeric attempt.');
  }
  const path = `repos/${repository}/actions/runs/${runID}/attempts/${attempt}`;
  const run = api(path);
  if (String(run.id) !== String(runID) || String(run.run_attempt) !== String(attempt)) {
    throw new Error('API returned a different run attempt.');
  }
  const jobs = [];
  let total;
  for (let page = 1; ; page++) {
    const response = api(`${path}/jobs?per_page=100&page=${page}`);
    if (!Array.isArray(response.jobs) || !Number.isInteger(response.total_count) || response.total_count < 0 ||
        (total !== undefined && total !== response.total_count)) {
      throw new Error('Invalid or changing Jobs API pagination.');
    }
    total = response.total_count;
    jobs.push(...response.jobs);
    if (jobs.length >= total) break;
    if (response.jobs.length === 0) throw new Error('Incomplete Jobs API pagination.');
  }
  if (jobs.length !== total) throw new Error('Unexpected Jobs API count.');
  return summarizeValidation(run, jobs);
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    const [repository, runID, attempt, ...extra] = process.argv.slice(2);
    if (extra.length || !repository || !runID || !attempt) {
      throw new Error('Usage: node scripts/collect-validation-metrics.mjs OWNER/REPO RUN_ID ATTEMPT');
    }
    const report = collectValidation(repository, runID, attempt, path => JSON.parse(
      execFileSync('gh', ['api', path], {encoding: 'utf8', maxBuffer: 20 * 1024 * 1024}),
    ));
    console.log(JSON.stringify(report, null, 2));
  } catch (error) {
    console.error(error.message);
    process.exitCode = 1;
  }
}

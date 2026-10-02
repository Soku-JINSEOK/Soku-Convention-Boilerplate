import assert from 'node:assert/strict';
import test from 'node:test';
import {collectValidation, summarizeValidation} from './collect-validation-metrics.mjs';

const run = {id: 42, run_attempt: 2, name: 'Validation', status: 'completed', head_sha: 'abc'};
const job = (id, name, seconds, conclusion = 'success') => ({
  id, name, run_id: 42, run_attempt: 2, status: 'completed', conclusion,
  started_at: '2026-10-01T00:00:00Z', completed_at: `2026-10-01T00:00:${String(seconds).padStart(2, '0')}Z`,
});
const jobs = () => [job(1, 'Quick validation / Quick / Soku', 10), job(2, 'CI Quick Gate', 1),
  job(3, 'Full repository validation / Soku', 20), job(4, 'Validation Gate', 1),
  job(5, 'Full runtime-template validation / Python', 10), job(6, 'Security validation / Audit', 10)];

test('sums parallel job time separately from critical duration', () => {
  const report = summarizeValidation(run, jobs());
  assert.equal(report.result, 'quick-pass/full-pass');
  assert.equal(report.quick.runnerSeconds, 11);
  assert.equal(report.full.criticalSeconds, 20);
  assert.equal(report.criticalRatio, 0.5);
  assert.equal(report.runnerReduction, 1 - 11 / 41);
  assert.equal(report.cacheHitRate, null);
  assert.equal(report.qualifiesForGateTransition, false);
});

test('retains Quick pass / Full fail misses and both-fail attempts', () => {
  const failed = jobs();
  failed[2].conclusion = 'failure';
  failed[3].conclusion = 'failure';
  assert.equal(summarizeValidation(run, failed).possibleMiss, true);
  failed[0].conclusion = 'failure';
  assert.equal(summarizeValidation(run, failed).result, 'quick-fail/full-fail');
});

test('cancelled or absent gate is incomplete, missing times are not zero', () => {
  const cancelled = jobs();
  cancelled[0].conclusion = 'cancelled';
  cancelled[0].completed_at = null;
  const report = summarizeValidation(run, cancelled);
  assert.equal(report.possibleMiss, null);
  assert.equal(report.quick.runnerSeconds, null);
  assert.equal(report.criticalRatio, null);
  assert.equal(summarizeValidation(run, jobs().slice(0, 3)).full.outcome, 'incomplete');
});

test('rejects mixed attempts and duplicated jobs', () => {
  const mixed = jobs();
  mixed[0].run_attempt = 1;
  assert.throws(() => summarizeValidation(run, mixed), /one run attempt/);
  assert.throws(() => summarizeValidation(run, [...jobs(), jobs()[0]]), /duplicates/);
});

test('fetches every page for the requested attempt', () => {
  const calls = [];
  const report = collectValidation('owner/repo', 42, 2, path => {
    calls.push(path);
    if (!path.includes('/jobs?')) return run;
    return {total_count: 6, jobs: path.endsWith('page=1') ? jobs().slice(0, 2) : jobs().slice(2)};
  });
  assert.equal(calls.length, 3);
  assert.ok(calls.every(path => path.includes('/attempts/2')));
  assert.equal(report.full.jobCount, 4);
  assert.throws(() => collectValidation('owner/repo', '42;echo', 2, () => {}), /Expected/);
});

test('a renamed or absent group cannot silently improve the comparison', () => {
  const renamed = jobs();
  renamed[4].name = 'Renamed template job';
  const report = summarizeValidation(run, renamed);
  assert.equal(report.full.outcome, 'incomplete');
  assert.equal(report.unclassifiedJobCount, 1);
  assert.equal(report.runnerReduction, null);
});

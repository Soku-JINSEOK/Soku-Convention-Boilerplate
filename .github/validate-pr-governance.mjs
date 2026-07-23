#!/usr/bin/env node
import {runPullRequestPolicy} from '../scripts/pull-request-policy.mjs';

process.exitCode = runPullRequestPolicy();

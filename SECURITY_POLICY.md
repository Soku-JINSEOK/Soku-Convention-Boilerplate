# Security Policy

## Purpose

This document defines the baseline security posture for repositories based on `Soku-Convention-Boilerplate`.

Security should be treated as an operating concern from the beginning, not as a later compliance add-on.

## Security Principles

Repository security should prioritize:

- least privilege
- explicit access boundaries
- safe defaults
- traceability
- fast remediation

## Minimum Expectations

At minimum, repositories should:

- avoid committing secrets
- document sensitive configuration handling
- use environment-specific credentials
- keep dependencies reviewable
- define a response path for security issues

## Secret Management

Secrets must not be stored directly in source control.

Use:

- environment variables
- cloud secret managers
- CI platform secret stores
- documented local development overrides that are excluded from version control

## Access Control

Access to infrastructure, production systems, and deployment workflows should follow least-privilege principles.

Prefer:

- role-based access
- scoped service accounts
- short-lived credentials where possible
- auditable permission changes

## Dependency Hygiene

Dependencies should be reviewed with security and maintenance in mind.

Projects should aim to:

- avoid abandoned packages
- update vulnerable dependencies promptly
- keep transitive risk visible
- document exceptions when upgrades are delayed

## Logging and Sensitive Data

Logs should be useful for diagnosis without leaking secrets or regulated information.

Avoid logging:

- access tokens
- passwords
- connection strings
- private keys
- personal or regulated data unless explicitly required and protected

## CI/CD Security

Pipelines should:

- protect secrets from untrusted contexts
- avoid over-privileged automation tokens
- separate validation from production deployment where appropriate
- keep deployment approval paths explicit

## Reporting and Remediation

Repositories should define how security issues are handled, including:

- where to report them
- who reviews them
- how severity is assessed
- how remediation is tracked

## Summary

A strong security policy does not require heavy ceremony.  
It requires consistent operational discipline, safe defaults, and clear responsibility boundaries.

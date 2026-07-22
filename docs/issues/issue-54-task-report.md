# Issue #54 Task Report — Public Provider API Mirror

## Scope

Publish the reviewed `ci-cd-control-plane-v1` declarative bundle and composite
action in the public Boilerplate repository without exposing private
control-plane access, credentials, executable provider code, or mutable refs.

The first publication phase deliberately omitted the caller workflow. Its
squash merge produced commit
`c5435ea36d88dbe3b4b2c373265206943c53fcbf`; the second phase pins the caller
to that immutable commit and marks the public mirror as published.

## Delivered contract

- The public bundle contains only strict provider metadata, configuration
  schema, reviewed example configuration, and one declared literal template.
- The external provenance ledger binds the reviewed control-plane baseline,
  the three merged delivery commits for Issues #23–#25, and every public file
  by raw-byte SHA-256.
- The composite action accepts a safe workspace-relative YAML path, the exact
  public source, and a lowercase 40-character commit ref. It emits only the
  three `soku --integration-*` values.
- Go conformance tests load the public bundle through the production decoder,
  reject unknown files, verify raw-byte provenance, and initialize a downstream
  repository using the published provider contract.
- Python tests reject mutable or uppercase refs, unsafe paths, unknown sources,
  symlinks, non-UTF-8 input, and literal-byte tampering.
- The public caller invokes the immutable action and passes only its three
  bounded outputs to the corresponding `soku --integration-*` arguments.
- Caller conformance rejects mutable action refs, extra integration arguments,
  secret-like inputs, private access, network fetches, and remote code execution.

## Verification

- `go test ./internal/initcmd ./internal/lifecyclee2e`
- `python3 -m unittest discover -s soku/actions -p 'test_*.py'`
- Provider and configuration JSON schema validation
- Raw-byte SHA-256 ledger verification
- Exact-ref caller conformance and prohibited-capability checks
- Markdown, YAML, actionlint, repository hygiene, and `git diff --check`
- Hosted full Validation and PR Metadata gates before merge

## Security and delivery boundary

The action does not fetch private repositories, receive secrets, execute remote
provider code, enable delivery, or create cloud resources. The caller is a
copyable example under `docs/callers/`, not an enabled repository workflow.
Merge, Project status changes, Issue closure, and delivery remain separate
approval boundaries.

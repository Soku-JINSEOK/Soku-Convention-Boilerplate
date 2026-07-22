# Issue #54 Task Report — Public Provider API Mirror

## Scope

Publish the reviewed `ci-cd-control-plane-v1` declarative bundle and composite
action in the public Boilerplate repository without exposing private
control-plane access, credentials, executable provider code, or mutable refs.

The first publication phase deliberately omitted the caller workflow. Its
immutable merge commit is now recorded in the provenance ledger and used as
the exact action revision in the separately reviewed caller.

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
- The manual caller uses read-only contents permission, consumes no secret, and
  pins the public action to `c5435ea36d88dbe3b4b2c373265206943c53fcbf`.

## Verification

- `go test ./internal/initcmd ./internal/lifecyclee2e`
- `python3 -m unittest discover -s soku/actions -p 'test_*.py'`
- Provider and configuration JSON schema validation
- Raw-byte SHA-256 ledger verification
- Markdown, YAML, actionlint, repository hygiene, and `git diff --check`
- Hosted full Validation and PR Metadata gates before merge

## Security and delivery boundary

The action and caller do not fetch private repositories, receive secrets,
execute remote provider code, enable delivery, or create cloud resources. The
caller is manual and read-only, and the provenance ledger keeps
`delivery_enabled` false.

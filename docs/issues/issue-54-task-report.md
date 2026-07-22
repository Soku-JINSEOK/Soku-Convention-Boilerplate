# Issue #54 Task Report — Public Provider API Mirror

## Scope

Publish the reviewed `ci-cd-control-plane-v1` declarative bundle and composite
action in the public Boilerplate repository without exposing private
control-plane access, credentials, executable provider code, or mutable refs.

This first publication phase deliberately omits the caller workflow. After this
PR merges, a second PR will pin that caller to this phase's immutable full
commit and will replace the pending mirror entry in the provenance ledger.

## Delivered contract

- The public bundle contains only strict provider metadata, configuration
  schema, reviewed example configuration, and one declared literal template.
- The external provenance ledger binds the reviewed control-plane baseline,
  bundle, public-contract review commits, and every public file by raw-byte
  SHA-256.
- The composite action accepts a safe workspace-relative YAML path, the exact
  public source, and a lowercase 40-character commit ref. It emits only the
  three `soku --integration-*` values.
- Go conformance tests load the public bundle through the production decoder,
  reject unknown files, verify raw-byte provenance, and initialize a downstream
  repository using the published provider contract.
- Python tests reject mutable or uppercase refs, unsafe paths, unknown sources,
  symlinks, non-UTF-8 input, and literal-byte tampering.

## Verification

- `go test ./internal/initcmd ./internal/lifecyclee2e`
- `python3 -m unittest discover -s soku/actions -p 'test_*.py'`
- Provider and configuration JSON schema validation
- Raw-byte SHA-256 ledger verification
- Markdown, YAML, actionlint, repository hygiene, and `git diff --check`
- Hosted full Validation and PR Metadata gates before merge

## Security and delivery boundary

The action does not fetch private repositories, receive secrets, execute remote
provider code, enable delivery, or create cloud resources. The public caller is
not valid until it is added in phase two with the immutable merge commit from
this phase. Merge, Project status changes, Issue closure, and delivery remain
separate approval boundaries.

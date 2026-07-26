# Release Signer Rotations

This append-only record defines the approved primary GPG fingerprint used to
sign new Boilerplate and `soku` release tags. The current value must match
`signing.activeFingerprint` in `release-identity.json`.

Current active fingerprint: `03944489C01275035F9D68049A359FC72B404DFC`

## Rotation Procedure

1. Add the replacement public key to the maintainer's GitHub account and verify
   its full 40-character primary fingerprint independently.
2. Add one row to the end of the history. Never edit, reorder, or remove an
   existing row.
3. Record the previous and replacement fingerprints, the effective source
   boundary, the reason, and the verification result.
4. In the same reviewed pull request, update
   `release-identity.json`, run the release-identity and signed-tag regression
   tests, and obtain the required review.
5. Merge that pull request before using the replacement signer. The effective
   commit is the first merged commit whose tree contains both the new row and
   the matching active fingerprint.
6. Keep old public keys available for historical tag verification. Removing an
   old private key does not authorize rewriting an existing public tag.

## History

| Previous fingerprint | New fingerprint | Effective source boundary | Reason | Verification result |
| --- | --- | --- | --- | --- |
| none | `03944489C01275035F9D68049A359FC72B404DFC` | First merged commit containing this row and the matching release identity | Establish the reviewed publication signer for #123 | Local secret-key and GitHub public-key fingerprints matched on 2026-07-26; approved and unapproved generated-key regression cases pass |

# Product user manual

This file is a core-managed structural example. Create project-owned
`docs/manual/USAGE.md`, `USAGE.ko.md`, and `USAGE.ja.md` for actual prose.

Reference stable capture IDs from `generated-index.md`. Do not edit generated
PNG files, the generated index, or `capture-report.json` manually.

Every caption must disclose documentation-only dialog overlays and any map
provider substitution. A successful doctor or capture run is technical
evidence, not a legal, licensing, privacy, or provider-policy compliance claim.

Run a reviewed local capture from the repository root with:

```bash
node tools/manual-capture/dist/cli.js capture \
  --config docs/manual/capture.yml
```

Use `--allow-dirty` only when recording the dirty worktree state is deliberate.

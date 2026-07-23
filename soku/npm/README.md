# `@soku-jinseok/soku`

Cross-platform launcher for the native `soku` CLI distributed from
[`Soku-Convention-Boilerplate`](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate).

This package is the npm installation path for `soku/v0.2.0` and later. The package
downloads the matching GitHub release asset at first run, verifies it with
`checksums.txt`, caches the native executable, and executes it with your provided
arguments.

## Install

```bash
npm install -g @soku-jinseok/soku@0.2.0
```

## Usage

```bash
soku --version
soku init --boilerplate-source https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate --help
```

The launcher accepts optional overrides for validation environments:

```bash
SOKU_GITHUB_REPOSITORY="owner/repo" soku --version
SOKU_LAUNCHER=1 soku status
```

## Verification

- Release assets are verified through the `checksums.txt` entry that matches the
  platform-specific archive in the target tag.
- Unsupported OS/arch combinations fail quickly with a clear message.
- Cache is stored under:

  - macOS/Linux: `~/.cache/soku/Soku-JINSEOK_Soku-Convention-Boilerplate/soku/vX.Y.Z/<os>/<arch>/`
  - Windows: `%USERPROFILE%\.cache\soku\...`

No external runtime dependency is required beyond a supported Node.js runtime.

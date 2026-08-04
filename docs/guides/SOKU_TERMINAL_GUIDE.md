# Soku Terminal and Completion Guide

[English](./SOKU_TERMINAL_GUIDE.md) | [한국어](./SOKU_TERMINAL_GUIDE.ko.md) | [日本語](./SOKU_TERMINAL_GUIDE.ja.md)

This guide explains Soku's terminal output, a safe daily workflow, automation,
and completion for Bash, Zsh, Fish, and PowerShell. Soku only prints completion
scripts. It never edits a shell profile, changes the default shell, or invokes a
plugin manager.

## Output and Color

`--color` accepts `auto` (the default), `always`, or `never`.

- `auto` uses ANSI styling only when stdout is a TTY, `TERM` is not `dumb`, and
  `NO_COLOR` is unset.
- `always` explicitly enables styling, including in a pipe, and overrides
  `NO_COLOR` and `TERM=dumb`.
- `never` always emits plain text.

Piped output is plain in `auto` mode. JSON envelopes, quiet output, prompts,
errors, and generated completion scripts do not receive color. Use `--json`
for a stable machine-readable envelope and `--quiet` when only the exit code is
needed.

```bash
soku status                         # color only on a capable TTY
soku status --color=never | less    # stable plain text
NO_COLOR=1 soku status              # disable automatic color
TERM=dumb soku status               # disable automatic color
soku status --color=always | less -R
soku status --json | jq '.data'
soku status --quiet
```

## Daily Personal Workflow

Inspect before changing managed files. Replace the example release with the
exact immutable release you intend to adopt.

```bash
soku status
soku diff --boilerplate-release v1.1.0
soku upgrade --boilerplate-release v1.1.0 --dry-run
# Review the report and repository diff, then apply deliberately:
soku upgrade --boilerplate-release v1.1.0 --yes
```

`status`, `diff`, and `--dry-run` do not apply managed-file changes. A real
upgrade requires explicit confirmation according to the lifecycle contract.

## Load Completion for One Session

Run the command for the shell that is already active:

```bash
# Bash
source <(soku completion bash)

# Zsh
autoload -Uz compinit && compinit
source <(soku completion zsh)

# Fish
soku completion fish | source
```

```powershell
# PowerShell
soku completion powershell | Out-String | Invoke-Expression
```

Try `soku <Tab>`, `soku docs <Tab>`, `soku docs manual <Tab>`,
`soku --color <Tab>`, or `soku init --profile <Tab>`. Candidates include
`docs manual`, `completion`, color modes, and `bootstrap`, `standard`, and
`scaled` profiles.

## Install in a User-Owned Path

Generate the script first, then connect it from your own profile. These commands
do not need administrator access.

### Bash

```bash
mkdir -p ~/.local/share/soku/completions
soku completion bash > ~/.local/share/soku/completions/soku.bash
printf '%s\n' 'source "$HOME/.local/share/soku/completions/soku.bash"' >> ~/.bashrc
source ~/.bashrc
```

Remove the `source` line from `~/.bashrc`, then remove
`~/.local/share/soku/completions/soku.bash` to uninstall it.

### Zsh

```zsh
mkdir -p ~/.zfunc
soku completion zsh > ~/.zfunc/_soku
```

Add the following lines to `~/.zshrc` before any existing `compinit` call (do
not add a second `compinit` when your setup already has one), then start a new
shell:

```zsh
fpath=(~/.zfunc $fpath)
autoload -Uz compinit && compinit
```

Remove the `fpath` line, remove the added `compinit` line only if you added it
for Soku, and remove `~/.zfunc/_soku` to uninstall.

### Fish

```fish
mkdir -p ~/.config/fish/completions
soku completion fish > ~/.config/fish/completions/soku.fish
```

Fish discovers that file automatically. Remove it to uninstall.

### PowerShell

```powershell
$completionDirectory = Join-Path $HOME ".config/soku/completions"
$completionFile = Join-Path $completionDirectory "soku.ps1"
New-Item -ItemType Directory -Force $completionDirectory | Out-Null
soku completion powershell | Set-Content $completionFile
Add-Content $PROFILE '. "$HOME/.config/soku/completions/soku.ps1"'
. $PROFILE
```

Remove the dot-source line from `$PROFILE`, then remove `$completionFile` to
uninstall. Soku does not create or edit `$PROFILE` itself.

## Automation Examples

```bash
# Consume JSON without parsing human text.
soku status --json | jq -e '.ok and (.command == "status")'

# Use only the documented exit code in a scheduled check.
if soku status --quiet; then
  echo "Soku state is clean"
else
  code=$?
  echo "Soku status exited with $code" >&2
fi

# Cache a deterministic script as an artifact and verify it has no ANSI bytes.
soku completion fish > soku.fish
LC_ALL=C grep -q $'\033' soku.fish && echo "unexpected ANSI" >&2
```

Completion candidates are local hints, not a substitute for validation. Scripts
contain no repository lookup or network result and can be regenerated whenever
the Soku binary changes.

## Troubleshooting

- Confirm the intended binary is first on `PATH` with `command -v soku` (or
  `Get-Command soku` in PowerShell), then check `soku --version`.
- Regenerate a saved completion script after upgrading Soku and start a new
  shell. Old scripts may still expose an older command tree.
- Bash requires a compatible `bash-completion` package. Load that framework
  before sourcing the generated file; macOS's system Bash does not provide all
  helpers by itself. Check registration with `complete -p soku`.
- Zsh must have the directory containing `_soku` in `fpath` before `compinit`.
  Remove a stale cache such as `~/.zcompdump*` only when your Zsh setup permits
  it, then run `compinit` again.
- Fish should place the file at
  `~/.config/fish/completions/soku.fish`. Test candidates with
  `complete -C 'soku '`.
- PowerShell must execute the generated script in the current session. Check
  `$PROFILE`, execution policy, and `Get-Command soku`; then reload with
  `. $PROFILE`. Test with `TabExpansion2 'soku ' 5`.
- If completion works in a clean shell but not in your normal one, inspect the
  ordering or cache behavior of your completion/plugin manager. Soku does not
  manage that tool.

## Safety Boundary

`soku completion` writes only to stdout unless you explicitly redirect it.
Soku does not modify Bash, Zsh, Fish, or PowerShell profiles; does not change
the login shell; and does not install, configure, or remove plugin managers.
Review generated scripts before choosing to persist them, and remove only the
profile line and user-owned file that you added.

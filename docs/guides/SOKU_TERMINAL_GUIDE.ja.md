# Soku ターミナルおよび Completion ガイド

[English](./SOKU_TERMINAL_GUIDE.md) | [한국어](./SOKU_TERMINAL_GUIDE.ko.md) | [日本語](./SOKU_TERMINAL_GUIDE.ja.md)

このガイドでは、Soku のターミナル出力、安全な日常フロー、自動化、Bash・Zsh・
Fish・PowerShell の completion を説明します。Soku は completion script を
出力するだけです。shell profile の編集、default shell の変更、plugin manager
の実行は行いません。

## 出力と色

`--color` には `auto`（デフォルト）、`always`、`never` を指定できます。

- `auto` は stdout が TTY、`TERM` が `dumb` ではなく、`NO_COLOR` が未設定の
  場合だけ ANSI style を使用します。
- `always` は pipe でも明示的に style を有効にし、`NO_COLOR` と
  `TERM=dumb` より優先されます。
- `never` は常に plain text を出力します。

Pipe 出力は `auto` では plain text です。JSON envelope、quiet 出力、prompt、
error、生成された completion script に色は付きません。安定した機械可読形式には
`--json`、exit code だけが必要な場合は `--quiet` を使用してください。

```bash
soku status                         # 対応 TTY だけで色を使用
soku status --color=never | less    # 安定した plain text
NO_COLOR=1 soku status              # 自動色付けを無効化
TERM=dumb soku status               # 自動色付けを無効化
soku status --color=always | less -R
soku status --json | jq '.data'
soku status --quiet
```

## 個人利用の日常フロー

管理対象ファイルを変更する前に検査してください。例の release は、導入する正確な
immutable release に置き換えてください。

```bash
soku status
soku diff --boilerplate-release v1.1.0
soku upgrade --boilerplate-release v1.1.0 --dry-run
# レポートとリポジトリの diff を確認してから、明示的に適用します。
soku upgrade --boilerplate-release v1.1.0 --yes
```

`status`、`diff`、`--dry-run` は管理対象ファイルの変更を適用しません。実際の
upgrade には lifecycle contract に従った明示的な確認が必要です。

## 1 セッションで Completion を読み込む

現在利用中の shell に対応するコマンドを実行してください。

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

`soku <Tab>`、`soku docs <Tab>`、`soku docs manual <Tab>`、
`soku --color <Tab>`、`soku init --profile <Tab>` を試してください。候補には
`docs manual`、`completion`、color mode、`bootstrap`、`standard`、
`scaled` profile が含まれます。

## ユーザー所有パスへのインストール

最初に script を生成し、自分で profile から接続してください。次のコマンドに
管理者権限は必要ありません。

### Bash

```bash
mkdir -p ~/.local/share/soku/completions
soku completion bash > ~/.local/share/soku/completions/soku.bash
printf '%s\n' 'source "$HOME/.local/share/soku/completions/soku.bash"' >> ~/.bashrc
source ~/.bashrc
```

削除するには `~/.bashrc` の `source` 行を消し、
`~/.local/share/soku/completions/soku.bash` を削除してください。

### Zsh

```zsh
mkdir -p ~/.zfunc
soku completion zsh > ~/.zfunc/_soku
```

既存の `compinit` 呼び出しより前に次の行を `~/.zshrc` に追加し、新しい shell を
起動してください。すでに `compinit` がある場合は二重に追加しないでください。

```zsh
fpath=(~/.zfunc $fpath)
autoload -Uz compinit && compinit
```

削除するには `fpath` 行を消し、Soku のために追加した場合だけ `compinit` 行を
消してから `~/.zfunc/_soku` を削除してください。

### Fish

```fish
mkdir -p ~/.config/fish/completions
soku completion fish > ~/.config/fish/completions/soku.fish
```

Fish はこのファイルを自動検出します。削除するにはファイルを消してください。

### PowerShell

```powershell
$completionDirectory = Join-Path $HOME ".config/soku/completions"
$completionFile = Join-Path $completionDirectory "soku.ps1"
New-Item -ItemType Directory -Force $completionDirectory | Out-Null
soku completion powershell | Set-Content $completionFile
Add-Content $PROFILE '. "$HOME/.config/soku/completions/soku.ps1"'
. $PROFILE
```

削除するには `$PROFILE` の dot-source 行を消し、`$completionFile` を削除して
ください。Soku 自体は `$PROFILE` を作成・編集しません。

## 自動化の例

```bash
# human text を parse せず JSON を利用します。
soku status --json | jq -e '.ok and (.command == "status")'

# 定期チェックでは文書化された exit code だけを利用します。
if soku status --quiet; then
  echo "Soku state is clean"
else
  code=$?
  echo "Soku status exited with $code" >&2
fi

# 決定的な script を artifact として保存し、ANSI byte がないことを確認します。
soku completion fish > soku.fish
LC_ALL=C grep -q $'\033' soku.fish && echo "unexpected ANSI" >&2
```

Completion 候補はローカルのヒントであり validation の代わりではありません。
Script にリポジトリ参照や network の結果は含まれず、Soku binary の変更時に
いつでも再生成できます。

## トラブルシューティング

- `command -v soku`（PowerShell は `Get-Command soku`）で意図した binary が
  `PATH` の先頭にあることを確認し、`soku --version` を確認してください。
- Soku の upgrade 後は保存済み completion script を再生成し、新しい shell を
  起動してください。古い script には以前の command tree が残る場合があります。
- Bash には互換性のある `bash-completion` package が必要です。macOS の system
  Bash だけではすべての helper が提供されないため、framework を先に読み込んで
  から生成ファイルを source してください。`complete -p soku` で登録を確認して
  ください。
- Zsh は `compinit` より前に `_soku` の directory を `fpath` に含める必要が
  あります。Zsh 設定で許される場合だけ `~/.zcompdump*` などの古い cache を
  削除し、`compinit` を再実行してください。
- Fish は `~/.config/fish/completions/soku.fish` にファイルを置きます。
  `complete -C 'soku '` で候補を確認できます。
- PowerShell は現在の session で生成 script を実行する必要があります。
  `$PROFILE`、execution policy、`Get-Command soku` を確認し、`. $PROFILE` で
  再読み込みしてください。`TabExpansion2 'soku ' 5` で確認できます。
- clean shell では動き、通常の shell では動かない場合、completion/plugin
  manager の順序や cache を確認してください。Soku はそのツールを管理しません。

## 安全境界

`soku completion` はユーザーが明示的に redirect しない限り stdout だけに
書き込みます。Soku は Bash、Zsh、Fish、PowerShell profile を変更せず、login
shell を変更せず、plugin manager をインストール・設定・削除しません。永続化を
選ぶ前に生成 script を確認し、削除時は自分が追加した profile 行とユーザー所有
ファイルだけを削除してください。

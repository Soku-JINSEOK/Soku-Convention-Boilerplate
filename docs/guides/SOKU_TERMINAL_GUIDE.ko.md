# Soku 터미널 및 Completion 가이드

[English](./SOKU_TERMINAL_GUIDE.md) | [한국어](./SOKU_TERMINAL_GUIDE.ko.md) | [日本語](./SOKU_TERMINAL_GUIDE.ja.md)

이 가이드는 Soku의 터미널 출력, 안전한 일상 흐름, 자동화, Bash·Zsh·Fish·
PowerShell completion을 설명합니다. Soku는 completion script만 출력합니다.
shell profile을 편집하거나 기본 shell을 변경하거나 plugin manager를 실행하지 않습니다.

## 출력과 색상

`--color`는 `auto`(기본값), `always`, `never`를 받습니다.

- `auto`는 stdout이 TTY이고 `TERM`이 `dumb`가 아니며 `NO_COLOR`가 설정되지
  않은 경우에만 ANSI style을 사용합니다.
- `always`는 pipe에서도 style을 명시적으로 사용하며 `NO_COLOR`와
  `TERM=dumb`보다 우선합니다.
- `never`는 항상 plain text를 출력합니다.

Pipe 출력은 `auto`에서 plain text입니다. JSON envelope, quiet 출력, prompt,
error, 생성된 completion script에는 색상이 적용되지 않습니다. 안정적인 기계 판독
형식은 `--json`, exit code만 필요할 때는 `--quiet`를 사용하십시오.

```bash
soku status                         # 지원되는 TTY에서만 색상 사용
soku status --color=never | less    # 안정적인 plain text
NO_COLOR=1 soku status              # 자동 색상 비활성화
TERM=dumb soku status               # 자동 색상 비활성화
soku status --color=always | less -R
soku status --json | jq '.data'
soku status --quiet
```

## 개인용 일상 흐름

관리 파일을 변경하기 전에 먼저 검사하십시오. 예제 release는 실제 도입하려는 정확한
immutable release로 바꾸십시오.

```bash
soku status
soku diff --boilerplate-release v1.1.0
soku upgrade --boilerplate-release v1.1.0 --dry-run
# 보고서와 저장소 diff를 검토한 뒤 의도적으로 적용합니다.
soku upgrade --boilerplate-release v1.1.0 --yes
```

`status`, `diff`, `--dry-run`은 관리 파일 변경을 적용하지 않습니다. 실제 upgrade는
lifecycle 계약에 따른 명시적 확인이 필요합니다.

## 검토된 Core 파일 하나 Handoff

현재 core-managed 파일에 검토된 의도적인 project 변경이 있을 때 normalized
SHA-256을 계산한 뒤 manifest-only handoff 계획을 검토합니다.

```bash
soku ownership handoff \
  --path .prettierignore \
  --expected-sha256 <lowercase-64-character-sha256> \
  --dry-run
```

정확한 계획을 `--yes` 또는 interactive confirmation으로 적용합니다. Command는
canonical path 하나만 허용하고 해당 파일의 bytes나 mode를 변경하지 않으며 향후
core rendering 억제를 manifest v3에 기록합니다. Clean, stale, missing, symlink,
mergeable, provider-managed, project-owned, repeated path는 write 전에 거부합니다.

## 한 세션에서 Completion 로드

현재 사용 중인 shell에 해당하는 명령을 실행하십시오.

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

`soku <Tab>`, `soku ownership <Tab>`, `soku docs <Tab>`, `soku docs manual <Tab>`,
`soku --color <Tab>`, `soku init --profile <Tab>`을 사용해 보십시오. 후보에는
`ownership handoff`, `docs manual`, `completion`, color mode와 `bootstrap`, `standard`, `scaled`
profile이 포함됩니다.

## 사용자 소유 경로에 설치

먼저 script를 생성한 뒤 직접 profile에서 연결하십시오. 아래 명령에는 관리자 권한이
필요하지 않습니다.

### Bash

```bash
mkdir -p ~/.local/share/soku/completions
soku completion bash > ~/.local/share/soku/completions/soku.bash
printf '%s\n' 'source "$HOME/.local/share/soku/completions/soku.bash"' >> ~/.bashrc
source ~/.bashrc
```

제거하려면 `~/.bashrc`의 `source` 줄을 지운 뒤
`~/.local/share/soku/completions/soku.bash`를 삭제하십시오.

### Zsh

```zsh
mkdir -p ~/.zfunc
soku completion zsh > ~/.zfunc/_soku
```

기존 `compinit` 호출보다 앞에 다음 줄을 `~/.zshrc`에 추가한 뒤 새 shell을
시작하십시오. 설정에 이미 `compinit`이 있으면 두 번째 호출을 추가하지 마십시오.

```zsh
fpath=(~/.zfunc $fpath)
autoload -Uz compinit && compinit
```

제거하려면 `fpath` 줄을 지우고, Soku 때문에 추가한 경우에만 `compinit` 줄을 지운
뒤 `~/.zfunc/_soku`를 삭제하십시오.

### Fish

```fish
mkdir -p ~/.config/fish/completions
soku completion fish > ~/.config/fish/completions/soku.fish
```

Fish가 이 파일을 자동으로 찾습니다. 제거하려면 파일을 삭제하십시오.

### PowerShell

```powershell
$completionDirectory = Join-Path $HOME ".config/soku/completions"
$completionFile = Join-Path $completionDirectory "soku.ps1"
New-Item -ItemType Directory -Force $completionDirectory | Out-Null
soku completion powershell | Set-Content $completionFile
Add-Content $PROFILE '. "$HOME/.config/soku/completions/soku.ps1"'
. $PROFILE
```

제거하려면 `$PROFILE`의 dot-source 줄을 지운 뒤 `$completionFile`을 삭제하십시오.
Soku 자체는 `$PROFILE`을 생성하거나 편집하지 않습니다.

## 자동화 예제

```bash
# human text를 parsing하지 않고 JSON을 사용합니다.
soku status --json | jq -e '.ok and (.command == "status")'

# 예약 검사에서는 문서화된 exit code만 사용합니다.
if soku status --quiet; then
  echo "Soku state is clean"
else
  code=$?
  echo "Soku status exited with $code" >&2
fi

# 결정적인 script를 artifact로 저장하고 ANSI byte가 없는지 확인합니다.
soku completion fish > soku.fish
LC_ALL=C grep -q $'\033' soku.fish && echo "unexpected ANSI" >&2
```

Completion 후보는 로컬 힌트이며 validation을 대신하지 않습니다. Script에는 저장소
조회나 network 결과가 없고 Soku binary가 바뀔 때마다 다시 생성할 수 있습니다.

## 문제 해결

- `command -v soku`(PowerShell은 `Get-Command soku`)로 의도한 binary가
  `PATH`에서 먼저 선택되는지 확인하고 `soku --version`을 확인하십시오.
- Soku upgrade 후 저장된 completion script를 다시 생성하고 새 shell을
  시작하십시오. 오래된 script에는 이전 command tree가 남을 수 있습니다.
- Bash에는 호환되는 `bash-completion` package가 필요합니다. macOS system
  Bash만으로는 모든 helper가 제공되지 않으므로 framework를 먼저 로드한 뒤 생성
  파일을 source하십시오. `complete -p soku`로 등록 상태를 확인하십시오.
- Zsh는 `compinit` 전에 `_soku`가 있는 directory를 `fpath`에 포함해야 합니다.
  Zsh 설정에서 허용할 때만 `~/.zcompdump*` 같은 오래된 cache를 제거하고
  `compinit`을 다시 실행하십시오.
- Fish는 파일을 `~/.config/fish/completions/soku.fish`에 두어야 합니다.
  `complete -C 'soku '`로 후보를 검사하십시오.
- PowerShell은 현재 session에서 생성 script를 실행해야 합니다. `$PROFILE`,
  execution policy, `Get-Command soku`를 확인한 뒤 `. $PROFILE`로 다시
  로드하십시오. `TabExpansion2 'soku ' 5`로 검사할 수 있습니다.
- clean shell에서는 되지만 평소 shell에서 안 되면 completion/plugin manager의
  순서나 cache를 확인하십시오. Soku는 해당 도구를 관리하지 않습니다.

## 안전 경계

`soku completion`은 사용자가 명시적으로 redirect하지 않는 한 stdout에만 씁니다.
Soku는 Bash, Zsh, Fish, PowerShell profile을 수정하지 않고 login shell을
바꾸지 않으며 plugin manager를 설치·설정·제거하지 않습니다. 영구 적용을 선택하기
전에 생성 script를 검토하고, 제거할 때는 자신이 추가한 profile 줄과 사용자 소유
파일만 제거하십시오.

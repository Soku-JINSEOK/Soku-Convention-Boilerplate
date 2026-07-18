# Pull Request

## 🔗 Common Metadata

<!-- An issue is required. Use `Closes #N` for final work or `Related to #N` for partial work; `No-Issue` is not allowed. -->

- **Issue:** `<Closes #N for final work | Related to #N for partial work>`
- **Task report:** `docs/issues/issue-<n>-task-report.md`
- **Governance profile:** `<profile name or None>`

## 🇬🇧 English — Normative Source

### 🎯 Goal

<!-- State the outcome and rationale. -->

### 📦 Scope

<!-- List included and excluded changes, constraints, and approval boundaries. -->

### ✅ Acceptance Criteria

<!-- Record observable completion criteria. -->

### 🔒️ Security Boundary

<!-- Confirm downstream customization, user data, secrets, approval boundaries, and delivery state are preserved. -->

### 🧪 Verification

<!-- Select only checks actually run and record their results. -->

- [ ] `node --test templates/_shared/commitlint/*.test.mjs`
- [ ] `npx --yes markdownlint-cli2@0.22.1 --config .markdownlint.jsonc "**/*.md" "#node_modules"`
- [ ] `npx --yes yaml-lint@1.7.0 .github/*.yml .github/**/*.yml`
- [ ] `scripts/verify-sync-parity.sh`
- [ ] Relevant stack or template validation:
- [ ] `git diff --check`

### ⚠️ Risks and Follow-up

<!-- Record compatibility, migration, cost, delivery, and remaining risks. -->

## 🇰🇷 한국어 요약

### 🎯 목표

<!-- 목표와 기대 결과를 요약합니다. -->

### 📦 핵심 범위

<!-- 포함 범위, 제외 범위, 제약 조건과 승인 경계를 요약합니다. -->

### 🧪 검증

<!-- 실제 실행한 검증과 결과를 요약합니다. -->

### 🔒️ 비파괴 조건

<!-- 다운스트림 수정, 사용자 데이터, 비밀정보, 승인 경계와 delivery 상태를 요약합니다. -->

### ⚠️ 잔여 위험과 후속 작업

<!-- 호환성, migration, 비용, delivery와 잔여 위험을 요약합니다. -->

## 🇯🇵 日本語の要約

### 🎯 目標

<!-- 目標と期待する結果を要約します。 -->

### 📦 主な範囲

<!-- 対象、対象外、制約、承認境界を要約します。 -->

### 🧪 検証

<!-- 実際に実行した検証と結果を要約します。 -->

### 🔒️ 非破壊条件

<!-- ダウンストリームの変更、ユーザーデータ、secret、承認境界、delivery状態を要約します。 -->

### ⚠️ 残存リスクと後続作業

<!-- compatibility、migration、cost、delivery、残存リスクを要約します。 -->

## Gitmoji Checklist

- [ ] ✨ Feature (`feat`)
- [ ] 🐛 Fix (`fix`)
- [ ] ♻️ Refactor (`refactor`)
- [ ] 🎨 Style (`style`)
- [ ] 📚 Documentation (`docs`)
- [ ] ✅ Test (`test`)
- [ ] 🔧 Chore (`chore`)
- [ ] 👷 CI (`ci`)
- [ ] 📦 Build (`build`)
- [ ] 🚀 Performance (`perf`)
- [ ] 🔥 Remove (`remove`)
- [ ] 🚑 Hotfix (`hotfix`)
- [ ] 🔒️ Security (`security`)
- [ ] ⏪️ Revert (`revert`)
- [ ] 🔄 Sync (`sync`)
- [ ] 🔖 Release (`release`)
- [ ] 💥 Breaking Change (`feat!`, `fix!`, etc.)
- [ ] 🔒️ Security boundary maintained
- [ ] Delivery enabled

## 🤖 AI Assistance

<!-- Replace with the actual tool used or `None`; do not leave placeholders. -->

- **Planning/implementation/drafting:** `<actual tool or None>`

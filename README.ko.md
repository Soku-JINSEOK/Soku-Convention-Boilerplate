# 🧩 Soku-Convention-Boilerplate

> 읽기 쉬운 코드, 안정적인 구조, 그리고 장기적인 유지보수성을 중요하게 생각하는 팀과 AI 에이전트를 위한 재사용 가능한 컨벤션 베이스라인입니다.

[English](./README.md) | [日本語](./README.ja.md)

## 👋 개요

`Soku-Convention-Boilerplate`는 모든 프로젝트에서 일관된 코드 스타일, 구조 및 협업 기준을 유지하기 위한 재사용 가능한 베이스 템플릿입니다.  
단순한 시작 템플릿이 아니라 가독성과 장기적인 유지보수성을 중심으로 하는 개발 문화를 구축할 수 있는 반복 가능한 운영 기반으로 설계되었습니다.

## 🗺️ 마스터 블루프린트

표준화된 설계와 운영 모델에 대해서는 [BLUEPRINT.md](./BLUEPRINT.md) 문서를 참고하십시오.

---

## 📌 한눈에 보기

| 영역 | 표준 |
| --- | --- |
| 스타일 베이스라인 | `Google Style Guide` |
| 사람을 위한 문서 | 영어, 한국어, 일본어 기본 지원 |
| 룰 및 거버넌스 문서 | 영어로만 제공 |
| 핵심 목표 | 여러 프로젝트 간의 일관성 유지 |
| 코드 리뷰 최우선 순위 | 논리, 명확성, 유지보수성 |
| 규칙 강제화 방식 | 포맷터 + 린터 + 문서화 자동 검사 |
| 커밋 스타일 | Gitmoji + Conventional Commits |
| 릴리즈 태그 | Verified 상태를 위한 서명된 태그 (`git tag -s`) |

## 🤔 Google Style Guide를 채택한 이유

이 프로젝트는 기본 스타일 가이드로 `Google Style Guide`를 채택합니다.  
첫 번째 이유는 코드는 작성되는 횟수보다 읽히는 횟수가 훨씬 많다는 철학에 깊이 공감하기 때문입니다. 우리는 단기적인 구현 속도보다 장기적인 명확성과 유지보수성에 더 큰 가치를 둡니다.

두 번째 이유는 자동화의 용이성입니다. Google 스타일 가이드는 포맷터, 린터, 정적 분석 도구와 유기적으로 연동되므로 주관적인 스타일 논쟁을 최소화하고 개발자가 비즈니스 로직과 아키텍처 구현에 집중할 수 있는 환경을 만들어 줍니다.

---

## 💭 철학

이 보일러플레이트는 프로젝트의 규모, 연차, 작업자에 무관하게 코드베이스가 일관되게 이해되고 예측 가능해야 한다는 신념 위에 세워졌습니다.

우리는 코드 컨벤션을 단순한 시각적 취향으로 다루지 않습니다. 컨벤션은 모호성을 줄이고 협업을 강화하며, 동일한 코드베이스에서 일하는 사람들과 AI 에이전트 모두의 작업 효율성을 끌어올리는 최소한의 운영 안전장치입니다.

우리의 목표는 새로운 도메인에서 프로젝트를 시작할 때마다 팀 문화를 매번 새로 구성하지 않고, 재사용 가능한 견고한 기반 위에서 빠르게 시작할 수 있도록 돕는 것입니다.

## ✅ 원칙

1. 읽기 쉬운 코드가 현란한 코드보다 낫습니다.
2. 일관성은 개인적인 코딩 취향보다 훨씬 소중합니다.
3. 자동화 도구로 스타일을 규정할 수 있다면 반드시 그렇게 해야 합니다.
4. 프로젝트 구조는 여러 저장소에 걸쳐 예측 가능해야 합니다.
5. 코드 리뷰는 서식 다툼이 아닌 논리, 비헤이비어, 설계에 역량을 쏟아야 합니다.
6. 문서는 단순히 동작 구조를 설명하는 것을 넘어 작성자의 의도(Intent)를 설명해야 합니다.
7. 모든 컨벤션은 현재 작업자뿐만 아니라 미래의 유지보수자에게 도움을 주어야 합니다.

## ⚙️ 운영 표준

### 1. 스타일 베이스라인

이 보일러플레이트를 기반으로 생성된 모든 저장소는 명시적인 예외 사유가 문서화되지 않는 한 `Google Style Guide`를 기본 스타일 베이스라인으로 삼습니다.

### 2. 포맷팅 및 린팅

코드 포맷팅과 린팅은 의무적으로 자동화되어야 하며, 로컬 개발 및 빌드 과정의 필수 워크플로우로 취급되어야 합니다. 도구를 통해 규칙을 검사할 수 있다면 코드 리뷰 시 수동 지적 대신 자동 린터를 통해 사전에 걸러내야 합니다.

### 3. 저장소 일관성

저장소 간의 디렉토리 구조, 네이밍 규칙, 문서 배치 양식을 표준화하여 개발자가 다른 저장소로 전환해 작업할 때 발생하는 인지적 비용(Cognitive Overhead)을 최소화합니다.

### 4. 문서 작성 규칙

문서 작성 언어는 [BLUEPRINT.md의 Language Policy](./BLUEPRINT.md#language-policy)를 따릅니다. 사용자를 위한 개요 문서는 영어, 한국어, 일본어를 기본 지원하지만, 규칙, 정책, 기술 지침서 및 AI 지침은 영어로만 일관되게 기재합니다.

### 5. 리뷰 규율

풀 리퀘스트와 코드 리뷰는 다음 사항을 최우선적으로 다룹니다:

- 동작의 올바름 (Correctness)
- 유지보수성 (Maintainability)
- 아키텍처적 명확성 (Architectural clarity)
- 테스트 가능성 (Testability)
- 의사결정의 트레이드오프

서식 및 줄 바꿈 등의 스타일링 쟁점은 자동화 린트/포맷 툴에게 전적으로 위임합니다.

### 6. 컨벤션의 범용성

이 보일러플레이트에 추가되는 모든 룰은 여러 다양한 저장소에서 복제하여 재사용할 수 있는 범용성을 갖추어야 합니다. 특정 개별 프로젝트에만 유효한 규칙은 이 보일러플레이트가 아니라 해당 프로젝트 내부 개발 문서에 개별적으로 기재되어야 합니다.

### 7. AI 에이전트 호환성

이 저장소는 AI 에이전트가 다음 사항들을 직관적으로 파악할 수 있도록 최적화되어 있어야 합니다:

- 프로젝트의 원래 의도 및 아키텍처 목적
- 코드 소유권 범위 및 경계선
- 구조적 컨벤션 및 네이밍 규칙
- 문서 작성 기준
- 빌드, 테스트 및 검증 워크플로우

이를 실현하기 위해 글로벌 가이드와 지침은 직접적이고 명료한 영어로 일관되게 작성됩니다.

## 🎯 기대 효과 및 용도

이 보일러플레이트는 다음 역할을 수행하기 위해 존재합니다:

- 신규 저장소를 시작할 때의 표준 스켈레톤 (스타팅 포인트)
- 개인 및 팀 프로젝트 전체에 적용되는 일관된 컨벤션 공유 계층
- 가독성 높고 읽기 쉬운 견고한 코드를 작성하기 위한 훈련장
- 자동화를 통해 서식 관련 마찰을 제거하는 쾌적한 로컬 개발 도구 모음
- 사람 개발자와 AI 에이전트가 함께 소통하고 이해할 수 있는 안정된 저장소 Shape 구축

## 📚 관련 문서 목록

- [README.md](./README.md): 다국어 개요 및 프로젝트 포지셔닝
- [BLUEPRINT.md](./BLUEPRINT.md): 공식 저장소 아키텍처 설계도 및 권한 계층 정보
- [CONTRIBUTING.md](./CONTRIBUTING.md): 기여 방식 워크플로우 및 커밋 메시지 규칙서
- [AGENTS.md](./AGENTS.md): AI 에이전트 동작 안내서 (영어 전용)
- [LICENSE](./LICENSE): 보일러플레이트 기본 라이선스 문서 (MIT)
- [SECURITY.md](./SECURITY.md): 보안 결함 제보 창구

### 📏 `docs/standards/` — 구조 및 프로세스 규범 지침서

- [CODE_STYLE.md](./docs/standards/CODE_STYLE.md): 기본 코드 작성 스타일 표준
- [PROJECT_STRUCTURE.md](./docs/standards/PROJECT_STRUCTURE.md): 디렉토리 및 모듈 레이아웃 규칙
- [GITHUB_STANDARDS.md](./docs/standards/GITHUB_STANDARDS.md): 이슈, PR, 리뷰, 레이블 및 템플릿 운영 기준
- [RELEASE_AND_SYNC.md](./docs/standards/RELEASE_AND_SYNC.md): 버전 릴리즈 태그 배포 및 다운스트림 동기화 절차
- [SOKU_LIFECYCLE.md](./docs/standards/SOKU_LIFECYCLE.md): `soku` CLI, 매니페스트, 소유권, 공급자 및 트랜잭션 라이프사이클 규범
- [CICD_STANDARDS.md](./docs/standards/CICD_STANDARDS.md): 지속적 통합 및 배포(CI/CD) 수용 규칙

### 🛡️ `docs/policy/` — 영역별 기본 정책 선언서

- [LICENSE_POLICY.md](./docs/policy/LICENSE_POLICY.md): 오픈소스 라이선스 선택 가이드라인
- [SECURITY_POLICY.md](./docs/policy/SECURITY_POLICY.md): 소스, 비밀값(Secret), 배포 단계별 기본 보안 수칙
- [CLOUD_POLICY.md](./docs/policy/CLOUD_POLICY.md): 실무 클라우드(GCP, AWS, Azure) 공급업체 결정 규칙

### 🧭 `docs/guides/` — 모범 사례 및 안내 가이드 문서

- [STACK_EXAMPLES.md](./docs/guides/STACK_EXAMPLES.md): 다국어, 프레임워크, DB 및 클라우드 구축 모범 사례 모음
- [STACK_CONFIGS.md](./docs/guides/STACK_CONFIGS.md): 스택별 설정 파일 스타터 템플릿 복제 가이드
- [README_GUIDE.md](./docs/guides/README_GUIDE.md): 프로젝트 README를 읽기 쉽고 명확하게 유지하는 작성법
- [INIT_GUIDE.md](./docs/guides/INIT_GUIDE.md): 다운스트림 프로젝트를 부트스트랩할 때 에이전트가 참고하는 환경 탐색 체크리스트
- [APPLICABILITY.md](./docs/guides/APPLICABILITY.md): 개인 프로젝트와 팀 프로젝트에서 각각 적용 가능한 수준 분리 기준
- [LANGUAGE_SELECTION.md](./docs/guides/LANGUAGE_SELECTION.md): 새 프로젝트나 기능에 사용할 프로그래밍 언어를 선택하는 기준

### 📝 `docs/issues/` — 작업 보고서 아티팩트

- [TASK_REPORT_TEMPLATE.md](./docs/issues/TASK_REPORT_TEMPLATE.md): 구현 착수 전 문서화하고 승인받는 계획 템플릿

## 🧱 시작 스택 지원 범위

이 보일러플레이트는 다양한 기술 스택을 담아낼 수 있도록 지속적으로 확장 가능한 상태를 유지합니다. 기본 지원 가이드는 다음을 커버합니다:

- `JavaScript` / `TypeScript` / `Node.js`
- `Python`
- `Go`
- `Java` / `Spring`
- `MySQL` / `PostgreSQL`
- `gcloud` / `AWS` / `Azure`

## 🛠️ 설정 템플릿 세트

각 개발 환경 설정 파일(`package.json`, `pyproject.toml`, `Makefile` 등)의 복사본은 [STACK_CONFIGS.md](./docs/guides/STACK_CONFIGS.md) 및 `templates/` 디렉토리를 참조하십시오.

## 🎬 요약

이 보일러플레이트는 단순한 프로젝트 시작 지점을 넘어서서, 다수의 프로젝트에 걸쳐 팀의 기술 일관성과 코드 가치를 흔들림 없이 수호하기 위한 신뢰할 수 있는 재사용 운영 기반입니다.

# 실제 런타임 사용자 매뉴얼 캡처 요약

정식 규칙은
[`REAL_RUNTIME_MANUAL_CAPTURE.md`](../standards/REAL_RUNTIME_MANUAL_CAPTURE.md)에
있습니다.

`docs-manual` component는 실제 frontend를 로컬 Chromium에서 실행해 PNG,
stable capture ID, source/fixture/hook/image hash와 credential이 제거된 report를
만듭니다. `soku docs manual plan`은 읽기 전용 계획, `doctor`는 환경 진단,
`init --dry-run|--yes`는 이미 초기화된 저장소에 고정 runner와 schema만
transactionally 설치합니다.

실제 `capture.yml`, synthetic fixture, hook, 설명문, PNG, report와 PDF는
project-owned이며 자동 생성하거나 덮어쓰지 않습니다. 원본 frontend와 map
provider를 유지하고, backend·GAS bridge·dialog overlay·provider 대체는 모두
caption/report에 공개하며 authenticity를 `runtime-authentic-with-adapters`로
낮춥니다.

OSM은 attribution을 표시한 로컬 수동 단일 viewport와 제한된 request만
허용합니다. Google Maps JavaScript는 `GOOGLE_MAPS_API_KEY` 환경 변수 이름,
검토된 제한 key, billing owner, 1회 map load, request ceiling, readiness
event/hook, attribution 표시가 필요합니다. key 값, cloud ID, billing 상태와
credential URL은 저장하지 않으며 cloud/billing/key 설정을 만들거나 변경하지
않습니다.

초기화는 `npm ci`, browser, OS package, font를 설치하지 않습니다. 실제 live
provider capture, CI 자동화, PDF 생성·배포와 공개 release는 별도 승인 범위입니다.

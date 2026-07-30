# 実ランタイム利用マニュアル・キャプチャの概要

正式な規則は
[`REAL_RUNTIME_MANUAL_CAPTURE.md`](../standards/REAL_RUNTIME_MANUAL_CAPTURE.md)
を参照してください。

`docs-manual` component は実際の frontend をローカル Chromium で実行し、
PNG、stable capture ID、source/fixture/hook/image hash、credential を除去した
report を生成します。`soku docs manual plan` は read-only plan、`doctor` は
環境診断、`init --dry-run|--yes` は初期化済み repository に固定 runner と
schema のみを transactionally install します。

実際の `capture.yml`、synthetic fixture、hook、manual prose、PNG、report、
PDF は project-owned で、自動生成・上書きしません。元の frontend と map
provider を維持し、backend、GAS bridge、dialog overlay、provider 置換は
caption/report に開示して authenticity を
`runtime-authentic-with-adapters` に下げます。

OSM は attribution を表示する local/manual の単一 viewport と制限された
request のみ許可します。Google Maps JavaScript は
`GOOGLE_MAPS_API_KEY` という環境変数名、review 済み制限 key、billing owner、
1 回の map load、request ceiling、readiness event/hook、attribution 表示が
必要です。key 値、cloud ID、billing status、credential URL は保存せず、
cloud/billing/key 設定を作成・変更しません。

初期化は `npm ci`、browser、OS package、font を install しません。live
provider capture、CI automation、PDF 生成・公開 release は別の承認範囲です。

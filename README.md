# edinet-cli

EDINET API v2 の Go 製 CLI ラッパー。AIエージェントによる自律的な日本企業の開示書類取得・財務分析に最適化。

[![CI](https://github.com/beatinaniwa/edinet-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/beatinaniwa/edinet-cli/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8.svg)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## AIエージェント向けクイックリファレンス

> **AI エージェントはまず `edinet-cli schema commands` を実行してください。** 全コマンドのフラグ・型・使用例が JSON で返ります。

### 出力規約

| ストリーム | 内容 | パース方法 |
|-----------|------|-----------|
| stdout | 構造化データ（JSON / table） | `json.Unmarshal` 等 |
| stderr | エラー（JSON）、進捗、デバッグ | 人間向け、パース不要 |

### 終了コード

| コード | 意味 | 対処 |
|--------|------|------|
| 0 | 成功（部分成功を含む） | 結果をそのまま利用 |
| 1 | 一般エラー | stderr を確認 |
| 2 | バリデーションエラー（引数不正） | コマンドの使い方を修正 |
| 3 | 認証エラー | `EDINET_API_KEY` を確認 |
| 4 | APIエラー（400/404/500） | リクエスト内容を見直し |

### スキーマ自己記述（`schema` コマンド群）

AIエージェントが CLI の能力を事前に把握するための機械可読メタデータ:

```bash
edinet-cli schema commands             # 全コマンド・フラグ・型・使用例（JSON）
edinet-cli schema doc-types            # 書類種別コード一覧（42種、例: 120=有価証券報告書）
edinet-cli schema sections             # テキスト抽出可能なセクション一覧
edinet-cli schema financial-elements   # 財務XBRL要素マッピング一覧
edinet-cli schema derived-metrics      # 派生財務指標（ROE, ROA等）の計算式一覧
```

### 典型的なワークフロー例

```bash
# 1. 企業を特定
edinet-cli company search トヨタ

# 2. 直近の有価証券報告書を探す
edinet-cli company filings 7203 --doc-type 120 --limit 3

# 3. 財務データを構造化抽出（複数期間）
edinet-cli company financials 7203 --periods 5 --summary-only

# 4. 特定書類のリスク情報をテキスト抽出
edinet-cli doc text S100XXXX --section risk
```

---

## 概要

[EDINET](https://disclosure.edinet-fsa.go.jp/) は金融庁が運営する有価証券報告書等の電子開示システムです。`edinet-cli` はその API v2 を薄くラップし、以下を提供します:

- **構造化出力**: stdout は常に JSON（`--format table` で人間向けテーブル）
- **決定論的な終了コード**: 自動化・エラーハンドリングが容易
- **スキーマ自己記述**: `schema` コマンドで全機能を機械可読形式で取得
- **財務データ抽出**: BS/PL/CF の構造化抽出と ROE・ROA 等の派生指標自動計算
- **ローカルキャッシュ**: API レスポンスを自動キャッシュし不要なリクエストを削減

## インストール

### 前提条件

- Go 1.26 以上
- EDINET API キー（[EDINET API 利用申請ページ](https://disclosure.edinet-fsa.go.jp/)から取得）

### go install

```bash
go install github.com/beatinaniwa/edinet-cli@latest
```

### ソースからビルド

```bash
git clone https://github.com/beatinaniwa/edinet-cli.git
cd edinet-cli
make build
```

## クイックスタート

```bash
# API キーを設定（環境変数のみ。ファイルには保存されない）
export EDINET_API_KEY="your-api-key"

# 今日の開示書類を一覧表示
edinet-cli doc list --date 2026-03-31

# 有価証券報告書のみに絞り込み
edinet-cli doc list --date 2026-03-31 --doc-type 120

# 企業を検索
edinet-cli company search トヨタ

# 書類をPDFでダウンロード
edinet-cli doc get S100ABCD --type pdf

# 企業の財務サマリーを複数期間取得
edinet-cli company financials 7203 --periods 3 --summary-only
```

---

## コマンドリファレンス

### グローバルフラグ

| フラグ | 型 | デフォルト | 説明 |
|--------|------|-----------|------|
| `--format` | string | `json` | 出力形式（`json` / `table`） |
| `--debug` | bool | `false` | デバッグ出力を stderr に表示 |
| `--no-cache` | bool | `false` | ローカルキャッシュをバイパス |

---

### `doc list` --- 開示書類の一覧取得

日付または期間を指定して開示書類を一覧取得します。

```bash
# 単一日付
edinet-cli doc list --date 2025-06-20

# 期間指定（有価証券報告書のみ）
edinet-cli doc list --from 2025-06-01 --to 2025-06-30 --doc-type 120

# 提出者名と書類説明でフィルタ
edinet-cli doc list --date 2025-06-20 --filer-name トヨタ --doc-description 有価証券報告書
```

| フラグ | 型 | 説明 |
|--------|------|------|
| `--date` | string | 単一日付（YYYY-MM-DD）。`--from`/`--to` と排他 |
| `--from` | string | 期間の開始日（`--to` と同時に指定） |
| `--to` | string | 期間の終了日（`--from` と同時に指定） |
| `--doc-type` | string | 書類種別コードでフィルタ（例: `120`=有価証券報告書） |
| `--sec-code` | string | 証券コードでフィルタ（例: `72030`） |
| `--edinet-code` | string | EDINET コードでフィルタ |
| `--filer-name` | string | 提出者名で部分一致フィルタ |
| `--doc-description` | string | 書類説明で部分一致フィルタ |
| `--rate-limit` | int | リクエスト間隔ミリ秒（デフォルト: 100） |

---

### `doc get` --- 書類のダウンロード

指定した書類 ID のファイルをダウンロードします。

```bash
edinet-cli doc get S100ABCD --type pdf
edinet-cli doc get S100ABCD --type csv --out ./data/
```

| フラグ | 型 | 必須 | デフォルト | 説明 |
|--------|------|------|-----------|------|
| `--type` | string | **必須** | --- | `pdf` / `xbrl`(zip) / `csv`(zip) / `attach`(zip) / `english`(zip) |
| `--out` | string | --- | `.` | 出力ディレクトリ |

---

### `doc financial` --- 財務データの構造化抽出

書類の CSV データから BS（貸借対照表）/ PL（損益計算書）/ CF（キャッシュフロー計算書）を構造化抽出し、ROE・ROA 等の派生指標も自動計算します。

```bash
# 全財務諸表を抽出
edinet-cli doc financial S100ABCD

# 損益計算書のみ
edinet-cli doc financial S100ABCD --statement pl

# サマリー指標のみ（詳細行なし）
edinet-cli doc financial S100ABCD --summary-only

# 単体財務諸表を優先
edinet-cli doc financial S100ABCD --non-consolidated
```

| フラグ | 型 | デフォルト | 説明 |
|--------|------|-----------|------|
| `--statement` | string | `all` | 財務諸表の種類: `bs` / `pl` / `cf` / `all` |
| `--non-consolidated` | bool | `false` | 単体（非連結）財務諸表を優先 |
| `--summary-only` | bool | `false` | サマリー指標のみ出力（詳細行を省略） |

---

### `doc text` --- テキストの抽出

HTML 書類からテキストを抽出します。セクション指定で特定の章のみ取得可能。

```bash
# 全テキストを抽出
edinet-cli doc text S100ABCD

# 特定セクションを抽出
edinet-cli doc text S100ABCD --section risk

# 利用可能なセクション一覧を表示
edinet-cli doc text --list-sections
```

| フラグ | 型 | 説明 |
|--------|------|------|
| `--section` | string | セクション ID またはヘッディングパターン |
| `--list-sections` | bool | 利用可能なセクション一覧を表示（docID 不要） |

**利用可能なセクション:**

| ID | 対応する見出し |
|-----|--------------|
| `business` | 事業の内容 |
| `risk` | 事業等のリスク |
| `mda` | 経営者による財政状態・経営成績等の状況 |
| `governance` | コーポレート・ガバナンス |
| `financial` | 財務諸表 |
| `employees` | 従業員の状況 |
| `facilities` | 設備の状況 |
| `history` | 沿革 |
| `shares` | 株式等の状況 |
| `dividends` | 配当政策 |

---

### `doc data` --- 財務データ抽出（実験的、非推奨）

> **注意:** `doc financial` の利用を推奨します。`doc data` は後方互換のために残されています。

```bash
edinet-cli doc data S100ABCD
```

---

### `company search` --- 企業検索

EDINET コードリスト（13,000社以上）から企業を検索します。名称・証券コード・EDINET コードで検索可能。

```bash
edinet-cli company search トヨタ
edinet-cli company search 7203
edinet-cli company search トヨタ --industry 自動車
```

| フラグ | 型 | 説明 |
|--------|------|------|
| `--industry` | string | 業種名でフィルタ（部分一致） |

---

### `company filings` --- 企業の提出書類一覧

指定した企業の提出書類を一覧取得します。証券コードまたは EDINET コードで指定。

```bash
edinet-cli company filings 7203 --doc-type 120 --limit 5
edinet-cli company filings E00010 --from 2025-01-01 --to 2025-03-31
```

| フラグ | 型 | デフォルト | 説明 |
|--------|------|-----------|------|
| `--doc-type` | string | --- | 書類種別コードでフィルタ |
| `--from` | string | 365日前 | 期間の開始日 |
| `--to` | string | 今日 | 期間の終了日 |
| `--limit` | int | `0` | 最大件数（0=無制限） |

---

### `company financials` --- 複数期間の財務データ取得

企業コード（証券コード・EDINET コード・企業名）を指定して、複数決算期の財務データを一括取得します。内部で有価証券報告書の検索・ダウンロード・抽出を自動で行います。

```bash
# 直近3期分の全財務諸表
edinet-cli company financials 7203

# 直近5期分の損益計算書サマリー
edinet-cli company financials 7203 --periods 5 --statement pl --summary-only

# 企業名でも指定可能
edinet-cli company financials トヨタ --periods 3
```

| フラグ | 型 | デフォルト | 説明 |
|--------|------|-----------|------|
| `--periods` | int | `3` | 取得する決算期数（1-10） |
| `--statement` | string | `all` | 財務諸表の種類: `bs` / `pl` / `cf` / `all` |
| `--non-consolidated` | bool | `false` | 単体（非連結）財務諸表を優先 |
| `--summary-only` | bool | `false` | サマリー指標のみ出力 |

---

### `company update` --- EDINET コードリストの更新

EDINET コードリスト（企業マスタ）をダウンロードしてキャッシュを更新します。

```bash
edinet-cli company update
```

---

### `schema` --- CLI メタデータの自己記述

AIエージェントが CLI の全機能を自動把握するための機械可読メタデータ。

| サブコマンド | 説明 |
|-------------|------|
| `schema commands` | 全コマンドのフラグ・型・必須/任意・使用例 |
| `schema doc-types` | 書類種別コード一覧（42種） |
| `schema sections` | テキスト抽出セクション一覧 |
| `schema financial-elements` | 財務 XBRL 要素マッピング（BS/PL/CF 各勘定科目） |
| `schema derived-metrics` | 派生財務指標と計算式（ROE, ROA, FCF 等 9指標） |

---

## 設定

### 環境変数

| 変数名 | 必須 | 説明 |
|--------|------|------|
| `EDINET_API_KEY` | **必須** | EDINET API サブスクリプションキー。ディスクには保存されない |
| `EDINET_CONFIG_DIR` | --- | 設定ディレクトリ（デフォルト: `~/.config/edinet-cli`） |
| `EDINET_CACHE_DIR` | --- | キャッシュディレクトリ（デフォルト: `~/.cache/edinet-cli`） |

### キャッシュ

API レスポンスは自動キャッシュされます:

- **書類一覧**: 過去日は24時間、当日は5分間
- **EDINET コードリスト**: 7日間

`--no-cache` でキャッシュをバイパスできます。

---

## 主な書類種別コード

| コード | 名称 | 用途 |
|--------|------|------|
| `120` | 有価証券報告書 | 年次の包括的開示（財務諸表・事業リスク等） |
| `130` | 訂正有価証券報告書 | 有価証券報告書の訂正 |
| `140` | 四半期報告書 | 四半期ごとの開示 |
| `160` | 半期報告書 | 半期の開示 |
| `180` | 臨時報告書 | 重要事象発生時の開示 |
| `350` | 大量保有報告書 | 5%以上の株式保有報告 |

全42種のコード一覧は `edinet-cli schema doc-types` で取得できます。

## 派生財務指標

`doc financial` / `company financials` が自動計算する指標:

| 指標 | 計算式 |
|------|--------|
| `gross_margin` | 売上総利益 / 売上高 |
| `operating_margin` | 営業利益 / 売上高 |
| `net_margin` | 当期純利益 / 売上高 |
| `roe` | 当期純利益 / 自己資本 |
| `roa` | 当期純利益 / 総資産 |
| `equity_ratio` | 自己資本 / 総資産 |
| `current_ratio` | 流動資産 / 流動負債 |
| `fcf` | 営業CF + 投資CF |
| `debt_to_equity` | 有利子負債 / 自己資本 |

全指標の詳細は `edinet-cli schema derived-metrics` で取得できます。

---

## アーキテクチャ

```
cmd/              Cobra コマンド（CLI アダプタ層）
internal/
  api/            EDINET API v2 HTTP クライアント & ワイヤーモデル
  cache/          キャッシュインターフェース & ファイルベース実装
  config/         環境変数ベースの設定読み込み
  company/        EDINET コードリスト & 企業検索
  extract/        ZIP / CSV / HTML の抽出・パース
  financial/      財務諸表パーサー（BS/PL/CF 抽出 & 派生指標計算）
  output/         JSON / テーブルフォーマッタ
  service/        ビジネスロジック層（書類一覧・ダウンロード・企業検索・財務分析）
  schema/         CLI 自己記述データ（コマンド・書類種別・セクション・財務要素）
  testutil/       テスト用共通ヘルパー（httptest モック）
```

## 開発

```bash
make build      # バイナリをビルド
make test       # テスト実行（-race 有効）
make lint       # go vet + golangci-lint
make coverage   # カバレッジレポート生成
make clean      # ビルド成果物を削除
```

## ライセンス

[MIT License](LICENSE)

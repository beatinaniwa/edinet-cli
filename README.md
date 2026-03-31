# edinet-cli

EDINET API v2 のコマンドラインクライアント。AI エージェントによる自律的な書類取得・分析に最適化されています。

[![CI](https://github.com/beatinaniwa/edinet-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/beatinaniwa/edinet-cli/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8.svg)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## 概要

[EDINET](https://disclosure.edinet-fsa.go.jp/)（Electronic Disclosure for Investors' NETwork）は、金融庁が運営する有価証券報告書等の開示書類の電子開示システムです。

`edinet-cli` は EDINET API v2 を Go で薄くラップした CLI ツールで、以下の特徴を持ちます：

- **構造化出力**: stdout にはすべて JSON（または `--format table` でテーブル形式）を出力
- **決定論的な終了コード**: 自動化やエラーハンドリングが容易
- **スキーマ自己記述**: `schema` コマンドで CLI の全機能を機械可読な形式で取得可能
- **ローカルキャッシュ**: API レスポンスをキャッシュし、不要なリクエストを削減

## 特徴

- 日付・期間指定による開示書類の一覧取得（書類種別・証券コード・提出者名等でフィルタリング）
- 書類のダウンロード（PDF はそのまま、XBRL / CSV / 英文 / 添付資料は ZIP アーカイブ）
- 財務 CSV データの構造化抽出（実験的機能）
- HTML 書類からのテキスト抽出（セクション指定対応）
- 13,000社以上の企業検索（名称・証券コード・EDINET コード・業種）
- 企業の提出書類一覧の取得
- AI エージェント向けの CLI スキーマ自己記述機能

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
# API キーを設定
export EDINET_API_KEY="your-api-key"

# 今日の開示書類を一覧表示
edinet-cli doc list --date 2026-03-31

# 有価証券報告書のみを絞り込み
edinet-cli doc list --date 2026-03-31 --doc-type 120

# 企業を検索
edinet-cli company search トヨタ

# 書類をPDFでダウンロード
edinet-cli doc get S100ABCD --type pdf
```

## コマンドリファレンス

### グローバルフラグ

| フラグ | 型 | デフォルト | 説明 |
|--------|------|-----------|------|
| `--format` | string | `json` | 出力形式（`json` または `table`） |
| `--debug` | bool | `false` | デバッグ出力を stderr に表示 |
| `--no-cache` | bool | `false` | ローカルキャッシュをバイパス |

---

### `doc list` — 開示書類の一覧取得

日付または期間を指定して開示書類を一覧取得します。

```bash
# 単一日付
edinet-cli doc list --date 2025-06-20

# 期間指定（有価証券報告書のみ）
edinet-cli doc list --from 2025-06-01 --to 2025-06-30 --doc-type 120
```

| フラグ | 型 | 説明 |
|--------|------|------|
| `--date` | string | 単一日付（YYYY-MM-DD）。`--from`/`--to` と排他 |
| `--from` | string | 期間の開始日（`--to` と同時に指定） |
| `--to` | string | 期間の終了日（`--from` と同時に指定） |
| `--doc-type` | string | 書類種別コードでフィルタ |
| `--sec-code` | string | 証券コードでフィルタ |
| `--edinet-code` | string | EDINET コードでフィルタ |
| `--filer-name` | string | 提出者名で部分一致フィルタ |
| `--rate-limit` | int | リクエスト間隔（ミリ秒、デフォルト: 100） |

---

### `doc get` — 書類のダウンロード

指定した書類 ID のファイルをダウンロードします。

```bash
edinet-cli doc get S100ABCD --type pdf
edinet-cli doc get S100ABCD --type csv --out ./data/
```

| フラグ | 型 | 必須 | デフォルト | 説明 |
|--------|------|------|-----------|------|
| `--type` | string | **必須** | — | `pdf` / `xbrl`(zip) / `csv`(zip) / `attach`(zip) / `english`(zip) |
| `--out` | string | — | `.` | 出力ディレクトリ |

---

### `doc data` — 財務データの抽出（実験的）

書類の CSV データを構造化して抽出します。

```bash
edinet-cli doc data S100ABCD
```

---

### `doc text` — テキストの抽出

HTML 書類からテキストを抽出します。セクション指定も可能です。

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
| `mda` | 経営者による財政状態 |
| `governance` | コーポレート・ガバナンス |
| `financial` | 財務諸表 |
| `employees` | 従業員の状況 |
| `facilities` | 設備の状況 |
| `history` | 沿革 |
| `shares` | 株式等の状況 |
| `dividends` | 配当政策 |

---

### `company search` — 企業検索

EDINET コードリストから企業を検索します。名称・証券コード・EDINET コードで検索可能です。

```bash
edinet-cli company search トヨタ
edinet-cli company search 7203
edinet-cli company search E00010 --industry 自動車
```

| フラグ | 型 | 説明 |
|--------|------|------|
| `--industry` | string | 業種名でフィルタ（部分一致） |

---

### `company filings` — 企業の提出書類一覧

指定した企業の提出書類を一覧取得します。証券コードまたは EDINET コードで指定できます。

```bash
edinet-cli company filings 7203 --doc-type 120 --limit 5
edinet-cli company filings E00010 --from 2025-01-01 --to 2025-03-31
```

| フラグ | 型 | デフォルト | 説明 |
|--------|------|-----------|------|
| `--doc-type` | string | — | 書類種別コードでフィルタ |
| `--from` | string | 365日前 | 期間の開始日 |
| `--to` | string | 今日 | 期間の終了日 |
| `--limit` | int | `0` | 最大件数（0=無制限） |

---

### `company update` — EDINET コードリストの更新

EDINET コードリスト（企業マスタ）をダウンロードしてキャッシュを更新します。

```bash
edinet-cli company update
```

---

### `schema` — CLI スキーマの自己記述

AI エージェントが CLI の機能を自動的に把握するための機械可読メタデータを出力します。

```bash
# 全コマンドのフラグ・例をJSON で取得
edinet-cli schema commands

# 書類種別コード一覧
edinet-cli schema doc-types

# テキスト抽出セクション一覧
edinet-cli schema sections
```

## 設定

### 環境変数

| 変数名 | 必須 | 説明 |
|--------|------|------|
| `EDINET_API_KEY` | **必須** | EDINET API のサブスクリプションキー。ディスクには保存されません |
| `EDINET_CONFIG_DIR` | — | 設定ディレクトリのパス（デフォルト: `~/.config/edinet-cli`） |
| `EDINET_CACHE_DIR` | — | キャッシュディレクトリのパス（デフォルト: `~/.cache/edinet-cli`） |

### キャッシュ

API レスポンスは自動的にローカルにキャッシュされます：

- 書類一覧: 過去日は24時間、当日は5分間
- EDINET コードリスト: 7日間

キャッシュを無視する場合は `--no-cache` フラグを使用してください。

## AI エージェント連携

`edinet-cli` は AI エージェント（Claude、ChatGPT 等）からの自律的な利用を想定して設計されています。

### 出力規約

- **stdout**: 常に構造化データ（JSON）。パース可能
- **stderr**: エラー（JSON 形式）、進捗情報、デバッグ出力
- **終了コード**: 自動化に対応した決定論的なコード体系

| 終了コード | 意味 |
|-----------|------|
| 0 | 成功（部分成功を含む） |
| 1 | 一般エラー |
| 2 | バリデーションエラー（引数不正等） |
| 3 | 認証エラー（API キー未設定・無効） |
| 4 | API エラー（400/404/500） |

### スキーマによる自己発見

AI エージェントは `schema` コマンドで CLI の全機能を事前に把握できます：

```bash
# 利用可能なコマンド・フラグ・使用例を取得
edinet-cli schema commands

# 書類種別コードのマッピングを取得
edinet-cli schema doc-types

# テキスト抽出可能なセクションを確認
edinet-cli schema sections
```

## アーキテクチャ

```
cmd/              Cobra コマンド（CLI アダプタ層）
internal/
  api/            EDINET API v2 HTTP クライアント & ワイヤーモデル
  cache/          キャッシュインターフェース & ファイルベース実装
  config/         環境変数ベースの設定読み込み
  company/        EDINET コードリスト & 企業検索
  extract/        ZIP / CSV / HTML の抽出・パース
  output/         JSON / テーブルフォーマッタ
  service/        ビジネスロジック層（書類一覧・ダウンロード・企業検索）
  schema/         CLI 自己記述データ（コマンド・書類種別・セクション）
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

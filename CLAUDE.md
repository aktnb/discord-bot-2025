# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

このリポジトリは Go 言語で書かれた Discord ボットです。音声チャンネルとテキストチャンネルをリンクする機能を提供します。

## 開発コマンド

### アプリケーションの実行

```bash
go run cmd/bot/main.go
```

### テスト実行

```bash
go test ./...
```

### 特定パッケージのテスト

```bash
go test ./internal/domain/
```

### データベース関連

#### データベース起動（Docker）

```bash
docker-compose up -d
```

#### マイグレーション実行

```bash
migrate -path db/migrations -database "$DATABASE_URL" up
```

#### 新しいマイグレーションファイルの作成

```bash
migrate create -ext sql -dir db/migrations -seq <migration_name>
```

## アーキテクチャ

### ディレクトリ構造

プロジェクトはクリーンアーキテクチャ/ヘキサゴナルアーキテクチャを採用しています。

#### コアレイヤー

- `cmd/bot/`: アプリケーションのエントリーポイント
- `internal/config/`: 設定管理（環境変数、.env）

#### ドメイン層

- `internal/domain/voicetext/`: ボイステキストリンク機能のドメインモデル
  - `model.go`: エンティティとビジネスロジック
  - `repository.go`: リポジトリインターフェース
  - `errors.go`: ドメインエラー定義
- `internal/domain/faker/`: Faker 伝説エピソード機能のドメインモデル
  - `model.go`: エピソード一覧とランダム選択ロジック
- `internal/shared/discordid/`: 共有ドメインオブジェクト（Discord ID 型定義）

#### アプリケーション層

- `internal/application/voicetext/`: ボイステキストリンク機能のユースケース
  - `service.go`: アプリケーションサービス
  - `commands.go`: Discord スラッシュコマンドの実装
- `internal/application/faker/`: Faker 伝説エピソード機能のユースケース
  - `service.go`: アプリケーションサービス

#### スラッシュコマンド一覧

| コマンド | 説明 | パッケージ |
|---|---|---|
| `/ping` | 疎通確認 | `internal/infrastructure/discord/commands/ping/` |
| `/cat` | ランダムな猫画像を表示 | `internal/infrastructure/discord/commands/cat/` |
| `/dog` | ランダムな犬画像を表示 | `internal/infrastructure/discord/commands/dog/` |
| `/mahjong` | 麻雀牌をランダムに引く | `internal/infrastructure/discord/commands/mahjong/` |
| `/omikuji` | 今日の運勢を占う（ユーザー＋日付で決定的） | `internal/infrastructure/discord/commands/omikuji/` |
| `/collatz` | コラッツ予想の計算 | `internal/infrastructure/discord/commands/collatz/` |
| `/faker` | LOL プロプレイヤー Faker の伝説エピソードをランダムに紹介 | `internal/infrastructure/discord/commands/faker/` |

#### インターフェース層（ポート定義）

- `internal/interfaces/discord/`: Discord 統合のためのポートインターフェース
- `internal/interfaces/db/`: データベース統合のためのポートインターフェース

#### インフラストラクチャ層（アダプター実装）

- `internal/infrastructure/discord/`: Discord API との統合
  - `ready_handler.go`: Bot 起動時のイベントハンドラー
  - `interaction_handler.go`: スラッシュコマンドのイベントハンドラー
  - `voice_state_handler.go`: ボイスチャンネル状態変更のイベントハンドラー
  - `command_registrar.go`: コマンド登録機能
  - `commands.go`: コマンド定義
  - `adapter.go`: Discord アダプター実装
- `internal/infrastructure/persistence/`: データ永続化
  - `voicetext_repository.go`: リポジトリ実装
  - `postgres.go`: PostgreSQL 接続管理

#### その他

- `db/migrations/`: データベースマイグレーションファイル
- `docs/`: プロジェクトドキュメント

### 依存関係

- Discord API: `github.com/bwmarrin/discordgo`
- PostgreSQL: `github.com/jackc/pgx/v5`
- 環境変数管理: `github.com/joho/godotenv`
- UUID 生成: `github.com/google/uuid`
- マイグレーション: `golang-migrate`

### 設定

以下の環境変数が必要です：

- `DISCORD_TOKEN`: Discord ボットのトークン
- `DATABASE_URL`: PostgreSQL の接続 URL

### アーキテクチャパターン

このプロジェクトはクリーンアーキテクチャとヘキサゴナルアーキテクチャ（ポート&アダプター）の原則に従っています。

#### 依存関係の方向

- 外側のレイヤーは内側のレイヤーに依存
- 内側のレイヤーは外側のレイヤーを知らない
- ドメイン層は他のどのレイヤーにも依存しない
- インフラストラクチャ層はドメイン層が定義したインターフェース（ポート）を実装

#### データフロー

1. Discord イベント受信 → `infrastructure/discord/*_handler.go`（各イベントハンドラー）
2. ハンドラーがアプリケーションサービスを呼び出し → `application/voicetext/service.go`
3. サービスがドメインモデルとリポジトリを使用 → `domain/voicetext/`
4. リポジトリ実装がデータベースにアクセス → `infrastructure/persistence/`

### データベース設計

`voice_text_links`テーブルでボイスチャンネルとテキストチャンネルのマッピングを管理しています。

## Pull Request ルール

### Base Branch の決定

作業ブランチ名に応じて適切な base branch を選択してください：

#### 1. `feature/**` の場合

- **Base branch**: `develop`
- **例**: `feature/add-command` → `develop` への PR

#### 2. `<specific word>/feature/**` の場合

- **Base branch**: `<specific word>/develop`
- **例**: `voice-text-link/feature/add-database` → `voice-text-link/develop` への PR

#### 3. 依存関係がある `<specific word>/feature/**` の場合

作業ブランチが `<specific word>/feature/<other specific word>` から分岐している場合（つまり別の PR の作業内容に依存している場合）:

- **Base branch**: `<specific word>/feature/<other specific word>`
- **例**: `voice-text-link/feature/add-ui` が `voice-text-link/feature/add-database` から分岐 → `voice-text-link/feature/add-database` への PR

### PR 作成時の注意点

- 分岐元ブランチを確認して適切な base branch を選択する
- 依存関係がある場合は、依存先の PR が先にマージされる必要がある

## ブランチ作業ルール

### ブランチの作成について

指示を受けて作業するとき、現在のブランチに応じて適切にブランチを切って作業してください：

#### 1. 現在のブランチが `develop` の場合

- `feature/**` パターンのブランチを作成します
- **例**: `feature/add-new-command`

#### 2. 現在のブランチが `<specific word>/develop` の場合

- `<specific word>/feature/**` パターンのブランチを作成します
- **例**: `voice-text-link/develop` → `voice-text-link/feature/add-new-feature`

### ブランチ作業前の確認事項

#### 新しいブランチを切る前

1. remote に push していないコミットがないか確認する
2. push していないコミットが見つかった場合は pull してからブランチを切る

#### すでに作業ブランチにいる場合

1. 作業開始前に remote に push していないコミットがないか確認する
2. 必要に応じて pull してから作業を開始する

### ブランチ確認コマンド

```bash
# リモートとの差分確認
git status

# リモートの最新情報を取得
git fetch

# リモートブランチとの比較
git log --oneline origin/HEAD..HEAD
```

# discord-bot-go

Go 言語で書かれた Discord ボットです。音声チャンネルとテキストチャンネルをリンクする機能や、各種スラッシュコマンドを提供します。

## セットアップ

### 環境変数の準備

`.env.example` をコピーして `.env` を作成し、必要な値を設定します。

```bash
cp .env.example .env
```

| 変数名 | 説明 |
|---|---|
| `DISCORD_TOKEN` | Discord ボットのトークン |
| `DATABASE_URL` | PostgreSQL の接続 URL |
| `POSTGRES_USER` | PostgreSQL のユーザー名 |
| `POSTGRES_PASSWORD` | PostgreSQL のパスワード |
| `POSTGRES_DB` | PostgreSQL のデータベース名 |

> **注意**: `DATABASE_URL` のホスト部分は Docker Compose で起動する場合は `db`、ローカルで直接実行する場合は `localhost` を指定してください。

## ローカル環境での起動

### Docker を使った起動（推奨）

Docker Compose で PostgreSQL とボットをまとめて起動できます。マイグレーションも自動で実行されます。

```bash
docker-compose up
```

バックグラウンドで起動する場合：

```bash
docker-compose up -d
```

停止する場合：

```bash
docker-compose down
```

### ローカルで直接実行する場合

先に DB だけ起動し、マイグレーションを適用してからボットを実行します。

```bash
# DB 起動
docker-compose up -d db

# マイグレーション適用
migrate -path db/migrations -database "$DATABASE_URL" up

# ボット起動
go run cmd/bot/main.go
```

## テスト実行

```bash
go test ./...
```

## スラッシュコマンド一覧

| コマンド | 説明 |
|---|---|
| `/ping` | 疎通確認 |
| `/cat` | ランダムな猫画像を表示 |
| `/dog` | ランダムな犬画像を表示 |
| `/mahjong` | 麻雀牌をランダムに引く |
| `/omikuji` | 今日の運勢を占う（ユーザー＋日付で決定的） |
| `/collatz` | コラッツ予想の計算 |
| `/faker` | LOL プロプレイヤー Faker の伝説エピソードをランダムに紹介 |
| `/jeff-dean` | Google のエンジニア Jeff Dean の伝説をランダムに紹介 |

## データベース（Migration）

golang-migrate を使用。

### インストール方法

- Mac
  ```bash
  brew install golang-migrate
  ```

### Migration ファイルの作成

```bash
migrate create -ext sql -dir db/migrations -seq create_voice_text_links
```

### Migration ファイル適用

```bash
migrate -path db/migrations -database "$DATABASE_URL" up
```

### 本番環境の Migration

以下を実行してからマイグレーションする

```sql
SET ROLE mybot_owner;
SELECT current_user, session_user;
```

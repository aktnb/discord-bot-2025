# discord-bot-go

Go 言語で書かれた Discord ボットです。音声チャンネルとテキストチャンネルをリンクする機能や、各種スラッシュコマンドを提供します。

## 開発コマンド

### アプリケーションの実行

```bash
go run cmd/bot/main.go
```

### テスト実行

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

## 環境変数

| 変数名 | 説明 |
|---|---|
| `DISCORD_TOKEN` | Discord ボットのトークン |
| `DATABASE_URL` | PostgreSQL の接続 URL |

## データベース（Migration）

golang-migrate を使用．

#### インストール方法

- Mac
  ```bash
  brew install golang-migrate
  ```

#### Migration ファイルの作成

```bash
migrate create -ext sql -dir db/migrations -seq create_voice_text_links
```

#### Migration ファイル適用

```bash
migrate -path db/migrations -database "$DATABASE_URL" up
```

#### 本番環境の Migration

以下を実行してからマイグレーションする

```sql
SET ROLE mybot_owner;
SELECT current_user, session_user;
```

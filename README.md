## Migration

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

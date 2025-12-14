# Discord Bot デプロイメントガイド

このドキュメントでは、Discord Botを本番環境（ARM64 Ubuntu 24.04.3）にデプロイする手順を説明します。

## 概要

このデプロイメント構成では、以下の機能を提供します：

- **自動デプロイ**: GitHub Releasesから最新版を自動的に検出してデプロイ
- **Blue-Green デプロイメント**: ダウンタイムを最小限に抑えたデプロイ
- **自動ロールバック**: デプロイ失敗時に前バージョンへ自動復旧
- **定期チェック**: 2分ごとに新しいリリースをチェック

## ディレクトリ構造

```
/opt/bot-user/
├── shared/
│   ├── bot.env          # 環境変数ファイル
│   └── last_tag         # 最後にデプロイしたタグ
├── releases/
│   ├── v1.0.0/          # リリースごとのディレクトリ
│   │   ├── discord-bot-linux-arm64
│   │   ├── discord-bot-linux-arm64.tar.gz
│   │   └── SHA256SUMS.txt
│   └── v1.0.1/
├── current -> releases/v1.0.1/discord-bot-linux-arm64  # 現在のバージョン
└── previous -> releases/v1.0.0/discord-bot-linux-arm64 # 前のバージョン（ロールバック用）
```

## 初回セットアップ

### 1. bot_user の作成

```bash
sudo useradd -r -m -d /opt/bot-user -s /bin/bash bot_user
```

### 2. 必要なディレクトリの作成

```bash
sudo mkdir -p /opt/bot-user/shared
sudo mkdir -p /opt/bot-user/releases
sudo chown -R bot_user:bot_user /opt/bot-user
```

### 3. 環境変数ファイルの作成

```bash
sudo tee /opt/bot-user/shared/bot.env > /dev/null <<EOF
DISCORD_TOKEN=your_discord_token_here
DATABASE_URL=postgres://user:password@localhost:5432/dbname
EOF

sudo chmod 600 /opt/bot-user/shared/bot.env
sudo chown bot_user:bot_user /opt/bot-user/shared/bot.env
```

### 4. bot-update スクリプトのインストール

```bash
sudo cp deployment/bot-update /usr/local/bin/bot-update
sudo chmod +x /usr/local/bin/bot-update
```

### 5. systemd サービスのインストール

```bash
# サービスファイルをコピー
sudo cp deployment/systemd/bot.service /etc/systemd/system/
sudo cp deployment/systemd/bot-update.service /etc/systemd/system/
sudo cp deployment/systemd/bot-update.timer /etc/systemd/system/

# systemd をリロード
sudo systemctl daemon-reload
```

### 6. サービスの有効化と起動

```bash
# bot.service を有効化（自動起動設定）
sudo systemctl enable bot.service

# bot-update.timer を有効化して起動
sudo systemctl enable bot-update.timer
sudo systemctl start bot-update.timer
```

### 7. 初回デプロイの実行

```bash
# 手動で初回デプロイを実行
sudo /usr/local/bin/bot-update

# サービスを起動
sudo systemctl start bot.service
```

## 運用

### サービスの状態確認

```bash
# bot サービスの状態確認
sudo systemctl status bot.service

# bot-update タイマーの状態確認
sudo systemctl status bot-update.timer

# bot-update サービスの最終実行結果確認
sudo systemctl status bot-update.service
```

### ログの確認

```bash
# bot サービスのログ
sudo journalctl -u bot.service -f

# bot-update のログ
sudo journalctl -u bot-update.service -f

# タイマーのログ
sudo journalctl -u bot-update.timer -f
```

### 手動デプロイ

自動デプロイを待たずに、手動で最新版をデプロイする場合：

```bash
sudo systemctl start bot-update.service
```

または直接スクリプトを実行：

```bash
sudo /usr/local/bin/bot-update
```

### サービスの再起動

```bash
sudo systemctl restart bot.service
```

### サービスの停止

```bash
sudo systemctl stop bot.service
```

### 自動更新の一時停止

```bash
# タイマーを停止
sudo systemctl stop bot-update.timer

# タイマーを再開
sudo systemctl start bot-update.timer
```

### 手動ロールバック

自動ロールバックが機能しなかった場合の手動ロールバック：

```bash
# previous リンクを確認
ls -la /opt/bot-user/previous

# current と previous を入れ替え
sudo rm -f /opt/bot-user/current
sudo mv /opt/bot-user/previous /opt/bot-user/current

# サービスを再起動
sudo systemctl restart bot.service

# last_tag を更新（例：v1.0.0 にロールバック）
echo "v1.0.0" | sudo tee /opt/bot-user/shared/last_tag
```

### 特定バージョンへの手動デプロイ

```bash
# 利用可能なリリースを確認
ls -la /opt/bot-user/releases/

# 特定バージョンを指定してシンボリックリンクを作成
sudo rm -f /opt/bot-user/previous
sudo mv /opt/bot-user/current /opt/bot-user/previous
sudo ln -sf /opt/bot-user/releases/v1.0.0/discord-bot-linux-arm64 /opt/bot-user/current

# サービスを再起動
sudo systemctl restart bot.service

# last_tag を更新
echo "v1.0.0" | sudo tee /opt/bot-user/shared/last_tag
```

## トラブルシューティング

### サービスが起動しない

1. ログを確認：
   ```bash
   sudo journalctl -u bot.service -n 50 --no-pager
   ```

2. 環境変数が正しく設定されているか確認：
   ```bash
   sudo cat /opt/bot-user/shared/bot.env
   ```

3. バイナリが存在し、実行権限があるか確認：
   ```bash
   ls -la /opt/bot-user/current
   ```

### 自動更新が動作しない

1. タイマーが有効化されているか確認：
   ```bash
   sudo systemctl is-enabled bot-update.timer
   sudo systemctl is-active bot-update.timer
   ```

2. 次回実行予定を確認：
   ```bash
   sudo systemctl list-timers bot-update.timer
   ```

3. 最後の実行結果を確認：
   ```bash
   sudo journalctl -u bot-update.service -n 50 --no-pager
   ```

4. 手動で実行してエラーを確認：
   ```bash
   sudo /usr/local/bin/bot-update
   ```

### ダウンロードが失敗する

1. ネットワーク接続を確認：
   ```bash
   curl -I https://github.com
   ```

2. GitHub APIへのアクセスを確認：
   ```bash
   curl -s https://api.github.com/repos/aktnb/discord-bot-2025/releases/latest
   ```

3. ディスク容量を確認：
   ```bash
   df -h /opt/bot-user
   ```

### チェックサム検証が失敗する

ダウンロードが破損している可能性があります：

```bash
# 該当リリースを削除
sudo rm -rf /opt/bot-user/releases/v1.0.x

# 再度ダウンロードを試行
sudo systemctl start bot-update.service
```

## セキュリティ考慮事項

1. **環境変数ファイルのパーミッション**:
   - `bot.env` は 600 (所有者のみ読み書き可能) に設定
   - 所有者は `bot_user` に設定

2. **実行権限**:
   - bot サービスは `bot_user` で実行（最小権限の原則）
   - bot-update サービスは `root` で実行（systemctl コマンド実行のため）

3. **systemd セキュリティ設定**:
   - `NoNewPrivileges=true`: 特権昇格を防止
   - `PrivateTmp=true`: プライベートな /tmp を使用
   - `ProtectSystem=strict`: システムディレクトリを読み取り専用に
   - `ProtectHome=true`: ホームディレクトリへのアクセスを制限

4. **ファイアウォール**:
   - 必要なポートのみを開放
   - データベースへのアクセスはローカルネットワークのみに制限

## 参考情報

### systemd コマンド

```bash
# サービスの有効化/無効化
sudo systemctl enable bot.service
sudo systemctl disable bot.service

# サービスの起動/停止/再起動
sudo systemctl start bot.service
sudo systemctl stop bot.service
sudo systemctl restart bot.service

# サービスの状態確認
sudo systemctl status bot.service
sudo systemctl is-active bot.service
sudo systemctl is-enabled bot.service

# systemd 設定のリロード
sudo systemctl daemon-reload
```

### ログコマンド

```bash
# リアルタイムでログを監視
sudo journalctl -u bot.service -f

# 最新50行を表示
sudo journalctl -u bot.service -n 50

# 特定の時間範囲のログを表示
sudo journalctl -u bot.service --since "2025-01-01 00:00:00" --until "2025-01-01 23:59:59"

# エラーのみを表示
sudo journalctl -u bot.service -p err
```

## ライセンス

このプロジェクトのライセンスについては、リポジトリのLICENSEファイルを参照してください。

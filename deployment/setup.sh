#!/usr/bin/env bash
# setup.sh - Discord Bot 初回サーバーセットアップスクリプト
#
# 使い方:
#   sudo bash deployment/setup.sh
#
# このスクリプトはべき等（何度実行しても同じ結果になる）に設計されています。

set -euo pipefail

# ─── 定数 ────────────────────────────────────────────────────────────────────

BOT_USER="bot_user"
BOT_DIR="/opt/bot-user"
UPDATE_SCRIPT_SRC="$(cd "$(dirname "$0")" && pwd)/bot-update"
UPDATE_SCRIPT_DST="/usr/local/bin/bot-update"
SYSTEMD_SRC="$(cd "$(dirname "$0")" && pwd)/systemd"
SYSTEMD_DST="/etc/systemd/system"

# ─── ヘルパー関数 ────────────────────────────────────────────────────────────

log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"; }
die() { echo "ERROR: $*" >&2; exit 1; }

require_root() {
  [[ "$(id -u)" -eq 0 ]] || die "このスクリプトはrootで実行してください: sudo bash $0"
}

# ─── システムパッケージ ───────────────────────────────────────────────────────

install_packages() {
  log "システムパッケージをインストールします..."
  apt-get update -qq
  apt-get install -y --no-install-recommends \
    ffmpeg \
    curl \
    ca-certificates

  log "ffmpeg バージョン: $(ffmpeg -version 2>&1 | head -1)"
}

# ─── bot_user の作成 ─────────────────────────────────────────────────────────

create_bot_user() {
  if id "$BOT_USER" &>/dev/null; then
    log "ユーザー '$BOT_USER' はすでに存在します（スキップ）"
  else
    log "ユーザー '$BOT_USER' を作成します..."
    useradd -r -m -d "$BOT_DIR" -s /usr/sbin/nologin "$BOT_USER"
  fi
}

# ─── ディレクトリ構造の作成 ──────────────────────────────────────────────────

create_directories() {
  log "ディレクトリを作成します..."
  mkdir -p "$BOT_DIR/shared"
  mkdir -p "$BOT_DIR/releases"
  chown -R "$BOT_USER:$BOT_USER" "$BOT_DIR"
}

# ─── 環境変数ファイルの作成 ──────────────────────────────────────────────────

create_env_file() {
  local env_file="$BOT_DIR/shared/bot.env"

  if [[ -f "$env_file" ]]; then
    log "環境変数ファイルはすでに存在します（スキップ）: $env_file"
    return
  fi

  log "環境変数ファイルのテンプレートを作成します: $env_file"
  cat > "$env_file" <<'EOF'
DISCORD_TOKEN=your_discord_token_here
DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
EOF
  chmod 600 "$env_file"
  chown "$BOT_USER:$BOT_USER" "$env_file"

  log ""
  log ">>> $env_file を編集して、実際のトークンとデータベースURLを設定してください <<<"
  log ""
}

# ─── bot-update スクリプトのインストール ────────────────────────────────────

install_update_script() {
  log "bot-update スクリプトをインストールします..."
  [[ -f "$UPDATE_SCRIPT_SRC" ]] || die "bot-update スクリプトが見つかりません: $UPDATE_SCRIPT_SRC"
  cp "$UPDATE_SCRIPT_SRC" "$UPDATE_SCRIPT_DST"
  chmod +x "$UPDATE_SCRIPT_DST"
}

# ─── systemd サービスのインストール ─────────────────────────────────────────

install_systemd_services() {
  log "systemd サービスをインストールします..."

  for file in bot.service bot-update.service bot-update.timer; do
    [[ -f "$SYSTEMD_SRC/$file" ]] || die "systemd ファイルが見つかりません: $SYSTEMD_SRC/$file"
    cp "$SYSTEMD_SRC/$file" "$SYSTEMD_DST/$file"
  done

  systemctl daemon-reload

  log "bot.service を有効化します..."
  systemctl enable bot.service

  log "bot-update.timer を有効化して起動します..."
  systemctl enable bot-update.timer
  systemctl start bot-update.timer
}

# ─── 初回デプロイ ────────────────────────────────────────────────────────────

run_initial_deploy() {
  log "初回デプロイを実行します..."
  if "$UPDATE_SCRIPT_DST"; then
    log "デプロイ成功。bot.service を起動します..."
    systemctl start bot.service
  else
    log "警告: 初回デプロイが失敗しました。"
    log "  'sudo $UPDATE_SCRIPT_DST' を手動で実行してデプロイしてください。"
  fi
}

# ─── メイン ──────────────────────────────────────────────────────────────────

main() {
  require_root

  log "=== Discord Bot セットアップ開始 ==="

  install_packages
  create_bot_user
  create_directories
  create_env_file
  install_update_script
  install_systemd_services
  run_initial_deploy

  log ""
  log "=== セットアップ完了 ==="
  log ""
  log "次のステップ:"
  log "  1. $BOT_DIR/shared/bot.env を編集して環境変数を設定する"
  log "  2. sudo systemctl restart bot.service でボットを再起動する"
  log "  3. sudo journalctl -u bot.service -f でログを確認する"
}

main "$@"

---
name: create-pr
description: >
  ブランチを作成して GitHub の Pull Request を作成するスキル。
  ユーザーが PR 作成、ブランチ作成、GitHub への変更提出を依頼したときは必ずこのスキルを使用すること。
  「PR を作って」「プルリクを出して」「ブランチを切って PR を作りたい」「feature ブランチに切り替えて」などの
  依頼があった場合に適用する。gh コマンドを使って PR を作成する方法を提供する。
---

# PR 作成スキル

このスキルは、ブランチ戦略に従ってブランチを作成し、GitHub CLI (`gh`) を使って Pull Request を作成するワークフローを提供する。

## ブランチ作成ルール

現在のブランチに応じて、作成するブランチパターンが決まる。

| 現在のブランチ | 作成するブランチ |
|---|---|
| `master` | `feature/<作業内容>` |
| `develop` | `feature/<作業内容>` |
| `<project>/develop` | `<project>/feature/<作業内容>` |
| `<project>/feature/<X>`（依存関係あり） | `<project>/feature/<Y>` |

## PR の base branch ルール

PR を作成するとき、作業ブランチに応じて base branch を決定する。

### ケース 1: `feature/**` ブランチの場合

- **Base branch**: `develop`（または `master`）
- **例**: `feature/add-command` → `develop` への PR

### ケース 2: `<project>/feature/**` ブランチの場合

- **Base branch**: `<project>/develop`
- **例**: `voice-text-link/feature/add-database` → `voice-text-link/develop` への PR

### ケース 3: 依存関係がある `<project>/feature/**` の場合

作業ブランチが別の feature ブランチから分岐している場合（その PR の作業内容に依存している場合）:

- **Base branch**: `<project>/feature/<依存先>`
- **例**: `voice-text-link/feature/add-ui` が `voice-text-link/feature/add-database` から分岐 → `voice-text-link/feature/add-database` への PR
- 依存先の PR がマージされるまで、base はその feature ブランチのまま

## 作業前の確認（必須）

ブランチを切る前に必ず実施する:

```bash
# リモートの最新情報を取得
git fetch

# 未 push コミットがないか確認
git status
git log --oneline origin/HEAD..HEAD

# 未 push コミットがある場合は push してからブランチを切る
git push
```

すでに作業ブランチにいる場合も、作業開始前に同様に確認する。

## ブランチ作成

```bash
git checkout -b <branch-name>
# または
git switch -c <branch-name>
```

## gh コマンドを使った PR 作成

### 基本的な PR 作成

```bash
gh pr create \
  --title "PR のタイトル" \
  --body "PR の説明" \
  --base <base-branch>
```

### 主要なオプション

| オプション | 短縮形 | 説明 |
|---|---|---|
| `--title` | `-t` | PR のタイトル |
| `--body` | `-b` | PR の本文（Markdown 可） |
| `--base` | `-B` | マージ先のブランチ |
| `--head` | `-H` | PR の head ブランチ（省略時は現在のブランチ） |
| `--draft` | `-d` | ドラフト PR として作成 |
| `--reviewer` | `-r` | レビュワーを指定（カンマ区切りで複数可） |
| `--assignee` | `-a` | アサイニーを指定（`@me` で自分自身） |
| `--label` | `-l` | ラベルを追加 |
| `--fill` | `-f` | コミットメッセージからタイトル・本文を自動入力 |
| `--web` | `-w` | ブラウザで PR 作成画面を開く |

### ベストプラクティス: HEREDOC でボディを渡す

改行や特殊文字を含む場合は HEREDOC を使う:

```bash
gh pr create \
  --title "feat: 新機能を追加" \
  --body "$(cat <<'EOF'
## 概要
- 新機能 A を追加
- バグ B を修正

## テスト方法
1. `go test ./...` を実行
2. 動作を確認

🤖 Generated with Claude Code
EOF
)" \
  --base master
```

### ドラフト PR の作成

```bash
gh pr create \
  --title "WIP: 作業中" \
  --body "作業中の PR です" \
  --base master \
  --draft
```

### PR の確認・操作

```bash
# PR 一覧表示
gh pr list

# 現在のブランチの PR を表示
gh pr view

# PR をブラウザで開く
gh pr view --web

# PR のステータス（CI など）を確認
gh pr checks

# PR をマージ
gh pr merge

# PR をクローズ
gh pr close <PR番号>
```

### リモートに push していない場合

`gh pr create` 実行時、未 push のコミットがある場合は自動的に push するか確認プロンプトが出る。
事前に push しておくと安全:

```bash
git push -u origin <branch-name>
gh pr create --title "..." --body "..." --base master
```

## 完全なワークフロー例

### master から feature ブランチを切って PR を作る

```bash
# 1. リモートを最新化
git fetch
git log --oneline origin/master..HEAD  # 未 push コミットがないか確認

# 2. ブランチ作成
git checkout -b feature/add-new-command

# 3. 作業 → コミット
git add .
git commit -m "feat: 新しいコマンドを追加"

# 4. push
git push -u origin feature/add-new-command

# 5. PR 作成
gh pr create \
  --title "feat: 新しいコマンドを追加" \
  --body "$(cat <<'EOF'
## 概要
新しいスラッシュコマンドを追加しました。

## 変更内容
- `/newcommand` コマンドを追加

## テスト
- [ ] `go test ./...` が通ることを確認
EOF
)" \
  --base master
```

### `<project-name>/develop` から feature ブランチを切って PR を作る

```bash
# 1. develop ブランチに切り替えてリモートを最新化
git checkout voice-text-link/develop
git fetch
git pull

# 2. feature ブランチ作成
git checkout -b voice-text-link/feature/add-database

# 3. 作業 → コミット
git add .
git commit -m "feat: データベーステーブルを追加"

# 4. push
git push -u origin voice-text-link/feature/add-database

# 5. PR 作成（base は develop ブランチ）
gh pr create \
  --title "feat: データベーステーブルを追加" \
  --body "## 概要\nデータベーステーブルを追加しました。" \
  --base voice-text-link/develop
```

## PR 作成後の確認

```bash
# 作成した PR の URL を表示
gh pr view

# CI ステータスを監視
gh pr checks --watch
```

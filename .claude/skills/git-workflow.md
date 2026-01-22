---
description: ブランチ戦略、コミットフロー、PR作成手順
---

# Git Workflow

VRC Shift Scheduler のブランチ戦略と開発フロー。

---

## ブランチ構成

| ブランチ | 環境 | 用途 |
|---------|------|------|
| `main` | 本番環境 | リリース済みの安定版コード |
| `develop` | ステージング環境 | 次回リリース候補のコード |
| `feature/*` | - | 機能開発用 |
| `fix/*` | - | バグ修正用 |

---

## 基本ワークフロー

```
feature/xxx  →  develop (STG確認)  →  main (本番リリース)
```

### 1. 機能開発・バグ修正

```bash
# developから新しいブランチを作成
git checkout develop
git pull origin develop
git checkout -b feature/add-new-feature

# 開発・コミット
git add .
git commit -m "feat: 新機能を追加"

# プッシュ
git push -u origin feature/add-new-feature
```

### 2. PRを作成（develop向け）

```bash
gh pr create --base develop --title "feat: 新機能を追加" --body "## Summary
- 機能の説明

## Test plan
- [ ] テスト項目"
```

### 3. マージ後、本番リリース

```bash
# develop → main のPRを作成
gh pr create --base main --head develop --title "Release: v0.x.x"
```

---

## 重要ルール

### develop → main のマージ

**必ず「Create a merge commit」を使用する**

| マージ方法 | 使用可否 |
|-----------|---------|
| Create a merge commit | ✅ 推奨 |
| Squash and merge | ❌ 禁止 |
| Rebase and merge | ❌ 禁止 |

**理由**: スカッシュマージを使用すると、次回マージ時にコンフリクトが発生する

### feature/fix → develop のマージ

どのマージ方法でもOK。コミット履歴を整理したい場合は「Squash and merge」可。

---

## コミットメッセージ規約

```
<type>: <description>

[optional body]

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
```

### Type

| Type | 用途 |
|------|------|
| `feat` | 新機能 |
| `fix` | バグ修正 |
| `docs` | ドキュメント |
| `style` | フォーマット（機能変更なし） |
| `refactor` | リファクタリング |
| `test` | テスト追加・修正 |
| `chore` | ビルド、CI、依存関係 |

### 例

```bash
git commit -m "$(cat <<'EOF'
feat: 出欠確認機能を追加

- 公開URL経由での回答機能
- 回答締切機能

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## PR作成手順

### 標準的なPR

```bash
gh pr create --base develop --title "feat: 機能名" --body "$(cat <<'EOF'
## Summary
- 変更内容1
- 変更内容2

## Test plan
- [ ] テスト項目1
- [ ] テスト項目2

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

### リリースPR（develop → main）

```bash
gh pr create --base main --head develop --title "Release: v0.x.x" --body "$(cat <<'EOF'
## Summary
- 含まれる変更の概要

## Changelog
- feat: 機能1
- fix: バグ修正1

## Test plan
- [ ] ステージング環境で動作確認済み

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

---

## タグ付け

デプロイ成功後、ローカルからタグを付与：

```bash
# 最新のmainを取得
git checkout main
git pull origin main

# 現在のバージョンタグを確認
git tag --list 'v*' --sort=-v:refname | head -5

# 新しいタグを作成
git tag -a v0.2.0 -m "Release v0.2.0: 機能追加・バグ修正"

# タグをリモートにプッシュ
git push origin v0.2.0
```

### タグ命名規則

```
v<MAJOR>.<MINOR>.<PATCH>
```

| セグメント | 用途 |
|-----------|------|
| MAJOR | 破壊的変更 |
| MINOR | 新機能追加 |
| PATCH | バグ修正 |

---

## トラブルシューティング

### コンフリクトが発生した場合

```bash
# mainブランチをチェックアウト
git checkout main
git pull origin main

# developをマージ（コンフリクト発生）
git merge origin/develop --no-commit

# developの内容を優先して解決
git checkout --theirs <conflicted-files>
git add <conflicted-files>

# マージコミットを作成
git commit -m "Merge branch 'develop' into main"
git push origin main
```

### 履歴をリセットする場合（最終手段）

```bash
# ⚠️ 注意: mainの独自変更が失われます
git checkout main
git reset --hard origin/develop
git push --force origin main
```

---

## 禁止事項

1. **main への直接push** - PR経由のみ
2. **develop → main でのスカッシュマージ** - コンフリクトの原因
3. **force push（特別な理由がない限り）**

---

## 関連ドキュメント

- `docs/BRANCH_STRATEGY.md` - 詳細なブランチ運用ガイド
- `docs/PRODUCTION_DEPLOYMENT.md` - 本番デプロイ手順

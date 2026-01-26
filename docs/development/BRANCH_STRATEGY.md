# ブランチ運用ガイド

## 現状

このプロジェクトでは以下のブランチ構成を採用しています：

| ブランチ | 環境 | 用途 |
|---------|------|------|
| `main` | 本番環境 | リリース済みの安定版コード |
| `develop` | ステージング環境 | 次回リリース候補のコード |
| `feature/*`, `fix/*` | - | 機能開発・バグ修正用 |

### 基本的なワークフロー

```
feature/xxx  →  develop (STG確認)  →  main (本番リリース)
```

1. `develop` から feature/fix ブランチを作成
2. 開発完了後、`develop` へPRを作成してマージ
3. ステージング環境で動作確認
4. 確認完了後、`develop` → `main` へPRを作成してマージ
5. 本番環境へデプロイ

## 問題点：スカッシュマージによるコンフリクト

### 発生した問題

2026年1月17日、`develop` → `main` のマージ時にコンフリクトが発生しました。

**原因**: PR #164 で `develop` → `main` をマージする際に **スカッシュマージ** を使用したため。

### なぜスカッシュマージが問題を引き起こすか

```
【正常なマージコミット】
develop: A → B → C → D → E
                      ↓ (merge commit: 2つの親を持つ)
main:    A → B → C → M
                    ↑
         親1: C (main)
         親2: D (develop)

→ 次回マージ時、Gitは「Mで既にDまでマージ済み」と認識
→ E以降の差分のみをマージ

【スカッシュマージ】
develop: A → B → C → D → E
                      ↓ (squash: 1つの親のみ)
main:    A → B → C → S
                    ↑
         親: C のみ
         S = D+Eの内容を1つにまとめた新規コミット

→ 次回マージ時、Gitは「CまでしかマージしてないからDとEもマージ対象」と誤認識
→ 同じ変更が両方のブランチにあるとコンフリクト
```

### 実際に発生したコンフリクト

```
CONFLICT: backend/internal/domain/shift/shift_slot.go
CONFLICT: backend/internal/domain/shift/shift_slot_test.go
CONFLICT: backend/internal/infra/db/shift_slot_repository.go
CONFLICT: web-frontend/src/pages/ShiftSlotList.tsx
CONFLICT: web-frontend/tests/api/shift-slot.spec.ts
```

これらはすべてPR #164で既にmainに取り込まれていた変更でしたが、スカッシュマージのためGitが認識できませんでした。

## 解決策

### 今後のルール

**`develop` → `main` のマージ時は必ず「Create a merge commit」を使用する**

GitHub上でPRをマージする際：

1. 「Merge pull request」ボタンの横の ▼ をクリック
2. **「Create a merge commit」** を選択（デフォルト）
3. 「Squash and merge」や「Rebase and merge」は使用しない

### なぜ「Create a merge commit」を使うべきか

| マージ方法 | 履歴 | develop→main での使用 |
|-----------|------|----------------------|
| Create a merge commit | 両ブランチの履歴を保持 | ✅ 推奨 |
| Squash and merge | 複数コミットを1つに圧縮 | ❌ 禁止 |
| Rebase and merge | 履歴を直線化 | ❌ 禁止 |

### feature/fix → develop のマージ

feature/fix ブランチから `develop` へのマージでは、どのマージ方法でも問題ありません。
コミット履歴を整理したい場合は「Squash and merge」を使用しても構いません。

## ゴール

- `main` と `develop` は常に同期された状態を保つ
- `develop` → `main` のマージでコンフリクトが発生しない
- 本番リリース時のオペレーションがシンプルで安全

## トラブルシューティング

### コンフリクトが発生した場合

もしスカッシュマージによりコンフリクトが発生した場合：

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

履歴が複雑になりすぎた場合、mainをdevelopに強制同期することも可能です：

```bash
# ⚠️ 注意: mainの独自変更が失われます
git checkout main
git reset --hard origin/develop
git push --force origin main
```

## 関連インシデント

- 2026-01-17: PR #164 スカッシュマージによるコンフリクト発生
  - 対応: 手動でコンフリクト解決、マージコミット作成

## 参考

- [GitHub: About merge methods](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/about-merge-methods-on-github)

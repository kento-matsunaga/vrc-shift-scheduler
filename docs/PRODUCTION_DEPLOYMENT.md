# 本番デプロイ運用ガイド

> VRC Shift Scheduler の本番環境デプロイ・運用手順書

---

## 1. 運用方針

### 1.1 基本方針

| 項目 | 方針 |
|------|------|
| リポジトリ構成 | 1リポジトリ運用（モノレポ） |
| 本番ブランチ | `main` = 本番反映ブランチ |
| デプロイ方式 | サーバーで `git pull` → `docker compose up` |
| タグ付け | デプロイ成功**後**にローカルからタグを付与 |
| 機密情報 | `.env.prod` 等はGitに入れず、サーバーにのみ配置 |

### 1.2 タグ付けルール

```
デプロイ成功 → 動作確認 → タグ付け（v0.x.y）
```

タグはデプロイが成功し、動作確認が完了した**後**に付与します。これにより、タグが付いたコミットは「本番で動作実績のあるバージョン」として信頼できます。

---

## 2. 前提条件

### 2.1 本番サーバー要件

- **OS**: Ubuntu 22.04 LTS 以上（または同等のLinuxディストリビューション）
- **Docker**: 24.0 以上
- **Docker Compose**: v2.20 以上（`docker compose` コマンドが使用可能）
- **Git**: 2.30 以上
- **メモリ**: 2GB 以上推奨
- **ディスク**: 20GB 以上推奨

### 2.2 事前準備（初回のみ）

#### リポジトリのクローン

```bash
cd /opt
sudo git clone https://github.com/<org>/vrc-shift-scheduler.git
sudo chown -R $USER:$USER vrc-shift-scheduler
cd vrc-shift-scheduler
```

#### 環境変数ファイルの作成

```bash
# .env.prod を作成（このファイルはGitに含めない）
cp .env.example .env.prod
vim .env.prod
```

**.env.prod の必須設定:**

```bash
# データベース
DATABASE_URL=postgres://vrcshift:<強力なパスワード>@db:5432/vrcshift?sslmode=disable
POSTGRES_USER=vrcshift
POSTGRES_PASSWORD=<強力なパスワード>
POSTGRES_DB=vrcshift

# 認証
JWT_SECRET=<64文字以上のランダム文字列>

# アプリケーション
PORT=8080
NODE_ENV=production

# フロントエンド
VITE_API_BASE_URL=https://your-domain.com
```

**JWT_SECRET の生成例:**

```bash
openssl rand -base64 64 | tr -d '\n'
```

---

## 3. ブランチ運用

### 3.1 ブランチ構成

```
main          ← 本番環境に反映されるブランチ
  ↑
develop       ← 開発用統合ブランチ（任意）
  ↑
feature/*     ← 機能開発ブランチ
fix/*         ← バグ修正ブランチ
```

### 3.2 運用ルール

| ルール | 説明 |
|--------|------|
| main への直接push | **禁止**（GitHub設定でブランチ保護を推奨） |
| main へのマージ | Pull Request 経由のみ |
| マージ前の確認 | ローカルでのビルド・テスト成功を確認 |

### 3.3 推奨ワークフロー

```bash
# 1. 機能ブランチを作成
git checkout main
git pull origin main
git checkout -b feature/add-new-feature

# 2. 開発・コミット
git add .
git commit -m "feat: 新機能を追加"

# 3. プッシュしてPRを作成
git push origin feature/add-new-feature
# → GitHub でPRを作成、レビュー後にマージ
```

---

## 4. 本番デプロイ手順

### 4.1 デプロイ前チェックリスト

- [ ] PRがマージされ、`main` ブランチが最新である
- [ ] ローカルでビルド・テストが成功している
- [ ] 破壊的変更がある場合、マイグレーション手順を確認済み
- [ ] `.env.prod` の設定が最新である

### 4.2 デプロイコマンド（サーバーで実行）

```bash
# 1. プロジェクトディレクトリに移動
cd /opt/vrc-shift-scheduler

# 2. 最新のコードを取得
git fetch origin
git checkout main
git pull origin main

# 3. 現在のコミットハッシュを記録（ロールバック用）
git rev-parse HEAD > /tmp/deploy_commit.txt
echo "Deploying commit: $(cat /tmp/deploy_commit.txt)"

# 4. コンテナを再ビルド・起動
docker compose -f docker-compose.prod.yml --env-file .env.prod down
docker compose -f docker-compose.prod.yml --env-file .env.prod build --no-cache
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d

# 5. コンテナの起動確認
docker compose -f docker-compose.prod.yml ps

# 6. ログの確認（エラーがないことを確認）
docker compose -f docker-compose.prod.yml logs --tail=50 backend
docker compose -f docker-compose.prod.yml logs --tail=50 frontend
```

### 4.3 マイグレーションの実行（必要な場合）

```bash
# マイグレーションを実行
docker compose -f docker-compose.prod.yml exec backend ./migrate up

# マイグレーション状態の確認
docker compose -f docker-compose.prod.yml exec backend ./migrate version
```

### 4.4 動作確認

```bash
# ヘルスチェック
curl -s http://localhost:8080/health
# 期待: {"status":"ok"}

# ログイン確認（任意）
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}'
```

---

## 5. デプロイ後のタグ付け

### 5.1 タグ付け手順（ローカルPCで実行）

デプロイが成功し、動作確認が完了した後、**ローカルPC**から以下を実行します。

```bash
# 1. 最新のmainを取得
git checkout main
git pull origin main

# 2. 現在のバージョンタグを確認
git tag --list 'v*' --sort=-v:refname | head -5

# 3. 新しいタグを作成（例: v0.2.0）
git tag -a v0.2.0 -m "Release v0.2.0: 機能追加・バグ修正"

# 4. タグをリモートにプッシュ
git push origin v0.2.0
```

### 5.2 タグ命名規則

```
v<MAJOR>.<MINOR>.<PATCH>
```

| セグメント | 用途 | 例 |
|-----------|------|-----|
| MAJOR | 破壊的変更、大規模リニューアル | v1.0.0 |
| MINOR | 新機能追加、後方互換あり | v0.2.0 |
| PATCH | バグ修正、軽微な改善 | v0.1.3 |

### 5.3 タグメッセージの例

```bash
# 機能追加の場合
git tag -a v0.3.0 -m "Release v0.3.0: 出欠確認機能を追加"

# バグ修正の場合
git tag -a v0.2.1 -m "Release v0.2.1: シフト割り当ての満員チェックを修正"

# 複数の変更がある場合
git tag -a v0.4.0 -m "Release v0.4.0

- feat: 日程調整機能を追加
- fix: メンバー一覧のロール表示を修正
- perf: シフト枠取得のクエリを最適化"
```

---

## 6. バージョニングルール

### 6.1 セマンティックバージョニング（簡易版）

本プロジェクトでは、v1.0.0 リリース前は以下のルールで運用します。

| 変更内容 | バージョン更新 | 例 |
|---------|---------------|-----|
| 新機能追加 | MINOR を上げる | v0.2.0 → v0.3.0 |
| バグ修正・軽微な改善 | PATCH を上げる | v0.2.0 → v0.2.1 |
| API破壊的変更 | MINOR を上げる（v1.0.0前） | v0.5.0 → v0.6.0 |
| 大規模リニューアル | MAJOR を上げる | v0.9.0 → v1.0.0 |

### 6.2 v1.0.0 以降のルール

正式リリース後は、以下の厳格なルールに移行します。

- **MAJOR**: 後方互換性のない変更
- **MINOR**: 後方互換性のある機能追加
- **PATCH**: 後方互換性のあるバグ修正

---

## 7. ロールバック手順

### 7.1 タグを使用したロールバック（推奨）

```bash
# 1. プロジェクトディレクトリに移動
cd /opt/vrc-shift-scheduler

# 2. 利用可能なタグを確認
git tag --list 'v*' --sort=-v:refname | head -10

# 3. 戻したいバージョンをチェックアウト
git fetch origin
git checkout v0.2.0

# 4. コンテナを再起動
docker compose -f docker-compose.prod.yml --env-file .env.prod down
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d

# 5. 動作確認
curl -s http://localhost:8080/health
```

### 7.2 コミットハッシュを使用したロールバック

タグがない場合、またはタグ間のコミットに戻す場合：

```bash
# 1. コミット履歴を確認
git log --oneline -20

# 2. 特定のコミットをチェックアウト
git checkout abc1234

# 3. コンテナを再起動
docker compose -f docker-compose.prod.yml --env-file .env.prod down
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d
```

### 7.3 マイグレーションのロールバック（必要な場合）

```bash
# 1つ前のバージョンに戻す
docker compose -f docker-compose.prod.yml exec backend ./migrate down 1

# 特定のバージョンまで戻す
docker compose -f docker-compose.prod.yml exec backend ./migrate down -to 20231215120000
```

### 7.4 ロールバック後の対応

```bash
# ロールバック完了後、mainブランチに戻る
git checkout main

# 問題を修正してから再デプロイ
```

---

## 8. よくあるミスと注意点

### 8.1 デプロイ時のミス

| ミス | 対策 |
|-----|------|
| `git pull` を忘れる | デプロイスクリプト化を検討 |
| `.env.prod` の更新漏れ | 新しい環境変数追加時はドキュメント化 |
| マイグレーション実行忘れ | デプロイ手順にチェックリストを含める |
| `--no-cache` を忘れて古いイメージが使われる | 常に `--no-cache` を付ける |

### 8.2 タグ付けのミス

| ミス | 対策 |
|-----|------|
| デプロイ前にタグを付けてしまう | タグは必ず動作確認**後**に付ける |
| タグ名の重複 | 既存タグを確認してから付ける |
| タグのpush忘れ | `git push origin <tag>` を忘れずに |

### 8.3 機密情報の漏洩防止

```bash
# .gitignore に含まれていることを確認
cat .gitignore | grep -E "\.env|\.prod"

# 誤ってコミットしていないか確認
git log --all --full-history -- "*.env*"
git log --all --full-history -- ".env.prod"
```

### 8.4 緊急時の対応

```bash
# コンテナが起動しない場合
docker compose -f docker-compose.prod.yml logs backend
docker compose -f docker-compose.prod.yml logs db

# ディスク容量不足の場合
docker system prune -a --volumes  # 注意: 未使用のすべてを削除

# データベース接続エラーの場合
docker compose -f docker-compose.prod.yml exec db psql -U vrcshift -d vrcshift -c '\l'
```

---

## 9. デプロイチェックリスト

### 9.1 デプロイ前

- [ ] `main` ブランチにマージ済み
- [ ] ローカルでビルド成功
- [ ] 破壊的変更がある場合、マイグレーション確認済み
- [ ] `.env.prod` の設定確認済み

### 9.2 デプロイ中

- [ ] `git pull origin main` 実行
- [ ] `docker compose build --no-cache` 実行
- [ ] `docker compose up -d` 実行
- [ ] コンテナ起動確認（`docker compose ps`）
- [ ] マイグレーション実行（必要な場合）

### 9.3 デプロイ後

- [ ] ヘルスチェック成功（`/health`）
- [ ] ログインテスト成功
- [ ] 主要機能の動作確認
- [ ] タグ付け完了（ローカルから）
- [ ] タグをリモートにpush

---

## 10. 参考: デプロイスクリプト例

繰り返しのデプロイを簡略化するためのスクリプト例です。

**scripts/deploy.sh:**

```bash
#!/bin/bash
set -e

COMPOSE_FILE="docker-compose.prod.yml"
ENV_FILE=".env.prod"

echo "=== VRC Shift Scheduler Deploy Script ==="

# 1. 最新コードを取得
echo "[1/5] Fetching latest code..."
git fetch origin
git checkout main
git pull origin main

# 2. コミット情報を表示
COMMIT=$(git rev-parse --short HEAD)
echo "[2/5] Deploying commit: $COMMIT"
echo "$COMMIT" > /tmp/last_deploy_commit.txt

# 3. コンテナを停止
echo "[3/5] Stopping containers..."
docker compose -f $COMPOSE_FILE down

# 4. 再ビルド・起動
echo "[4/5] Building and starting containers..."
docker compose -f $COMPOSE_FILE --env-file $ENV_FILE build --no-cache
docker compose -f $COMPOSE_FILE --env-file $ENV_FILE up -d

# 5. ヘルスチェック
echo "[5/5] Health check..."
sleep 5
curl -sf http://localhost:8080/health > /dev/null && echo "Health check: OK" || echo "Health check: FAILED"

echo "=== Deploy complete ==="
echo "Commit: $COMMIT"
echo "Don't forget to tag this release after verification!"
```

**使用方法:**

```bash
chmod +x scripts/deploy.sh
./scripts/deploy.sh
```

---

## 11. 関連ドキュメント

- [環境変数一覧](./ENVIRONMENT_VARIABLES.md)
- [ポートと環境設定](./verification/PORT_AND_ENV_CHECK.md)
- [サーバー要件](./DEPLOYMENT_SERVER_REQUIREMENTS.md)

---

**作成日**: 2025-12-19
**更新者**: 運用ドキュメント

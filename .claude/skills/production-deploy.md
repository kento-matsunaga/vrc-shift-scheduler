---
description: 本番デプロイ手順、ロールバック、トラブルシューティング
---

# Production Deployment

VRC Shift Scheduler の本番環境デプロイ・運用手順。

---

## 運用方針

| 項目 | 方針 |
|------|------|
| 本番ブランチ | `main` |
| デプロイ方式 | サーバーで `git pull` → `docker compose up` |
| タグ付け | デプロイ成功**後**にローカルからタグを付与 |
| 機密情報 | `.env.prod` はGitに入れずサーバーにのみ配置 |

---

## 本番サーバー情報

- **サーバー**: ConoHa VPS (163.44.103.76)
- **デプロイパス**: /opt/vrcshift
- **重要**: 必ず `docker-compose.prod.yml` を使用

---

## デプロイ前チェックリスト

- [ ] PRがマージされ、`main` ブランチが最新
- [ ] ローカルでビルド・テストが成功
- [ ] 破壊的変更がある場合、マイグレーション手順を確認済み
- [ ] `.env.prod` の設定が最新

---

## デプロイコマンド（サーバーで実行）

```bash
# 1. プロジェクトディレクトリに移動
cd /opt/vrcshift

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

# 6. ログの確認
docker compose -f docker-compose.prod.yml logs --tail=50 backend
```

---

## マイグレーション

```bash
# マイグレーションを実行
docker compose -f docker-compose.prod.yml exec backend ./migrate up

# マイグレーション状態の確認
docker compose -f docker-compose.prod.yml exec backend ./migrate version
```

---

## 動作確認

```bash
# ヘルスチェック
curl -s http://localhost:8080/health
# 期待: {"status":"ok"}

# ログイン確認
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}'
```

---

## デプロイ後のタグ付け（ローカルPCで実行）

```bash
# 1. 最新のmainを取得
git checkout main
git pull origin main

# 2. 現在のバージョンタグを確認
git tag --list 'v*' --sort=-v:refname | head -5

# 3. 新しいタグを作成
git tag -a v0.2.0 -m "Release v0.2.0: 機能追加・バグ修正"

# 4. タグをリモートにプッシュ
git push origin v0.2.0
```

### タグ命名規則

```
v<MAJOR>.<MINOR>.<PATCH>
```

| セグメント | 用途 | 例 |
|-----------|------|-----|
| MAJOR | 破壊的変更 | v1.0.0 |
| MINOR | 新機能追加 | v0.2.0 |
| PATCH | バグ修正 | v0.1.3 |

---

## ロールバック手順

### タグを使用（推奨）

```bash
cd /opt/vrcshift

# 利用可能なタグを確認
git tag --list 'v*' --sort=-v:refname | head -10

# 戻したいバージョンをチェックアウト
git fetch origin
git checkout v0.2.0

# コンテナを再起動
docker compose -f docker-compose.prod.yml --env-file .env.prod down
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d
```

### マイグレーションのロールバック

```bash
# 1つ前のバージョンに戻す
docker compose -f docker-compose.prod.yml exec backend ./migrate down 1
```

---

## トラブルシューティング

### コンテナが起動しない

```bash
docker compose -f docker-compose.prod.yml logs backend
docker compose -f docker-compose.prod.yml logs db
```

### データベース接続エラー

```bash
docker compose -f docker-compose.prod.yml exec db psql -U vrcshift -d vrcshift -c '\l'
```

### ディスク容量不足

```bash
docker system prune -a --volumes  # 注意: 未使用のすべてを削除
```

### サーバーログの確認

```bash
# バックエンドログ
tail -f /tmp/backend.log

# PostgreSQLコンテナの状態
docker ps | grep postgres
```

### データベース直接確認

```bash
# 管理者一覧
docker exec vrc-shift-scheduler-db-1 psql -U vrcshift -d vrcshift \
  -c "SELECT admin_id, email, display_name, role FROM admins WHERE deleted_at IS NULL;"

# マイグレーション状態
docker exec vrc-shift-scheduler-db-1 psql -U vrcshift -d vrcshift \
  -c "SELECT migration_id, applied_at FROM schema_migrations ORDER BY migration_id;"
```

---

## 環境変数（.env.prod）

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

**JWT_SECRET 生成:**
```bash
openssl rand -base64 64 | tr -d '\n'
```

---

## テストアカウント（開発環境）

| Email | Password | Role |
|-------|----------|------|
| admin1@example.com | password123 | owner |

---

## 関連ドキュメント

- `docs/PRODUCTION_DEPLOYMENT.md` - 完全なデプロイ手順
- `docs/ENVIRONMENT_VARIABLES.md` - 環境変数一覧
- `DEBUG_GUIDE.md` - デバッグガイド

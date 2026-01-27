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
- **コマンド**: `docker-compose`（ハイフン付き）を使用（`docker compose`ではない）

---

## デプロイ前チェックリスト

- [ ] PRがマージされ、`main` ブランチが最新
- [ ] ローカルでビルド・テストが成功
- [ ] 破壊的変更がある場合、マイグレーション手順を確認済み
- [ ] `.env.prod` の設定が最新
- [ ] 新しいバイナリ（batch等）がある場合、Dockerfileに含まれている
- [ ] Cronジョブの確認（新規バッチタスク追加時）

---

## デプロイコマンド（サーバーで実行）

> **注意**: サーバーでは `docker-compose`（ハイフン付き）を使用

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
docker-compose -f docker-compose.prod.yml --env-file .env.prod down
docker-compose -f docker-compose.prod.yml --env-file .env.prod build --no-cache
docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d

# 5. コンテナの起動確認
docker-compose -f docker-compose.prod.yml ps

# 6. ログの確認
docker-compose -f docker-compose.prod.yml logs --tail=50 backend
```

---

## マイグレーション

```bash
# マイグレーション状態の確認（デプロイ前に必ず確認）
docker exec vrc-shift-backend /app/migrate -action=status

# マイグレーションを実行
docker exec vrc-shift-backend /app/migrate -action=up
```

---

## 環境変数の変更

**重要**: `.env.prod`を変更した場合、`restart`ではなくコンテナの再作成が必要

```bash
# NG: restartでは環境変数が反映されない
docker-compose -f docker-compose.prod.yml restart backend

# OK: コンテナを削除して再作成
docker stop vrc-shift-backend
docker rm vrc-shift-backend
docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d backend
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
docker-compose -f docker-compose.prod.yml --env-file .env.prod down
docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d
```

### マイグレーションのロールバック

```bash
# 1つ前のバージョンに戻す
docker exec vrc-shift-backend /app/migrate -action=down -steps=1
```

---

## トラブルシューティング

### コンテナが起動しない

```bash
docker logs vrc-shift-backend
docker logs vrc-shift-db
```

### データベース接続エラー

```bash
docker exec vrc-shift-db psql -U vrcshift -d vrcshift -c '\l'
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

## Stripe本番デプロイ（月額課金機能）

### ⚠️ 重要: 開発環境と本番環境の違い

| 項目 | 開発環境 | 本番環境 |
|------|---------|---------|
| Webhook配信 | CLI経由でlocalhost転送 | Stripeから直接公開URLへ |
| CLI起動 | ✅ **必要**（`stripe listen`） | ❌ **不要** |
| Webhook Secret | CLIが生成（`stripe listen`出力） | **Dashboardから取得（別物！）** |
| APIキー | `sk_test_...` | `sk_live_...` |

### ⚠️ 最重要: Webhook Secretは環境ごとに異なる

**CLIの`stripe listen`で表示されるSecretと、Dashboardで取得するSecretは完全に別物です。**

```
# 開発用（CLI生成）- 本番では使用不可
STRIPE_WEBHOOK_SECRET=whsec_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# 本番用（Dashboard取得）- これを使う
STRIPE_WEBHOOK_SECRET=whsec_yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy
```

> ❌ 間違い: 開発で動いたからそのままwhsec_を本番にコピー
> ✅ 正解: Dashboardでエンドポイント作成後、新しいwhsec_を取得

### 本番Stripe設定手順

#### 1. Stripeダッシュボードでの設定

1. **本番APIキー取得**
   - https://dashboard.stripe.com/apikeys
   - 「本番用シークレットキー」（`sk_live_...`）をコピー

2. **商品・価格作成**
   - Products → 「+商品を追加」
   - 商品名: 「VRCシフト管理 月額プラン」
   - 価格: ¥200/月（継続課金）
   - Price ID（`price_live_...`）をメモ

3. **Webhookエンドポイント作成**
   - 開発者 → Webhook → 「エンドポイントを追加」
   - URL: `https://api.vrcshift.com/api/v1/stripe/webhook`
   - イベント選択:
     - `checkout.session.completed`
     - `customer.subscription.created`
     - `customer.subscription.updated`
     - `customer.subscription.deleted`
     - `invoice.payment_succeeded`
     - `invoice.payment_failed`
   - **作成後「署名シークレットを表示」から`whsec_...`を取得**

4. **Customer Portal設定**
   - 設定 → Customer Portal
   - サブスクリプションのキャンセル: ✅
   - 支払い方法の更新: ✅

#### 2. .env.prod への追記

```bash
# ============================================
# Stripe Configuration (Production)
# ============================================
# 本番用シークレットキー（sk_live_で始まる）
STRIPE_SECRET_KEY=sk_live_...

# Dashboardから取得したWebhook署名シークレット
# ⚠️ CLIの whsec_ とは別物！Dashboardから新規取得すること！
STRIPE_WEBHOOK_SECRET=whsec_...

# 本番用Price ID（price_live_で始まる）
STRIPE_PRICE_SUB_200=price_live_...

# リダイレクトURL
STRIPE_SUCCESS_URL=https://vrcshift.com/subscribe/complete
STRIPE_CANCEL_URL=https://vrcshift.com/subscribe/cancel
BILLING_PORTAL_RETURN_URL=https://vrcshift.com/admin/settings
```

#### 3. Cronジョブ設定（サーバーで実行）

```bash
crontab -e
```

```cron
# 猶予期間終了チェック - 毎日午前2時（JST）
0 2 * * * docker exec vrc-shift-backend /app/batch -task=grace-expiry >> /var/log/vrcshift/batch.log 2>&1

# Webhookログクリーンアップ - 毎週日曜午前3時
0 3 * * 0 docker exec vrc-shift-backend /app/batch -task=webhook-cleanup >> /var/log/vrcshift/batch.log 2>&1

# 支払い待ちテナント削除 - 毎日午前3時30分
30 3 * * * docker exec vrc-shift-backend /app/batch -task=pending-cleanup >> /var/log/vrcshift/batch.log 2>&1
```

```bash
mkdir -p /var/log/vrcshift
```

### Stripeトラブルシューティング

#### Webhook署名検証エラー

```bash
# よくある原因: 開発用SecretをそのままコピーしているD
# 対処: Dashboardからエンドポイントの署名シークレットを再取得

# ログ確認
docker logs vrc-shift-backend --tail=100 | grep -i stripe
```

#### Webhookが届かない

```bash
# Stripeダッシュボードでイベント配信状況を確認
# 開発者 → Webhook → エンドポイント → イベント

# テストイベント送信
# 「テストイベントを送信」ボタンで checkout.session.completed を送信
```

#### Stripeを無効化（緊急時）

```bash
# .env.prodのStripe設定をコメントアウト
nano .env.prod
# STRIPE_SECRET_KEY=...  ← コメントアウト

# コンテナを再作成（restartではなく再作成が必要）
docker stop vrc-shift-backend
docker rm vrc-shift-backend
docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d backend
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

# ============================================
# Stripe Configuration (Production)
# ============================================
# ⚠️ 必ずDashboardから本番用の値を取得すること
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...  # ← Dashboardのエンドポイントから取得！CLIの値は使えない！
STRIPE_PRICE_SUB_200=price_live_...
STRIPE_SUCCESS_URL=https://vrcshift.com/subscribe/complete
STRIPE_CANCEL_URL=https://vrcshift.com/subscribe/cancel
BILLING_PORTAL_RETURN_URL=https://vrcshift.com/admin/settings
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

---

## SSHアクセス（非対話的環境から）

Claude Code等の非対話的環境からSSH接続する場合、`setsid`が必要。

```bash
# SSH接続（出力を取得する場合は -w オプション）
DISPLAY=:0 SSH_ASKPASS=/tmp/askpass.sh SSH_ASKPASS_REQUIRE=force \
  setsid -w ssh -o StrictHostKeyChecking=no root@163.44.103.76 "command"

# SCP転送
DISPLAY=:0 SSH_ASKPASS=/tmp/askpass.sh SSH_ASKPASS_REQUIRE=force \
  setsid scp -o StrictHostKeyChecking=no /path/to/file root@163.44.103.76:/path/

# askpass.sh の内容
#!/bin/bash
echo 'パスワード'
```

**注意**: `setsid`がないとSSH_ASKPASSが動作しない

---

## Tarballデプロイ方式

gitが使えない場合やクリーンな状態でデプロイしたい場合。

```bash
# 1. ローカルでtarball作成
tar -czf /tmp/vrcshift-deploy.tar.gz \
  --exclude='node_modules' \
  --exclude='.git' \
  --exclude='*.tar.gz' \
  --exclude='.vite' \
  --exclude='dist' \
  .

# 2. サーバーにアップロード
scp /tmp/vrcshift-deploy.tar.gz root@163.44.103.76:/tmp/

# 3. サーバーで展開（.env.prod保持）
ssh root@163.44.103.76 "
  cp /opt/vrcshift/.env.prod /tmp/.env.prod.bak && \
  rm -rf /opt/vrcshift && \
  mkdir -p /opt/vrcshift && \
  cd /opt/vrcshift && \
  tar -xzf /tmp/vrcshift-deploy.tar.gz && \
  mv /tmp/.env.prod.bak .env.prod
"

# 4. Docker再ビルド・再起動
ssh root@163.44.103.76 "
  cd /opt/vrcshift && \
  docker-compose -f docker-compose.prod.yml --env-file .env.prod down && \
  docker-compose -f docker-compose.prod.yml --env-file .env.prod build --no-cache && \
  docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d
"
```

---

## 関連ドキュメント

- `LOCAL_DEPLOY_GUIDE.md` - 詳細なデプロイ手順（パスワード等含む）
- `docs/deployment/PRODUCTION_DEPLOYMENT.md` - 完全なデプロイ手順
- `docs/setup/ENVIRONMENT_VARIABLES.md` - 環境変数一覧
- `docs/guides/DEBUG_GUIDE.md` - デバッグガイド
- `docs/incidents/` - 過去のインシデントレポート

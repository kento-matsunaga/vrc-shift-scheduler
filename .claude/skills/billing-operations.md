---
description: 課金システム運用、BOOTH/Stripe連携、ライセンスキー管理
---

# Billing Operations

VRC Shift Scheduler の課金システム運用ガイド。

---

## 概要

2つの収益化チャネルに対応:

| チャネル | 種類 | 説明 |
|---------|------|------|
| BOOTH | 買い切り | ワンタイム購入ライセンスキー |
| Stripe | 継続課金 | サブスクリプション |

---

## アーキテクチャ

```
┌─────────────────────────────────────────────────────────────┐
│ テナントアプリ (web-frontend)                                │
│ http://localhost:5173                                       │
│ - 一般のテナントユーザーがアクセス                            │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ 管理コンソール (admin-frontend)                              │
│ http://localhost:5174                                       │
│ - 運営者のみアクセス可能                                      │
│ - ライセンスキー管理、テナント管理、監査ログ                    │
└─────────────────────────────────────────────────────────────┘
```

---

## 環境変数

```bash
# Stripe 設定
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PRICE_ID=price_...

# BOOTH 設定
LICENSE_KEY_PREFIX=VRCSS-  # デフォルト

# Cloudflare Access 設定（本番環境）
CF_ACCESS_TEAM_DOMAIN=yourteam.cloudflareaccess.com
CF_ACCESS_POLICY_AUD=your-policy-aud-from-cloudflare
```

---

## ライセンスキー操作

### キー生成（管理者）

```bash
curl -X POST http://localhost:8080/api/v1/admin/license-keys \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "count": 10,
    "memo": "BOOTH 2025年12月バッチ"
  }'
```

### ライフサイクル

| 状態 | 説明 |
|------|------|
| `unused` | 初期状態、顧客がクレーム可能 |
| `used` | クレーム済み、テナントに紐付け |
| `revoked` | 管理者が手動で無効化 |

### キークレーム（顧客）

```bash
curl -X POST http://localhost:8080/api/v1/public/license/claim \
  -H "Content-Type: application/json" \
  -d '{
    "email": "customer@example.com",
    "password": "securepassword",
    "display_name": "顧客名",
    "tenant_name": "組織名",
    "license_key": "VRCSS-ABCD1234-EFGH5678-IJKL9012-MNOP"
  }'
```

作成されるもの:
- `active` ステータスの新規テナント
- `owner` ロールの管理者ユーザー
- `LIFETIME` plan_code のエンタイトルメント

---

## Stripe Webhook

### ⚠️ 開発環境 vs 本番環境の違い（重要）

| 項目 | 開発環境 | 本番環境 |
|------|---------|---------|
| Webhook配信 | Stripe CLI経由 | Stripeから直接 |
| CLI起動 | ✅ `stripe listen --forward-to localhost:8080/api/v1/stripe/webhook` | ❌ **不要** |
| Webhook Secret | `stripe listen` の出力値 | **Dashboardから取得（別物！）** |
| APIキー | `sk_test_...` | `sk_live_...` |

### ⚠️ 絶対に間違えてはいけない点

**CLIの`stripe listen`で表示される`whsec_`と、Dashboardで取得する`whsec_`は完全に別物！**

```bash
# 開発用（CLIが生成）→ 本番では使えない
stripe listen --forward-to localhost:8080/api/v1/stripe/webhook
> Ready! Your webhook signing secret is whsec_xxxxxxxx (開発専用)

# 本番用（Dashboardから取得）→ これを使う
# Stripeダッシュボード → Webhook → エンドポイント → 「署名シークレットを表示」
# whsec_yyyyyyyy (本番専用)
```

### 設定（本番環境）

Stripe ダッシュボードでエンドポイントを設定:
- URL: `https://api.vrcshift.com/api/v1/stripe/webhook`
- イベント:
  - `checkout.session.completed`
  - `customer.subscription.created`
  - `customer.subscription.updated`
  - `customer.subscription.deleted`
  - `invoice.payment_succeeded`
  - `invoice.payment_failed`

**エンドポイント作成後、「署名シークレットを表示」から`whsec_...`を取得して`.env.prod`に設定すること。**

### サブスクリプションライフサイクル

| イベント | アクション |
|---------|-----------|
| `customer.subscription.created` | テナントとエンタイトルメント作成 |
| `invoice.payment_succeeded` | サブスクリプション更新 |
| `invoice.payment_failed` | テナントを `grace` に（14日間） |
| `customer.subscription.deleted` | エンタイトルメント無効化 |

---

## テナントステータス管理

### ステータス遷移

```
active → grace（支払い失敗）
grace → active（支払い成功）
grace → suspended（猶予期間終了）
suspended → active（手動再有効化）
```

### 猶予期間

- 期間: **14日間**
- 設定タイミング: 支払い失敗時
- チェック: `batch-grace-expiry` ジョブ

---

## バッチジョブ

### 猶予期間終了チェック（毎日実行）

```bash
# ドライラン
make batch-grace-expiry-dry

# 実行
make batch-grace-expiry

# 手動実行
DATABASE_URL="..." go run ./cmd/batch/main.go -task=grace-expiry
```

### Webhook ログクリーンアップ（毎週実行）

```bash
# ドライラン
make batch-webhook-cleanup-dry

# 実行
make batch-webhook-cleanup

# 手動実行
DATABASE_URL="..." go run ./cmd/batch/main.go -task=webhook-cleanup
```

### Cron 設定

```cron
# 猶予期間終了チェック - 毎日午前2時（JST）
0 2 * * * cd /path/to/app && DATABASE_URL="..." ./bin/batch -task=grace-expiry

# Webhook ログクリーンアップ - 毎週日曜日午前3時（JST）
0 3 * * 0 cd /path/to/app && DATABASE_URL="..." ./bin/batch -task=webhook-cleanup
```

---

## 管理コンソール

### 起動

```bash
# Docker Compose
docker compose --profile admin up

# または個別に
cd admin-frontend && npm run dev
```

URL: http://localhost:5174

### 機能

| ページ | 機能 |
|--------|------|
| ライセンスキー | キーの発行、一覧、失効 |
| テナント | テナント検索、ステータス変更、詳細表示 |
| 監査ログ | 操作履歴の確認、フィルタリング |

---

## 監査証跡

すべての課金操作は `billing_audit_logs` に記録:

| アクション | 説明 |
|--------|------|
| `license_generated` | ライセンスキー作成 |
| `license_claimed` | ライセンスキーを使用した登録 |
| `license_revoked` | ライセンスキーの手動無効化 |
| `entitlement_created` | サブスクリプション/ライセンスの有効化 |
| `entitlement_revoked` | サブスクリプションのキャンセル |
| `tenant_status_changed` | ステータス遷移 |
| `tenant_suspended` | バッチジョブによる自動停止 |

---

## トラブルシューティング

### ライセンスキーの問題

```sql
-- ライセンスキーのステータスを確認
SELECT * FROM license_keys WHERE key_id = '...';
```

### Stripe Webhook の問題

```sql
-- 最近の Webhook イベントを確認
SELECT * FROM stripe_webhook_logs ORDER BY received_at DESC LIMIT 20;

-- 重複処理を確認
SELECT event_id, COUNT(*) FROM stripe_webhook_logs
GROUP BY event_id HAVING COUNT(*) > 1;
```

### テナントステータスの問題

```sql
-- テナントのエンタイトルメントを確認
SELECT t.tenant_name, t.status, t.grace_until, e.*
FROM tenants t
LEFT JOIN entitlements e ON t.tenant_id = e.tenant_id
WHERE t.tenant_id = '...';

-- テナントの監査ログを確認
SELECT * FROM billing_audit_logs
WHERE target_id = '...'
ORDER BY created_at DESC;
```

---

## 開発環境でのStripeテスト

### 前提条件

1. Stripe CLIがインストール済み
2. Stripeアカウントにログイン済み（`stripe login`）
3. テスト用APIキーが`.env`に設定済み

### 起動手順

```bash
# ターミナル1: Docker (DB)
docker compose up -d

# ターミナル2: Backend
cd backend && ./server

# ターミナル3: Frontend
cd web-frontend && npm run dev

# ターミナル4: Stripe CLI（Webhook転送）← 開発環境でのみ必要！
stripe listen --forward-to localhost:8080/api/v1/stripe/webhook
# 出力された whsec_ を backend/.env の STRIPE_WEBHOOK_SECRET に設定
```

### テストカード

| カード番号 | 用途 |
|-----------|------|
| `4242 4242 4242 4242` | 成功 |
| `4000 0000 0000 0002` | 拒否 |
| `4000 0000 0000 9995` | 残高不足 |

### 手動でWebhookイベント送信

```bash
# テストイベントをトリガー
stripe trigger checkout.session.completed
stripe trigger invoice.payment_succeeded
stripe trigger customer.subscription.deleted
```

---

## セキュリティ

1. **管理API分離**: `/api/v1/admin/*` はテナント認証から完全に分離
2. **Cloudflare Access**: 本番環境では管理APIとコンソールを保護
3. **レート制限**: ライセンスクレームは IP あたり 5リクエスト/分
4. **Webhook 検証**: `STRIPE_WEBHOOK_SECRET` で署名を検証
5. **ライセンスキー**: SHA-256 ハッシュとして保存

---

## 必要なテーブル

- `license_keys` - BOOTH ライセンスキー保存
- `entitlements` - テナントごとの有効なサブスクリプション/ライセンス
- `billing_audit_logs` - 課金操作の監査ログ
- `stripe_webhook_logs` - Webhook の冪等性追跡

---

## 関連ドキュメント

- `docs/BILLING_OPERATIONS.md` - 完全な課金運用ガイド
- `docs/ENVIRONMENT_VARIABLES.md` - 環境変数一覧

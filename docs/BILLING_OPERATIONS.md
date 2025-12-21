# 課金運用ガイド

本ドキュメントでは、VRC Shift Scheduler の課金システムの運用手順について説明します。

## 概要

課金システムは2つの収益化チャネルに対応しています：
1. **BOOTH（買い切り）** - ワンタイム購入ライセンスキー
2. **Stripe** - 継続課金サブスクリプション

## アーキテクチャ

課金管理機能はテナントアプリとは完全に分離された専用の管理コンソールで提供されます。

```
┌─────────────────────────────────────────────────────────────┐
│ テナントアプリ (web-frontend)                                │
│ http://localhost:5173                                       │
│ - 一般のテナントユーザーがアクセス                            │
│ - 課金管理機能は含まれていない                                │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ 管理コンソール (admin-frontend)                              │
│ http://localhost:5174                                       │
│ - 運営者のみアクセス可能                                      │
│ - Cloudflare Access で保護（本番環境）                       │
│ - ライセンスキー管理、テナント管理、監査ログ                    │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ バックエンド API                                             │
│ http://localhost:8080                                       │
│ /api/v1/*        → テナント JWT 認証                         │
│ /api/v1/admin/*  → Cloudflare Access 認証（運営専用）        │
└─────────────────────────────────────────────────────────────┘
```

## 環境変数

`.env` ファイルに以下を追加してください：

```bash
# Stripe 設定
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PRICE_ID=price_...

# BOOTH 設定（ライセンスキー形式）
LICENSE_KEY_PREFIX=VRCSS-  # 省略可、デフォルトは VRCSS-

# Cloudflare Access 設定（本番環境で必須）
CF_ACCESS_TEAM_DOMAIN=yourteam.cloudflareaccess.com
CF_ACCESS_POLICY_AUD=your-policy-aud-from-cloudflare
CF_ACCESS_ALLOWED_EMAILS=admin1@example.com,admin2@example.com  # 省略可
```

## データベースマイグレーション

課金テーブルが作成されていることを確認してください：

```bash
# マイグレーション実行
make migrate

# または手動で
DATABASE_URL="postgres://user:pass@localhost:5432/vrcshift" go run ./cmd/migrate/main.go
```

必要なテーブル：
- `license_keys` - BOOTH ライセンスキー保存
- `entitlements` - テナントごとの有効なサブスクリプション/ライセンス
- `billing_audit_logs` - 課金操作の監査ログ
- `stripe_webhook_logs` - Webhook の冪等性追跡

## ライセンスキー操作

### ライセンスキーの生成（管理者）

管理者 API を使用して BOOTH 販売用のライセンスキーを生成します：

```bash
# 10個のキーを生成
curl -X POST http://localhost:8080/api/v1/admin/license-keys \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "count": 10,
    "memo": "BOOTH 2025年12月バッチ"
  }'
```

レスポンスには生成されたキーが含まれます - 安全に保管してください。

### ライセンスキーのライフサイクル

1. **unused** - 初期状態、顧客がクレーム可能
2. **used** - 顧客がクレーム済み、テナントに紐付け
3. **revoked** - 管理者が手動で無効化

### ライセンスキーのクレーム（顧客）

顧客は公開エンドポイントを使用して登録します：

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

これにより以下が作成されます：
- `active` ステータスの新規テナント
- `owner` ロールの管理者ユーザー
- `LIFETIME` plan_code のエンタイトルメント

## Stripe 操作

### Webhook 設定

1. Stripe ダッシュボードで Webhook エンドポイントを設定：
   - URL: `https://your-domain.com/api/v1/stripe/webhook`
   - イベント：
     - `customer.subscription.created`
     - `customer.subscription.updated`
     - `customer.subscription.deleted`
     - `invoice.payment_succeeded`
     - `invoice.payment_failed`

2. Webhook 署名シークレットを `STRIPE_WEBHOOK_SECRET` にコピー

### サブスクリプションのライフサイクル

1. **customer.subscription.created** - テナントとエンタイトルメントを作成
2. **invoice.payment_succeeded** - サブスクリプションを更新
3. **invoice.payment_failed** - テナントを `grace` ステータスに設定（14日間）
4. **customer.subscription.deleted** - エンタイトルメントを無効化

## テナントステータス管理

### ステータス遷移

```
active → grace（支払い失敗）
grace → active（支払い成功）
grace → suspended（猶予期間終了）
suspended → active（支払い後の手動再有効化）
```

### 猶予期間

- 期間：14日間
- 設定タイミング：支払い失敗時
- チェック：`batch-grace-expiry` ジョブ

## バッチジョブ

### 猶予期間終了チェック

猶予期間が終了したテナントを停止するため毎日実行：

```bash
# ドライラン（プレビューのみ）
make batch-grace-expiry-dry

# 実行
make batch-grace-expiry

# または手動で
DATABASE_URL="..." go run ./cmd/batch/main.go -task=grace-expiry
```

### Webhook ログクリーンアップ

30日以上前の Webhook ログを削除：

```bash
# ドライラン
make batch-webhook-cleanup-dry

# 実行
make batch-webhook-cleanup

# または手動で
DATABASE_URL="..." go run ./cmd/batch/main.go -task=webhook-cleanup
```

### Cron 設定

crontab に追加：

```cron
# 猶予期間終了チェック - 毎日午前2時（JST）
0 2 * * * cd /path/to/app && DATABASE_URL="..." ./bin/batch -task=grace-expiry >> /var/log/vrcshift/batch.log 2>&1

# Webhook ログクリーンアップ - 毎週日曜日午前3時（JST）
0 3 * * 0 cd /path/to/app && DATABASE_URL="..." ./bin/batch -task=webhook-cleanup >> /var/log/vrcshift/batch.log 2>&1
```

## 管理コンソール

### ローカル開発での起動

```bash
# 管理コンソールを含めて起動
docker compose --profile admin up

# または個別に起動
cd admin-frontend && npm run dev
```

管理コンソールは http://localhost:5174 でアクセスできます。

### 本番環境でのアクセス

1. Cloudflare Access でアプリケーションを作成
   - Application Type: Self-hosted
   - Application URL: `https://admin.your-domain.com`
   - Policy: Google IdP で許可するメールアドレスを指定

2. 環境変数を設定
   - `CF_ACCESS_TEAM_DOMAIN`: Cloudflare Access のチームドメイン
   - `CF_ACCESS_POLICY_AUD`: アプリケーションの AUD タグ

### 機能一覧

| ページ | 機能 |
|--------|------|
| ライセンスキー | キーの発行、一覧、失効 |
| テナント | テナント検索、ステータス変更、詳細表示 |
| 監査ログ | 操作履歴の確認、フィルタリング |

## 監査証跡

すべての課金操作は `billing_audit_logs` に記録されます：

| アクション | 説明 |
|--------|-------------|
| `license_generated` | ライセンスキー作成 |
| `license_claimed` | ライセンスキーを使用した登録 |
| `license_revoked` | ライセンスキーの手動無効化 |
| `entitlement_created` | サブスクリプション/ライセンスの有効化 |
| `entitlement_revoked` | サブスクリプションのキャンセル |
| `tenant_status_changed` | ステータス遷移 |
| `tenant_suspended` | バッチジョブによる自動停止 |

## セキュリティ考慮事項

1. **管理API分離**: `/api/v1/admin/*` はテナント認証から完全に分離
2. **Cloudflare Access**: 本番環境では管理APIとコンソールをCloudflare Accessで保護
3. **レート制限**: ライセンスクレームエンドポイントは IP あたり 5リクエスト/分 に制限
4. **Stripe Webhook 検証**: `STRIPE_WEBHOOK_SECRET` を使用して署名を検証
5. **ライセンスキー**: SHA-256 ハッシュとして保存、元のキーは生成時のみ表示
6. **多層防御**: Cloudflare Access + アプリ側でのメール検証（`CF_ACCESS_ALLOWED_EMAILS`）

## トラブルシューティング

### ライセンスキーの問題

```sql
-- ライセンスキーのステータスを確認
SELECT * FROM license_keys WHERE key_id = '...';

-- ハッシュでライセンスを検索（キーがある場合）
-- キーの SHA-256 を使用して検索
```

### Stripe Webhook の問題

```sql
-- 最近の Webhook イベントを確認
SELECT * FROM stripe_webhook_logs ORDER BY received_at DESC LIMIT 20;

-- 重複処理を確認
SELECT event_id, COUNT(*) FROM stripe_webhook_logs GROUP BY event_id HAVING COUNT(*) > 1;
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

## バックアップ推奨

毎日のバックアップ対象：
- `license_keys` テーブル
- `entitlements` テーブル
- `billing_audit_logs` テーブル

毎週のバックアップ対象：
- `stripe_webhook_logs` テーブル（クリーンアップ前）

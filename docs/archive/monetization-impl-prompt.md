# Claude Code 用：マネタイズ基盤 実装プロンプト

あなたは Claude Code として、このリポジトリの既存実装に合わせて **マネタイズ基盤（BOOTH買い切りキー／Stripeサブスク／tenants.status／Cloudflare Access前提の管理画面／SES通知／監査ログ／バックアップ運用）** を追加実装してください。

以下の仕様は確定です。既存のディレクトリ構成・フレームワーク・DB層・ルーティング・DTO・レスポンス形式・トランザクション管理に合わせて実装し、勝手に大規模リファクタをしないでください。

最初に必ずコードベースを探索し、同じ流儀で差分を最小にします。
探索結果は `docs/exploration-notes.md` に出力し、以降の実装判断の根拠として参照すること。

---

## 0. 絶対ルール

- 今ある設計（レイヤ/DTO/レスポンス形式/tx管理/命名規則）に合わせる
- 既存の「作法」を変えない（例：レスポンス包み、エラーハンドリング、ID形式、Repositoryパターンなど）
- 大規模な置き換えは禁止。必要な追加のみ
- 実装は Commit単位で進める（小さく、ビルド/テストが通る状態を維持）
- 追加した変更点にはログ/監査を入れる（audit_logs）

---

## 1. 事前探索（最初に必ず実行して結果を要約）

### 1-1. プロジェクト構造の確認

- バックエンド言語/フレームワーク（Go/chi）
- ルーティング実装箇所（`internal/interface/rest/router.go`）
- DBアクセス層（Repository パターン、`internal/infra/db/`）
- 認証・認可の仕組み（JWT、`internal/infra/security/`）
- 既存の `tenants` テーブル構造
- 既存の「管理画面」があるか（`/admin` 相当）
- メール送信機構があるか（SES未実装なら追加）
- 既存の決済/外部連携の実装があるか

### 1-2. 既存のID/タイムスタンプ/tx方式

- ID生成方式: **ULID** (`common.NewULID()`)
- tx管理: `pgxpool.Pool` + `Begin/Commit/Rollback`
- マイグレーション: `internal/infra/db/migrations/` 配下、ファイル名は `YYYYMMDDHHMMSS_description.up.sql`

### 1-3. 既存レスポンス形式

- RESTレスポンス: `{"data": {...}}` 形式
- エラー形式: `{"error": {"code": "ERR_XXX", "message": "..."}}`
- Public API: `/api/v1/public/...`
- 認証必要API: `/api/v1/...`

### 1-4. 既存構造の補足（探索で確認すること）

- `tenants` テーブルに `status` カラムは未存在（要ALTER）
- 認証: JWT + `admins` テーブル（tenant所属）
- Router: chi を使用、`/api/v1` 配下

**探索結果を短くまとめ、以降の実装方針を「既存流儀に合わせて」宣言してから作業に入ること。**

---

## 2. 確定仕様（変更禁止）

### 2-1. インフラ前提

| 項目 | 内容 |
|------|------|
| 本番サーバ | ConoHa VPS |
| Cloudflare | DNS + 管理画面だけ Access で保護 |
| メール | Amazon SES |
| DB | PostgreSQL（当面同居、将来分離できるように） |
| 開発環境 | local のみ |

### 2-2. 販売

| 方式 | 詳細 |
|------|------|
| サブスク | Stripe（月額980円・1プラン） |
| 買い切り | BOOTH（利用開始キー方式、在庫方式から開始） |

### 2-3. 状態（tenants.status）と権利（entitlements）

**責務分離を厳守：**

- `tenants.status`: アクセス可否の大枠
  - `active` / `grace` / `suspended`
- `entitlements`: プラン・権利
  - `plan_code` は `LIFETIME` と `SUB_980` の2つ
  - `features_json` は初期 `{}` でOK（判定は `plan_code` 分岐）

**revoked：**
- `entitlements.revoked_at != NULL` は最優先で即停止（status に関係なく）

**優先順位（同一テナントに複数 entitlement がある場合）：**
1. `revoked_at != NULL` → 即停止
2. `ends_at IS NULL`（LIFETIME）→ 優先
3. `ends_at` が最も遠い SUB_980

### 2-4. grace/suspended 仕様（確定）

| 項目 | 内容 |
|------|------|
| grace 期間 | 14日 |
| grace 中 | 閲覧OK / 新規作成・編集禁止 / 支払い導線を強制表示 |
| suspended | 閲覧のみ |
| 復帰 | `invoice.paid` で自動で `active` に戻す |
| データ削除 | 当面なし |

### 2-5. webhook_events 保持

- Stripe webhook の冪等性のため `event_id` を保存
- 30日経過した webhook_events は定期削除（cron or バッチ）

### 2-6. タイムゾーン

- DB保存: すべて **UTC**
- 表示時: JST 変換（フロントエンド側で処理）

---

## 3. DB変更（マイグレーション追加）

既存のマイグレーション方式に合わせて追加する。既存テーブルがあるなら ALTER で差分のみ。

### 3-1. tenants（ALTER）

```sql
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'active' NOT NULL;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS grace_until TIMESTAMPTZ;
```

### 3-2. plans（新規、seed で初期2件）

```sql
CREATE TABLE plans (
    plan_code VARCHAR(50) PRIMARY KEY,
    plan_type VARCHAR(20) NOT NULL, -- 'lifetime' or 'subscription'
    display_name VARCHAR(100) NOT NULL,
    price_jpy INTEGER,
    features_json JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Seed data
INSERT INTO plans (plan_code, plan_type, display_name, price_jpy) VALUES
    ('LIFETIME', 'lifetime', '買い切りプラン', NULL),
    ('SUB_980', 'subscription', '月額プラン', 980);
```

### 3-3. entitlements（新規）

```sql
CREATE TABLE entitlements (
    entitlement_id VARCHAR(26) PRIMARY KEY,
    tenant_id VARCHAR(26) NOT NULL REFERENCES tenants(tenant_id),
    plan_code VARCHAR(50) NOT NULL REFERENCES plans(plan_code),
    source VARCHAR(50) NOT NULL, -- 'booth' or 'stripe'
    starts_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ends_at TIMESTAMPTZ, -- NULL for lifetime
    revoked_at TIMESTAMPTZ,
    revoked_reason TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_entitlements_tenant_id ON entitlements(tenant_id);
CREATE INDEX idx_entitlements_plan_code ON entitlements(plan_code);
```

### 3-4. subscriptions（Stripe用、新規）

```sql
CREATE TABLE subscriptions (
    subscription_id VARCHAR(26) PRIMARY KEY,
    tenant_id VARCHAR(26) NOT NULL UNIQUE REFERENCES tenants(tenant_id),
    stripe_customer_id VARCHAR(100) NOT NULL,
    stripe_subscription_id VARCHAR(100) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL, -- 'active', 'past_due', 'canceled', etc.
    current_period_end TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_stripe_subscription_id ON subscriptions(stripe_subscription_id);
```

### 3-5. webhook_events（新規）

```sql
CREATE TABLE webhook_events (
    id SERIAL PRIMARY KEY,
    provider VARCHAR(50) NOT NULL, -- 'stripe'
    event_id VARCHAR(100) NOT NULL,
    received_at TIMESTAMPTZ DEFAULT NOW(),
    payload_json JSONB,
    UNIQUE(provider, event_id)
);

CREATE INDEX idx_webhook_events_received_at ON webhook_events(received_at);
```

### 3-6. license_keys（BOOTH用、新規）

```sql
CREATE TABLE license_keys (
    key_id VARCHAR(26) PRIMARY KEY,
    key_hash VARCHAR(64) NOT NULL UNIQUE, -- SHA-256 hash
    status VARCHAR(20) NOT NULL DEFAULT 'unused', -- 'unused', 'used', 'revoked'
    issued_batch_id VARCHAR(26),
    used_at TIMESTAMPTZ,
    used_tenant_id VARCHAR(26) REFERENCES tenants(tenant_id),
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_license_keys_status ON license_keys(status);
CREATE INDEX idx_license_keys_batch ON license_keys(issued_batch_id);
```

### 3-7. audit_logs（新規）

```sql
CREATE TABLE audit_logs (
    log_id VARCHAR(26) PRIMARY KEY,
    actor_type VARCHAR(20) NOT NULL, -- 'admin', 'system', 'stripe'
    actor_id VARCHAR(26),
    action VARCHAR(100) NOT NULL,
    target_type VARCHAR(50),
    target_id VARCHAR(26),
    before_json JSONB,
    after_json JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_actor ON audit_logs(actor_type, actor_id);
CREATE INDEX idx_audit_logs_target ON audit_logs(target_type, target_id);
```

---

## 4. バックエンド実装（既存流儀で）

### 4-1. 共通認可ガード（最重要）

書き込み系API（POST/PUT/PATCH/DELETE）全部に適用。

**判定順：**
1. `entitlement.revoked_at != NULL` → 403 禁止
2. `tenant.status IN ('grace', 'suspended')` → 403 禁止（読み取りは許可）
3. `active` → OK

**読み取りAPIは `grace`/`suspended` でもOK（`revoked` は不可）**

既存の middleware や usecase 入口に入れる。
「どこに入れるのが既存流儀か」は探索結果に従うこと。

### 4-2. Stripe Webhook

#### 4-2-1. 実装内容

- endpoint: `POST /api/v1/stripe/webhook`（router.go に追加）
- `event_id` で冪等: `webhook_events` へ INSERT できたら処理、重複なら即 200 で終了
- 署名検証（Stripe の Webhook signing secret）を入れる

**MVP で処理するイベント：**

| イベント | 処理 |
|----------|------|
| `invoice.paid` | tenant `active` + `grace_until` NULL + subscription 更新 |
| `invoice.payment_failed` | tenant `grace` + `grace_until = now + 14days` |
| `customer.subscription.deleted` | `period_end` 経過後に `suspended`（バッチで判定でもOK） |

#### 4-2-2. 初期リリース時の扱い

Stripe 連携は将来利用のため実装するが、初期リリース時点では未使用。

- 環境変数が未設定の場合、Stripe 関連の処理はスキップする（エラーにしない）
- 起動時ログに `[INFO] Stripe integration disabled (STRIPE_WEBHOOK_SECRET not set)` を出力
- Webhook エンドポイントは作成するが、本番では当面アクセスされない想定
- サブスク関連のUI（支払い導線など）は、`entitlements.plan_code == 'SUB_980'` のテナントが存在する場合のみ表示

### 4-3. BOOTH キー Claim（在庫方式）

#### 4-3-1. エンドポイント

- `POST /api/v1/public/license/claim`（既存の public API 流儀に合わせる）

#### 4-3-2. Input

```json
{
  "email": "user@example.com",
  "password": "securepassword",
  "display_name": "管理者名",
  "tenant_name": "テナント名",
  "license_key": "XXXX-XXXX-XXXX-XXXX"
}
```

#### 4-3-3. 処理フロー（必ずトランザクション）

1. `license_keys` から `key_hash` でロック取得（`FOR UPDATE`）
2. ステータスが `unused` であることを確認
3. ステータスを `used` に更新
4. `tenants` 作成（`status = 'active'`）
5. `admins` 作成（tenant admin、パスワードハッシュ化）
6. `entitlements` 作成（`plan_code = 'LIFETIME'`, `source = 'booth'`）
7. `audit_logs` 追記
8. エラー時はロールバック

#### 4-3-4. キー形式

- 表示用: `XXXX-XXXX-XXXX-XXXX`（16文字 + ハイフン）
- DB保存: SHA-256 ハッシュのみ（平文保存禁止）

#### 4-3-5. セキュリティ

- **レート制限**: IP 単位で 5回/分
- **タイミング攻撃対策**: 失敗時は 1秒遅延レスポンス
- **ログ**: 成功/失敗とも `audit_logs` に記録

### 4-4. 管理画面用 API（運営専用）

#### 4-4-1. パス

- `/api/v1/admin/...` 配下（既存 router.go に追加）
- Cloudflare Access で保護するが、アプリ側でも owner 権限チェック

#### 4-4-2. 機能（MVP）

| エンドポイント | 機能 |
|----------------|------|
| `POST /api/v1/admin/license-keys/generate` | キーバッチ生成（N個生成して平文キー一覧を返す、DBにはhashのみ） |
| `POST /api/v1/admin/license-keys/{key_id}/revoke` | キー失効 |
| `POST /api/v1/admin/license-keys/batch/{batch_id}/revoke` | バッチ単位失効 |
| `GET /api/v1/admin/tenants` | テナント検索 |
| `PATCH /api/v1/admin/tenants/{tenant_id}/status` | 状態変更（active/suspended） |
| `GET /api/v1/admin/audit-logs` | 監査ログ閲覧（ページネーション） |

---

## 5. フロントエンド（既存UIに最小追加）

### 5-1. grace/suspended 表示

ログイン後にテナント状態を取得し、状態に応じた表示を行う。

| 状態 | 表示 |
|------|------|
| `grace` | 支払い導線（リンク）を常に表示、作成/編集ボタンは `disabled` |
| `suspended` | 閲覧のみの説明、支払い復帰導線 |
| `revoked` | ログイン不可 or 操作不可表示 |

### 5-2. BOOTH 利用開始ページ

- パス: `/claim`（未ログイン状態でアクセス可能）
- `license_key` + `email` + `password` + `display_name` + `tenant_name` で Claim
- 成功後はログインページへ誘導

### 5-3. 管理画面

- パス: `/admin/*`（新規ルート）
- 運営用の最小UI
  - キー発行（件数指定 → 平文キー一覧表示 → CSV ダウンロード）
  - キー失効
  - テナント検索・状態変更
  - 監査ログ閲覧

### 5-4. 既存UIへの影響

- `Settings.tsx` に「プラン情報」セクションを追加
  - 現在のプラン表示
  - サブスクの場合は次回請求日表示

---

## 6. Cloudflare Access 設定（手順書として出力）

### 6-1. 保護対象

- `/admin` と `/admin-api` を保護する

### 6-2. 設定内容

- IdP: Google
- Allowlist: 運営最大3人のメールアドレス
- 顧客は追加しない

### 6-3. 手順

`docs/cloudflare-access-setup.md` に出力すること。

---

## 7. SES（手順と最小実装）

### 7-1. 送信するメール

| タイミング | 内容 |
|------------|------|
| Claim 成功 | 初回ログイン案内 |
| 支払い失敗 | grace 開始通知 |
| 復帰 | active 復帰通知 |

### 7-2. 実装

- 既存メール基盤がなければ SES 送信モジュールを追加（既存流儀に合わせる）
- `internal/infra/email/ses.go` に実装
- 環境変数: `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `SES_FROM_ADDRESS`

### 7-3. 手順書

`docs/ses-setup.md` に出力すること。

---

## 8. バックアップ運用（手順書とスクリプト）

### 8-1. 要件

- 日次 `pg_dump` を cron で取得
- 保管先: Cloudflare R2
- 30日保持（30日より古いファイルは自動削除）

### 8-2. 環境変数

```
R2_ACCOUNT_ID
R2_ACCESS_KEY_ID
R2_SECRET_ACCESS_KEY
R2_BUCKET
R2_ENDPOINT  # https://<account_id>.r2.cloudflarestorage.com
```

### 8-3. 出力物

- `scripts/backup.sh` - バックアップスクリプト
- `scripts/restore.sh` - リストアスクリプト
- `docs/backup-setup.md` - 手順書（cron 設定含む）

---

## 9. 定期バッチ

### 9-1. webhook_events 削除

- 30日経過したレコードを削除
- 日次実行（cron）

### 9-2. suspended 判定

- `grace_until` を過ぎた `grace` テナントを `suspended` に変更
- 時間単位実行（cron）

### 9-3. 出力物

- `scripts/cleanup-webhook-events.sh`
- `scripts/check-grace-expiry.sh`
- `docs/cron-setup.md`

---

## 10. 受け入れテスト（必須）

### 10-1. LIFETIME

- [ ] Claim でテナント発行され全機能OK
- [ ] 無効なキーでエラー
- [ ] 使用済みキーでエラー
- [ ] レート制限が動作する

### 10-2. Stripe

- [ ] `invoice.payment_failed` → `grace`（書き込み不可、閲覧OK）
- [ ] `invoice.paid` → `active` 復帰（書き込みOK）
- [ ] `grace` 期限超過 → `suspended`（閲覧のみ）

### 10-3. revoked

- [ ] `revoked_at` が入ると即停止

### 10-4. webhook 冪等

- [ ] 同 `event_id` が複数回来ても二重処理しない

### 10-5. その他

- [ ] `webhook_events` の 30日削除バッチが動く
- [ ] `audit_logs` が主要操作で必ず残る
- [ ] 管理APIが owner 権限チェックを通過する

---

## 11. 作業の進め方（必ずこの順で）

### Phase 1: 基盤

1. 探索 → 要約 → 実装方針宣言（既存流儀の確認）
2. DB マイグレーション追加 → 適用 → ビルド/テスト
3. エンティティ/DTO/Repository 追加（最小）
4. 認可ガード（書き込み禁止）追加

### Phase 2: 決済連携

5. BOOTH Claim 実装（Tx 必須）
6. Stripe Webhook（署名/冪等）実装

### Phase 3: 管理機能

7. 管理 API（キー発行/失効/監査）実装
8. フロント UI（Claim ページ、状態表示、管理UI）

### Phase 4: 運用

9. `webhook_events` 削除バッチ、`suspended` 判定バッチ
10. 手順書（Cloudflare Access/SES/バックアップ）出力

---

## 12. 出力形式（最後に必ず出す）

### A. 変更したファイル一覧（パス付き）

### B. マイグレーション一覧と適用方法

### C. 動作確認手順（local）

### D. 本番反映手順（ConoHa）

### E. 想定落とし穴

特に以下に注意：
- Tx のロールバック漏れ
- 冪等性の破れ
- 状態判定の順序ミス
- キー平文の漏洩
- レート制限のバイパス

---

## 付録: 環境変数一覧

```bash
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/dbname

# JWT
JWT_SECRET=your-jwt-secret

# Stripe (初期リリースでは未設定でOK)
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...

# AWS SES
AWS_REGION=ap-northeast-1
AWS_ACCESS_KEY_ID=...
AWS_SECRET_ACCESS_KEY=...
SES_FROM_ADDRESS=noreply@example.com

# Cloudflare R2
R2_ACCOUNT_ID=...
R2_ACCESS_KEY_ID=...
R2_SECRET_ACCESS_KEY=...
R2_BUCKET=backup-bucket
R2_ENDPOINT=https://xxx.r2.cloudflarestorage.com
```

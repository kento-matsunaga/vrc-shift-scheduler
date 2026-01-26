---
description: Stripe API連携パターン、Webhook処理、サブスクリプション管理
---

# Stripe Integration

Stripe決済連携の実装パターン。

---

## アーキテクチャ

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Stripe     │────▶│   Webhook    │────▶│   Usecase    │
│   Dashboard  │     │   Handler    │     │              │
└──────────────┘     └──────────────┘     └──────────────┘
                            │                    │
                            ▼                    ▼
                     ┌──────────────┐     ┌──────────────┐
                     │   Webhook    │     │   Tenant     │
                     │   Event Log  │     │   Repo       │
                     └──────────────┘     └──────────────┘
```

---

## Webhook処理パターン

### 1. 署名検証（必須）

```go
func (h *WebhookHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
    payload, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    // 署名検証
    event, err := webhook.ConstructEvent(
        payload,
        r.Header.Get("Stripe-Signature"),
        h.webhookSecret,
    )
    if err != nil {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }

    // 処理を委譲
    processed, err := h.usecase.HandleWebhook(ctx, event, string(payload))
    // ...
}
```

### 2. 冪等性保証

```go
// Webhook イベントの重複処理を防止
isNew, err := uc.webhookEventRepo.TryInsert(ctx, "stripe", event.ID, &rawPayload)
if err != nil {
    return false, fmt.Errorf("failed to check webhook idempotency: %w", err)
}

if !isNew {
    // 既に処理済み - スキップ
    log.Printf("[Stripe Webhook] Duplicate event ignored: %s", event.ID)
    return false, nil
}
```

### 3. イベントタイプ別処理

```go
switch event.Type {
case "invoice.paid":
    return true, uc.handleInvoicePaid(ctx, now, event)
case "invoice.payment_failed":
    return true, uc.handleInvoicePaymentFailed(ctx, now, event)
case "customer.subscription.deleted":
    return true, uc.handleSubscriptionDeleted(ctx, now, event)
default:
    log.Printf("[Stripe Webhook] Unknown event type: %s", event.Type)
    return true, nil
}
```

---

## サブスクリプションライフサイクル

### ステータス遷移

```
incomplete → active → past_due → canceled
                 ↓
              trialing → active
```

### テナントステータスとの連動

| Stripeイベント | テナントステータス |
|---------------|-------------------|
| `invoice.paid` | `active` |
| `invoice.payment_failed` | `grace`（14日間） |
| `subscription.deleted` | `grace` → `suspended` |

---

## ドメインモデル

### Subscription エンティティ

```go
type Subscription struct {
    subscriptionID       SubscriptionID
    tenantID             common.TenantID
    stripeCustomerID     string
    stripeSubscriptionID string
    status               SubscriptionStatus
    currentPeriodEnd     *time.Time
    createdAt            time.Time
    updatedAt            time.Time
}
```

### ステータス遷移バリデーション

```go
var validSubscriptionTransitions = map[SubscriptionStatus][]SubscriptionStatus{
    SubscriptionStatusIncomplete: {SubscriptionStatusActive, SubscriptionStatusCanceled},
    SubscriptionStatusTrialing:   {SubscriptionStatusActive, SubscriptionStatusPastDue, SubscriptionStatusCanceled},
    SubscriptionStatusActive:     {SubscriptionStatusPastDue, SubscriptionStatusCanceled, SubscriptionStatusUnpaid},
    SubscriptionStatusPastDue:    {SubscriptionStatusActive, SubscriptionStatusCanceled, SubscriptionStatusUnpaid},
    SubscriptionStatusUnpaid:     {SubscriptionStatusActive, SubscriptionStatusCanceled},
    SubscriptionStatusCanceled:   {}, // 終端状態
}

func (s SubscriptionStatus) CanTransitionTo(newStatus SubscriptionStatus) bool {
    if s == newStatus {
        return true // 同一ステータスへの遷移は許可（更新時）
    }
    allowed, ok := validSubscriptionTransitions[s]
    // ...
}
```

---

## Customer Portal

### セッション作成

```go
func (uc *StripePortalUsecase) CreatePortalSession(ctx context.Context, tenantID common.TenantID) (*PortalSessionOutput, error) {
    // サブスクリプションからStripe Customer IDを取得
    sub, err := uc.subscriptionRepo.FindByTenantID(ctx, tenantID)
    if err != nil {
        return nil, err
    }

    // Stripe Customer Portalセッションを作成
    params := &stripe.BillingPortalSessionParams{
        Customer:  stripe.String(sub.StripeCustomerID()),
        ReturnURL: stripe.String(returnURL),
    }

    session, err := uc.stripeClient.BillingPortalSessions.New(params)
    // ...
}
```

---

## テスト用データ

### テストカード

| カード番号 | 結果 |
|-----------|------|
| 4242424242424242 | 成功 |
| 4000000000000002 | 拒否 |
| 4000000000009995 | 残高不足 |

### Webhook テスト

```bash
# Stripe CLIでローカルテスト
stripe listen --forward-to localhost:8080/api/v1/stripe/webhook

# テストイベント送信
stripe trigger invoice.paid
stripe trigger invoice.payment_failed
```

---

## 監査ログ

全てのStripe操作は `billing_audit_logs` に記録:

```go
auditLog, err := billing.NewBillingAuditLog(
    now,
    billing.ActorTypeStripe,  // アクター種別
    nil,                       // 管理者IDはnil
    string(billing.BillingAuditActionPaymentSucceeded),
    strPtr("tenant"),
    &tenantIDStr,
    nil,
    &afterJSON,
    nil,
    nil,
)
```

---

## 環境変数

```bash
# 必須
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PRICE_ID=price_...

# オプション
STRIPE_GRACE_PERIOD_DAYS=14
```

---

## 関連ファイル

- `backend/internal/domain/billing/subscription.go` - Subscriptionエンティティ
- `backend/internal/app/payment/stripe_webhook_usecase.go` - Webhook処理
- `backend/internal/app/payment/stripe_portal_usecase.go` - Customer Portal
- `backend/internal/interface/rest/stripe_handler.go` - HTTPハンドラー

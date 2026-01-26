# サブスクリプションシナリオ一覧

Stripe月額サブスクリプションにおける全ての正常系・異常系シナリオを記載します。

## 用語定義

| 用語 | 説明 |
|------|------|
| `current_period_end` | 現在の請求期間の終了日（次回請求日） |
| `cancel_at_period_end` | 期間終了時にキャンセル予約されているか |
| `cancel_at` | キャンセル予定日時（通常は `current_period_end` と同じ） |
| `grace_until` | 猶予期間の終了日（`current_period_end` + 14日） |
| Grace Period | 支払い失敗後にサービスを継続利用できる猶予期間（14日間） |

## テナントステータス一覧

| ステータス | 説明 | ログイン | API利用 |
|-----------|------|---------|---------|
| `active` | 有効なサブスクリプション | 可 | 可 |
| `pending_payment` | 初回決済待ち（Checkout Session作成後） | 不可 | 不可 |
| `grace` | 猶予期間中（支払い失敗後14日間） | 可 | 可 |
| `suspended` | 停止（猶予期間終了後） | 不可 | 不可 |

## サブスクリプションステータス一覧

| ステータス | 説明 |
|-----------|------|
| `active` | 有効（正常に支払い済み） |
| `trialing` | トライアル期間中 |
| `past_due` | 支払い遅延（リトライ中） |
| `canceled` | キャンセル済み |
| `unpaid` | 未払い（リトライ終了） |
| `incomplete` | 初回決済未完了 |

---

## 正常系シナリオ

### シナリオ N1: 新規サブスクリプション契約

**前提条件**: テナントが存在しない、または `pending_payment` ステータス

**フロー**:
1. ユーザーが `/api/v1/payment/subscribe` を呼び出す
2. システムが Stripe Checkout Session を作成
3. テナントステータスを `pending_payment` に設定
4. ユーザーが Stripe Checkout で支払い完了
5. Webhook `checkout.session.completed` を受信
6. テナントステータスを `active` に変更
7. エンタイトルメントを付与

**結果**:
- テナントステータス: `active`
- サブスクリプションステータス: `active`

---

### シナリオ N2: 月次自動更新成功

**前提条件**: テナントが `active` ステータス

**タイムライン例**:
```
1/1  - 月額支払い完了
2/1  - 自動更新支払い成功
```

**フロー**:
1. Stripe が自動的に請求を実行
2. 支払い成功
3. Webhook `invoice.paid` を受信
4. `current_period_end` を更新

**結果**:
- テナントステータス: `active` （変更なし）
- サブスクリプションステータス: `active`

---

### シナリオ N3: サブスクリプションキャンセル予約

**前提条件**: テナントが `active` ステータス、サブスクリプションが有効

**タイムライン例**:
```
1/1  - 月額支払い完了
1/15 - ユーザーがキャンセル申請
1/31 - current_period_end（この日まで利用可能）
2/1  - サービス停止
```

**フロー**:
1. ユーザーが Customer Portal でキャンセル申請
2. Stripe が `cancel_at_period_end` を `true` に設定
3. Webhook `customer.subscription.updated` を受信
4. DB の `cancel_at_period_end` と `cancel_at` を更新
5. Admin 管理画面に「キャンセル予約中」バッジを表示

**キャンセル予約中の状態**:
- テナントステータス: `active`
- サブスクリプションステータス: `active`
- `cancel_at_period_end`: `true`
- `cancel_at`: `current_period_end` の値

**期間終了時**:
6. `current_period_end` に達する
7. Webhook `customer.subscription.deleted` を受信
8. テナントステータスを `grace` に変更（猶予期間開始）
9. `grace_until` を設定（`current_period_end` + 14日）

**結果**:
- テナントステータス: `grace`
- サブスクリプションステータス: `canceled`

---

### シナリオ N4: キャンセル予約の取り消し

**前提条件**: キャンセル予約中（`cancel_at_period_end` = `true`）

**フロー**:
1. ユーザーが Customer Portal でキャンセル取り消し
2. Stripe が `cancel_at_period_end` を `false` に設定
3. Webhook `customer.subscription.updated` を受信
4. DB の `cancel_at_period_end` を `false` に、`cancel_at` を `null` に更新

**結果**:
- テナントステータス: `active`
- サブスクリプションステータス: `active`
- `cancel_at_period_end`: `false`
- 通常の自動更新が継続

---

### シナリオ N5: 猶予期間中のサービス利用

**前提条件**: テナントが `grace` ステータス

**タイムライン例**:
```
1/31 - サブスクリプション終了、grace 開始
2/7  - ユーザーがサービスを利用（可能）
2/14 - grace_until（猶予期間終了）
2/15 - サービス停止
```

**フロー**:
1. ユーザーはログイン可能
2. 全 API が利用可能
3. `grace_until` に達するとバッチ処理で `suspended` に変更

**結果**:
- 猶予期間中は通常通りサービス利用可能

---

### シナリオ N6: Customer Portal でのカード情報更新

**前提条件**: テナントが `active` ステータス

**フロー**:
1. ユーザーが `/api/v1/payment/billing-portal` を呼び出す
2. システムが Stripe Customer Portal Session を作成
3. ユーザーが Customer Portal でカード情報を更新
4. Stripe がカード情報を更新（Webhook なし）

**結果**:
- カード情報が更新される
- テナントステータス: 変更なし
- 次回の請求は新しいカードで行われる

---

### シナリオ N7: 猶予期間中の再契約

**前提条件**: テナントが `grace` ステータス

**フロー**:
1. ユーザーが `/api/v1/payment/subscribe` を呼び出す
2. 新しい Checkout Session が作成される
3. 支払い完了後、テナントステータスが `active` に戻る
4. `grace_until` がクリアされる

**結果**:
- テナントステータス: `active`
- 猶予期間中に再契約すればサービス中断なし

---

## 異常系シナリオ

### シナリオ E1: 初回 Checkout Session 未完了

**前提条件**: Checkout Session 作成後、支払いが完了しない

**フロー**:
1. ユーザーが `/api/v1/payment/subscribe` を呼び出す
2. Checkout Session を作成
3. テナントステータスを `pending_payment` に設定
4. ユーザーが支払いを完了しない（ブラウザを閉じる等）

**結果**:
- テナントステータス: `pending_payment` のまま
- Stripe Checkout Session は 24 時間で自動期限切れ
- ユーザーは再度 `/api/v1/payment/subscribe` を呼び出して新しい Session を作成可能

**注意**: 孤立した Checkout Session は自動的に期限切れになるため、クリーンアップ処理は不要

---

### シナリオ E2: 月次更新時の支払い失敗（リトライ中）

**前提条件**: テナントが `active` ステータス

**タイムライン例**:
```
2/1  - 自動更新支払い失敗
2/1  - Stripe がリトライ開始（Smart Retries）
2/3  - 1回目のリトライ
2/6  - 2回目のリトライ
...
```

**フロー**:
1. Stripe が自動的に請求を実行
2. 支払い失敗（カード期限切れ、残高不足等）
3. Webhook `invoice.payment_failed` を受信
4. Stripe が Smart Retries を開始
5. この段階ではテナントステータスは変更しない

**結果**:
- テナントステータス: `active`（リトライ中は変更なし）
- サブスクリプションステータス: `past_due`

---

### シナリオ E3: 支払いリトライ全て失敗

**前提条件**: リトライが全て失敗

**タイムライン例**:
```
2/1  - 自動更新支払い失敗
2/15 - 全リトライ失敗、サブスクリプション canceled
2/15 - grace 期間開始
3/1  - grace_until（current_period_end + 14日）
3/2  - suspended
```

**フロー**:
1. 全てのリトライが失敗
2. Webhook `customer.subscription.deleted` を受信
3. テナントステータスを `grace` に変更
4. `grace_until` を設定（`current_period_end` + 14日）

**結果**:
- テナントステータス: `grace`
- サブスクリプションステータス: `canceled`

---

### シナリオ E4: 猶予期間終了後の自動停止

**前提条件**: テナントが `grace` ステータス、`grace_until` を過ぎている

**フロー**:
1. バッチ処理 `grace-expiry` が定期実行される
2. `grace_until` < 現在時刻 のテナントを検索
3. テナントステータスを `suspended` に変更

**結果**:
- テナントステータス: `suspended`
- ユーザーはログイン不可
- 全 API がブロック（403 Forbidden）

---

### シナリオ E5: 停止後の再契約

**前提条件**: テナントが `suspended` ステータス

**フロー**:
1. ユーザーがログインを試みる → 拒否される
2. 運営がユーザーに連絡、または自動メール通知（将来実装）
3. ユーザーが再度 `/api/v1/payment/subscribe` を呼び出す
4. 新しい Checkout Session が作成される
5. 支払い完了後、テナントステータスが `active` に戻る

**結果**:
- テナントステータス: `active`
- 新しいサブスクリプションが開始

---

### シナリオ E6: Webhook 署名検証失敗

**前提条件**: 不正な Webhook リクエスト

**フロー**:
1. 外部から不正な Webhook リクエストが送信される
2. 署名検証に失敗
3. HTTP 400 Bad Request を返す
4. データベースは変更されない

**結果**:
- 不正なリクエストは全て拒否される
- セキュリティが保護される

---

### シナリオ E7: DB トランザクション失敗

**前提条件**: Webhook 処理中に DB エラー発生

**フロー**:
1. 正当な Webhook を受信
2. 署名検証成功
3. DB 更新処理でエラー発生
4. トランザクションがロールバック
5. HTTP 500 Internal Server Error を返す
6. Stripe が自動的にリトライ（最大 72 時間）

**結果**:
- 一時的なエラーは Stripe のリトライで復旧可能
- 永続的なエラーの場合は手動対応が必要

---

### シナリオ E8: 重複 Webhook 配信

**前提条件**: Stripe が同じイベントを複数回配信

**フロー**:
1. Webhook が正常に処理される（HTTP 200）
2. ネットワークの問題で Stripe がレスポンスを受け取れない
3. Stripe が同じイベントを再送信
4. システムが再度 Webhook を処理

**対策**:
- `checkout.session.completed`: テナントが既に `active` なら何もしない
- `invoice.paid`: `current_period_end` の更新は冪等性あり
- `customer.subscription.updated`: フィールド更新は冪等性あり
- `customer.subscription.deleted`: テナントが既に `grace` または `suspended` なら何もしない

**結果**:
- 重複処理は冪等性により安全に処理される

---

### シナリオ E9: サブスクリプションが DB に存在しない

**前提条件**: Webhook でサブスクリプション ID を受け取るが、DB にレコードがない

**発生ケース**:
- `checkout.session.completed` の前に他の Webhook が到着
- 手動でサブスクリプションを作成した場合

**フロー**:
1. `customer.subscription.updated` を受信
2. `stripe_subscription_id` で DB 検索
3. レコードが見つからない
4. ログを記録し、HTTP 200 を返す（エラーにはしない）

**結果**:
- 未知のサブスクリプションは無視される
- `checkout.session.completed` で正しいレコードが作成される

---

### シナリオ E10: Webhook の到着順序逆転

**前提条件**: ネットワーク遅延により Webhook の到着順序が逆転

**例**:
```
Stripe 送信順: checkout.session.completed → invoice.paid
実際の到着順: invoice.paid → checkout.session.completed
```

**フロー**:
1. `invoice.paid` が先に到着
2. サブスクリプションが DB に存在しない
3. E9 と同様に無視される
4. `checkout.session.completed` が到着
5. サブスクリプションが作成される
6. 次回の `invoice.paid` で正常に処理される

**結果**:
- 到着順序の逆転は安全に処理される
- 初回の `invoice.paid` は無視されるが、次回以降は正常に処理

---

## 状態遷移図

```
                                    ┌─────────────────┐
                                    │   新規登録      │
                                    └────────┬────────┘
                                             │
                                             ▼
                               ┌─────────────────────────┐
                               │   pending_payment       │
                               │   (Checkout待ち)        │
                               └────────────┬────────────┘
                                            │
                          ┌─────────────────┼─────────────────┐
                          │                 │                 │
                   Checkout完了        Session期限切れ    支払い失敗
                          │                 │                 │
                          ▼                 ▼                 ▼
                    ┌──────────┐      (再試行可能)        (再試行可能)
                    │  active  │◄───────────────────────────────┘
                    └────┬─────┘
                         │
           ┌─────────────┼─────────────┐
           │             │             │
      自動更新成功  キャンセル予約  支払い失敗
           │             │             │
           ▼             ▼             │
      (active維持)  (cancel_at_      │
                    period_end=true)  │
                         │             │
                   期間終了時      全リトライ失敗
                         │             │
                         └──────┬──────┘
                                │
                                ▼
                         ┌──────────┐
                         │  grace   │
                         │ (猶予中) │
                         └────┬─────┘
                              │
                    ┌─────────┼─────────┐
                    │                   │
              猶予期間中            猶予期間終了
              (利用可能)                │
                                        ▼
                                 ┌───────────┐
                                 │ suspended │
                                 │  (停止)   │
                                 └─────┬─────┘
                                       │
                                  再契約成功
                                       │
                                       ▼
                                 (active に戻る)
```

---

## Webhook イベント一覧

| イベント | トリガー | システムの処理 |
|---------|---------|---------------|
| `checkout.session.completed` | 初回支払い完了 | テナントを `active` に、エンタイトルメント付与 |
| `invoice.paid` | 月次更新成功 | `current_period_end` 更新 |
| `invoice.payment_failed` | 支払い失敗 | ログ記録（ステータス変更なし） |
| `customer.subscription.updated` | サブスクリプション更新 | `cancel_at_period_end`、`cancel_at` 更新 |
| `customer.subscription.deleted` | サブスクリプション終了 | テナントを `grace` に、`grace_until` 設定 |

---

## 管理者向け操作

### Admin 管理画面での表示

1. **テナントステータス**: active / grace / suspended / pending_payment
2. **サブスクリプションステータス**: active / past_due / canceled 等
3. **キャンセル予約中バッジ**: `cancel_at_period_end` が `true` の場合に表示
4. **キャンセル予定日**: `cancel_at` の日時を表示

### 手動ステータス変更

管理者は以下の操作が可能：
- `active` → `suspended`: 強制停止
- `suspended` → `active`: 手動復活
- `grace` → `active`: 猶予期間中に手動復活

---

## トラブルシューティング

### Q1: ユーザーが「ログインできない」と報告

1. Admin 管理画面でテナントステータスを確認
2. `suspended` の場合 → 支払い状況を確認、必要に応じて再契約を案内
3. `pending_payment` の場合 → Checkout を完了するよう案内

### Q2: Webhook が処理されていない

1. Stripe Dashboard で Webhook ログを確認
2. 署名検証エラー → `STRIPE_WEBHOOK_SECRET` の設定を確認
3. DB エラー → サーバーログを確認、必要に応じて手動でステータス更新

### Q3: キャンセル予約が反映されない

1. `customer.subscription.updated` Webhook が受信されているか確認
2. DB の `cancel_at_period_end` カラムを確認
3. 必要に応じて手動で更新

---

## アーキテクチャ: 集約関係

### Tenant と Subscription の関係

```
┌─────────────────────┐     参照      ┌─────────────────────┐
│      Tenant         │◄─────────────│    Subscription     │
│  (集約ルート)        │  tenantID    │   (集約ルート)       │
├─────────────────────┤              ├─────────────────────┤
│ - tenantID (PK)     │              │ - subscriptionID    │
│ - status            │              │ - tenantID (FK)     │
│ - graceUntil        │              │ - stripeCustomerID  │
│ - ...               │              │ - status            │
└─────────────────────┘              └─────────────────────┘
```

### 設計方針

**Subscription は Tenant とは別の独立した集約**として設計されています。

**理由**:
1. **Stripe が真実の源**: サブスクリプションデータの正確な状態は Stripe が保持
2. **Webhook による同期**: 状態変更は Webhook イベントで同期される
3. **ドメイン境界の明確化**: 課金とテナント管理は異なる関心事
4. **トランザクション境界**: Webhook ハンドラが両方の更新を調整

**トレードオフ**:
- 一貫性は Webhook 処理の正確さに依存
- 同一トランザクションで両方を更新する必要がある

### Webhook ハンドラの役割

```go
// stripe_webhook_usecase.go での調整パターン

func (uc *StripeWebhookUsecase) handleEvent(...) {
    return uc.txManager.WithTx(ctx, func(txCtx context.Context) error {
        // 1. Subscription を更新
        sub, _ := uc.subscriptionRepo.FindBy...(txCtx, ...)
        sub.UpdateStatus(...)
        uc.subscriptionRepo.Save(txCtx, sub)

        // 2. Tenant を更新（Subscription の状態に基づいて）
        tenant, _ := uc.tenantRepo.FindByID(txCtx, sub.TenantID())
        tenant.SetStatusActive(...)  // または SetStatusGrace など
        uc.tenantRepo.Save(txCtx, tenant)

        // 3. 監査ログ
        uc.auditLogRepo.Save(txCtx, auditLog)

        return nil  // トランザクションコミット
    })
}
```

この設計により、Subscription と Tenant の状態は常に整合性が保たれます。

---

## 関連ファイル

| ファイル | 説明 |
|---------|------|
| `internal/app/payment/stripe_webhook_usecase.go` | Webhook 処理ロジック（集約調整） |
| `internal/interface/rest/stripe_webhook_handler.go` | Webhook エンドポイント |
| `internal/domain/billing/subscription.go` | サブスクリプションドメインモデル |
| `internal/domain/tenant/tenant.go` | テナントドメインモデル |
| `internal/infra/db/subscription_repository.go` | リポジトリ実装 |
| `cmd/batch/main.go` | grace-expiry バッチ処理 |

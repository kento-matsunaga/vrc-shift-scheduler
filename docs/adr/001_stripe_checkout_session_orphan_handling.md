# ADR-001: Stripe Checkout Session の孤立セッション処理

## ステータス

承認済み (Accepted)

## コンテキスト

新規サブスクリプション登録フローでは、以下の処理が必要:

1. Stripe Checkout Session の作成
2. データベースへのテナント・管理者の保存

これらの処理の実行順序と、失敗時のリカバリ方法を決定する必要がある。

### 問題点

- DB保存が失敗した場合、Stripe側に孤立したCheckout Sessionが残る可能性がある
- Stripe API呼び出しが失敗した場合、DBに不整合なデータが残る可能性がある

## 検討したオプション

### オプション1: DB先行（Session IDをプレースホルダで保存）

```
1. DBにテナント作成（pending_stripe_session_id = "pending"）
2. Stripe API呼び出し
3. DBを更新（pending_stripe_session_id = 実際のSession ID）
```

**メリット:**
- DBトランザクションで一貫性を保てる

**デメリット:**
- 2段階更新が必要で複雑
- ステップ2-3間で障害が発生すると、"pending"のまま残る

### オプション2: Stripe先行（採用）

```
1. Stripe Checkout Session 作成
2. DBトランザクションでテナント・管理者を保存
```

**メリット:**
- シンプルな実装
- DB障害時のリカバリが容易（Stripeセッションは自動期限切れ）

**デメリット:**
- DB保存失敗時にStripe側に孤立セッションが残る

### オプション3: Sagaパターン（補償トランザクション）

```
1. Stripe Checkout Session 作成
2. DB保存
3. 失敗時: Stripe Session をキャンセル
```

**メリット:**
- 完全な一貫性

**デメリット:**
- 実装が複雑
- キャンセル処理自体が失敗する可能性

## 決定

**オプション2（Stripe先行）を採用**

## 理由

1. **Stripe Checkout Sessionの自動期限切れ**
   - 設定された期限（デフォルト24時間、`CHECKOUT_SESSION_EXPIRE_MINUTES`で設定可能）で自動的に期限切れになる
   - 特別なクリーンアップ処理が不要

2. **シンプルさの優先**
   - 複雑な補償トランザクションより、シンプルな実装を優先
   - 障害からのリカバリが容易

3. **ビジネスへの影響が限定的**
   - 孤立セッションはStripe側で自動処理される
   - 課金は発生しない（セッション完了前）
   - ユーザーは再度登録を試行可能

## 影響

### 許容される状態

- **正常系**: Stripe Session作成 → DB保存成功 → ユーザーがCheckoutで決済
- **DB失敗時**: Stripe Session作成 → DB保存失敗 → 孤立Session（24時間後に自動期限切れ）

### 監視

- Stripeダッシュボードで期限切れSessionの件数を確認可能
- 大量の孤立Sessionが発生する場合は、DB障害の可能性を調査

### 許容範囲

- 孤立Sessionの発生頻度: 月間数件程度であれば正常
- それ以上の場合は、DBやネットワークの問題を調査

## 関連ファイル

- `backend/internal/app/payment/subscribe_usecase.go` - 実装箇所

## 参考

- [Stripe Checkout Session API](https://stripe.com/docs/api/checkout/sessions)
- [Stripe Session Expiration](https://stripe.com/docs/payments/checkout/custom#session-expiration)

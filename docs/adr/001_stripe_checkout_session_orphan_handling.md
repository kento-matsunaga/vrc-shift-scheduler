# ADR-001: Stripe Checkout Session 孤立化の許容

## ステータス

承認済み（Accepted）

## コンテキスト

VRC Shift Scheduler では、新規テナント登録時に Stripe Checkout Session を使用して決済を行う。
この処理には以下の2つの操作が含まれる：

1. Stripe API で Checkout Session を作成
2. DB トランザクションでテナントと管理者を作成

これらの操作は原子的に実行できないため、どちらかが失敗した場合のリカバリ方法を決定する必要がある。

## 検討した選択肢

### Option A: Session 作成 → DB 保存（採用）

```
1. Stripe Checkout Session を作成（session_id を取得）
2. DB トランザクションでテナント作成（session_id を保存）
3. (失敗時) Stripe 側に孤立 Session が残る
```

**メリット:**
- シンプルな実装
- 孤立 Session は 24 時間で自動期限切れ
- テナント作成失敗時、ユーザーは再試行可能

**デメリット:**
- Stripe 側に一時的に孤立 Session が残る可能性

### Option B: DB 保存 → Session 作成

```
1. DB トランザクションでテナント作成（session_id = null）
2. Stripe Checkout Session を作成
3. DB で session_id を更新
4. (失敗時) DB に孤立テナントが残る
```

**メリット:**
- Stripe 側に孤立 Session が残らない

**デメリット:**
- 複雑な実装（3ステップ）
- DB に孤立テナントが残る可能性があり、クリーンアップが必要
- 孤立テナントのステータス管理が複雑

### Option C: Session 作成失敗時に明示的に expire

```
1. Stripe Checkout Session を作成
2. DB トランザクションでテナント作成
3. (DB失敗時) Stripe Session を明示的に expire
```

**メリット:**
- 孤立 Session を即座にクリーンアップ

**デメリット:**
- Stripe API 呼び出しが増える（expire 処理）
- expire API 呼び出しも失敗する可能性がある
- 結局 Option A と同等の孤立リスクが残る

## 決定

**Option A を採用する。**

理由：
1. **シンプルさ**: 実装がシンプルで理解しやすい
2. **Stripe の自動クリーンアップ**: Checkout Session は 24 時間で自動的に期限切れになる
3. **リスクの低さ**: 孤立 Session による実害がない（課金されない、ユーザーに影響なし）
4. **リカバリの容易さ**: DB トランザクション失敗時、ユーザーは単純に再試行できる

## 結果

### 影響

- `subscribe_usecase.go` で Session 作成を先に行い、その後 DB トランザクションを実行
- 孤立 Session は Stripe 側で 24 時間後に自動削除
- 特別なクリーンアップ処理は実装しない

### 監視

- Stripe Dashboard で未完了の Checkout Session 数を定期確認
- 異常に多い場合は DB トランザクション失敗率を調査

### 許容範囲

- 孤立 Session 数: 通常運用で 1-2 件/日 は許容範囲
- 大量発生時（10 件/日以上）: システム障害の可能性を調査

## 関連

- `backend/internal/app/payment/subscribe_usecase.go` - 実装箇所
- Stripe Documentation: [Checkout Session Expiration](https://stripe.com/docs/payments/checkout/how-checkout-works#session-expiration)

## 作成日

2026-01-26

## 作成者

Claude Opus 4.5 (Code Assistant)

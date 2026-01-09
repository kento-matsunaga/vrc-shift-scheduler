# Issue #140: Instance エンティティ移行手順書

## 概要

このドキュメントは、シフト枠（shift_slots）の `instance_name` 文字列フィールドを、
正規化された `Instance` エンティティへ移行する手順を説明します。

### 変更内容

- `instances` テーブルの新規作成
- `shift_slots` テーブルへの `instance_id` カラム追加
- 既存データの移行（instance_name → Instance エンティティ）

### 後方互換性

- `shift_slots.instance_name` カラムは残存（Deprecated）
- `shift_slots.instance_id` は nullable（移行期間中は NULL 許容）
- ロールバック可能な設計

---

## 前提条件

- PostgreSQL データベースへのアクセス権限
- マイグレーションツール（golang-migrate）がインストール済み
- バックアップが取得済み

---

## セットアップ手順

### Step 1: バックアップ取得（必須）

```bash
# 本番環境の場合、必ずバックアップを取得
pg_dump -h <host> -U <user> -d <database> -F c -f backup_before_140.dump
```

### Step 2: マイグレーション実行

```bash
# マイグレーションディレクトリに移動
cd backend/internal/infra/db/migrations

# 環境変数設定
export DATABASE_URL="postgres://<user>:<password>@<host>:<port>/<database>?sslmode=disable"

# マイグレーション実行（3つのファイルを順番に適用）
# 037: instances テーブル作成
# 038: shift_slots に instance_id カラム追加
# 039: 既存データ移行

# golang-migrate を使用する場合
migrate -path . -database "${DATABASE_URL}" up

# または Makefile を使用
cd ../../..  # backend ディレクトリへ
make migrate-up
```

### Step 3: 移行結果の確認

```sql
-- 作成された Instance の確認
SELECT
    i.instance_id,
    i.event_id,
    i.name,
    i.display_order,
    COUNT(ss.slot_id) as slot_count
FROM instances i
LEFT JOIN shift_slots ss ON ss.instance_id = i.instance_id
WHERE i.deleted_at IS NULL
GROUP BY i.instance_id, i.event_id, i.name, i.display_order
ORDER BY i.event_id, i.display_order;

-- instance_id が正しく紐付けられているか確認
SELECT
    COUNT(*) as total_slots,
    COUNT(instance_id) as slots_with_instance_id,
    COUNT(*) - COUNT(instance_id) as slots_without_instance_id
FROM shift_slots
WHERE deleted_at IS NULL
  AND instance_name IS NOT NULL
  AND instance_name != '';

-- 紐付けされていないシフト枠がないか確認
SELECT slot_id, instance_name, instance_id
FROM shift_slots
WHERE deleted_at IS NULL
  AND instance_name IS NOT NULL
  AND instance_name != ''
  AND instance_id IS NULL;
```

### Step 4: アプリケーションデプロイ

```bash
# バックエンドのビルドとデプロイ
cd backend
go build -o server ./cmd/server/
# サーバー再起動
```

---

## ロールバック手順

問題が発生した場合、以下の手順でロールバックできます。

### Option A: マイグレーションツールでロールバック

```bash
# 環境変数設定
export DATABASE_URL="postgres://<user>:<password>@<host>:<port>/<database>?sslmode=disable"

# 3つのマイグレーションをロールバック（逆順）
migrate -path backend/internal/infra/db/migrations -database "${DATABASE_URL}" down 3

# または個別にロールバック
migrate -path backend/internal/infra/db/migrations -database "${DATABASE_URL}" down 1  # 039
migrate -path backend/internal/infra/db/migrations -database "${DATABASE_URL}" down 1  # 038
migrate -path backend/internal/infra/db/migrations -database "${DATABASE_URL}" down 1  # 037
```

### Option B: 手動でロールバック

```sql
-- Step 1: shift_slots.instance_id を NULL に戻す
UPDATE shift_slots
SET instance_id = NULL,
    updated_at = NOW()
WHERE instance_id IS NOT NULL;

-- Step 2: shift_slots.instance_name のコメントを削除
COMMENT ON COLUMN shift_slots.instance_name IS NULL;

-- Step 3: 外部キー制約を削除
ALTER TABLE shift_slots DROP CONSTRAINT IF EXISTS fk_shift_slots_instance;

-- Step 4: instance_id カラムを削除
ALTER TABLE shift_slots DROP COLUMN IF EXISTS instance_id;

-- Step 5: instances テーブルを削除
DROP TABLE IF EXISTS instances;
```

### Option C: バックアップからリストア

```bash
# 完全なロールバックが必要な場合
pg_restore -h <host> -U <user> -d <database> -c backup_before_140.dump
```

---

## トラブルシューティング

### 問題: マイグレーション 039 が失敗する

**原因**: `shift_slots` と `event_business_days` の JOIN に問題がある可能性

**対処**:
```sql
-- 孤立した shift_slots を確認
SELECT ss.slot_id, ss.business_day_id
FROM shift_slots ss
LEFT JOIN event_business_days ebd ON ss.business_day_id = ebd.business_day_id
WHERE ebd.business_day_id IS NULL
  AND ss.deleted_at IS NULL;

-- 必要に応じて孤立データを削除
DELETE FROM shift_slots
WHERE business_day_id NOT IN (
    SELECT business_day_id FROM event_business_days
);
```

### 問題: 重複した Instance が作成される

**原因**: 同じ event_id + name の組み合わせが既に存在

**対処**:
```sql
-- 重複を確認
SELECT tenant_id, event_id, name, COUNT(*)
FROM instances
WHERE deleted_at IS NULL
GROUP BY tenant_id, event_id, name
HAVING COUNT(*) > 1;

-- 重複を解消（古い方を論理削除）
-- 具体的なクエリは状況に応じて調整
```

### 問題: アプリケーションエラー "instance_id not found"

**原因**: フロントエンドとバックエンドのバージョン不整合

**対処**:
1. フロントエンドとバックエンドを同時にデプロイ
2. または、古いバックエンドを使用して instance_id を無視するように設定

---

## 確認チェックリスト

移行完了後、以下を確認してください：

- [ ] `instances` テーブルが作成されている
- [ ] `shift_slots.instance_id` カラムが追加されている
- [ ] 既存の `instance_name` に対応する Instance エンティティが作成されている
- [ ] `shift_slots.instance_id` が正しく紐付けられている
- [ ] API エンドポイント `/api/v1/events/:eventId/instances` が動作する
- [ ] フロントエンドのインスタンス管理画面が表示される
- [ ] テンプレート適用時に Instance が自動作成される

---

## 関連ファイル

- `backend/internal/infra/db/migrations/037_create_instances.up.sql`
- `backend/internal/infra/db/migrations/037_create_instances.down.sql`
- `backend/internal/infra/db/migrations/038_add_instance_id_to_shift_slots.up.sql`
- `backend/internal/infra/db/migrations/038_add_instance_id_to_shift_slots.down.sql`
- `backend/internal/infra/db/migrations/039_migrate_instance_data.up.sql`
- `backend/internal/infra/db/migrations/039_migrate_instance_data.down.sql`

---

## 変更履歴

| 日付 | 内容 |
|------|------|
| 2026-01-08 | 初版作成（Issue #140） |

-- Migration: 039_migrate_instance_data
-- Description: 既存のshift_slots.instance_nameからInstanceエンティティを作成し、instance_idを紐付け
-- Issue #140

-- Step 1: 既存のinstance_nameからInstanceエンティティを作成
-- shift_slotsをevent_business_daysと結合して、event_idごとにユニークなinstance_nameを取得
INSERT INTO instances (instance_id, tenant_id, event_id, name, display_order, max_members, created_at, updated_at, deleted_at)
SELECT
    -- ULIDの代わりにUUIDを使用（PostgreSQL標準機能）
    -- ULIDフォーマット: 26文字の英数字（Crockford's Base32）
    -- ここではgen_random_uuid()でUUIDを生成し、適切な形式に変換
    UPPER(REPLACE(gen_random_uuid()::text, '-', '')) AS instance_id,
    ss.tenant_id,
    ebd.event_id,
    ss.instance_name AS name,
    ROW_NUMBER() OVER (PARTITION BY ss.tenant_id, ebd.event_id ORDER BY MIN(ss.created_at)) - 1 AS display_order,
    NULL AS max_members,
    MIN(ss.created_at) AS created_at,
    NOW() AS updated_at,
    NULL AS deleted_at
FROM shift_slots ss
INNER JOIN event_business_days ebd ON ss.business_day_id = ebd.business_day_id
WHERE ss.instance_name IS NOT NULL
  AND ss.instance_name != ''
  AND ss.deleted_at IS NULL
  AND ebd.deleted_at IS NULL
GROUP BY ss.tenant_id, ebd.event_id, ss.instance_name
-- 既に同名のインスタンスが存在しない場合のみ挿入
ON CONFLICT DO NOTHING;

-- Step 2: shift_slotsのinstance_idを更新
-- 作成したInstanceエンティティのIDを紐付け
UPDATE shift_slots ss
SET instance_id = i.instance_id,
    updated_at = NOW()
FROM event_business_days ebd
INNER JOIN instances i ON ebd.event_id = i.event_id
                      AND ebd.tenant_id = i.tenant_id
                      AND ss.instance_name = i.name
WHERE ss.business_day_id = ebd.business_day_id
  AND ss.instance_name IS NOT NULL
  AND ss.instance_name != ''
  AND ss.deleted_at IS NULL
  AND i.deleted_at IS NULL;

-- Step 3: 確認用コメント
COMMENT ON COLUMN shift_slots.instance_name IS 'Deprecated: instance_idを使用してください。移行完了後に削除予定。';

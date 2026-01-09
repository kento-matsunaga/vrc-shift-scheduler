-- Migration: 039_migrate_instance_data (rollback)
-- Description: データ移行をロールバック
-- Issue #140

-- Step 1: shift_slotsのinstance_idをNULLに戻す
-- instance_nameは残っているので、データは保持される
UPDATE shift_slots
SET instance_id = NULL,
    updated_at = NOW()
WHERE instance_id IS NOT NULL;

-- Step 2: 移行で作成されたInstanceエンティティを削除
-- 注意: 手動で作成されたInstanceは削除しない（移行前には存在しなかったため）
-- 移行で作成されたInstanceは、instance_nameに対応するshift_slotが存在する
-- ただし、このrollbackでは全てのInstanceを削除する（安全のため）
-- 本番環境では慎重に実行すること

-- コメント解除: 移行で作成されたインスタンスのみを削除する場合
-- DELETE FROM instances
-- WHERE instance_id IN (
--     SELECT DISTINCT i.instance_id
--     FROM instances i
--     INNER JOIN shift_slots ss ON ss.instance_id = i.instance_id
--     WHERE ss.instance_name IS NOT NULL
--       AND ss.instance_name = i.name
-- );

-- Step 3: COMMENTを削除
COMMENT ON COLUMN shift_slots.instance_name IS NULL;

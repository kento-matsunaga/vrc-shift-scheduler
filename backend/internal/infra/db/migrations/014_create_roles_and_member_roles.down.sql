-- Migration: 014_create_roles_and_member_roles (Rollback)
-- Description: ロールシステムのロールバック

-- 1. インデックスを削除
DROP INDEX IF EXISTS uq_role_name_per_tenant;
DROP INDEX IF EXISTS idx_member_roles_role_id;
DROP INDEX IF EXISTS idx_roles_deleted_at;
DROP INDEX IF EXISTS idx_roles_display_order;
DROP INDEX IF EXISTS idx_roles_tenant_id;

-- 2. メンバーロール関連テーブルを削除
DROP TABLE IF EXISTS member_roles;

-- 3. ロールテーブルを削除
DROP TABLE IF EXISTS roles;

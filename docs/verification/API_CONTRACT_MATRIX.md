# API契約（JSON形式）の整合性チェック

> 最終更新: 2025-12-19

## 概要

このドキュメントは、主要エンティティについて「DBマイグレーション → Goドメイン構造体 → RESTレスポンス → TypeScript型」の間でフィールド名・型・必須/任意が揃っているかを確認したものです。

---

## 1. Event（イベント）

### フィールドマトリクス

| フィールド名 | DB型 | Go型 | REST | TS型 | 状態 |
|-------------|------|------|------|------|------|
| event_id | CHAR(26) | EventID | ✅ | ✅ | ✅ OK |
| tenant_id | CHAR(26) | TenantID | ✅ | ✅ | ✅ OK |
| event_name | VARCHAR(255) | string | ✅ | ✅ | ✅ OK |
| event_type | VARCHAR(20) | EventType | ✅ | ✅ | ✅ OK |
| description | TEXT | string | ✅ | ✅ | ✅ OK |
| is_active | BOOLEAN | bool | ✅ | ✅ | ✅ OK |
| created_at | TIMESTAMPTZ | time.Time | ✅ | ✅ | ✅ OK |
| updated_at | TIMESTAMPTZ | time.Time | ✅ | ✅ | ✅ OK |
| deleted_at | TIMESTAMPTZ NULL | *time.Time | - | - | ✅ OK（除外）|

### 定期パターン関連（追加フィールド）

| フィールド名 | 説明 | 状態 |
|-------------|------|------|
| recurrence_type | 定期パターン種別 | ✅ 実装済 |
| recurrence_start_date | 開始日 | ✅ 実装済 |
| recurrence_day_of_week | 曜日 | ✅ 実装済 |
| default_start_time | デフォルト開始時刻 | ✅ 実装済 |
| default_end_time | デフォルト終了時刻 | ✅ 実装済 |

---

## 2. EventBusinessDay（営業日）

### フィールドマトリクス

| フィールド名 | DB型 | Go型 | REST | TS型 | 状態 |
|-------------|------|------|------|------|------|
| business_day_id | CHAR(26) | BusinessDayID | ✅ | ✅ | ✅ OK |
| tenant_id | CHAR(26) | TenantID | ✅ | ✅ | ✅ OK |
| event_id | CHAR(26) | EventID | ✅ | ✅ | ✅ OK |
| target_date | DATE | time.Time | ✅ | ✅ | ✅ OK |
| start_time | TIME | time.Time | ✅ | ✅ | ✅ OK |
| end_time | TIME | time.Time | ✅ | ✅ | ✅ OK |
| occurrence_type | VARCHAR(20) | string | ✅ | ✅ | ✅ OK |
| is_active | BOOLEAN | bool | ✅ | ✅ | ✅ OK |
| created_at | TIMESTAMPTZ | time.Time | ✅ | ✅ | ✅ OK |
| updated_at | TIMESTAMPTZ | time.Time | ✅ | ✅ | ✅ OK |

---

## 3. ShiftSlot（シフト枠）

### フィールドマトリクス

| フィールド名 | DB型 | Go型 | REST | TS型 | 状態 |
|-------------|------|------|------|------|------|
| slot_id | CHAR(26) | SlotID | ✅ | ✅ | ✅ OK |
| tenant_id | CHAR(26) | TenantID | ✅ | ✅ | ✅ OK |
| business_day_id | CHAR(26) | BusinessDayID | ✅ | ✅ | ✅ OK |
| position_id | CHAR(26) | PositionID | ✅ | ✅ | ✅ OK |
| slot_name | VARCHAR(255) | string | ✅ | ✅ | ✅ OK |
| instance_name | VARCHAR(255) NULL | string | ✅ | ✅ | ✅ OK |
| start_time | TIME | time.Time | ✅ | ✅ | ✅ OK |
| end_time | TIME | time.Time | ✅ | ✅ | ✅ OK |
| required_count | INT | int | ✅ | ✅ | ✅ OK |
| **assigned_count** | (JOIN) | int | ✅ | ✅ | ✅ 実装済 |
| priority | INT | int | ✅ | ✅ | ✅ OK |
| is_overnight | (計算値) | bool | ✅ | ✅ | ✅ OK |

---

## 4. ShiftAssignment（シフト割り当て）

### フィールドマトリクス

| フィールド名 | DB型 | Go型 | REST | TS型 | 状態 |
|-------------|------|------|------|------|------|
| assignment_id | CHAR(26) | AssignmentID | ✅ | ✅ | ✅ OK |
| tenant_id | CHAR(26) | TenantID | ✅ | ✅ | ✅ OK |
| slot_id | CHAR(26) | SlotID | ✅ | ✅ | ✅ OK |
| member_id | CHAR(26) | MemberID | ✅ | ✅ | ✅ OK |
| **member_display_name** | (JOIN) | string | ✅ | ✅ | ✅ 実装済 |
| **slot_name** | (JOIN) | string | ✅ | ✅ | ✅ 実装済 |
| **target_date** | (JOIN) | string | ✅ | ✅ | ✅ 実装済 |
| **start_time** | (JOIN) | string | ✅ | ✅ | ✅ 実装済 |
| **end_time** | (JOIN) | string | ✅ | ✅ | ✅ 実装済 |
| assignment_status | VARCHAR(20) | string | ✅ | ✅ | ✅ OK |
| assignment_method | VARCHAR(20) | string | ✅ | ✅ | ✅ OK |
| is_outside_preference | BOOLEAN | bool | ✅ | ✅ | ✅ OK |
| assigned_at | TIMESTAMPTZ | time.Time | ✅ | ✅ | ✅ OK |
| cancelled_at | TIMESTAMPTZ NULL | *time.Time | ✅ | ✅ | ✅ OK |
| note | TEXT NULL | string | ✅ | ✅ | ✅ OK |

### フィルタ機能

| パラメータ | 説明 | 状態 |
|-----------|------|------|
| member_id | メンバーIDでフィルタ | ✅ 実装済 |
| slot_id | シフト枠IDでフィルタ | ✅ 実装済 |
| assignment_status | ステータスでフィルタ | ✅ 実装済 |
| **start_date** | 日付範囲（開始） | ✅ 実装済 |
| **end_date** | 日付範囲（終了） | ✅ 実装済 |

---

## 5. Member（メンバー）

### フィールドマトリクス

| フィールド名 | DB型 | Go型 | REST | TS型 | 状態 |
|-------------|------|------|------|------|------|
| member_id | CHAR(26) | MemberID | ✅ | ✅ | ✅ OK |
| tenant_id | CHAR(26) | TenantID | ✅ | ✅ | ✅ OK |
| display_name | VARCHAR(255) | string | ✅ | ✅ | ✅ OK |
| discord_user_id | VARCHAR(100) NULL | string | ✅ | ✅ | ✅ OK |
| email | VARCHAR(255) NULL | string | ✅ | ✅ | ✅ OK |
| is_active | BOOLEAN | bool | ✅ | ✅ | ✅ OK |
| **roles** | (JOIN) | []Role | ✅ | ✅ | ✅ 実装済 |

---

## 6. Role（ロール）

### フィールドマトリクス

| フィールド名 | DB型 | Go型 | REST | TS型 | 状態 |
|-------------|------|------|------|------|------|
| role_id | CHAR(26) | RoleID | ✅ | ✅ | ✅ OK |
| tenant_id | CHAR(26) | TenantID | ✅ | ✅ | ✅ OK |
| name | VARCHAR(100) | string | ✅ | ✅ | ✅ OK |
| description | VARCHAR(500) NULL | string | ✅ | ✅ | ✅ OK |
| color | VARCHAR(20) NULL | string | ✅ | ✅ | ✅ OK |
| display_order | INT | int | ✅ | ✅ | ✅ OK |

---

## 7. AttendanceCollection（出欠確認）

### フィールドマトリクス

| フィールド名 | DB型 | Go型 | REST | TS型 | 状態 |
|-------------|------|------|------|------|------|
| collection_id | CHAR(26) | CollectionID | ✅ | ✅ | ✅ OK |
| tenant_id | CHAR(26) | TenantID | ✅ | ✅ | ✅ OK |
| title | VARCHAR(255) | string | ✅ | ✅ | ✅ OK |
| description | TEXT NULL | string | ✅ | ✅ | ✅ OK |
| target_type | VARCHAR(20) | string | ✅ | ✅ | ✅ OK |
| target_id | CHAR(26) | string | ✅ | ✅ | ✅ OK |
| public_token | UUID | string | ✅ | ✅ | ✅ OK |
| status | VARCHAR(20) | string | ✅ | ✅ | ✅ OK |
| deadline | TIMESTAMPTZ NULL | *time.Time | ✅ | ✅ | ✅ OK |

---

## 8. DateSchedule（日程調整）

### フィールドマトリクス

| フィールド名 | DB型 | Go型 | REST | TS型 | 状態 |
|-------------|------|------|------|------|------|
| schedule_id | CHAR(26) | ScheduleID | ✅ | ✅ | ✅ OK |
| tenant_id | CHAR(26) | TenantID | ✅ | ✅ | ✅ OK |
| title | VARCHAR(255) | string | ✅ | ✅ | ✅ OK |
| description | TEXT NULL | string | ✅ | ✅ | ✅ OK |
| event_id | CHAR(26) NULL | *EventID | ✅ | ✅ | ✅ OK |
| public_token | UUID | string | ✅ | ✅ | ✅ OK |
| status | VARCHAR(20) | string | ✅ | ✅ | ✅ OK |
| deadline | TIMESTAMPTZ NULL | *time.Time | ✅ | ✅ | ✅ OK |
| decided_date | CHAR(26) NULL | *CandidateDateID | ✅ | ✅ | ✅ OK |

---

## 9. まとめ

### 現在の状態

すべての主要エンティティにおいて、DB → Go → REST → TypeScript の整合性が取れています。

### 以前の問題点（解決済み）

以下の問題は実装により解決済みです：

| 問題 | 状態 |
|------|------|
| ShiftSlot.assigned_count が常に0 | ✅ 解決済（JOINで取得） |
| ShiftAssignment のJOINフィールド未実装 | ✅ 解決済 |
| ShiftAssignment の日付範囲フィルタ未実装 | ✅ 解決済 |
| Member のロール情報未取得 | ✅ 解決済 |

---

**作成日**: 2025-12-19
**更新者**: ドキュメント検証

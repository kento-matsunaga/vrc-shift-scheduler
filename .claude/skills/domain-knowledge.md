---
description: 各ドメイン領域の業務知識、エンティティ、制約条件
---

# Domain Knowledge

VRC Shift Scheduler の各ドメイン領域の業務知識。機能開発時のドメイン理解に使用。

---

## 1. テナント・イベント領域

### テナント（Tenant）

VRChat内で活動する団体・店舗・イベント運営チームの単位。

| 属性 | 説明 |
|------|------|
| tenantID | ULID形式の一意識別子 |
| tenantName | テナント名 |
| timezone | タイムゾーン（デフォルト: Asia/Tokyo） |
| isActive | アクティブ状態 |

**重要**: マルチテナント設計により、各テナントのデータは完全に分離される。

### イベント（Event）

テナントが運営する営業ブランド・企画の単位。

| 属性 | 説明 |
|------|------|
| eventID | ULID形式の一意識別子 |
| tenantID | 所属するテナントのID |
| eventName | イベント名 |
| eventType | `normal`（通常営業）/ `special`（特別営業） |
| recurrenceType | `none` / `weekly` / `biweekly` |

### 営業日（EventBusinessDay）

実際にシフトを組む対象となる「1回分の営業日」。

| 属性 | 説明 |
|------|------|
| businessDayID | ULID形式の一意識別子 |
| eventID | 関連するイベントのID |
| targetDate | 営業日（DATE） |
| startTime / endTime | 開始・終了時刻（TIME） |
| occurrenceType | `recurring`（定期）/ `special`（特別） |

### 制約条件

1. **テナント境界**: イベントは必ず1つのテナントに属する
2. **深夜営業対応**: 終了時刻 < 開始時刻 の場合、日付をまたぐ
3. **有効期間**: `validFrom`と`validTo`は両方設定するか、両方null

---

## 2. メンバー・ロール領域

### メンバー（Member）

テナントに所属するキャスト・店長・スタッフ。

| 属性 | 説明 |
|------|------|
| memberID | ULID形式の一意識別子 |
| tenantID | 所属するテナントのID |
| displayName | 表示名（キャスト名） |
| discordUserID | Discord連携用ID |
| isActive | アクティブ状態 |

### ロール（Role）

メンバーの役割分類。テナントごとにカスタム定義可能。

| 属性 | 説明 |
|------|------|
| roleID | ULID形式の一意識別子 |
| name | ロール名（例: キャスト、スタッフ） |
| color | UI表示用の色コード |
| displayOrder | 表示順序 |

### 管理者ロール（Admin Role）

| 値 | 説明 |
|-----|------|
| `owner` | テナントの最終責任者。すべての操作が可能 |
| `manager` | テナントの管理者。通常のシフト管理操作が可能 |

---

## 3. シフト枠・ポジション領域

### ポジション（Position）

営業時に必要となる役割の種別。

| 属性 | 説明 |
|------|------|
| positionID | ULID形式の一意識別子 |
| positionName | ポジション名 |
| displayOrder | 表示順序 |

**用途例**: 受付、案内、配信、MC、カウンター、テーブル

### シフト枠（ShiftSlot）

特定の営業日における「時間帯×インスタンス×ポジション」の組み合わせで定義される「1人分の席」（複数枠で複数人配置）。

| 属性 | 説明 |
|------|------|
| slotID | ULID形式の一意識別子 |
| businessDayID | 関連する営業日のID |
| positionID | 関連するポジションのID |
| slotName | 枠名 |
| instanceName | インスタンス名（例: 第1インスタンス） |
| startTime / endTime | 開始・終了時刻 |
| requiredCount | 必要人数（1以上） |

### 制約条件

1. **必要人数**: `requiredCount` は1以上
2. **ポジション参照**: 削除されたポジションを新規シフト枠に指定不可
3. **深夜営業**: `IsOvernight()` で判定

---

## 4. シフト割り当て領域

### シフト割り当て（ShiftAssignment）

メンバーをシフト枠に配置した結果。

| 属性 | 説明 |
|------|------|
| assignmentID | ULID形式の一意識別子 |
| slotID | シフト枠ID |
| memberID | 配置されたメンバーのID |
| assignmentStatus | `confirmed` / `cancelled` |
| assignmentMethod | `auto` / `manual` |
| isOutsidePreference | 希望外配置フラグ |

### ステータスの違い

| 状態 | 説明 |
|------|------|
| `confirmed` | 有効な割り当て |
| `cancelled` | メンバーがキャンセル（履歴として保持） |
| `deleted_at` あり | 管理者が誤作成を削除（履歴から除外） |

### 割り当てフロー

1. シフト枠確認
2. 出欠確認（参加可能なメンバーを把握）
3. メンバー選択
4. 割り当て作成
5. 確認・調整

### 制約条件

1. **重複防止**: 同一メンバーを同一シフト枠に重複割り当て不可
2. **配置上限**: `requiredCount` 超過は警告（エラーではない）

---

## 5. 出欠確認領域

### 出欠確認（AttendanceCollection）

営業日に対するメンバーの出欠を収集する機能。

| 属性 | 説明 |
|------|------|
| collectionID | ULID形式の一意識別子 |
| publicToken | 公開URL用トークン（UUID） |
| targetType | `event` / `business_day` |
| status | `open` / `closed` |
| deadline | 回答締切 |

### 出欠回答（AttendanceResponse）

| 値 | 説明 | 表示 |
|----|------|------|
| `attending` | 出席 | - |
| `absent` | 欠席 | - |
| `maybe` | 未定 | - |

---

## 6. 日程調整領域

### 日程調整（DateSchedule）

複数の候補日を提示し、参加可否を収集してイベント開催日を決定。

| 属性 | 説明 |
|------|------|
| scheduleID | ULID形式の一意識別子 |
| publicToken | 公開URL用トークン |
| status | `open` / `closed` / `decided` |
| decidedDate | 決定した候補日ID |

### 日程可否（DateAvailabilityType）

| 値 | 説明 | 表示 |
|----|------|------|
| `available` | 参加可能 | ○ |
| `unavailable` | 参加不可 | × |
| `maybe` | 未定 | △ |

---

## 7. 監査ログ領域

### 監査ログ（AuditLog）

シフト管理における重要な操作の履歴。

| 属性 | 説明 |
|------|------|
| auditLogID | ULID形式の一意識別子 |
| entityType | 操作対象エンティティ種別 |
| operationType | `CREATE` / `UPDATE` / `DELETE` |
| actorID | 操作者 |
| beforeData / afterData | 変更前後データ（JSON） |

**保持期間**: 1年（設定可能）

---

## エンティティ関連図

```
Tenant
  └── Event
        └── EventBusinessDay
              └── ShiftSlot
                    └── ShiftAssignment ←── Member

Tenant
  └── Member
        └── Role（多対多）

Tenant
  └── AttendanceCollection
        └── AttendanceResponse ←── Member

Tenant
  └── DateSchedule
        └── CandidateDate
              └── DateScheduleResponse ←── Member
```

---

## 詳細ドキュメント

各ドメインの詳細は `docs/domain/` 配下を参照:

- `10_tenant-and-event/` - テナント・イベント・営業日
- `20_member-and-role/` - メンバー・ロール
- `30_shift-frame-and-slot-definition/` - シフト枠・ポジション
- `50_shift-plan-and-assignment/` - シフト割り当て
- `40_attendance-collection/` - 出欠確認
- `45_date-schedule/` - 日程調整
- `55_audit-log/` - 監査ログ

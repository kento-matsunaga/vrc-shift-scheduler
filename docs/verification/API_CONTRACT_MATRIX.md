# API契約（JSON形式）の整合性チェック

## 概要

このドキュメントは、Event、EventBusinessDay、ShiftSlot、ShiftAssignment、Member の5種類のエンティティについて、「DBマイグレーション → Goドメイン構造体 → RESTレスポンス → TypeScript型 → 画面コンポーネント」の間でフィールド名・型・必須/任意が揃っているかを確認したものです。

---

## 1. Event（イベント）

### 1.1 フィールドマトリクス

| フィールド名 | DBカラム (型) | Goドメイン構造体 (型) | RESTレスポンス構造体 (型) | TS型 (型) | 画面での利用有無 | コメント |
|--------------|--------------|------------------------|---------------------------|-----------|------------------|----------|
| id | event_id (CHAR(26)) | EventID (string) | EventResponse.EventID (string) | Event.event_id (string) | 一覧で表示、詳細へのリンク | ✅ OK |
| tenant_id | tenant_id (CHAR(26)) | TenantID (string) | EventResponse.TenantID (string) | Event.tenant_id (string) | 非表示（認証ヘッダーで使用） | ✅ OK |
| name | event_name (VARCHAR(255)) | EventName (string) | EventResponse.EventName (string) | Event.event_name (string) | 一覧で表示 | ✅ OK |
| type | event_type (VARCHAR(20)) | EventType (string) | EventResponse.EventType (string) | Event.event_type ('normal' \| 'special') | 一覧で表示 | ✅ OK |
| description | description (TEXT) | Description (string) | EventResponse.Description (string) | Event.description (string) | 詳細画面（未実装） | ✅ OK |
| is_active | is_active (BOOLEAN) | IsActive (bool) | EventResponse.IsActive (bool) | Event.is_active (boolean) | 非表示（将来のフィルタ用） | ✅ OK |
| created_at | created_at (TIMESTAMPTZ) | CreatedAt (time.Time) | EventResponse.CreatedAt (string) | Event.created_at (string) | 非表示 | ✅ OK |
| updated_at | updated_at (TIMESTAMPTZ) | UpdatedAt (time.Time) | EventResponse.UpdatedAt (string) | Event.updated_at (string) | 非表示 | ✅ OK |
| deleted_at | deleted_at (TIMESTAMPTZ NULL) | DeletedAt (*time.Time) | - (レスポンスに含まれない) | - (型定義にない) | 非表示（論理削除） | ✅ OK |

### 1.2 問題点

**なし** - すべてのフィールドが整合している。

### 1.3 補足

- **日時フォーマット**: RESTレスポンスでは `time.RFC3339` 形式（例: `2006-01-02T15:04:05Z07:00`）で返される
- **論理削除**: `deleted_at` はレスポンスに含まれない（正常）

---

## 2. EventBusinessDay（営業日）

### 2.1 フィールドマトリクス

| フィールド名 | DBカラム (型) | Goドメイン構造体 (型) | RESTレスポンス構造体 (型) | TS型 (型) | 画面での利用有無 | コメント |
|--------------|--------------|------------------------|---------------------------|-----------|------------------|----------|
| id | business_day_id (CHAR(26)) | BusinessDayID (string) | BusinessDayResponse.BusinessDayID (string) | BusinessDay.business_day_id (string) | 一覧で表示、詳細へのリンク | ✅ OK |
| tenant_id | tenant_id (CHAR(26)) | TenantID (string) | BusinessDayResponse.TenantID (string) | BusinessDay.tenant_id (string) | 非表示（認証ヘッダーで使用） | ✅ OK |
| event_id | event_id (CHAR(26)) | EventID (string) | BusinessDayResponse.EventID (string) | BusinessDay.event_id (string) | 一覧で表示 | ✅ OK |
| target_date | target_date (DATE) | TargetDate (time.Time) | BusinessDayResponse.TargetDate (string) | BusinessDay.target_date (string) | 一覧で表示 | ✅ OK |
| start_time | start_time (TIME) | StartTime (time.Time) | BusinessDayResponse.StartTime (string) | BusinessDay.start_time (string) | 一覧で表示 | ⚠️ 要確認 |
| end_time | end_time (TIME) | EndTime (time.Time) | BusinessDayResponse.EndTime (string) | BusinessDay.end_time (string) | 一覧で表示 | ⚠️ 要確認 |
| occurrence_type | occurrence_type (VARCHAR(20)) | OccurrenceType (string) | BusinessDayResponse.OccurrenceType (string) | BusinessDay.occurrence_type ('recurring' \| 'special') | 一覧で表示 | ✅ OK |
| is_active | is_active (BOOLEAN) | IsActive (bool) | BusinessDayResponse.IsActive (bool) | BusinessDay.is_active (boolean) | 非表示（将来のフィルタ用） | ✅ OK |
| created_at | created_at (TIMESTAMPTZ) | CreatedAt (time.Time) | BusinessDayResponse.CreatedAt (string) | BusinessDay.created_at (string) | 非表示 | ✅ OK |
| updated_at | updated_at (TIMESTAMPTZ) | UpdatedAt (time.Time) | - (レスポンスに含まれない) | BusinessDay.updated_at (string) | 非表示 | ⚠️ 不整合 |
| recurring_pattern_id | recurring_pattern_id (CHAR(26) NULL) | RecurringPatternID (*EventID) | - (レスポンスに含まれない) | - (型定義にない) | 非表示 | ✅ OK |
| valid_from | valid_from (DATE NULL) | ValidFrom (*time.Time) | - (レスポンスに含まれない) | - (型定義にない) | 非表示 | ✅ OK |
| valid_to | valid_to (DATE NULL) | ValidTo (*time.Time) | - (レスポンスに含まれない) | - (型定義にない) | 非表示 | ✅ OK |
| deleted_at | deleted_at (TIMESTAMPTZ NULL) | DeletedAt (*time.Time) | - (レスポンスに含まれない) | - (型定義にない) | 非表示（論理削除） | ✅ OK |

### 2.2 問題点

#### 問題1: 時刻フォーマットの不整合

**詳細**:
- **RESTレスポンス**: `15:04` 形式（例: `21:30`）で返される（`business_day_handler.go` 265-266行目）
- **TypeScript型**: `HH:MM:SS` 形式を期待（`api.ts` 37-38行目）
- **実際のレスポンス**: `HH:MM` 形式

**影響**: フロントエンドで時刻をパースする際に、秒部分がないためエラーになる可能性がある。

**修正候補**:
- **オプション1**: RESTレスポンスを `HH:MM:SS` 形式に変更（推奨）
  ```go
  // business_day_handler.go 265-266行目
  StartTime: bd.StartTime().Format("15:04:05"),  // 15:04 から変更
  EndTime:   bd.EndTime().Format("15:04:05"),     // 15:04 から変更
  ```
- **オプション2**: TypeScript型のコメントを `HH:MM` に変更（現状に合わせる）

#### 問題2: updated_at フィールドの不整合

**詳細**:
- **Goドメイン**: `UpdatedAt()` メソッドが存在
- **RESTレスポンス**: `updated_at` が含まれない（`business_day_handler.go` 269行目）
- **TypeScript型**: `updated_at` が定義されている（`api.ts` 42行目）

**影響**: フロントエンドで `updated_at` にアクセスしようとすると `undefined` になる。

**修正候補**:
- **オプション1**: RESTレスポンスに `updated_at` を追加（推奨）
  ```go
  // business_day_handler.go 259-270行目
  return BusinessDayResponse{
      // ... 既存フィールド
      UpdatedAt: bd.UpdatedAt().Format(time.RFC3339),
  }
  ```
- **オプション2**: TypeScript型から `updated_at` を削除（現状に合わせる）

### 2.3 補足

- **日付フォーマット**: `target_date` は `YYYY-MM-DD` 形式で返される（正常）
- **論理削除**: `deleted_at`、`recurring_pattern_id`、`valid_from`、`valid_to` はレスポンスに含まれない（正常）

---

## 3. ShiftSlot（シフト枠）

### 3.1 フィールドマトリクス

| フィールド名 | DBカラム (型) | Goドメイン構造体 (型) | RESTレスポンス構造体 (型) | TS型 (型) | 画面での利用有無 | コメント |
|--------------|--------------|------------------------|---------------------------|-----------|------------------|----------|
| id | slot_id (CHAR(26)) | SlotID (string) | ShiftSlotResponse.SlotID (string) | ShiftSlot.slot_id (string) | 一覧で表示、割り当てへのリンク | ✅ OK |
| tenant_id | tenant_id (CHAR(26)) | TenantID (string) | ShiftSlotResponse.TenantID (string) | ShiftSlot.tenant_id (string) | 非表示（認証ヘッダーで使用） | ✅ OK |
| business_day_id | business_day_id (CHAR(26)) | BusinessDayID (string) | ShiftSlotResponse.BusinessDayID (string) | ShiftSlot.business_day_id (string) | 一覧で表示 | ✅ OK |
| position_id | position_id (CHAR(26)) | PositionID (string) | ShiftSlotResponse.PositionID (string) | ShiftSlot.position_id (string) | 一覧で表示 | ✅ OK |
| slot_name | slot_name (VARCHAR(255)) | SlotName (string) | ShiftSlotResponse.SlotName (string) | ShiftSlot.slot_name (string) | 一覧で表示 | ✅ OK |
| instance_name | instance_name (VARCHAR(255) NULL) | InstanceName (string) | ShiftSlotResponse.InstanceName (string) | ShiftSlot.instance_name (string) | 一覧で表示 | ✅ OK |
| start_time | start_time (TIME) | StartTime (time.Time) | ShiftSlotResponse.StartTime (string) | ShiftSlot.start_time (string) | 一覧で表示 | ✅ OK |
| end_time | end_time (TIME) | EndTime (time.Time) | ShiftSlotResponse.EndTime (string) | ShiftSlot.end_time (string) | 一覧で表示 | ✅ OK |
| required_count | required_count (INT) | RequiredCount (int) | ShiftSlotResponse.RequiredCount (int) | ShiftSlot.required_count (number) | 一覧で表示 | ✅ OK |
| assigned_count | - (JOINで取得) | - (ドメインにない) | ShiftSlotResponse.AssignedCount (int, omitempty) | ShiftSlot.assigned_count? (number) | 一覧で表示（満員チェック） | ⚠️ 実装不足 |
| priority | priority (INT) | Priority (int) | ShiftSlotResponse.Priority (int) | ShiftSlot.priority (number) | 非表示（将来の自動割り当て用） | ✅ OK |
| is_overnight | - (計算値) | IsOvernight() (bool) | ShiftSlotResponse.IsOvernight (bool) | ShiftSlot.is_overnight (boolean) | 非表示 | ✅ OK |
| created_at | created_at (TIMESTAMPTZ) | CreatedAt (time.Time) | ShiftSlotResponse.CreatedAt (string) | ShiftSlot.created_at (string) | 非表示 | ✅ OK |
| updated_at | updated_at (TIMESTAMPTZ) | UpdatedAt (time.Time) | ShiftSlotResponse.UpdatedAt (string) | ShiftSlot.updated_at (string) | 非表示 | ✅ OK |
| deleted_at | deleted_at (TIMESTAMPTZ NULL) | DeletedAt (*time.Time) | - (レスポンスに含まれない) | - (型定義にない) | 非表示（論理削除） | ✅ OK |

### 3.2 問題点

#### 問題1: assigned_count が常に 0 を返す

**詳細**:
- **RESTレスポンス**: `assigned_count` が常に `0` で固定（`shift_slot_handler.go` 231行目）
- **TypeScript型**: `assigned_count?` が定義されている（`api.ts` 61行目）
- **実装**: `shift_slot_handler.go` 215行目に TODO コメントあり

**影響**: フロントエンドで満員チェックができない。

**修正候補**:
- `shift_slot_handler.go` の `GetShiftSlots` と `GetShiftSlotDetail` で、`shift_assignments` テーブルを JOIN して実際の割り当て数を取得する実装が必要。

**実装例**:
```go
// shift_slot_handler.go の GetShiftSlots メソッド内
// TODO: assigned_count を JOIN で取得
// SELECT ss.*, COUNT(sa.assignment_id) as assigned_count
// FROM shift_slots ss
// LEFT JOIN shift_assignments sa ON ss.slot_id = sa.slot_id 
//   AND sa.assignment_status = 'confirmed' AND sa.deleted_at IS NULL
// WHERE ss.business_day_id = $1 AND ss.tenant_id = $2 AND ss.deleted_at IS NULL
// GROUP BY ss.slot_id
```

### 3.3 補足

- **時刻フォーマット**: `start_time` と `end_time` は `HH:MM:SS` 形式で返される（正常）
- **日時フォーマット**: `created_at` と `updated_at` は `RFC3339` 形式で返される（正常）

---

## 4. ShiftAssignment（シフト割り当て）

### 4.1 フィールドマトリクス

| フィールド名 | DBカラム (型) | Goドメイン構造体 (型) | RESTレスポンス構造体 (型) | TS型 (型) | 画面での利用有無 | コメント |
|--------------|--------------|------------------------|---------------------------|-----------|------------------|----------|
| id | assignment_id (CHAR(26)) | AssignmentID (string) | ShiftAssignmentResponse.AssignmentID (string) | ShiftAssignment.assignment_id (string) | 一覧で表示 | ✅ OK |
| tenant_id | tenant_id (CHAR(26)) | TenantID (string) | - (レスポンスに含まれない) | ShiftAssignment.tenant_id (string) | 非表示（認証ヘッダーで使用） | ⚠️ 不整合 |
| plan_id | plan_id (CHAR(26)) | PlanID (string) | - (レスポンスに含まれない) | - (型定義にない) | 非表示（内部管理用） | ✅ OK |
| slot_id | slot_id (CHAR(26)) | SlotID (string) | ShiftAssignmentResponse.SlotID (string) | ShiftAssignment.slot_id (string) | 一覧で表示 | ✅ OK |
| member_id | member_id (CHAR(26)) | MemberID (string) | ShiftAssignmentResponse.MemberID (string) | ShiftAssignment.member_id (string) | 一覧で表示 | ✅ OK |
| member_display_name | - (JOINで取得) | - (ドメインにない) | ShiftAssignmentResponse.MemberDisplayName (string, omitempty) | ShiftAssignment.member_display_name? (string) | 一覧で表示 | ⚠️ 実装不足 |
| slot_name | - (JOINで取得) | - (ドメインにない) | ShiftAssignmentResponse.SlotName (string, omitempty) | ShiftAssignment.slot_name? (string) | 一覧で表示 | ⚠️ 実装不足 |
| target_date | - (JOINで取得) | - (ドメインにない) | ShiftAssignmentResponse.TargetDate (string, omitempty) | ShiftAssignment.target_date? (string) | 一覧で表示 | ⚠️ 実装不足 |
| start_time | - (JOINで取得) | - (ドメインにない) | ShiftAssignmentResponse.StartTime (string, omitempty) | ShiftAssignment.start_time? (string) | 一覧で表示 | ⚠️ 実装不足 |
| end_time | - (JOINで取得) | - (ドメインにない) | ShiftAssignmentResponse.EndTime (string, omitempty) | ShiftAssignment.end_time? (string) | 一覧で表示 | ⚠️ 実装不足 |
| assignment_status | assignment_status (VARCHAR(20)) | AssignmentStatus (string) | ShiftAssignmentResponse.AssignmentStatus (string) | ShiftAssignment.assignment_status ('confirmed' \| 'cancelled') | 一覧で表示 | ✅ OK |
| assignment_method | assignment_method (VARCHAR(20)) | AssignmentMethod (string) | ShiftAssignmentResponse.AssignmentMethod (string) | ShiftAssignment.assignment_method ('auto' \| 'manual') | 一覧で表示 | ✅ OK |
| is_outside_preference | is_outside_preference (BOOLEAN) | IsOutsidePreference (bool) | - (レスポンスに含まれない) | ShiftAssignment.is_outside_preference (boolean) | 非表示 | ⚠️ 不整合 |
| assigned_at | assigned_at (TIMESTAMPTZ) | AssignedAt (time.Time) | ShiftAssignmentResponse.AssignedAt (string) | ShiftAssignment.assigned_at (string) | 一覧で表示 | ✅ OK |
| cancelled_at | cancelled_at (TIMESTAMPTZ NULL) | CancelledAt (*time.Time) | - (レスポンスに含まれない) | ShiftAssignment.cancelled_at? (string) | 一覧で表示 | ⚠️ 不整合 |
| created_at | created_at (TIMESTAMPTZ) | CreatedAt (time.Time) | - (レスポンスに含まれない) | ShiftAssignment.created_at (string) | 非表示 | ⚠️ 不整合 |
| updated_at | updated_at (TIMESTAMPTZ) | UpdatedAt (time.Time) | - (レスポンスに含まれない) | ShiftAssignment.updated_at (string) | 非表示 | ⚠️ 不整合 |
| deleted_at | deleted_at (TIMESTAMPTZ NULL) | DeletedAt (*time.Time) | - (レスポンスに含まれない) | - (型定義にない) | 非表示（論理削除） | ✅ OK |
| notification_sent | - (計算値) | - (ドメインにない) | ShiftAssignmentResponse.NotificationSent (bool) | - (型定義にない) | 非表示 | ⚠️ 不整合 |

### 4.2 問題点

#### 問題1: JOIN フィールドが実装されていない

**詳細**:
- **TypeScript型**: `member_display_name`、`slot_name`、`target_date`、`start_time`、`end_time` がオプショナルで定義されている（`api.ts` 79-83行目）
- **RESTレスポンス**: これらのフィールドが `omitempty` で定義されているが、実際には返されない（`shift_assignment_handler.go` 36-50行目）

**影響**: フロントエンドで「自分のシフト一覧」を表示する際に、シフトの詳細情報（日付、時刻、役職名など）が表示できない。

**修正候補**:
- `GetAssignments` メソッドで、`members`、`shift_slots`、`event_business_days` テーブルを JOIN して必要な情報を取得する実装が必要。

#### 問題2: フィールドの不整合

**詳細**:
- **tenant_id**: TypeScript型に定義されているが、RESTレスポンスに含まれない
- **is_outside_preference**: TypeScript型に定義されているが、RESTレスポンスに含まれない
- **cancelled_at**: TypeScript型に定義されているが、RESTレスポンスに含まれない（`cancelled` の場合のみ必要）
- **created_at / updated_at**: TypeScript型に定義されているが、RESTレスポンスに含まれない
- **notification_sent**: RESTレスポンスに含まれるが、TypeScript型に定義されていない

**影響**: フロントエンドでこれらのフィールドにアクセスしようとすると、`undefined` になる、または型エラーが発生する。

**修正候補**:
- **オプション1**: RESTレスポンスに不足しているフィールドを追加（推奨）
  ```go
  // shift_assignment_handler.go の GetAssignments メソッド内
  assignments = append(assignments, ShiftAssignmentResponse{
      AssignmentID:        a.AssignmentID().String(),
      TenantID:            a.TenantID().String(),  // 追加
      SlotID:              a.SlotID().String(),
      MemberID:            a.MemberID().String(),
      AssignmentStatus:    map[bool]string{true: "cancelled", false: "confirmed"}[a.IsCancelled()],
      AssignmentMethod:    "manual",
      IsOutsidePreference: a.IsOutsidePreference(),  // 追加
      AssignedAt:          a.AssignedAt().Format("2006-01-02T15:04:05Z07:00"),
      CancelledAt:         func() *string {  // 追加
          if a.CancelledAt() != nil {
              s := a.CancelledAt().Format("2006-01-02T15:04:05Z07:00")
              return &s
          }
          return nil
      }(),
      CreatedAt:           a.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),  // 追加
      UpdatedAt:            a.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),  // 追加
  })
  ```
- **オプション2**: TypeScript型から不足しているフィールドを削除（現状に合わせる）

#### 問題3: 日付範囲フィルタが未実装

**詳細**:
- **フロントエンド**: `MyShifts.tsx` で `start_date` と `end_date` パラメータを送信している
- **バックエンド**: `shift_assignment_handler.go` 160行目に TODO コメントあり

**影響**: 「今後のシフト」「過去のシフト」のフィルタリングが正しく動作しない。

**修正候補**:
- `GetAssignments` メソッドで、`start_date` と `end_date` パラメータを受け取り、`event_business_days` テーブルと JOIN して日付範囲でフィルタリングする実装が必要。

### 4.3 補足

- **日時フォーマット**: `assigned_at` は `RFC3339` 形式で返される（正常）
- **論理削除**: `deleted_at` と `plan_id` はレスポンスに含まれない（正常）

---

## 5. Member（メンバー）

### 5.1 フィールドマトリクス

| フィールド名 | DBカラム (型) | Goドメイン構造体 (型) | RESTレスポンス構造体 (型) | TS型 (型) | 画面での利用有無 | コメント |
|--------------|--------------|------------------------|---------------------------|-----------|------------------|----------|
| id | member_id (CHAR(26)) | MemberID (string) | MemberResponse.MemberID (string) | Member.member_id (string) | 一覧で表示、選択 | ✅ OK |
| tenant_id | tenant_id (CHAR(26)) | TenantID (string) | MemberResponse.TenantID (string) | Member.tenant_id (string) | 非表示（認証ヘッダーで使用） | ✅ OK |
| display_name | display_name (VARCHAR(255)) | DisplayName (string) | MemberResponse.DisplayName (string) | Member.display_name (string) | 一覧で表示、選択 | ✅ OK |
| discord_user_id | discord_user_id (VARCHAR(100) NULL) | DiscordUserID (string) | MemberResponse.DiscordUserID (string, omitempty) | Member.discord_user_id? (string) | 非表示（将来のDiscord連携用） | ✅ OK |
| email | email (VARCHAR(255) NULL) | Email (string) | MemberResponse.Email (string, omitempty) | Member.email? (string) | 非表示（将来の通知用） | ✅ OK |
| is_active | is_active (BOOLEAN) | IsActive (bool) | MemberResponse.IsActive (bool) | Member.is_active (boolean) | 非表示（将来のフィルタ用） | ✅ OK |
| created_at | created_at (TIMESTAMPTZ) | CreatedAt (time.Time) | MemberResponse.CreatedAt (string) | Member.created_at (string) | 非表示 | ✅ OK |
| updated_at | updated_at (TIMESTAMPTZ) | UpdatedAt (time.Time) | MemberResponse.UpdatedAt (string) | Member.updated_at (string) | 非表示 | ✅ OK |
| deleted_at | deleted_at (TIMESTAMPTZ NULL) | DeletedAt (*time.Time) | - (レスポンスに含まれない) | - (型定義にない) | 非表示（論理削除） | ✅ OK |

### 5.2 問題点

**なし** - すべてのフィールドが整合している。

### 5.3 補足

- **日時フォーマット**: `created_at` と `updated_at` は `RFC3339` 形式で返される（正常）
- **論理削除**: `deleted_at` はレスポンスに含まれない（正常）

---

## 6. まとめ

### 6.1 重大な問題（修正必須）

1. **ShiftSlot.assigned_count が常に 0**
   - 影響: 満員チェックができない
   - 修正: `shift_slot_handler.go` で JOIN クエリを実装

2. **ShiftAssignment の JOIN フィールドが未実装**
   - 影響: 「自分のシフト一覧」で詳細情報が表示できない
   - 修正: `shift_assignment_handler.go` で JOIN クエリを実装

3. **ShiftAssignment の日付範囲フィルタが未実装**
   - 影響: 「今後のシフト」「過去のシフト」のフィルタリングが動作しない
   - 修正: `shift_assignment_handler.go` で日付範囲フィルタを実装

### 6.2 軽微な問題（修正推奨）

1. **BusinessDay.start_time / end_time のフォーマット不整合**
   - 影響: フロントエンドで時刻をパースする際にエラーになる可能性
   - 修正: RESTレスポンスを `HH:MM:SS` 形式に変更、または TypeScript型のコメントを修正

2. **BusinessDay.updated_at がレスポンスに含まれない**
   - 影響: フロントエンドで `updated_at` にアクセスすると `undefined` になる
   - 修正: RESTレスポンスに `updated_at` を追加、または TypeScript型から削除

3. **ShiftAssignment のフィールド不整合**
   - 影響: フロントエンドで一部のフィールドにアクセスできない、または型エラーが発生
   - 修正: RESTレスポンスに不足しているフィールドを追加、または TypeScript型から削除

### 6.3 修正優先度

1. **高**: ShiftSlot.assigned_count、ShiftAssignment の JOIN フィールド、日付範囲フィルタ
2. **中**: BusinessDay の時刻フォーマット、updated_at
3. **低**: ShiftAssignment のその他のフィールド不整合

---

**作成日**: 2025-01-XX  
**作成者**: 検証専用アシスタント（Auto）




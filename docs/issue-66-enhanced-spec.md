# Issue #66 詳細仕様書 - 本出勤一括インポート機能

## 調査結果サマリ

### DB スキーマ確認結果

#### shift_assignments テーブル
```sql
CREATE TABLE shift_assignments (
    assignment_id CHAR(26) PRIMARY KEY,
    tenant_id CHAR(26) NOT NULL,
    plan_id CHAR(26) NULL,              -- Migration 016 で NULL許可に変更済み
    slot_id CHAR(26) NOT NULL,          -- shift_slots への FK（必須）
    member_id CHAR(26) NOT NULL,        -- members への FK（必須）
    assignment_status VARCHAR(20) NOT NULL DEFAULT 'confirmed',
    assignment_method VARCHAR(20) NOT NULL DEFAULT 'manual',
    ...
);
```

**重要**: `plan_id` は NULL 許可（手動割り当て対応）。インポート時は NULL で作成可能。

#### shift_slots テーブル
```sql
CREATE TABLE shift_slots (
    slot_id CHAR(26) PRIMARY KEY,
    tenant_id CHAR(26) NOT NULL,
    business_day_id CHAR(26) NOT NULL,  -- event_business_days への FK
    position_id CHAR(26) NOT NULL,      -- positions への FK（必須！）
    slot_name VARCHAR(255) NOT NULL,
    instance_name VARCHAR(255) NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    required_count INT NOT NULL DEFAULT 1,
    ...
);
```

**制約**: `position_id` は NOT NULL。シフト枠を作成する場合は必ずポジションが必要。

#### positions テーブル
```sql
CREATE TABLE positions (
    position_id CHAR(26) PRIMARY KEY,
    tenant_id CHAR(26) NOT NULL,
    position_name VARCHAR(255) NOT NULL,
    description TEXT,
    display_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    ...
);
-- テナント内で position_name は一意
```

---

## 仕様修正・補足

### CSV カラム定義（修正版）

| カラム名 | 必須 | 型 | 説明 | 例 |
|---------|------|-----|------|-----|
| `date` | ○ | string | 営業日（YYYY-MM-DD） | `2025-01-15` |
| `member_name` | ○ | string | メンバー表示名 | `たろう` |
| `event_name` | △ | string | イベント名（同一日に複数イベントがある場合必須） | `週末イベント` |
| `slot_name` | × | string | シフト枠名（省略時はデフォルト枠を使用/作成） | `受付` |
| `position_name` | △ | string | ポジション名（`create_missing_slots=true` 時は必須） | `スタッフ` |
| `start_time` | × | string | 開始時刻（HH:MM）（`create_missing_slots=true` 時は必須） | `20:00` |
| `end_time` | × | string | 終了時刻（HH:MM）（`create_missing_slots=true` 時は必須） | `22:00` |
| `note` | × | string | 備考 | `代打参加` |

### 必須条件の整理

1. **必ず必要**: `date`, `member_name`
2. **条件付き必須**:
   - `event_name`: 同一日に複数イベントがある場合、または `default_event_id` 未指定時
   - `position_name`: `create_missing_slots=true` かつ該当スロットが存在しない場合
   - `start_time`, `end_time`: `create_missing_slots=true` かつ該当スロットが存在しない場合

### ActualAttendanceRow 構造体（修正版）

```go
// csv_parser.go に追加・修正
type ActualAttendanceRow struct {
    RowNumber    int
    Date         string  // YYYY-MM-DD（必須）
    MemberName   string  // メンバー表示名（必須）
    EventName    string  // イベント名（オプション）
    SlotName     string  // シフト枠名（オプション）
    PositionName string  // ポジション名（オプション）← 追加
    StartTime    string  // HH:MM（オプション）
    EndTime      string  // HH:MM（オプション）
    Note         string  // 備考
}
```

---

## 処理フロー詳細

### Phase 1: 初期処理

```
1. ImportJob作成・開始
2. CSVパース（10,000行上限チェック）
3. CSV Injection対策（sanitizeCSVValue適用）
4. 行数バリデーション
```

### Phase 2: データ一括取得（N+1対策）

```go
// 1回のクエリで必要データをすべて取得
members := memberRepo.FindByTenantID(ctx, tenantID)          // 全メンバー
positions := positionRepo.FindByTenantID(ctx, tenantID)      // 全ポジション
businessDays := businessDayRepo.FindByTenantIDAndDateRange(ctx, tenantID, minDate, maxDate)
events := eventRepo.FindByTenantID(ctx, tenantID)            // 全イベント

// メモリ上でルックアップマップを構築
memberByDisplayName := buildMemberMap(members)
positionByName := buildPositionMap(positions)
businessDayByDateAndEvent := buildBusinessDayMap(businessDays)
```

### Phase 3: 行ごとの処理

```
FOR each row in CSV:
    1. メンバー検索
       - memberByDisplayName から検索
       - fuzzy_match=true の場合は曖昧検索も実行
       - 見つからない場合 → エラー記録、次の行へ

    2. 営業日検索
       - date + event_name で businessDayByDateAndEvent から検索
       - event_name が空の場合:
         - default_event_id があれば使用
         - なければその日の唯一のイベントを使用（複数あればエラー）
       - 見つからない場合:
         - create_missing_business_days=true → 営業日作成
         - それ以外 → エラー記録、次の行へ

    3. シフト枠検索
       - business_day_id + slot_name で検索
       - slot_name が空の場合:
         - その営業日のデフォルトシフト枠を使用
         - デフォルト枠がない場合 → 「通常枠」として作成（create_missing_slots=true時）
       - 見つからない場合:
         - create_missing_slots=true の場合:
           - position_name が必須（未指定ならエラー）
           - start_time, end_time が必須（未指定ならエラー）
           - 新規シフト枠作成
         - それ以外 → エラー記録、次の行へ

    4. 重複チェック
       - slot_id + member_id で既存の confirmed 割り当てを検索
       - 重複あり:
         - skip_existing=true → スキップカウント++、次の行へ
         - update_existing=true → 既存レコード更新
         - それ以外 → エラー記録、次の行へ

    5. ShiftAssignment作成
       - plan_id = NULL（手動割り当て）
       - assignment_status = 'confirmed'
       - assignment_method = 'manual'
       - is_outside_preference = false

    6. 成功カウント++
```

### Phase 4: 完了処理

```
1. ImportJob完了（ステータス更新）
2. 結果レスポンス返却
```

---

## エラーハンドリング詳細

| エラー種別 | 条件 | 対応 | メッセージ例 |
|-----------|------|------|-------------|
| メンバー未存在 | 検索結果なし | エラー記録、スキップ | `行3: メンバー 'さとう' が見つかりません` |
| 複数イベント該当 | 同日に複数イベント & event_name未指定 | エラー記録、スキップ | `行5: 2025-01-15 に複数のイベントがあります。event_name を指定してください` |
| 営業日未存在 | 検索結果なし & create_missing_business_days=false | エラー記録、スキップ | `行7: 日付 '2025-01-20' の営業日が見つかりません` |
| シフト枠未存在 | 検索結果なし & create_missing_slots=false | エラー記録、スキップ | `行9: シフト枠 '受付' が見つかりません` |
| ポジション未指定 | create_missing_slots=true & position_name空 | エラー記録、スキップ | `行11: シフト枠作成には position_name が必要です` |
| ポジション未存在 | position_name指定あり & 検索結果なし | エラー記録、スキップ | `行13: ポジション 'スタッフ' が見つかりません` |
| 時刻未指定 | create_missing_slots=true & start/end_time空 | エラー記録、スキップ | `行15: シフト枠作成には start_time, end_time が必要です` |
| 日付形式エラー | YYYY-MM-DD 形式でない | エラー記録、スキップ | `行17: 日付形式が不正です（YYYY-MM-DD形式で入力）` |
| 時刻形式エラー | HH:MM 形式でない | エラー記録、スキップ | `行19: 時刻形式が不正です（HH:MM形式で入力）` |
| 重複割り当て | skip_existing=false & update_existing=false | エラー記録、スキップ | `行21: 'たろう' は既に 2025-01-15 に割り当て済みです` |
| 行数超過 | 10,000行超 | インポート中止 | `行数が上限を超えています: 15000行 (上限: 10000行)` |

---

## リポジトリ追加メソッド

### EventBusinessDayRepository
```go
// 日付範囲で営業日を一括取得
FindByTenantIDAndDateRange(ctx context.Context, tenantID common.TenantID, startDate, endDate time.Time) ([]*EventBusinessDay, error)
```

### ShiftSlotRepository
```go
// 営業日IDとスロット名で検索
FindByBusinessDayIDAndSlotName(ctx context.Context, businessDayID event.BusinessDayID, slotName string) (*ShiftSlot, error)
```

### ShiftAssignmentRepository
```go
// スロットIDとメンバーIDで確定済み割り当てを検索
FindConfirmedBySlotIDAndMemberID(ctx context.Context, slotID SlotID, memberID common.MemberID) (*ShiftAssignment, error)
```

---

## ImportOptions 拡張

```go
type ImportOptions struct {
    SkipExisting              bool   `json:"skip_existing"`
    UpdateExisting            bool   `json:"update_existing"`
    FuzzyMemberMatch          bool   `json:"fuzzy_match"`
    CreateMissingSlots        bool   `json:"create_missing_slots"`
    CreateMissingBusinessDays bool   `json:"create_missing_business_days"`
    DefaultEventID            string `json:"default_event_id"`
}
```

---

## セキュリティ考慮事項

1. **テナント分離**: すべてのクエリで tenant_id スコープ必須
2. **CSV Injection対策**: sanitizeCSVValue() で =, +, -, @ プレフィックス処理
3. **行数制限**: 10,000行上限（DoS対策）
4. **認可チェック**: AdminID 必須、ImportJob に created_by 記録

---

## テストケース

### 正常系
1. 最小構成（date, member_name のみ）でインポート
2. 全カラム指定でインポート
3. fuzzy_match=true でひらがな/カタカナ混在名前をマッチ
4. skip_existing=true で重複スキップ
5. create_missing_slots=true でシフト枠自動作成

### 異常系
1. 存在しないメンバー名
2. 存在しないポジション名
3. 同一日複数イベントでevent_name未指定
4. create_missing_slots=true だが position_name 未指定
5. 日付形式エラー
6. 10,001行のCSV

---

## 実装ファイル一覧

### Domain層
- `internal/domain/import/csv_parser.go` - ActualAttendanceRow に position_name 追加
- `internal/domain/import/import_job.go` - ImportType に "actual_attendance" 追加
- `internal/domain/import/attendance_import_service.go` - 新規作成

### Application層
- `internal/app/import/import_actual_attendance_usecase.go` - 新規作成

### Infrastructure層
- `internal/infra/db/shift_slot_repository.go` - FindByBusinessDayIDAndSlotName 追加
- `internal/infra/db/shift_assignment_repository.go` - FindConfirmedBySlotIDAndMemberID 追加
- `internal/infra/db/business_day_repository.go` - FindByTenantIDAndDateRange 追加

### Interface層
- `internal/interface/rest/import_handler.go` - ImportActualAttendance ハンドラー追加
- `internal/interface/rest/router.go` - POST /api/v1/imports/actual-attendance ルート追加

### Frontend
- `web-frontend/src/lib/api/importApi.ts` - importActualAttendanceFromCSV 関数追加
- `web-frontend/src/components/BulkImport.tsx` - 本出勤タブ追加

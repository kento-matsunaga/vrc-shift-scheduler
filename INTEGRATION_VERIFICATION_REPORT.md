# VRC Shift Scheduler - 統合検証レポート

**作成日**: 2025-01-XX  
**作成者**: テストエンジニアAI（Auto）

---

## 1. 今のコードで動くはずの機能（まとめ）

SUMMARY.mdと実際のコードを確認した結果、以下の機能が現時点で動作可能です（Auth V2は未実装のため除外）：

1. **ログイン機能（簡易認証）**
   - 表示名による簡易ログイン
   - テナントIDの取得（URLパラメータ or 環境変数）
   - メンバーの自動登録（POST /api/v1/members）

2. **イベント管理**
   - イベント一覧表示（GET /api/v1/events）
   - イベント作成（POST /api/v1/events）
   - イベント詳細取得（GET /api/v1/events/:event_id）

3. **営業日管理**
   - 営業日一覧表示（GET /api/v1/events/:event_id/business-days）
   - 営業日作成（POST /api/v1/events/:event_id/business-days）
   - 営業日詳細取得（GET /api/v1/business-days/:business_day_id）

4. **シフト枠管理**
   - シフト枠一覧表示（GET /api/v1/business-days/:business_day_id/shift-slots）
   - シフト枠作成（POST /api/v1/business-days/:business_day_id/shift-slots）
   - シフト枠詳細取得（GET /api/v1/shift-slots/:slot_id）

5. **シフト割り当て**
   - シフト割り当て確定（POST /api/v1/shift-assignments）
   - シフト割り当て一覧取得（GET /api/v1/shift-assignments?member_id=...）
   - シフト割り当て詳細取得（GET /api/v1/shift-assignments/:assignment_id）

6. **自分のシフト一覧**
   - シフト一覧表示（GET /api/v1/shift-assignments?member_id=...&assignment_status=confirmed）

7. **メンバー管理**
   - メンバー一覧取得（GET /api/v1/members）
   - メンバー詳細取得（GET /api/v1/members/:member_id）

---

## 2. ポート・環境変数の現状と推奨設定

### 現状

#### バックエンド
- **実コード上のデフォルトポート**: `8080`（`backend/cmd/server/main.go` 24-27行目）
- **docker-compose.yml のホストポート**: `8090:8080`（29行目）
- **docker-compose.prod.yml のホストポート**: `${BACKEND_PORT:-8080}:8080`（47行目）

#### フロントエンド
- **apiClient.ts のデフォルト baseURL**: `http://localhost:8080`（10行目）
- **.env で上書き可能かどうか**: `VITE_API_BASE_URL` 環境変数で上書き可能
- **.env ファイルの存在**: 存在しない（`.env.local` や `.env.example` も未作成）

### 問題になりそうな点

1. **開発環境でのポート不整合**
   - `docker-compose.yml` ではバックエンドがホストの `8090` ポートにマッピングされている
   - フロントエンドのデフォルトは `http://localhost:8080` を参照
   - **影響**: Docker Compose で起動した場合、フロントエンドからバックエンドに接続できない

2. **環境変数ファイルの未作成**
   - フロントエンド側で `.env.local` や `.env.example` が存在しない
   - **影響**: 開発環境ごとに異なるポート設定に対応できない

### 推奨

**開発環境（ローカル開発モード）**:
- バックエンド: `:8080`（デフォルト）
- フロントエンド: `:5173`（Vite デフォルト）
- フロントからは `http://localhost:8080` を叩く

**推奨修正**:
1. `docker-compose.yml` の29行目を `8090:8080` から `8080:8080` に変更
2. `web-frontend/.env.local.example` を作成して、開発者向けの設定例を提供

---

## 3. API 契約の差分一覧

### 1) GET /api/v1/events

**フロント期待フィールド**（`web-frontend/src/types/api.ts` 15-24行目）:
- `event_id: string`
- `tenant_id: string`
- `event_name: string`
- `event_type: 'normal' | 'special'`
- `description: string`
- `is_active: boolean`
- `created_at: string`
- `updated_at: string`

**バックエンド実フィールド**（`backend/internal/interface/rest/event_handler.go` 35-44行目）:
- `event_id: string`
- `tenant_id: string`
- `event_name: string`
- `event_type: string`
- `description: string`
- `is_active: bool`
- `created_at: string`
- `updated_at: string`

**差分**: ✅ **整合している** - すべてのフィールドが一致

---

### 2) GET /api/v1/events/:event_id/business-days

**フロント期待フィールド**（`web-frontend/src/types/api.ts` 32-43行目）:
- `business_day_id: string`
- `tenant_id: string`
- `event_id: string`
- `target_date: string` // YYYY-MM-DD
- `start_time: string` // HH:MM:SS（コメント）
- `end_time: string` // HH:MM:SS（コメント）
- `occurrence_type: 'recurring' | 'special'`
- `is_active: boolean`
- `created_at: string`
- `updated_at: string`

**バックエンド実フィールド**（`backend/internal/interface/rest/business_day_handler.go` 38-48行目、259-270行目）:
- `business_day_id: string`
- `tenant_id: string`
- `event_id: string`
- `target_date: string` // YYYY-MM-DD
- `start_time: string` // **HH:MM** 形式（265行目: `Format("15:04")`）
- `end_time: string` // **HH:MM** 形式（266行目: `Format("15:04")`）
- `occurrence_type: string`
- `is_active: bool`
- `created_at: string`
- `updated_at: string` // **レスポンスに含まれない**（48行目の構造体定義にない）

**差分**:
- ⚠️ **優先度 中**: `start_time` と `end_time` のフォーマット不一致
  - フロントは `HH:MM:SS` を期待（コメント）
  - バックエンドは `HH:MM` を返す
  - **影響**: フロントエンドで時刻をパースする際にエラーになる可能性
- ⚠️ **優先度 低**: `updated_at` がレスポンスに含まれない
  - フロントの型定義にはあるが、バックエンドのレスポンス構造体にない
  - **影響**: フロントエンドで `updated_at` にアクセスすると `undefined` になる

---

### 3) GET /api/v1/business-days/:business_day_id/shift-slots

**フロント期待フィールド**（`web-frontend/src/types/api.ts` 51-66行目）:
- `slot_id: string`
- `tenant_id: string`
- `business_day_id: string`
- `position_id: string`
- `slot_name: string`
- `instance_name: string`
- `start_time: string` // HH:MM:SS
- `end_time: string` // HH:MM:SS
- `required_count: number`
- `assigned_count?: number` // オプショナル
- `priority: number`
- `is_overnight: boolean`
- `created_at: string`
- `updated_at: string`

**バックエンド実フィールド**（`backend/internal/interface/rest/shift_slot_handler.go` 41-56行目、221-236行目）:
- `slot_id: string`
- `tenant_id: string`
- `business_day_id: string`
- `position_id: string`
- `slot_name: string`
- `instance_name: string`
- `start_time: string` // HH:MM:SS（229行目: `Format("15:04:05")`）
- `end_time: string` // HH:MM:SS（230行目: `Format("15:04:05")`）
- `required_count: int`
- `assigned_count: int` // **常に 0 で固定**（231行目）
- `priority: int`
- `is_overnight: bool`
- `created_at: string`
- `updated_at: string`

**差分**:
- ⚠️ **優先度 高**: `assigned_count` が常に `0` を返す
  - バックエンドの215行目に TODO コメントあり
  - **影響**: 満員チェックができない、シフト枠の充足状況が分からない

---

### 4) GET /api/v1/shift-assignments?member_id=...

**フロント期待フィールド**（`web-frontend/src/types/api.ts` 74-91行目）:
- `assignment_id: string`
- `tenant_id: string`
- `slot_id: string`
- `member_id: string`
- `member_display_name?: string` // JOIN で取得する場合
- `slot_name?: string` // JOIN で取得する場合
- `target_date?: string` // JOIN で取得する場合
- `start_time?: string` // JOIN で取得する場合
- `end_time?: string` // JOIN で取得する場合
- `assignment_status: 'confirmed' | 'cancelled'`
- `assignment_method: 'auto' | 'manual'`
- `is_outside_preference: boolean`
- `assigned_at: string`
- `cancelled_at?: string`
- `created_at: string`
- `updated_at: string`

**バックエンド実フィールド**（`backend/internal/interface/rest/shift_assignment_handler.go` 37-50行目、178-188行目、204-214行目）:
- `assignment_id: string`
- `slot_id: string`
- `member_id: string`
- `member_display_name: string` // **omitempty だが実際には返されない**
- `slot_name: string` // **omitempty だが実際には返されない**
- `target_date: string` // **omitempty だが実際には返されない**
- `start_time: string` // **omitempty だが実際には返されない**
- `end_time: string` // **omitempty だが実際には返されない**
- `assignment_status: string`
- `assignment_method: string`
- `assigned_at: string`
- `notification_sent: bool` // **フロントの型定義にない**
- `tenant_id: string` // **レスポンスに含まれない**
- `is_outside_preference: bool` // **レスポンスに含まれない**
- `cancelled_at: string` // **レスポンスに含まれない**
- `created_at: string` // **レスポンスに含まれない**
- `updated_at: string` // **レスポンスに含まれない**

**差分**:
- ⚠️ **優先度 高**: JOIN フィールドが未実装
  - `member_display_name`、`slot_name`、`target_date`、`start_time`、`end_time` がレスポンスに含まれない
  - **影響**: 「自分のシフト一覧」で詳細情報が表示できない
- ⚠️ **優先度 中**: フィールドの不整合
  - `tenant_id`、`is_outside_preference`、`cancelled_at`、`created_at`、`updated_at` がレスポンスに含まれない
  - `notification_sent` がレスポンスに含まれるが、フロントの型定義にない
  - **影響**: フロントエンドでこれらのフィールドにアクセスできない、または型エラーが発生
- ⚠️ **優先度 高**: 日付範囲フィルタが未実装
  - `start_date`、`end_date` パラメータが未実装（160行目のTODOコメント）
  - **影響**: 「今後のシフト」「過去のシフト」のフィルタリングが正しく動作しない

---

### 5) GET /api/v1/members

**フロント期待フィールド**（`web-frontend/src/types/api.ts` 99-108行目）:
- `member_id: string`
- `tenant_id: string`
- `display_name: string`
- `discord_user_id?: string`
- `email?: string`
- `is_active: boolean`
- `created_at: string`
- `updated_at: string`

**バックエンド実フィールド**（`backend/internal/interface/rest/member_handler.go` 34-43行目、185-194行目）:
- `member_id: string`
- `tenant_id: string`
- `display_name: string`
- `discord_user_id: string` // omitempty
- `email: string` // omitempty
- `is_active: bool`
- `created_at: string`
- `updated_at: string`

**差分**: ✅ **整合している** - すべてのフィールドが一致

---

## 4. バックエンド統合テスト候補

BACKEND_INTEGRATION_TEST_PLAN.md を確認した結果、以下の統合テストシナリオを推奨します：

### シナリオ1: Happy Path - シフト割り当てフロー

**前提**:
- テナント1が存在
- メンバーが1人以上存在
- ポジションが1つ以上存在

**手順（API シーケンス）**:
1. POST /api/v1/events でイベント作成
2. POST /api/v1/events/{event_id}/business-days で営業日作成
3. POST /api/v1/business-days/{business_day_id}/shift-slots でシフト枠作成
4. POST /api/v1/shift-assignments で割り当て
5. GET /api/v1/shift-assignments?member_id=... で反映を確認

**検証ポイント**:
- 各 API が 2xx を返すこと
- DB の `events`、`event_business_days`、`shift_slots`、`shift_assignments` テーブルにレコードが増えていること
- JSON のフィールドが API_CONTRACT_MATRIX.md と一致していること

**既存テストコード**: なし（`backend/internal/interface/rest/` 配下に `_test.go` ファイルが存在しない）

---

### シナリオ2: 満員時の割り当てで 409 が返るか

**前提**:
- シナリオ1で作成したシフト枠が存在
- `required_count` が `2` で、既に `2` 人割り当てられている

**手順（API シーケンス）**:
1. POST /api/v1/shift-assignments で3人目の割り当てを試みる

**検証ポイント**:
- HTTP ステータスコードが `409 Conflict` であること
- エラーレスポンスの `error.code` が `ERR_SLOT_FULL` であること

**注意**: 現状、`assigned_count` が実装されていないため、このテストは実行できない可能性がある

---

### シナリオ3: テナント境界の確認

**前提**:
- テナント1とテナント2が存在
- テナント1にイベントが1つ存在

**手順（API シーケンス）**:
1. テナント1のヘッダーで GET /api/v1/events を実行
2. テナント2のヘッダーで GET /api/v1/events を実行

**検証ポイント**:
- テナント1のリクエストではイベントが返る
- テナント2のリクエストではイベントが返らない（空配列）

---

### シナリオ4: バリデーションエラーの確認

**前提**:
- テナント1が存在

**手順（API シーケンス）**:
1. POST /api/v1/events で `event_name` が空のリクエストを送信

**検証ポイント**:
- HTTP ステータスコードが `400 Bad Request` であること
- エラーレスポンスの `error.code` が `ERR_INVALID_REQUEST` であること

---

### シナリオ5: 重複チェックの確認

**前提**:
- テナント1が存在
- イベント名 "Test Event" が既に存在

**手順（API シーケンス）**:
1. POST /api/v1/events で同じ名前のイベントを作成しようとする

**検証ポイント**:
- HTTP ステータスコードが `409 Conflict` であること
- エラーレスポンスの `error.code` が `ERR_CONFLICT` であること

---

## 5. 手動E2Eチェックリストの修正提案

MANUAL_E2E_CHECKLIST.md と実際のルーティング（`web-frontend/src/App.tsx`）を照合した結果、以下の修正が必要です：

### 修正不要（整合している）

- ✅ `/login` - ログイン画面
- ✅ `/events` - イベント一覧画面
- ✅ `/events/:eventId/business-days` - 営業日一覧画面
- ✅ `/business-days/:businessDayId/shift-slots` - シフト枠一覧画面
- ✅ `/shift-slots/:slotId/assign` - シフト割り当て画面
- ✅ `/my-shifts` - 自分のシフト一覧画面

### 修正提案

**なし** - すべてのルーティングが整合している。

**補足**: MANUAL_E2E_CHECKLIST.md の記述は実際のルーティングと一致しています。ただし、以下の注意点があります：

1. **ステップ2（イベント一覧画面）**: ドキュメントでは「新規作成」ボタンと記載されているが、実際のコードでは「新規作成」ボタンが存在することを確認済み
2. **ステップ5（シフト割り当て画面）**: ドキュメントでは「割り当て」ボタンと記載されているが、実際のコードでは「割り当て」ボタンが存在することを確認済み

---

## 6. すぐ直すべきポイント（優先度付き TODO）

### 優先度 高：今のままだとテスターが確実にハマる or 画面が動かないもの

1. **ShiftSlot.assigned_count が常に 0 を返す**
   - **ファイル**: `backend/internal/interface/rest/shift_slot_handler.go`
   - **箇所**: 215-231行目（GetShiftSlots）、280-293行目（GetShiftSlotDetail）
   - **修正内容**: `shift_assignments` テーブルを JOIN して実際の割り当て数を取得
   - **影響**: 満員チェックができない、シフト枠の充足状況が分からない

2. **ShiftAssignment の JOIN フィールドが未実装**
   - **ファイル**: `backend/internal/interface/rest/shift_assignment_handler.go`
   - **箇所**: 143-226行目（GetAssignments）
   - **修正内容**: `members`、`shift_slots`、`event_business_days` テーブルを JOIN して必要な情報を取得
   - **影響**: 「自分のシフト一覧」で詳細情報が表示できない

3. **ShiftAssignment の日付範囲フィルタが未実装**
   - **ファイル**: `backend/internal/interface/rest/shift_assignment_handler.go`
   - **箇所**: 160行目のTODOコメント
   - **修正内容**: `start_date`、`end_date` パラメータを受け取り、`event_business_days` テーブルと JOIN して日付範囲でフィルタリング
   - **影響**: 「今後のシフト」「過去のシフト」のフィルタリングが正しく動作しない

4. **docker-compose.yml のポート不整合**
   - **ファイル**: `docker-compose.yml`
   - **箇所**: 29行目
   - **修正内容**: `8090:8080` を `8080:8080` に変更
   - **影響**: Docker Compose で起動した場合、フロントエンドからバックエンドに接続できない

---

### 優先度 中：Public Alpha までには直したいもの

5. **BusinessDay の時刻フォーマット不整合**
   - **ファイル**: `backend/internal/interface/rest/business_day_handler.go`
   - **箇所**: 265-266行目
   - **修正内容**: `Format("15:04")` を `Format("15:04:05")` に変更
   - **影響**: フロントエンドで時刻をパースする際にエラーになる可能性

6. **BusinessDay の updated_at がレスポンスに含まれない**
   - **ファイル**: `backend/internal/interface/rest/business_day_handler.go`
   - **箇所**: 38-48行目（構造体定義）、259-270行目（toBusinessDayResponse）
   - **修正内容**: レスポンス構造体に `UpdatedAt` を追加し、`toBusinessDayResponse` で設定
   - **影響**: フロントエンドで `updated_at` にアクセスすると `undefined` になる

7. **ShiftAssignment のフィールド不整合**
   - **ファイル**: `backend/internal/interface/rest/shift_assignment_handler.go`
   - **箇所**: 37-50行目（構造体定義）、178-188行目、204-214行目（GetAssignments）
   - **修正内容**: レスポンスに `tenant_id`、`is_outside_preference`、`cancelled_at`、`created_at`、`updated_at` を追加
   - **影響**: フロントエンドでこれらのフィールドにアクセスできない、または型エラーが発生

---

### 優先度 低：時間があれば で良いもの

8. **フロントエンドの環境変数ファイルの作成**
   - **ファイル**: `web-frontend/.env.local.example`
   - **修正内容**: 開発者向けの設定例を提供
   - **影響**: 開発環境ごとに異なるポート設定に対応できない（現状はデフォルト値で動作するため影響は小さい）

9. **ShiftAssignment の notification_sent フィールドの型定義追加**
   - **ファイル**: `web-frontend/src/types/api.ts`
   - **箇所**: 74-91行目（ShiftAssignment インターフェース）
   - **修正内容**: `notification_sent?: boolean` を追加
   - **影響**: 型エラーが発生する可能性（現状は使用されていないため影響は小さい）

---

**以上**


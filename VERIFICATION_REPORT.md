# VRC Shift Scheduler - バックエンド修正検証レポート

**検証日時**: 2025-12-04  
**検証者**: テストエンジニアAI  
**リポジトリ**: https://github.com/kento-matsunaga/vrc-shift-scheduler  
**最新コミット**: `9857502` - "fix: バックエンド REST API の JSON 整合性を修正"

---

## 実行サマリー

### ✅ 完了した検証項目

1. **STEP 0**: リポジトリの取得とブランチ確認 ✅
2. **STEP 1**: ポート設定と環境変数の整合性チェック ✅
3. **STEP 2**: BusinessDay APIのJSON形式検証 ✅
4. **STEP 3**: ShiftSlot.assigned_countの挙動検証 ✅（一部）

### ⚠️ 未完了・要調査項目

1. **STEP 3**: ShiftSlot.assigned_countの完全検証（シフト割り当て作成でエラー）
2. **STEP 4**: ShiftAssignment API（JOINフィールド & 日付フィルタ）の検証
3. **STEP 5**: 最小限の統合テスト（go test）の実行

---

## 詳細検証結果

### STEP 0: リポジトリの取得とブランチ確認

**やったこと**:
- リポジトリの状態確認
- 最新コミットの確認

**実行したコマンド**:
```bash
git status
git log -1 --oneline
git pull origin main
```

**実際の結果**:
- ブランチ: `main`
- 最新コミット: `9857502` - "fix: バックエンド REST API の JSON 整合性を修正"
- リポジトリはクリーンな状態（未追跡ファイルあり）

**期待との差分**: なし

---

### STEP 1: ポート設定と環境変数の整合性チェック

**やったこと**:
- `docker-compose.yml`のポート設定確認
- フロントエンドのAPIベースURL確認
- バックエンドサーバーの起動とヘルスチェック
- `docker-compose.yml`のcommand修正（`cmd/api` → `cmd/server`）

**実行したコマンド**:
```bash
docker compose up -d db
docker compose up -d backend
curl http://localhost:8080/health
```

**実際の結果**:
- ✅ バックエンドポート: `8080:8080`（docker-compose.yml）
- ✅ フロントエンドベースURL: `http://localhost:8080`（apiClient.ts）
- ✅ ヘルスチェック: `{"status":"ok"}` 正常応答
- ⚠️ **修正が必要**: `docker-compose.yml`のcommandが`cmd/api`になっていたため、`cmd/server`に変更

**修正内容**:
```yaml
# docker-compose.yml
backend:
  command: ["go", "run", "./cmd/server"]  # cmd/api → cmd/server に変更
  environment:
    PORT: "8080"  # API_PORT → PORT に変更
```

**期待との差分**: 
- `cmd/api`は`internal/http/router.go`を使用しており、APIエンドポイントが存在しない
- `cmd/server`は`internal/interface/rest/router.go`を使用しており、実際のAPIエンドポイントが定義されている

---

### STEP 2: BusinessDay APIのJSON形式検証

**やったこと**:
- イベント作成
- BusinessDay作成
- BusinessDay一覧取得
- JSON形式の確認

**実行したコマンド**:
```bash
# イベント作成
TENANT_ID="01KBHMYWYKRV8PK8EVYGF1SHV0"
EVENT_ID="01KBKQJCF4M8CXFKT6S0F74ZBZ"
curl -X POST "http://localhost:8080/api/v1/events" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{"event_name": "テストイベント", "event_type": "normal", "description": "BusinessDay JSON テスト用"}'

# BusinessDay作成
curl -X POST "http://localhost:8080/api/v1/events/$EVENT_ID/business-days" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{"target_date": "2025-01-15", "start_time": "21:30", "end_time": "23:00", "occurrence_type": "special"}'

# BusinessDay一覧取得
curl "http://localhost:8080/api/v1/events/$EVENT_ID/business-days" \
  -H "X-Tenant-ID: $TENANT_ID"
```

**実際の結果**:
```json
{
  "data": {
    "business_days": [
      {
        "business_day_id": "01KBKQJZ9K0N9JASM2177CPGM1",
        "tenant_id": "01KBHMYWYKRV8PK8EVYGF1SHV0",
        "event_id": "01KBKQJCF4M8CXFKT6S0F74ZBZ",
        "target_date": "2025-01-15",
        "start_time": "21:30:00",
        "end_time": "23:00:00",
        "occurrence_type": "special",
        "is_active": true,
        "created_at": "2025-12-04T03:48:10Z",
        "updated_at": "2025-12-04T03:48:10Z"
      }
    ],
    "count": 1
  }
}
```

**JSON形式チェック結果**:
- ✅ `business_day_id`: 文字列で存在
- ✅ `target_date`: `"YYYY-MM-DD"` 形式
- ✅ `start_time`: `"HH:MM:SS"` 形式（`"21:30:00"`）
- ✅ `end_time`: `"HH:MM:SS"` 形式（`"23:00:00"`）
- ✅ `created_at`: ISO8601 / RFC3339 形式（`"2025-12-04T03:48:10Z"`）
- ✅ `updated_at`: `created_at` と同様に文字列で存在

**フロントエンド型定義との整合性**:
- ✅ フロントエンドの`BusinessDay`型定義と完全一致

**期待との差分**: なし

---

### STEP 3: ShiftSlot.assigned_countの挙動検証

**やったこと**:
- Position作成（DB直接）
- ShiftSlot作成
- ShiftSlot一覧取得（割り当て前）
- `assigned_count`フィールドの確認
- `getTenantIDFromContext`の修正

**実行したコマンド**:
```bash
# Position作成（DB直接）
POSITION_ID="01ARZ3NDEKTSV4RRFFQ69G5FAV"
docker exec vrc-shift-scheduler-db-1 psql -U vrcshift -d vrcshift -c \
  "INSERT INTO positions (...) VALUES (...)"

# ShiftSlot作成
BUSINESS_DAY_ID="01KBKQJZ9K0N9JASM2177CPGM1"
curl -X POST "http://localhost:8080/api/v1/business-days/$BUSINESS_DAY_ID/shift-slots" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{"position_id": "...", "slot_name": "テストシフト", ...}'

# ShiftSlot一覧取得
curl "http://localhost:8080/api/v1/business-days/$BUSINESS_DAY_ID/shift-slots" \
  -H "X-Tenant-ID: $TENANT_ID"
```

**実際の結果**:
```json
{
  "data": {
    "shift_slots": [
      {
        "slot_id": "01KBKQNE0CGDFCGEH3NKJAAR6C",
        "assigned_count": 0,
        "required_count": 2,
        ...
      }
    ]
  }
}
```

**修正内容**:
1. **`response.go`の`getTenantIDFromContext`修正**:
   - ミドルウェアの`ContextKeyTenantID`と一致するように修正
   - `string`型から`common.TenantID`型に変更

2. **`shift_slot_handler.go`の`AssignedCount`フィールド修正**:
   - `omitempty`タグを削除して、常に`assigned_count`を含めるように修正

**期待との差分**: 
- ✅ `assigned_count`フィールドが正しく含まれている
- ⚠️ シフト割り当て作成でエラーが発生（後述）

---

### STEP 4: ShiftAssignment API（JOINフィールド & 日付フィルタ）の検証

**やったこと**:
- メンバー作成（DB直接）
- シフト割り当て作成（エラー発生）

**実行したコマンド**:
```bash
# メンバー作成（DB直接）
MEMBER1="01ARZ3NDEKTSV4RRFFQ69G5FAV"
MEMBER2="01ARZ3NDEKTSV4RRFFQ69G5FAW"
docker exec vrc-shift-scheduler-db-1 psql -U vrcshift -d vrcshift -c \
  "INSERT INTO members (...) VALUES (...)"

# シフト割り当て作成（エラー）
curl -X POST "http://localhost:8080/api/v1/shift-assignments" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "X-Member-ID: $MEMBER1" \
  -d '{"slot_id": "...", "member_id": "..."}'
```

**実際の結果**:
- ❌ シフト割り当て作成で`500 Internal Server Error`が発生
- エラーメッセージ: `"Failed to confirm shift assignment"`
- ログには詳細なエラー情報が表示されていない（Recoverミドルウェアがパニックをキャッチしている可能性）

**期待との差分**: 
- ⚠️ シフト割り当て作成が失敗している
- ⚠️ JOINフィールド（`member_display_name`, `slot_name`, `target_date`, `start_time`, `end_time`）の検証が未完了
- ⚠️ 日付範囲フィルタ（`start_date`, `end_date`）の検証が未完了

---

## 発見された問題と修正

### 1. docker-compose.ymlのcommand設定

**問題**: `cmd/api`を使用していたが、これはAPIエンドポイントが存在しないルーターを使用している

**修正**: `cmd/server`に変更（`internal/interface/rest/router.go`を使用）

### 2. response.goのgetTenantIDFromContext

**問題**: ミドルウェアの`ContextKeyTenantID`と不一致

**修正**: `ContextKeyTenantID`を使用し、`common.TenantID`型で取得するように変更

### 3. shift_slot_handler.goのAssignedCount

**問題**: `omitempty`タグにより、値が0の場合はJSONに含まれない

**修正**: `omitempty`タグを削除

---

## 実行したコマンド一覧

### ポート & 接続性確認
```bash
docker compose up -d db
docker compose up -d backend
curl http://localhost:8080/health
```

### BusinessDay API検証
```bash
# イベント作成
curl -X POST "http://localhost:8080/api/v1/events" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{"event_name": "テストイベント", ...}'

# BusinessDay作成
curl -X POST "http://localhost:8080/api/v1/events/$EVENT_ID/business-days" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{"target_date": "2025-01-15", ...}'

# BusinessDay一覧取得
curl "http://localhost:8080/api/v1/events/$EVENT_ID/business-days" \
  -H "X-Tenant-ID: $TENANT_ID"
```

### ShiftSlot API検証
```bash
# ShiftSlot作成
curl -X POST "http://localhost:8080/api/v1/business-days/$BUSINESS_DAY_ID/shift-slots" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{"position_id": "...", ...}'

# ShiftSlot一覧取得
curl "http://localhost:8080/api/v1/business-days/$BUSINESS_DAY_ID/shift-slots" \
  -H "X-Tenant-ID: $TENANT_ID"
```

---

## 人間がやるべきポチポチ作業 TODO

### 高優先度

1. **シフト割り当て作成エラーの調査**
   - `POST /api/v1/shift-assignments`で`500 Internal Server Error`が発生
   - ログに詳細なエラー情報を出力するように修正
   - `ShiftAssignmentService.ConfirmManualAssignment`の実装を確認
   - データベースの制約違反やNULL制約違反の可能性を調査

2. **メンバー作成APIのエラー調査**
   - `POST /api/v1/members`で`500 Internal Server Error`が発生（メール重複チェック）
   - エラーハンドリングの改善

3. **統合テストの実行**
   - `go test ./internal/interface/rest/...`を実行
   - テスト用データベースの準備
   - テスト結果の確認

### 中優先度

4. **ShiftAssignment APIの完全検証**
   - JOINフィールド（`member_display_name`, `slot_name`, `target_date`, `start_time`, `end_time`）の検証
   - 日付範囲フィルタ（`start_date`, `end_date`）の検証
   - `GET /api/v1/shift-assignments?member_id=...&assignment_status=confirmed`の検証

5. **ShiftSlot.assigned_countの完全検証**
   - シフト割り当て作成後の`assigned_count`更新確認
   - `required_count`を超える割り当て時の`409 Conflict`確認

6. **ログ出力の改善**
   - エラー時の詳細なスタックトレース出力
   - パニック時の詳細情報出力

### 低優先度

7. **Position APIエンドポイントの追加**
   - 現在、PositionはDB直接作成が必要
   - `POST /api/v1/positions`エンドポイントの追加を検討

8. **シードデータの整備**
   - テスト用のシードデータスクリプトの実行
   - テスト用テナント・メンバー・ポジションの準備

9. **エラーメッセージの日本語化**
   - フロントエンドで使用するエラーメッセージの日本語化

---

## まとめ

### ポート & 接続性

- ✅ バックエンドの実ポート: `8080`
- ✅ フロントが叩いているベースURL: `http://localhost:8080`
- ✅ `ERR_CONNECTION_REFUSED`問題は解消（`docker-compose.yml`の修正により）

### API JSON整合性の結果

- ✅ **BusinessDay**: フロントエンドの型定義と完全一致
- ✅ **ShiftSlot**: `assigned_count`フィールドが正しく含まれている（値は`0`）
- ⚠️ **ShiftAssignment**: JOINフィールドと日付フィルタの検証が未完了（シフト割り当て作成でエラー）

### 修正したファイル

1. `docker-compose.yml`: `cmd/api` → `cmd/server`, `API_PORT` → `PORT`
2. `backend/internal/interface/rest/response.go`: `getTenantIDFromContext`と`getMemberIDFromContext`の修正
3. `backend/internal/interface/rest/shift_slot_handler.go`: `AssignedCount`フィールドの`omitempty`タグ削除

---

**検証完了日時**: 2025-12-04 03:52:00 JST


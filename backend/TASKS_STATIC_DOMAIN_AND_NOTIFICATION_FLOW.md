# 静的ドメイン基盤 + 通知・監査フロー + 完全REST API 実装計画 (VRC Shift Scheduler)

## 🎯 真のMVP：最初に動かす1本の流れ

**このタスクファイルの最優先ゴール**は、以下の縦スライスを動作させることです：

```
1. Event 作成（REST API または Discord コマンド）
   ↓
2. EventBusinessDay を手動で数日分生成
   ↓
3. BusinessDay に ShiftSlot を手動で作成
   ↓
4. 管理者が Member を指定して ShiftAssignment を確定
   ↓
5. 結果が DB に記録され、REST API で取得できる
```

**この段階での割り切り**:
- ✅ **実装する**: Event, EventBusinessDay, ShiftSlot, ShiftAssignment, Member（最小限）のドメイン + DB + REST API
- ✅ **テーブルは作る**: Notification / AuditLog（将来の拡張のため）
- ⏸️ **後回し**: RecurringPattern の営業日自動生成ロジック、Notification の実送信、FrequencyControl、Idempotency、Availability（希望収集）
- 🔨 **stub 実装**: Notification は「ログ出力のみ」、AuditLog は「重要操作のみ記録」

**理由**: Multi-tenant + DDD + 通知 + 監査 + 頻度制御 + 冪等性を**最初から全部実装**すると、動くものが出るまでの距離が長すぎる。まず「Event → ShiftAssignment の縦スライス」を完成させ、そこから横に広げる方が現実的。

---

## このタスクファイルで扱う範囲と目的（フルスコープ）

上記の「真のMVP」を最優先としつつ、将来的には以下まで拡張します：

1. **静的部分の基盤実装**: Event / RecurringPattern / EventBusinessDay / ShiftSlot のドメインモデル、DBテーブル、リポジトリを完成させ、API・Bot抜きの純ドメインテストで動作確認する
2. **通知・監査の縦串フロー実装**: シフト確定から通知発火、FrequencyControl チェック、NotificationLog / AuditLog 記録までの1本の流れを実装・検証する
3. **完全な REST API の提供**: フロントエンドアプリケーションが実際に使用できるレベルの、完全なCRUD + ビジネスロジックAPIを実装する（OpenAPI ドキュメント含む）
4. **Discord Bot 連携**: Backend API を薄くラップした Discord Bot の実装（ビジネスロジックは Backend に集約）

**最終目標**: フロントエンド開発者が OpenAPI 仕様を見ながら、すぐにアプリケーションを構築できる状態にする。

## タスクステータスマーカー

- `[ ]` To Do: まだ着手していないタスク
- `[~]` In Progress: 現在 *アクティブに作業中* のタスク（同時に1つだけ）
- `[x]` Done: 完了したタスク
- `[!]` Blocked/Needs Attention: 何らかの理由で進行できないタスク（要確認）

## タスク優先度マーカー

- 🔴 緊急（直近の開発で最優先）
- 🟡 重要（なるべく早めに着手）
- 🟢 通常（今のタスクが片付き次第）
- ⚪ 低（将来的な改善・余裕があるとき）

## MVP（Minimum Viable Product）マーカー

- `[真MVP]` - **真のMVP**（Event → ShiftAssignment の縦スライスに必須）
- `[MVP]` - MVP として実装する機能（ただし真のMVPより優先度は低い）
- `[v1.1]` - MVP 完成後の次期バージョンで実装予定
- `[Nice-to-have]` - 余裕があれば実装する機能

**重要**: 開発は `[真MVP]` → `[MVP]` → `[v1.1]` の順で進める。`[真MVP]` が完成して初めて「動くものが見える」状態になる。

---

## 📐 実装順序の基本戦略

### Step 1: マイグレーション（DB）を薄く全て作る 🥇

**理由**: ドメインの struct を作る前に、DB スキーマを確定させた方が全体像が見えやすい

**対象テーブル（真のMVP + 将来拡張分）**:
- ✅ **真のMVP**: tenants, events, recurring_patterns（テーブルのみ）, event_business_days, shift_slots, members（最小限）, shift_assignments
- ⏸️ **将来のため作る**: notification_logs, notification_templates, audit_logs, availabilities

**この段階では**:
- CHECK / INDEX / FK は設計通り全て実装
- ただし RecurringPattern, Notification, Audit のテーブルは「**使わないが存在する**」状態でOK

### Step 2: ドメインは Event / BusinessDay / ShiftSlot / ShiftAssignment の4つ優先 🥈

**真のMVPに必要な最小限のドメイン**:
- Event（集約ルート）
- EventBusinessDay（独立エンティティ）
- ShiftSlot（独立エンティティ）
- ShiftAssignment（ShiftPlan 集約内のエンティティ、ただし ShiftPlan 自体は簡易実装でもOK）
- Member（最小限: member_id, tenant_id, display_name のみ）

**この段階で後回し**:
- RecurringPattern の詳細実装（struct は作るが、営業日生成ロジックは v1.1）
- Notification / AuditLog のドメインロジック（ログ出力 stub で代替）

### Step 3: REST API もこの4つに絞る 🥉

**真のMVP の REST API**:
- POST /api/v1/events - Event 作成
- GET /api/v1/events - Event 一覧
- GET /api/v1/events/:event_id - Event 詳細
- POST /api/v1/events/:event_id/business-days - BusinessDay 手動作成（RecurringPattern からの自動生成は後回し）
- GET /api/v1/events/:event_id/business-days - BusinessDay 一覧
- POST /api/v1/business-days/:business_day_id/shift-slots - ShiftSlot 作成
- GET /api/v1/business-days/:business_day_id/shift-slots - ShiftSlot 一覧
- POST /api/v1/shift-assignments - ShiftAssignment 確定
- GET /api/v1/shift-assignments - ShiftAssignment 一覧

**OpenAPI は「今実装済みのエンドポイントのみ」を記述**し、未実装分はコメントアウトまたは未記載とする。

---

## 将来のタスク / バックログ

### フロントエンド開発（別プロジェクトとして）
- [ ] ⚪ React/Vue/Svelte によるWebフロントエンド実装
  - ダッシュボード（イベント一覧、シフト充足率）
  - イベント管理画面（CRUD）
  - シフト割り当て画面（カレンダービュー、ドラッグ&ドロップ）
  - 希望提出画面（メンバー向け）
  - 通知履歴・監査ログ閲覧画面（管理者向け）

### 機能拡張
- [ ] ⚪ シフト自動割り当てアルゴリズムの実装（優先度ベース、公平性考慮）
- [ ] ⚪ 通知テンプレートの多言語対応
- [ ] ⚪ リマインダーの定期実行バッチジョブ化
- [ ] ⚪ WebSocket によるリアルタイム更新通知
- [ ] ⚪ メール通知の実装（Discord 以外のチャネル）
- [ ] ⚪ CSV/Excel によるシフトデータのインポート/エクスポート
- [ ] ⚪ レポート機能（メンバーごとのシフト稼働統計、充足率推移）

---

## 進行中 / To Do

### 🔴 **親タスク 1: 静的ドメインの基盤実装（Event 〜 ShiftSlot）**

静的な構造（イベント・営業日・シフト枠）をドメインモデル・DB・リポジトリ・テストで完全に動かせる状態にする。

- [x] 🔴 **サブタスク 1.1: Event ドメインの実装** [MVP]
  - [x] 🔴 1.1.1: Event エンティティの Go struct 定義 [MVP]
    - *詳細:* `backend/internal/domain/event/event.go` に Event エンティティを定義
    - *不変条件:* EventName の必須性、期間の前後関係、TenantID の存在
    - *依存:* `docs/domain/10_tenant-and-event/ドメインモデル.md`
    - *見積もり:* 1〜2時間
    - *⚠️ Multi-Tenant設計:* 全エンティティに tenant_id を必須フィールドとして含める
  - [x] 🔴 1.1.2: Event 用の DB テーブル定義とマイグレーション作成 [MVP]
    - *詳細:* `backend/internal/infra/db/migrations/001_create_events_table.sql`
    - *カラム:* event_id (ULID), tenant_id (ULID), event_name, event_type (normal/special), description, is_active, created_at, updated_at
    - *制約:* 
      - PK(event_id)
      - FK(tenant_id) REFERENCES tenants(tenant_id)
      - UNIQUE(tenant_id, event_name)（同一テナント内でイベント名一意）
      - INDEX(tenant_id, is_active)（テナント内のアクティブイベント検索用）
    - *見積もり:* 1時間
  - [x] 🔴 1.1.3: EventRepository インターフェースの定義 [MVP]
    - *詳細:* `backend/internal/domain/event/repository.go`
    - *メソッド:* Save, FindByID, FindByTenantID, Delete
    - *見積もり:* 30分
    - *⚠️ Multi-Tenant前提:* 全メソッドで tenant_id を引数に取る（例: `FindByID(ctx, tenantID, eventID)`）。tenant境界を越えたアクセスを防ぐため必須
  - [x] 🟡 1.1.4: EventRepository の実装（PostgreSQL）
    - *詳細:* `backend/internal/infra/db/event_repository.go`
    - *依存:* サブタスク 1.1.2, 1.1.3
    - *見積もり:* 2〜3時間
    - *✅ 完了日:* 2025-12-03

- [x] 🔴 **サブタスク 1.2: RecurringPattern ドメインの実装** [MVP]
  - [x] 🔴 1.2.1: RecurringPattern エンティティの定義 [MVP]
    - *詳細:* `backend/internal/domain/event/recurring_pattern.go`
    - *属性:* pattern_id (ULID), event_id, pattern_type (enum), config (map/struct), created_at, updated_at
    - *パターン種別:* Weekly（曜日リスト）、MonthlyDate（日付リスト）、Custom（JSONB自由形式）
    - *不変条件:* パターンごとのバリデーション（例: Weekly なら曜日リスト必須、7個以内）
    - *依存:* `docs/domain/10_tenant-and-event/ドメインモデル.md`
    - *見積もり:* 2〜3時間
  - [x] 🔴 1.2.2: RecurringPattern 用の DB テーブル定義とマイグレーション [MVP]
    - *詳細:* `backend/internal/infra/db/migrations/001_create_events_and_recurring_patterns_tables.sql`（Event と同じマイグレーションファイルに含める）
    - *カラム:* pattern_id (ULID), tenant_id (ULID), event_id (ULID), pattern_type (weekly/monthly_date/custom), config (JSONB), created_at, updated_at
    - *制約:*
      - PK(pattern_id)
      - FK(tenant_id) REFERENCES tenants(tenant_id)
      - FK(event_id) REFERENCES events(event_id) ON DELETE CASCADE
      - UNIQUE(tenant_id, event_id)（1 Event につき 1 RecurringPattern）
      - CHECK(pattern_type IN ('weekly', 'monthly_date', 'custom'))
      - INDEX(tenant_id, event_id)
    - *config JSONB の例*:
      - Weekly: `{"day_of_weeks": ["MON", "FRI"], "start_time": "21:30", "end_time": "23:00"}`
      - MonthlyDate: `{"dates": [1, 15], "start_time": "21:30", "end_time": "23:00"}`
    - *✅ 決定事項1を反映*: 専用テーブル + JSONB のハイブリッド方式
    - *見積もり:* 1〜2時間
  - [x] 🟡 1.2.3: RecurringPattern の config シリアライズ/デシリアライズ実装 [MVP]
    - *詳細:* Go struct (RecurringPatternConfig) ⇔ JSONB の変換ロジック
    - *実装方針:* `encoding/json` を使い、pattern_type ごとに異なる struct にアンマーシャル
    - *見積もり:* 1〜2時間

- [x] 🔴 **サブタスク 1.3: EventBusinessDay ドメインの実装** [MVP]
  - [x] 🔴 1.3.1: EventBusinessDay エンティティの定義 [MVP]
    - *詳細:* `backend/internal/domain/event/event_business_day.go`
    - *属性:* business_day_id, event_id, target_date, day_of_week, is_active
    - *不変条件:* target_date が Event の期間内、日付の一意性
    - *依存:* サブタスク 1.1
    - *見積もり:* 1〜2時間
    - *⚠️ 集約境界の明確化:* Event と EventBusinessDay の関係を明文化する
      - **方針**: Event は EventBusinessDay を直接保持せず、ID参照のみ（Event集約 ≠ BusinessDay集約）
      - Event は「期間 + RecurringPattern」の定義、BusinessDay は「生成されたインスタンス」
      - BusinessDay の編集（is_active変更など）は Event の不変条件を壊さない範囲に限定
      - この方針をドメインモデルドキュメントに明記すること
  - [x] 🔴 1.3.2: EventBusinessDay 用の DB テーブル定義とマイグレーション [MVP]
    - *詳細:* `backend/internal/infra/db/migrations/002_create_event_business_days_table.sql`
    - *カラム:* business_day_id (ULID), tenant_id (ULID), event_id (ULID), target_date (DATE), start_time (TIME), end_time (TIME), occurrence_type (recurring/special), recurring_pattern_id (ULID, nullable), is_active, valid_from (DATE), valid_to (DATE), created_at, updated_at
    - *制約:*
      - PK(business_day_id)
      - FK(tenant_id) REFERENCES tenants(tenant_id)
      - FK(event_id) REFERENCES events(event_id) ON DELETE CASCADE
      - FK(recurring_pattern_id) REFERENCES recurring_patterns(pattern_id) ON DELETE SET NULL（通常営業の場合のみ）
      - UNIQUE(tenant_id, event_id, target_date, start_time)（同一テナント・イベント・日時で一意）
      - CHECK(start_time < end_time OR end_time < start_time)（深夜営業対応: 日付跨ぎを許可）
      - CHECK((occurrence_type = 'recurring' AND recurring_pattern_id IS NOT NULL) OR (occurrence_type = 'special' AND recurring_pattern_id IS NULL))
      - INDEX(tenant_id, target_date)（テナント内の日付検索用）
      - INDEX(event_id, target_date)（イベント内の営業日検索用）
    - *⚠️ tenant_id の必須性:* ドメインドキュメントで明示的に「tenant_id を直接保持」と記載あり
    - *見積もり:* 1〜2時間
  - [x] 🔴 1.3.3: EventBusinessDayRepository インターフェースの定義 [MVP]
    - *メソッド:* Save, FindByEventID, FindByID, FindByDateRange
    - *見積もり:* 30分
    - *⚠️ Multi-Tenant前提:* 全メソッドで tenant_id を引数に取る
  - [x] 🔴 1.3.4: EventBusinessDayRepository の実装（PostgreSQL） [真MVP]
    - *詳細:* `backend/internal/infra/db/business_day_repository.go`
    - *見積もり:* 2時間
    - *✅ 完了日:* 2025-12-03
  - [ ] 🟡 1.3.5: RecurringPattern から EventBusinessDay を生成するドメインサービス実装 [v1.1]
    - *詳細:* `backend/internal/domain/event/business_day_generator.go`
    - *ロジック:* Event + RecurringPattern → EventBusinessDay のリストを生成
    - *依存:* サブタスク 1.2, 1.3.1
    - *見積もり:* 3〜4時間
    - *⏸️ 真のMVPでは後回し*: 営業日は手動作成（API経由）で進める。自動生成は v1.1 で実装

- [x] 🔴 **サブタスク 1.4: ShiftSlot ドメインの実装**
  - [x] 🔴 1.4.1: ShiftSlot エンティティの定義
    - *詳細:* `backend/internal/domain/shift/shift_slot.go`
    - *属性:* slot_id, business_day_id, slot_name, start_time, end_time, required_count, priority
    - *不変条件:* 時刻の前後関係、required_count >= 0
    - *依存:* `docs/domain/50_shift-plan-and-assignment/ドメインモデル.md`
    - *見積もり:* 2時間
  - [x] 🔴 1.4.2: ShiftSlot 用の DB テーブル定義とマイグレーション [MVP]
    - *詳細:* `backend/internal/infra/db/migrations/003_create_shift_slots_table.sql`
    - *カラム:* slot_id (ULID), tenant_id (ULID), business_day_id (ULID), position_id (ULID), slot_name, instance_name, start_time (TIME), end_time (TIME), required_count (INT), priority, created_at, updated_at
    - *制約:*
      - PK(slot_id)
      - FK(tenant_id) REFERENCES tenants(tenant_id)
      - FK(business_day_id) REFERENCES event_business_days(business_day_id) ON DELETE CASCADE
      - FK(position_id) REFERENCES positions(position_id)
      - CHECK(start_time < end_time OR end_time < start_time)（深夜対応）
      - CHECK(required_count >= 1)（必要人数は1以上）
      - INDEX(tenant_id, business_day_id)（営業日内のシフト枠検索用）
      - INDEX(business_day_id, start_time)（時刻順ソート用）
    - *⚠️ required_count の制御*: このカラムだけでは同時確定制御はできない。Application Service で排他制御が必要（サブタスク 2.5.1 参照）
    - *見積もり:* 1時間
  - [x] 🔴 1.4.3: ShiftSlotRepository インターフェースの定義
    - *メソッド:* Save, FindByID, FindByBusinessDayID, Delete
    - *見積もり:* 30分
  - [x] 🟡 1.4.4: ShiftSlotRepository の実装（PostgreSQL）
    - *詳細:* `backend/internal/infra/db/shift_slot_repository.go`
    - *見積もり:* 2時間
    - *✅ 完了日:* 2025-12-03

- [x] 🟡 **サブタスク 1.5: 静的ドメインの純粋テスト実装**
  - [x] 🟡 1.5.1: Event ドメインの単体テスト
    - *詳細:* `backend/internal/domain/event/event_test.go`
    - *テストケース:* エンティティ生成、不変条件違反、バリデーション
    - *見積もり:* 1〜2時間
  - [ ] 🟡 1.5.2: RecurringPattern のテスト [v1.1]
    - *テストケース:* 各パターンの営業日生成ロジック（Daily, Weekly, etc.）
    - *見積もり:* 2〜3時間
    - *⏸️ 真のMVPでは後回し*: 営業日自動生成ロジックは v1.1
  - [x] 🟡 1.5.3: EventBusinessDay + ShiftSlot の統合テスト（ShiftSlot単体テストとして実装済み）
    - *シナリオ:* Event 作成 → RecurringPattern で営業日生成 → ShiftSlot 登録 → リポジトリで永続化・取得
    - *詳細:* `backend/internal/domain/shift/shift_slot_test.go`
    - *見積もり:* 3〜4時間
  - [x] 🟡 1.5.4: EventRepository の統合テスト
    - *詳細:* `backend/internal/infra/db/event_repository_integration_test.go`
    - *テストケース:* DB接続、Event作成・取得・更新、BusinessDay連携
    - *✅ 完了日:* 2025-12-03

- [x] 🔴 **サブタスク 1.6: Member ドメインの実装（真のMVP版）** [真MVP]
  - [x] 🔴 1.6.1: Member エンティティの定義（最小限）[真MVP]
    - *詳細:* `backend/internal/domain/member/member.go`
    - *属性:* member_id, tenant_id, display_name, discord_user_id, email
    - *見積もり:* 1時間
    - *✅ 完了日:* 2025-12-03
    - *⏸️ 最小限実装*: ロール・外部アカウント管理は v1.1
  - [x] 🔴 1.6.2: MemberRepository インターフェースの定義 [真MVP]
    - *詳細:* `backend/internal/domain/member/repository.go`
    - *メソッド:* Save, FindByID, FindByDiscordUserID, FindByEmail, FindAllByTenantID
    - *見積もり:* 30分
    - *✅ 完了日:* 2025-12-03
  - [x] 🔴 1.6.3: MemberRepository の実装（PostgreSQL）[真MVP]
    - *詳細:* `backend/internal/infra/db/member_repository.go`
    - *見積もり:* 2時間
    - *✅ 完了日:* 2025-12-03

- [x] 🔴 **サブタスク 1.7: Position ドメインの実装** [真MVP]
  - [x] 🔴 1.7.1: Position エンティティの定義 [真MVP]
    - *詳細:* `backend/internal/domain/shift/position.go`
    - *属性:* position_id, tenant_id, position_name, description, display_order, is_active
    - *見積もり:* 1時間
    - *✅ 完了日:* 2025-12-03
  - [x] 🔴 1.7.2: PositionRepository インターフェースの定義 [真MVP]
    - *詳細:* `backend/internal/domain/shift/position_repository.go`
    - *メソッド:* Save, FindByID, FindByTenantID, FindActiveByTenantID, Delete
    - *見積もり:* 30分
    - *✅ 完了日:* 2025-12-03
  - [x] 🔴 1.7.3: PositionRepository の実装（PostgreSQL）[真MVP]
    - *詳細:* `backend/internal/infra/db/position_repository.go`
    - *見積もり:* 2時間
    - *✅ 完了日:* 2025-12-03

---

### 🔴 **親タスク 2: ShiftAssignment + 通知・監査（stub）の実装**

ShiftAssignment（シフト確定）のドメイン実装と、通知・監査の**最小限の stub**を用意する。
真のMVPでは、通知は「ログ出力のみ」、監査は「重要操作のみ記録」で進める。

- [x] 🔴 **サブタスク 2.1: ShiftAssignment ドメインの実装** [MVP]
  - [x] 🔴 2.1.1: ShiftAssignment エンティティの定義 [MVP]
    - *詳細:* `backend/internal/domain/shift/shift_assignment.go`
    - *属性:* assignment_id, slot_id, member_id, status (confirmed/pending/cancelled)
    - *不変条件:* 同じ slot_id + member_id の重複禁止、status の遷移ルール、required_count を超えない
    - *依存:* `docs/domain/50_shift-plan-and-assignment/ドメインモデル.md`
    - *見積もり:* 2時間
    - *⚠️ 同時実行制御:* 同じ枠に複数人が同時に確定しようとした場合の排他戦略を決める
      - **推奨方針**: `SELECT ... FOR UPDATE` で該当 slot の assignments をロックしてから required_count チェック
      - または DB の UNIQUE 制約違反を catch して `409 Conflict` を返す
      - この方針をアプリケーションサービス実装時に明記すること
  - [x] 🔴 2.1.2: ShiftAssignment 用の DB テーブル定義とマイグレーション [MVP]
    - *詳細:* `backend/internal/infra/db/migrations/004_create_shift_assignments_table.sql`
    - *カラム:* assignment_id (ULID), tenant_id (ULID), plan_id (ULID), slot_id (ULID), member_id (ULID), assignment_status (confirmed/cancelled), assignment_method (auto/manual), is_outside_preference (BOOLEAN), assigned_at, cancelled_at (nullable), created_at, updated_at
    - *制約:*
      - PK(assignment_id)
      - FK(tenant_id) REFERENCES tenants(tenant_id)
      - FK(plan_id) REFERENCES shift_plans(plan_id) ON DELETE CASCADE
      - FK(slot_id) REFERENCES shift_slots(slot_id) ON DELETE CASCADE
      - FK(member_id) REFERENCES members(member_id)
      - CHECK(assignment_status IN ('confirmed', 'cancelled'))
      - CHECK(assignment_method IN ('auto', 'manual'))
      - INDEX(tenant_id, member_id, assignment_status)（メンバーの確定済みシフト検索用）
      - INDEX(slot_id, assignment_status)（シフト枠の充足状況確認用）
      - INDEX(plan_id)（ShiftPlan に紐づく割り当て検索用）
    - *⚠️ UNIQUE 制約の注意*: `UNIQUE(slot_id, member_id)` は履歴管理（キャンセル後の再割り当て）があるため、`UNIQUE(slot_id, member_id, assignment_status) WHERE assignment_status = 'confirmed'` の部分一意インデックスを推奨
    - *⚠️ required_count 制御*: この制約だけでは「同じ枠に required_count を超えて割り当てない」は保証できない。Application Service で `SELECT ... FOR UPDATE` を使った排他制御が必須
    - *見積もり:* 1〜2時間
  - [x] 🔴 2.1.3: ShiftAssignmentRepository インターフェースの定義
    - *メソッド:* Save, FindByID, FindBySlotID, FindByMemberID, UpdateStatus
    - *見積もり:* 30分
  - [x] 🟡 2.1.4: ShiftAssignmentRepository の実装（PostgreSQL）
    - *詳細:* `backend/internal/infra/db/shift_assignment_repository.go`
    - *見積もり:* 2時間
    - *✅ 完了日:* 2025-12-03

- [ ] 🟡 **サブタスク 2.2: Notification ドメインの実装（stub）** [v1.1]
  - [ ] 🟡 2.2.1: NotificationEvent エンティティの定義（stub） [v1.1]
    - *詳細:* `backend/internal/domain/notification/notification_event.go`
    - *属性:* event_id, event_type (SHIFT_CONFIRMED, REMINDER, etc.), payload (JSONB), triggered_at
    - *依存:* `docs/domain/60_notification-and-reminder/ドメインモデル.md`
    - *見積もり:* 2時間
    - *⏸️ 真のMVPでは stub*: struct だけ定義し、実際の発火ロジックは「ログ出力」のみ
  - [ ] 🟡 2.2.2: NotificationLog エンティティの定義（stub） [v1.1]
    - *詳細:* `backend/internal/domain/notification/notification_log.go`
    - *属性:* log_id, event_id, recipient_id, channel (Discord/Email), sent_at, status, retry_count
    - *見積もり:* 1〜2時間
    - *⏸️ 真のMVPでは stub*: テーブルは作成済みだが、実際のログ記録は最小限
  - [ ] 🔴 2.2.3: Notification 用の DB テーブル定義とマイグレーション [MVP]
    - *詳細:* `backend/internal/infra/db/migrations/005_create_notification_tables.sql`
    - *テーブル1: notification_logs* （通知送信履歴）
      - *カラム:* log_id (ULID), tenant_id (ULID), business_day_id (ULID, nullable), recipient_id (ULID), notification_type (enum), message_content (TEXT), delivery_channel (Discord/Email), delivery_status (success/failed/pending), error_message (TEXT, nullable), sent_at, created_at
      - *制約:*
        - PK(log_id)
        - FK(tenant_id) REFERENCES tenants(tenant_id)
        - FK(business_day_id) REFERENCES event_business_days(business_day_id) ON DELETE SET NULL（営業日関連通知の場合のみ）
        - FK(recipient_id) REFERENCES members(member_id)
        - CHECK(notification_type IN ('shift_recruitment', 'deadline_reminder', 'shift_confirmed', 'attendance_reminder', 'urgent_help'))
        - CHECK(delivery_status IN ('success', 'failed', 'pending'))
        - **INDEX(recipient_id, sent_at)**（FrequencyControl 用の必須インデックス - サブタスク 2.3.1 参照）
        - INDEX(tenant_id, business_day_id, notification_type)（営業日ごとの通知履歴検索用）
        - INDEX(tenant_id, notification_type, sent_at)（通知種別ごとの履歴検索用）
    - *テーブル2: notification_templates* （通知テンプレート）
      - *カラム:* template_id (ULID), tenant_id (ULID), template_type (enum), template_name, message_template (TEXT), variable_definitions (JSONB), created_at, updated_at
      - *制約:*
        - PK(template_id)
        - FK(tenant_id) REFERENCES tenants(tenant_id)
        - UNIQUE(tenant_id, template_type)（同一テナント内で種別一意）
    - *見積もり:* 2〜3時間
  - [ ] 🔴 2.2.4: NotificationRepository インターフェースの定義
    - *メソッド:* SaveEvent, SaveLog, FindLogsByRecipient, FindRecentLogs
    - *見積もり:* 30分
  - [ ] 🟡 2.2.5: NotificationRepository の実装（PostgreSQL）
    - *見積もり:* 2〜3時間

- [ ] ⚪ **サブタスク 2.3: FrequencyControl ポリシーの実装** [v1.1]
  - [ ] ⚪ 2.3.1: FrequencyControlPolicy ドメインサービスの定義 [v1.1]
    - *詳細:* `backend/internal/domain/notification/frequency_control_policy.go`
    - *ロジック:* 過去 N 分以内に同一 recipient への通知が X 件以上ある場合はスパム判定
    - *依存:* NotificationLog の取得
    - *見積もり:* 2〜3時間
    - *⚠️ パフォーマンス対策:* 「過去N分以内のログをrecipient_idで絞って数える」クエリに対応するインデックスを設計
      - **必須インデックス**: `notification_logs` テーブルに `(recipient_id, sent_at)` の複合インデックス
      - このインデックスを該当マイグレーションファイル（005_create_notification_tables.sql）に含めること
      - クエリ例: `SELECT COUNT(*) FROM notification_logs WHERE recipient_id = ? AND sent_at > ?`
  - [ ] 🟡 2.3.2: FrequencyControl の設定値管理
    - *詳細:* 設定ファイル or DB テーブルで管理（例: 10分以内に5件以上でスパム）
    - *方針確認:* 設定の持ち方（要ユーザー確認）
    - *見積もり:* 1時間
  - [ ] 🟡 2.3.3: FrequencyControlPolicy のテスト
    - *テストケース:* スパム判定される / されないケース
    - *見積もり:* 1〜2時間

- [ ] 🟡 **サブタスク 2.4: AuditLog ドメインの実装（最小限）** [v1.1]
  - [ ] 🟡 2.4.1: AuditLog エンティティの定義（stub） [v1.1]
    - *詳細:* `backend/internal/domain/audit/audit_log.go`
    - *属性:* log_id, entity_type, entity_id, action (CREATE/UPDATE/DELETE), actor_id, changed_data (JSONB), timestamp
    - *依存:* `docs/domain/60_notification-and-reminder/ドメインモデル.md`（監査ログ仕様）
    - *見積もり:* 1〜2時間
    - *⏸️ 真のMVPでは最小限*: ShiftAssignment の CREATE のみ記録、他は後回し
  - [ ] 🔴 2.4.2: AuditLog 用の DB テーブル定義とマイグレーション [真MVP]
    - *詳細:* `backend/internal/infra/db/migrations/006_create_audit_logs_table.sql`
    - *カラム:* log_id (ULID), tenant_id (ULID), entity_type (events/shift_assignments/etc.), entity_id (ULID), action (CREATE/UPDATE/DELETE), actor_id (ULID), changed_data_before (JSONB, nullable), changed_data_after (JSONB, nullable), timestamp, created_at
    - *制約:*
      - PK(log_id)
      - FK(tenant_id) REFERENCES tenants(tenant_id)
      - FK(actor_id) REFERENCES members(member_id)（操作者）
      - CHECK(action IN ('CREATE', 'UPDATE', 'DELETE'))
      - INDEX(tenant_id, entity_type, entity_id)（エンティティごとの変更履歴検索用）
      - INDEX(tenant_id, actor_id, timestamp)（操作者ごとの履歴検索用）
      - INDEX(timestamp)（時系列検索用）
    - *見積もり:* 1〜2時間
  - [ ] 🟡 2.4.3: AuditLogRepository インターフェースの定義
    - *メソッド:* Save, FindByEntityID, FindByActorID, FindByTimeRange
    - *見積もり:* 30分
  - [ ] 🟡 2.4.4: AuditLogRepository の実装（PostgreSQL）
    - *見積もり:* 2時間

- [ ] 🔴 **サブタスク 2.5: ShiftAssignment 確定フローの実装（通知・監査は stub）** [真MVP + v1.1]
  - [ ] 🔴 2.5.1: ShiftAssignmentService（アプリケーションサービス）の実装 [真MVP]
    - *詳細:* `backend/internal/app/shift_assignment_service.go`
    - *ロジック（真のMVP版）:*
      1. ShiftAssignment を作成・保存（排他制御付き）
      2. ログに「シフト確定」を出力（Notification stub）
      3. AuditLog に CREATE アクションを記録（最小限）
    - *ロジック（v1.1 で追加）:*
      - NotificationEvent 発火 → FrequencyControl チェック → NotificationLog 記録 → Discord 実送信
    - *依存:* 親タスク1の完了、サブタスク 2.1
    - *見積もり:* 3〜4時間（真のMVP版）+ 2〜3時間（v1.1 拡張）
    - *⚠️ トランザクション境界と通知の同期/非同期:*
      - **v1 実装方針（同期処理）**:
        - **トランザクション内**: ShiftAssignment の作成・保存、AuditLog（"確定した"という事実）の記録
        - **トランザクション外**: 実際の Discord 送信、NotificationLog（送信結果）の記録
      - **将来的な拡張性の担保**:
        - NotificationEvent の「発火」と「送信」は分離されたインターフェースにする
        - outbox パターンや非同期キュー（RabbitMQ/SQS）への移行を見越した設計
        - 送信失敗時のリトライロジックを別レイヤーに切り出せるようにする
      - **事故パターンの回避**:
        - DB コミット成功 → Discord 送信失敗 の場合、NotificationLog に失敗を記録
        - 二重送信を防ぐため、NotificationEvent に idempotency key を持たせる
  - [ ] 🔴 2.5.2: 統合シナリオテストの実装（真のMVP版） [真MVP]
    - *シナリオ（真のMVP版）:*
      1. Event 作成
      2. EventBusinessDay + ShiftSlot 作成
      3. ShiftAssignment 確定
      4. ログ出力確認（Notification stub）
      5. AuditLog 記録確認（CREATE のみ）
    - *シナリオ（v1.1 で追加）:*
      - NotificationEvent 発火確認
      - FrequencyControl ポリシーチェック（スパム判定含む）
      - NotificationLog 記録確認
      - Discord 実送信確認
    - *詳細:* `backend/internal/app/integration_test.go`
    - *見積もり:* 2〜3時間（真のMVP版）+ 2〜3時間（v1.1 拡張）

---

### 🔴 **親タスク 3A: REST API 基盤 + Event/BusinessDay API（真のMVP）**

真のMVPとして、Event → EventBusinessDay の作成・取得 API を実装する。
親タスク3は規模が大きいため、以下のように分割して段階的に実装する：
- **3A（真のMVP）**: API基盤 + Event/BusinessDay 管理 API（作成・一覧・詳細のみ）
- **3B（真のMVP）**: ShiftSlot/Assignment 管理 API（作成・一覧・詳細のみ）
- **3C（v1.1以降）**: 更新・削除 API、Member/Availability 管理 + 可視化 API

- [x] 🟢 **サブタスク 3.1: API 基盤の実装** [MVP]
  - [x] 🟢 3.1.1: HTTP ルーター / ミドルウェアの実装 [MVP]
    - *詳細:* `backend/internal/interface/rest/router.go`, `middleware.go`
    - *機能:* CORS設定、ロギング、エラーハンドリングミドルウェア、JSON レスポンスヘルパー、簡易認証（X-Tenant-ID/X-Member-IDヘッダー）
    - *見積もり:* 2〜3時間
    - *✅ 完了日:* 2025-12-03
    - *⚠️ DDD レイヤ保護ルール（重要）:*
      - **状態変更系 API（POST/PUT/PATCH/DELETE）は必ず Application Service 経由**
      - ハンドラから直接 Repository を呼び出して永続化してはいけない（集約の不変条件が破壊される）
      - **参照系 API（GET）のみ**、パフォーマンス目的で Repository 直接アクセスを許可
      - 例外的に Repository を直接触る場合は、必ずコードレビューで合意を得ること
  - [x] 🟢 3.1.2: API エラーレスポンスの標準化 [MVP]
    - *詳細:* `backend/internal/interface/rest/response.go`
    - *形式:* `{ "error": { "code": "ERR_xxx", "message": "...", "details": {...} } }`
    - *見積もり:* 1時間
    - *✅ 完了日:* 2025-12-03
    - *実装済みエラーコード:*
      - `ERR_INVALID_REQUEST` - バリデーションエラー
      - `ERR_NOT_FOUND` - リソースが存在しない
      - `ERR_CONFLICT` - 競合（同時実行、重複など）
      - `ERR_FORBIDDEN` - テナント境界違反
      - `ERR_INTERNAL` - サーバー内部エラー
  - [ ] 🟢 3.1.3: リクエストバリデーション共通機構 [v1.1]
    - *詳細:* `backend/internal/interface/rest/validator.go`
    - *機能:* struct tag ベースのバリデーション、カスタムルール
    - *見積もり:* 2時間
    - *⏸️ 真のMVPでは後回し*: 現在は各ハンドラ内で手動バリデーション、共通化は v1.1

- [x] 🔴 **サブタスク 3.2: Event 管理 API の実装（作成・一覧・詳細）** [真MVP]
  - [x] 🔴 3.2.1: POST /api/v1/events - Event 作成 [真MVP]
    - *詳細:* `backend/internal/interface/rest/event_handler.go`
    - *リクエスト:* `{ event_name, event_type, description }`（tenant_idはヘッダーから取得）
    - *レスポンス:* `{ data: { event_id, tenant_id, event_name, event_type, description, is_active, created_at, updated_at } }`
    - *バリデーション:* event_name の必須性、event_type の妥当性、重複チェック
    - *見積もり:* 2〜3時間
    - *✅ 完了日:* 2025-12-03
    - *⏸️ RecurringPattern は後回し*: 真のMVPでは Event 作成のみ、RecurringPattern は v1.1
  - [x] 🔴 3.2.2: GET /api/v1/events - Event 一覧取得 [真MVP]
    - *レスポンス:* `{ data: { events: [...], count: N } }`
    - *見積もり:* 2時間
    - *✅ 完了日:* 2025-12-03
    - *⏸️ ページネーションは v1.1*: 現在は全件取得、ページネーション対応は v1.1
  - [x] 🔴 3.2.3: GET /api/v1/events/:event_id - Event 詳細取得 [真MVP]
    - *レスポンス:* `{ data: { event_id, tenant_id, event_name, ... } }`
    - *見積もり:* 1時間
    - *✅ 完了日:* 2025-12-03
  - [ ] 🟡 3.2.4: PUT /api/v1/events/:event_id - Event 更新 [v1.1]
    - *リクエスト:* 更新可能フィールド（event_name, description, recurring_pattern）
    - *制約:* 既にシフトが確定している場合は期間変更不可
    - *見積もり:* 2〜3時間
  - [ ] 🟡 3.2.5: DELETE /api/v1/events/:event_id - Event 削除（論理削除） [v1.1]
    - *制約:* 確定済みシフトがある場合は削除不可（エラー返却）
    - *見積もり:* 1〜2時間

- [x] 🔴 **サブタスク 3.3: EventBusinessDay 管理 API（作成・一覧・詳細）** [真MVP]
  - [x] 🔴 3.3.1: POST /api/v1/events/:event_id/business-days - 営業日手動作成 [真MVP]
    - *詳細:* `backend/internal/interface/rest/business_day_handler.go`
    - *リクエスト:* `{ target_date, start_time, end_time, occurrence_type }`
    - *レスポンス:* `{ data: { business_day_id, tenant_id, event_id, target_date, start_time, end_time, occurrence_type, is_active, created_at } }`
    - *処理:* EventBusinessDay を手動で1件作成
    - *見積もり:* 1〜2時間
    - *✅ 完了日:* 2025-12-03
    - *⏸️ 自動生成は後回し*: RecurringPattern からの自動生成は v1.1
  - [x] 🔴 3.3.2: GET /api/v1/events/:event_id/business-days - 営業日一覧取得 [真MVP]
    - *クエリパラメータ:* `start_date`, `end_date`（日付範囲フィルタ対応）
    - *レスポンス:* `{ data: { business_days: [...], count: N } }`
    - *見積もり:* 1〜2時間
    - *✅ 完了日:* 2025-12-03
  - [x] 🔴 3.3.3: GET /api/v1/business-days/:business_day_id - 営業日詳細取得 [真MVP]
    - *レスポンス:* `{ data: { business_day_id, tenant_id, event_id, target_date, ... } }`
    - *見積もり:* 1時間
    - *✅ 完了日:* 2025-12-03
  - [ ] 🟡 3.3.4: POST /api/v1/events/:event_id/generate-business-days - 営業日一括生成 [v1.1]
    - *処理:* RecurringPattern に基づいて EventBusinessDay を生成
    - *レスポンス:* 生成された営業日数
    - *見積もり:* 2時間
  - [ ] 🟡 3.3.5: PATCH /api/v1/business-days/:business_day_id - 営業日のアクティブ状態変更 [v1.1]
    - *リクエスト:* `{ is_active: true/false }`
    - *用途:* 特定日の営業を休止する場合など
    - *見積もり:* 1時間
    - *⚠️ 集約境界の注意:* is_active 変更は Event の不変条件（期間の整合性など）を壊さないことを確認

---

### 🔴 **親タスク 3B: ShiftSlot/Assignment 管理 API（MVP 必須）**

シフト枠の作成・管理とシフト確定・キャンセルの API を提供する。通知・監査フローと連携。

- [ ] 🔴 **サブタスク 3.4: ShiftSlot 管理 API** [MVP]
  - [ ] 🔴 3.4.1: POST /api/v1/business-days/:business_day_id/shift-slots - ShiftSlot 作成 [MVP]
    - *リクエスト:* `{ slot_name, start_time, end_time, required_count, priority }`
    - *バリデーション:* 時刻の前後関係、同一営業日内での時刻重複チェック
    - *見積もり:* 2〜3時間
  - [ ] 🔴 3.4.2: GET /api/v1/business-days/:business_day_id/shift-slots - ShiftSlot 一覧取得 [MVP]
    - *レスポンス:* `{ shift_slots: [{ slot_id, slot_name, start_time, end_time, required_count, assigned_count, status }] }`
    - *見積もり:* 1〜2時間
  - [ ] 🔴 3.4.3: GET /api/v1/shift-slots/:slot_id - ShiftSlot 詳細取得 [MVP]
    - *レスポンス:* ShiftSlot情報 + 確定済みメンバー一覧 + 希望提出メンバー一覧
    - *見積もり:* 2時間
  - [ ] 🟡 3.4.4: PUT /api/v1/shift-slots/:slot_id - ShiftSlot 更新 [v1.1]
    - *更新可能:* slot_name, start_time, end_time, required_count, priority
    - *制約:* 確定済みシフトがある場合は時刻変更不可
    - *見積もり:* 2時間
  - [ ] 🟡 3.4.5: DELETE /api/v1/shift-slots/:slot_id - ShiftSlot 削除 [v1.1]
    - *制約:* 確定済みシフトがある場合は削除不可
    - *見積もり:* 1時間

- [ ] 🔴 **サブタスク 3.5: ShiftAssignment 管理 API** [MVP]
  - [ ] 🔴 3.5.1: POST /api/v1/shift-assignments - ShiftAssignment 確定 [MVP]
    - *リクエスト:* `{ slot_id, member_id, note }`
    - *処理:* ShiftAssignmentService 経由で通知・監査ログを記録（サブタスク2.5.1のフロー）
    - *レスポンス:* `{ assignment_id, notification_sent, status }`
    - *見積もり:* 2〜3時間
    - *⚠️ Application Service 経由必須:* handler から直接 Repository を触らないこと
  - [ ] 🔴 3.5.2: GET /api/v1/shift-assignments - ShiftAssignment 一覧取得 [MVP]
    - *クエリパラメータ:* `event_id`, `member_id`, `slot_id`, `status`, `start_date`, `end_date`
    - *用途:* メンバーごとのシフト一覧、営業日ごとの配置状況など
    - *見積もり:* 2〜3時間
  - [ ] 🔴 3.5.3: GET /api/v1/shift-assignments/:assignment_id - ShiftAssignment 詳細取得 [MVP]
    - *レスポンス:* 割り当て情報 + メンバー情報 + シフト枠情報 + 営業日情報
    - *見積もり:* 1時間
  - [ ] 🟡 3.5.4: PATCH /api/v1/shift-assignments/:assignment_id/status - ステータス変更 [v1.1]
    - *リクエスト:* `{ status: "confirmed" | "cancelled" | "pending", reason }`
    - *処理:* ステータス変更通知を発火（Application Service 経由）
    - *見積もり:* 2時間
  - [ ] 🟡 3.5.5: DELETE /api/v1/shift-assignments/:assignment_id - ShiftAssignment キャンセル [v1.1]
    - *処理:* 論理削除 + キャンセル通知発火（Application Service 経由）
    - *見積もり:* 1〜2時間

---

### 🟢 **親タスク 3C: Member/Availability 管理 + 可視化 API（v1.1 以降）**

メンバー管理、シフト希望提出、通知・監査ログの可視化 API を提供する。MVP後の優先実装対象。

- [ ] 🟡 **サブタスク 3.6: Member / Availability 関連 API（基本実装）** [v1.1]
  - [ ] 🟡 3.6.1: POST /api/v1/members - Member 作成 [v1.1]
    - *リクエスト:* `{ tenant_id, discord_user_id, display_name, email }`
    - *見積もり:* 1〜2時間
  - [ ] 🟡 3.6.2: GET /api/v1/members - Member 一覧取得 [v1.1]
    - *クエリパラメータ:* `tenant_id`, `is_active`
    - *見積もり:* 1時間
  - [ ] 🟡 3.6.3: GET /api/v1/members/:member_id - Member 詳細取得 [v1.1]
    - *レスポンス:* Member情報 + 役割 + 直近のシフト履歴
    - *見積もり:* 1〜2時間
  - [ ] 🟡 3.6.4: POST /api/v1/availabilities - シフト希望登録 [v1.1]
    - *リクエスト:* `{ member_id, slot_id, preference_level, note }`
    - *見積もり:* 2時間
  - [ ] 🟡 3.6.5: GET /api/v1/shift-slots/:slot_id/availabilities - シフト枠ごとの希望一覧 [v1.1]
    - *用途:* 誰が希望を出しているか確認
    - *見積もり:* 1時間

- [ ] 🟡 **サブタスク 3.7: Notification / Audit 可視化 API** [v1.1]
  - [ ] 🟡 3.7.1: GET /api/v1/notifications/logs - 通知ログ一覧取得 [v1.1]
    - *クエリパラメータ:* `recipient_id`, `event_type`, `start_date`, `end_date`, `status`
    - *用途:* 管理者が通知履歴を確認
    - *見積もり:* 2時間
  - [ ] 🟡 3.7.2: GET /api/v1/audit/logs - 監査ログ一覧取得 [v1.1]
    - *クエリパラメータ:* `entity_type`, `entity_id`, `actor_id`, `action`, `start_date`, `end_date`
    - *用途:* 誰がいつ何を変更したかの追跡
    - *見積もり:* 2〜3時間
  - [ ] 🟢 3.7.3: POST /api/v1/notifications/send - 手動通知送信（テスト/管理者用） [MVP]
    - *リクエスト:* `{ recipient_id, message, channel }`
    - *見積もり:* 1時間
    - *用途:* 開発・テスト時に通知フローを手動で確認するため、MVP に含める

- [ ] 🟡 **サブタスク 3.8: API ドキュメント生成** [v1.1]
  - [ ] 🟡 3.8.1: OpenAPI (Swagger) 定義ファイルの作成 [v1.1]
    - *詳細:* `backend/api/openapi.yaml`
    - *内容:* 全エンドポイント、リクエスト/レスポンススキーマ、エラーコード
    - *見積もり:* 3〜4時間
    - *段階的作成:* MVP API（3A/3B）の定義を優先、3C の API は後から追記
  - [ ] 🟡 3.8.2: Swagger UI のセットアップ [v1.1]
    - *エンドポイント:* GET /api/docs
    - *見積もり:* 1時間

---

### 🟢 **親タスク 4: Discord Bot 連携（薄いアダプタ実装）**

Backend API を利用した Discord Bot の実装。Bot はビジネスロジックを持たず、UIとしてのみ機能する。

- [ ] 🟢 **サブタスク 4.1: Backend API クライアントの実装**
  - [ ] 🟢 4.1.1: HTTPClient 基盤クラスの作成
    - *詳細:* `bot/src/services/backendClient.ts`
    - *機能:* 認証ヘッダー付与、エラーハンドリング、リトライロジック
    - *見積もり:* 2時間
  - [ ] 🟢 4.1.2: Event API クライアントの実装
    - *メソッド:* createEvent, getEvents, getEventDetail, updateEvent, deleteEvent
    - *見積もり:* 2時間
  - [ ] 🟢 4.1.3: ShiftSlot / ShiftAssignment API クライアントの実装
    - *メソッド:* createShiftSlot, getShiftSlots, confirmAssignment, getAssignments, cancelAssignment
    - *見積もり:* 2〜3時間
  - [ ] 🟢 4.1.4: Member / Availability API クライアントの実装
    - *メソッド:* registerMember, getMembers, submitAvailability, getAvailabilities
    - *見積もり:* 2時間

- [ ] 🟢 **サブタスク 4.2: Discord コマンドの実装（イベント管理）**
  - [ ] 🟢 4.2.1: `/event create` - イベント作成コマンド
    - *詳細:* `bot/src/commands/event/create.ts`
    - *UI:* Modal フォームでイベント情報を入力 → API 呼び出し → 結果を Embed で表示
    - *見積もり:* 3〜4時間
  - [ ] 🟢 4.2.2: `/event list` - イベント一覧表示コマンド
    - *UI:* ページネーション付き Embed、ボタンで詳細表示
    - *見積もり:* 2〜3時間
  - [ ] 🟢 4.2.3: `/event detail` - イベント詳細表示コマンド
    - *UI:* Event情報 + 営業日数 + シフト枠数を表示
    - *見積もり:* 1〜2時間
  - [ ] 🟢 4.2.4: `/event generate-days` - 営業日一括生成コマンド
    - *処理:* event_id を指定 → API 呼び出し → 生成結果を通知
    - *見積もり:* 1〜2時間

- [ ] 🟢 **サブタスク 4.3: Discord コマンドの実装（シフト管理）**
  - [ ] 🟢 4.3.1: `/shift create-slot` - シフト枠作成コマンド
    - *UI:* Modal で時刻・必要人数を入力
    - *見積もり:* 2〜3時間
  - [ ] 🟢 4.3.2: `/shift view` - シフト枠一覧表示コマンド
    - *UI:* 営業日を指定 → シフト枠一覧を Embed で表示（確定状況付き）
    - *見積もり:* 2〜3時間
  - [ ] 🟢 4.3.3: `/shift confirm` - シフト確定コマンド
    - *UI:* slot_id と member を選択 → 確定 → 通知送信結果を表示
    - *見積もり:* 2〜3時間
  - [ ] 🟢 4.3.4: `/shift cancel` - シフトキャンセルコマンド
    - *処理:* assignment_id を指定 → キャンセル理由入力 → API 呼び出し
    - *見積もり:* 2時間
  - [ ] 🟢 4.3.5: `/shift my-shifts` - 自分のシフト一覧表示コマンド
    - *UI:* 実行ユーザーの確定済みシフト一覧を表示
    - *見積もり:* 2時間

- [ ] 🟢 **サブタスク 4.4: Discord コマンドの実装（希望提出）**
  - [ ] 🟢 4.4.1: `/availability submit` - シフト希望提出コマンド
    - *UI:* 営業日を選択 → シフト枠一覧を表示 → 希望レベルを選択 → 提出
    - *見積もり:* 3〜4時間
  - [ ] 🟢 4.4.2: `/availability view` - 提出済み希望の確認コマンド
    - *UI:* 自分が提出した希望一覧を表示
    - *見積もり:* 2時間
  - [ ] 🟢 4.4.3: `/availability summary` - シフト枠ごとの希望集計コマンド（管理者用）
    - *UI:* 特定シフト枠に誰が希望を出しているかを表示
    - *見積もり:* 2〜3時間

- [ ] 🟢 **サブタスク 4.5: エンドツーエンド統合テスト**
  - [ ] 🟢 4.5.1: ローカル環境で backend + bot + db を起動
    - *確認:* backend /health が 200、bot が Discord に接続
    - *見積もり:* 1時間
  - [ ] 🟢 4.5.2: Discord 上でイベント作成から希望提出までの流れを実行
    - *シナリオ:*
      1. `/event create` でイベント作成
      2. `/event generate-days` で営業日生成
      3. `/shift create-slot` でシフト枠作成
      4. `/availability submit` で複数メンバーが希望提出
      5. `/availability summary` で希望を確認
      6. `/shift confirm` でシフト確定
      7. `/shift my-shifts` で確定内容確認
    - *見積もり:* 3〜4時間
  - [ ] 🟢 4.5.3: 通知・監査ログの記録確認
    - *確認:* DB に NotificationLog / AuditLog が正しく記録されているか
    - *見積もり:* 1時間

---

## ブロック中 / 要注意

*現時点でブロッカーなし*

---

## 完了したタスク

### ✅ Phase 1: マイグレーションとドメインエンティティ基盤（完了日: 2025-12-03）

**マイグレーション（6ファイル完成）:**
- [x] 001: tenants, events, recurring_patterns テーブル
- [x] 002: event_business_days テーブル  
- [x] 003: members, positions, shift_slots テーブル
- [x] 004: shift_plans, shift_assignments テーブル
- [x] 005: notification_logs, notification_templates テーブル
- [x] 006: audit_logs テーブル
- [x] マイグレーション実行ツール: `cmd/migrate/main.go`（filesystem ベース、Go 1.22+ 対応）

**ドメインエンティティ（全エンティティ完成）:**
- [x] Event エンティティ + Repository IF（`domain/event/`）
- [x] RecurringPattern エンティティ（Weekly/MonthlyDate/Custom対応、JSONB変換）
- [x] EventBusinessDay エンティティ + Repository IF（独立エンティティ、深夜営業対応）
- [x] ShiftSlot エンティティ + Repository IF（深夜営業判定、時刻フォーマット）
- [x] ShiftAssignment エンティティ + Repository IF（ステータス遷移、キャンセル機能）
- [x] Member エンティティ + Repository IF（真のMVP版、最小限実装）
- [x] Position エンティティ + Repository IF（役職管理、表示順対応）

**共通型:**
- [x] domain/common/id.go: ULID生成、型安全なID（TenantID, EventID, MemberID, PositionID）
- [x] domain/common/errors.go: ドメインエラー型定義

**ドメインテスト:**
- [x] domain/event/event_test.go: 13テストケース（正常系・異常系・操作系）
- [x] domain/shift/shift_slot_test.go: 12テストケース（深夜営業判定含む）

**設計決定事項:**
- [x] Multi-Tenant前提: 全エンティティ・全Repositoryメソッドで tenantID 必須
- [x] 論理削除: deleted_at カラム全テーブル実装
- [x] 型安全性: 専用ID型使用（生のstring禁止）
- [x] コンストラクタ強制: NewXXX() 経由でのみ生成可能
- [x] バリデーション: 不変条件チェック完備

**テスト結果:**
- ✅ Event ドメイン: 13テスト PASS (0.002s)
- ✅ ShiftSlot ドメイン: 12テスト PASS (0.003s)
- ✅ 全ドメイン層: PASS

---

### ✅ Phase 2: PostgreSQL リポジトリ実装（完了日: 2025-12-03）

**リポジトリ実装（全エンティティ対応）:**
- [x] EventRepository（`infra/db/event_repository.go`）
  - Save, FindByID, FindByTenantID, FindActiveByTenantID, Delete, ExistsByName
- [x] EventBusinessDayRepository（`infra/db/business_day_repository.go`）
  - Save, FindByID, FindByEventID, FindByEventIDAndDateRange, FindActiveByEventID, FindByTenantIDAndDate, ExistsByEventIDAndDate
- [x] ShiftSlotRepository（`infra/db/shift_slot_repository.go`）
  - Save, FindByID, FindByBusinessDayID, Delete
- [x] ShiftAssignmentRepository（`infra/db/shift_assignment_repository.go`）
  - Save, FindByID, FindBySlotID, FindByMemberID, UpdateStatus
- [x] MemberRepository（`infra/db/member_repository.go`）
  - Save, FindByID, FindByDiscordUserID, FindByEmail, FindAllByTenantID
- [x] PositionRepository（`infra/db/position_repository.go`）
  - Save, FindByID, FindByTenantID, FindActiveByTenantID, Delete

**統合テスト:**
- [x] EventRepository 統合テスト（`infra/db/event_repository_integration_test.go`）
  - DB接続確認、Event作成・取得・更新、BusinessDay連携

**技術スタック:**
- PostgreSQL 16 + pgx/v5（コネクションプーリング）
- ULID（oklog/ulid/v2）による型安全なID管理
- sql.NullTime, sql.NullString による nullable カラム対応

---

### ✅ Phase 3: REST API 基盤と真のMVP API（完了日: 2025-12-03）

**API 基盤（`internal/interface/rest/`）:**
- [x] router.go: go-chi/chi v5 ベースのルーティング
- [x] middleware.go: Logger, CORS, Auth（X-Tenant-ID/X-Member-ID ヘッダー）, Recover
- [x] response.go: 統一JSON レスポンス形式、ドメインエラー→HTTPステータス変換

**Event API（`event_handler.go`）:**
- [x] POST /api/v1/events - Event 作成
- [x] GET /api/v1/events - Event 一覧取得
- [x] GET /api/v1/events/:event_id - Event 詳細取得

**EventBusinessDay API（`business_day_handler.go`）:**
- [x] POST /api/v1/events/:event_id/business-days - 営業日手動作成
- [x] GET /api/v1/events/:event_id/business-days - 営業日一覧取得（日付範囲フィルタ対応）
- [x] GET /api/v1/business-days/:business_day_id - 営業日詳細取得

**APIサーバー（`cmd/server/main.go`）:**
- [x] Graceful shutdown 対応
- [x] 環境変数による設定（DATABASE_URL, PORT）
- [x] pgxpool によるコネクションプーリング

**動作確認:**
- ✅ ヘルスチェック: GET /health → `{"status":"ok"}`
- ✅ Event 作成・一覧・詳細取得
- ✅ BusinessDay 作成・一覧・詳細取得・日付範囲フィルタ
- ✅ Multi-Tenant 認証（X-Tenant-ID ヘッダー）
- ✅ エラーハンドリング（バリデーション、NotFound、Conflict）

**Go 環境:**
- Go 1.24.11（1.22 から自動アップグレード）
- go-chi/chi/v5, pgx/v5, oklog/ulid/v2

---

**次のステップ:** 
1. 🔴 親タスク 2: ShiftAssignment Application Service 実装（通知・監査 stub）
2. 🔴 親タスク 3B: ShiftSlot/Assignment 管理 API 実装
3. 🟢 親タスク 4: Discord Bot 連携

---

## データ設計の共通ポリシー

### 日付・時刻・タイムゾーンの扱い

**基本方針**:
- VRChatイベント運営は日本時間（JST）を前提とする
- ただし、将来的な国際化を見越した設計を採用

**具体的な型定義**:
- **営業日の日付**: `DATE` 型（例: `2025-12-05`）
  - テナントのローカル日付として扱う（JSTの日付）
- **シフト枠の時刻**: `TIME WITHOUT TIME ZONE` 型（例: `21:30:00`）
  - 深夜営業対応: 終了時刻が開始時刻より前の場合、日付をまたぐ営業として扱う（例: 21:30-02:00）
- **イベント発生日時**: `TIMESTAMP WITH TIME ZONE` 型
  - 通知送信日時、監査ログのタイムスタンプなど、正確な時刻記録が必要な場合
- **テナントのタイムゾーン**: tenants テーブルに `timezone` カラムを持つ（デフォルト: 'Asia/Tokyo'）
  - 将来的に海外テナントが追加された場合の拡張性を担保

**実装時の注意**:
- Go での時刻処理は `time.Time` を使用し、テナントのタイムゾーンを考慮
- API レスポンスの日時は ISO 8601 形式（`2025-12-05T21:30:00+09:00`）で返す
- フロントエンドは受信したタイムゾーン付き日時をユーザーのローカル時刻に変換

### Soft Delete（論理削除）の扱い

**基本方針**:
- 履歴保持が重要なエンティティ（Event, ShiftAssignment, Member など）は論理削除を採用
- 削除後に復旧の可能性があるデータは論理削除

**実装方法**:
- 全テーブルに `deleted_at TIMESTAMP WITH TIME ZONE NULL` カラムを追加
- `deleted_at IS NULL`: 有効なレコード
- `deleted_at IS NOT NULL`: 削除済みレコード

**クエリ時のルール**:
- **デフォルトの動作**: 一覧取得APIでは `WHERE deleted_at IS NULL` を自動的に適用
- **削除済みを含める場合**: クエリパラメータ `include_deleted=true` で明示的に指定
- **Repository 実装**: FindAll() は deleted_at IS NULL がデフォルト、FindAllIncludingDeleted() で削除済みを含める

**論理削除vs物理削除の使い分け**:
- **論理削除**: Event, Member, ShiftAssignment, NotificationTemplate
- **物理削除**: セッショントークン、一時データなど、履歴保持不要なもの
- **カスケード削除**: Event を削除 → EventBusinessDay も論理削除（FK制約で対応）

**ShiftAssignment の特殊ケース**:
- `assignment_status = 'cancelled'` と `deleted_at` の使い分け
  - **cancelled**: メンバーがキャンセルした（履歴として残し、UI にも表示可能）
  - **deleted_at**: 管理者が誤って作成した割り当てを削除（履歴から除外）

### エラーと冪等性の扱い

**冪等性の基本方針**:
- **POST リクエスト**: Idempotency Key ヘッダー（`Idempotency-Key: <UUID>`）を受け入れ、同じキーでの重複リクエストは同じレスポンスを返す
- **PUT / PATCH リクエスト**: リソースIDで一意に特定されるため、本質的に冪等
- **DELETE リクエスト**: 既に削除済みの場合も `204 No Content` を返す（エラーにしない）

**重複操作の扱い（ShiftAssignment の例）**:
- `POST /api/v1/shift-assignments` で同じ (slot_id, member_id) を二重に叩いた場合:
  - **DB 的**: UNIQUE 制約違反（部分一意インデックス）
  - **API 的**: `409 Conflict` を返し、既存の assignment_id を含める
  - **推奨**: クライアントは 409 を受け取ったら、既存リソースを使用

**ネットワークリトライ対応**:
- クライアントは `Idempotency-Key` を付与してリトライ
- サーバーは過去 24時間以内の同一キーのリクエストを記録し、同じレスポンスを返す
- Idempotency キャッシュは Redis または DB テーブル（`idempotency_keys`）で管理

---

## 実装詳細

### 🚨 重要な設計原則と注意事項（必読）

#### 1. Multi-Tenant 前提設計
- **全エンティティに `tenant_id` を必須フィールドとして含める**
- Repository の全メソッドで `tenant_id` を引数に取る（例: `FindByID(ctx, tenantID, eventID)`）
- テナント境界を越えたアクセスを防ぐため、WHERE 句に必ず `tenant_id` を含める
- テストデータも必ず `tenant_id` 付きで作成

#### 2. 集約境界の明確化
- **Event ≠ EventBusinessDay**: Event は「定義」、BusinessDay は「生成されたインスタンス」
- Event は EventBusinessDay を直接保持せず、ID参照のみ
- BusinessDay の編集（is_active変更など）は Event の不変条件を壊さない範囲に限定
- 集約を跨ぐ参照は ID ベースで行い、必要に応じて Repository 経由で取得

#### 3. 同時実行制御（ShiftAssignment）
- **排他制御戦略**: `SELECT ... FOR UPDATE` で該当 slot の assignments をロックしてから required_count チェック
- または DB の UNIQUE 制約違反を catch して `409 Conflict` を返す
- 同時確定による不整合を防ぐため、必ずトランザクション内で排他制御を実施

#### 4. トランザクション境界と通知の同期/非同期
- **v1 実装方針（同期処理）**:
  - トランザクション内: ShiftAssignment の作成・保存、AuditLog 記録
  - トランザクション外: Discord 送信、NotificationLog（送信結果）記録
- **将来的な拡張性**: NotificationEvent の「発火」と「送信」を分離、outbox パターン移行を見越す
- **事故パターン回避**: DBコミット成功→送信失敗の場合も NotificationLog に失敗を記録、idempotency key で二重送信防止

#### 5. FrequencyControl のパフォーマンス対策
- **必須インデックス**: `notification_logs` に `(recipient_id, sent_at)` の複合インデックス
- 「過去N分以内のログをrecipient_idで絞って数える」クエリが頻発するため、インデックスなしでは重くなる
- マイグレーションファイル作成時に必ず含めること

#### 6. DDD レイヤ保護ルール
- **状態変更系 API（POST/PUT/PATCH/DELETE）は必ず Application Service 経由**
- ハンドラから直接 Repository を呼び出して永続化してはいけない（集約の不変条件が破壊される）
- **参照系 API（GET）のみ**、パフォーマンス目的で Repository 直接アクセスを許可
- 例外的に Repository を直接触る場合は、コードレビューで合意を得ること

#### 7. MVP スコープの明確化

**MVP のゴール**: 「イベント作成 → 営業日生成 → シフト枠作成 → シフト確定 → 通知」の1本の流れを動かす

**MVP で実装する機能**:
- **ドメイン**: Event, EventBusinessDay, ShiftSlot, ShiftAssignment（手動割り当てのみ）, Member（基本CRUD）, Notification, AuditLog
- **API（親タスク 3A/3B）**:
  - Event 作成・一覧・詳細取得
  - EventBusinessDay 一覧・詳細取得・一括生成
  - ShiftSlot 作成・一覧・詳細取得
  - ShiftAssignment 確定（手動割り当て）・一覧・詳細取得
  - Member 作成・一覧・詳細取得（最低限）
  - Notification 手動送信（テスト用）
- **Discord Bot（親タスク 4 の基本部分）**:
  - `/event create` - イベント作成
  - `/event list` - イベント一覧
  - `/event generate-days` - 営業日一括生成
  - `/shift create-slot` - シフト枠作成
  - `/shift view` - シフト枠一覧
  - `/shift confirm` - シフト確定（管理者による手動割り当て）
  - `/shift my-shifts` - 自分のシフト一覧
  - `/member register` - メンバー登録（簡易版）

**MVP で実装しない機能（v1.1 以降）**:
- **ドメイン**: Availability（シフト希望）, 自動割り当てアルゴリズム
- **API（親タスク 3C）**:
  - Availability 登録・一覧取得
  - Member の詳細管理（ロール履歴、外部アカウント）
  - Event / ShiftSlot の更新・削除
  - 通知ログ・監査ログの可視化API
- **Discord Bot（親タスク 4 の高度な部分）**:
  - `/availability submit` - シフト希望提出
  - `/availability view` - 提出済み希望確認
  - `/availability summary` - 希望集計
  - `/shift auto-assign` - 自動割り当て

**MVP の割り切り**:
- メンバーは管理者が事前に DB に登録（または簡易登録コマンド）
- シフト確定は管理者が手動で member_id を指定して割り当て
- 希望収集フローは v1.1 で実装
- 更新・削除系APIは最小限（将来の拡張性は確保）

---

### アーキテクチャ概要

```
┌─────────────────────────────────────────────────┐
│ Discord Bot (Node/TS)                           │
│  - 薄いアダプタ: コマンド受付 → Backend API 呼び出し │
└─────────────────┬───────────────────────────────┘
                  │ HTTP (REST/RPC)
┌─────────────────┴───────────────────────────────┐
│ Backend (Go)                                    │
│  ┌─────────────────────────────────────────┐   │
│  │ Application Layer (UseCases/Services)   │   │
│  │  - ShiftAssignmentService               │   │
│  │  - NotificationService                  │   │
│  └────────────┬────────────────────────────┘   │
│               │                                 │
│  ┌────────────┴────────────────────────────┐   │
│  │ Domain Layer                            │   │
│  │  - Event / EventBusinessDay / ShiftSlot │   │
│  │  - ShiftAssignment                      │   │
│  │  - NotificationEvent / NotificationLog  │   │
│  │  - AuditLog                             │   │
│  │  - FrequencyControlPolicy               │   │
│  │  - BusinessDayGenerator (ドメインサービス) │   │
│  └────────────┬────────────────────────────┘   │
│               │                                 │
│  ┌────────────┴────────────────────────────┐   │
│  │ Infrastructure Layer                    │   │
│  │  - PostgreSQL Repositories              │   │
│  │  - DB Migrations                        │   │
│  └─────────────────────────────────────────┘   │
└─────────────────────────────────────────────────┘
```

### ドメイン境界と集約設計

**集約ルートと集約境界の明確化**:

- **Tenant 集約**: Tenant（集約ルート）
  - 全てのドメインの最上位境界
  
- **Event 集約**: Event（集約ルート）, RecurringPattern（エンティティ）
  - Event は EventBusinessDay を「直接保持せず、ID参照のみ」
  - EventBusinessDay は Event に属するが、独立したエンティティ（Event集約には含まれない）
  - 理由: Event は「営業の定義」、EventBusinessDay は「生成されたインスタンス」として分離
  
- **EventBusinessDay**: 独立したエンティティ
  - Event との関係は ID参照（event_id）
  - tenant_id を直接保持（テナント境界をDBレベルで表現）
  
- **ShiftSlot**: 独立したエンティティ
  - EventBusinessDay との関係は ID参照（business_day_id）
  - tenant_id を直接保持
  
- **ShiftPlan 集約**: ShiftPlan（集約ルート）, ShiftAssignment（エンティティ）
  - ShiftAssignment は ShiftPlan 内のエンティティ
  - ただし ShiftSlot, Member との関係は ID参照のみ
  
- **Member 集約**: Member（集約ルート）, MemberRole（エンティティ）, ExternalAccount（値オブジェクト）
  
- **Notification**: NotificationTemplate（エンティティ）, NotificationLog（エンティティ）
  - FrequencyControlPolicy はドメインサービスとして実装
  
- **Audit**: AuditLog（独立したエンティティ）

### データフロー（シフト確定通知の例）

1. Discord Bot: `/confirm-shift` コマンド受信
2. Bot → Backend API: `POST /shift-assignments` { slot_id, member_id }
3. Backend: ShiftAssignmentService.ConfirmShift()
   - ShiftAssignment を作成・保存
   - NotificationEvent を発火
   - FrequencyControlPolicy.Check() → スパム判定
   - NotificationLog を記録
   - AuditLog を記録
4. Backend → Bot: レスポンス返却
5. Bot → Discord: 結果をユーザーに通知

### REST API エンドポイント一覧（フロントエンド向け）

#### Event 管理
- `POST /api/v1/events` - イベント作成
- `GET /api/v1/events` - イベント一覧取得（ページネーション対応）
- `GET /api/v1/events/:event_id` - イベント詳細取得
- `PUT /api/v1/events/:event_id` - イベント更新
- `DELETE /api/v1/events/:event_id` - イベント削除

#### EventBusinessDay 管理
- `GET /api/v1/events/:event_id/business-days` - 営業日一覧取得
- `GET /api/v1/business-days/:business_day_id` - 営業日詳細取得
- `POST /api/v1/events/:event_id/generate-business-days` - 営業日一括生成
- `PATCH /api/v1/business-days/:business_day_id` - 営業日のアクティブ状態変更

#### ShiftSlot 管理
- `POST /api/v1/business-days/:business_day_id/shift-slots` - シフト枠作成
- `GET /api/v1/business-days/:business_day_id/shift-slots` - シフト枠一覧取得
- `GET /api/v1/shift-slots/:slot_id` - シフト枠詳細取得
- `PUT /api/v1/shift-slots/:slot_id` - シフト枠更新
- `DELETE /api/v1/shift-slots/:slot_id` - シフト枠削除

#### ShiftAssignment 管理
- `POST /api/v1/shift-assignments` - シフト確定
- `GET /api/v1/shift-assignments` - シフト割り当て一覧取得
- `GET /api/v1/shift-assignments/:assignment_id` - シフト割り当て詳細取得
- `PATCH /api/v1/shift-assignments/:assignment_id/status` - ステータス変更
- `DELETE /api/v1/shift-assignments/:assignment_id` - シフトキャンセル

#### Member 管理
- `POST /api/v1/members` - メンバー作成
- `GET /api/v1/members` - メンバー一覧取得
- `GET /api/v1/members/:member_id` - メンバー詳細取得

#### Availability 管理
- `POST /api/v1/availabilities` - シフト希望登録
- `GET /api/v1/shift-slots/:slot_id/availabilities` - シフト枠ごとの希望一覧

#### Notification / Audit
- `GET /api/v1/notifications/logs` - 通知ログ一覧取得
- `GET /api/v1/audit/logs` - 監査ログ一覧取得
- `POST /api/v1/notifications/send` - 手動通知送信（管理者/テスト用）

#### その他
- `GET /api/health` - ヘルスチェック
- `GET /api/docs` - Swagger UI（API ドキュメント閲覧）

### テスト戦略

- **単体テスト**: 各エンティティ・値オブジェクトの不変条件、バリデーションロジック
- **統合テスト**: ドメインサービス + リポジトリの組み合わせ（DB接続あり）
- **シナリオテスト**: Event 作成 → ShiftAssignment 確定 → 通知・監査の一連の流れ
- **API テスト**: 各エンドポイントのリクエスト/レスポンス検証（Postman / HTTPie）
- **E2Eテスト**: Discord Bot → Backend API → DB までの全体フロー

---

## 関連ファイル

### ドキュメント

- `docs/domain/10_tenant-and-event/ドメインモデル.md` - Event / RecurringPattern / EventBusinessDay の仕様
- `docs/domain/50_shift-plan-and-assignment/ドメインモデル.md` - ShiftSlot / ShiftAssignment の仕様
- `docs/domain/60_notification-and-reminder/ドメインモデル.md` - Notification / AuditLog の仕様

### Backend（予定）

#### ドメイン層
- `backend/internal/domain/event/event.go` - Event エンティティ
- `backend/internal/domain/event/recurring_pattern.go` - RecurringPattern 値オブジェクト
- `backend/internal/domain/event/event_business_day.go` - EventBusinessDay エンティティ
- `backend/internal/domain/event/business_day_generator.go` - 営業日生成ドメインサービス
- `backend/internal/domain/event/repository.go` - Event リポジトリIF
- `backend/internal/domain/shift/shift_slot.go` - ShiftSlot エンティティ
- `backend/internal/domain/shift/shift_assignment.go` - ShiftAssignment エンティティ
- `backend/internal/domain/shift/repository.go` - Shift リポジトリIF
- `backend/internal/domain/notification/notification_event.go` - NotificationEvent エンティティ
- `backend/internal/domain/notification/notification_log.go` - NotificationLog エンティティ
- `backend/internal/domain/notification/frequency_control_policy.go` - FrequencyControl ドメインサービス
- `backend/internal/domain/notification/repository.go` - Notification リポジトリIF
- `backend/internal/domain/audit/audit_log.go` - AuditLog エンティティ
- `backend/internal/domain/audit/repository.go` - AuditLog リポジトリIF
- `backend/internal/domain/member/member.go` - Member エンティティ
- `backend/internal/domain/member/repository.go` - Member リポジトリIF
- `backend/internal/domain/availability/availability.go` - Availability エンティティ
- `backend/internal/domain/availability/repository.go` - Availability リポジトリIF

#### アプリケーション層
- `backend/internal/app/shift_assignment_service.go` - シフト確定ユースケース
- `backend/internal/app/event_service.go` - イベント管理ユースケース
- `backend/internal/app/notification_service.go` - 通知送信ユースケース

#### インターフェース層（REST API）
- `backend/internal/interface/rest/router.go` - ルーティング定義
- `backend/internal/interface/rest/middleware.go` - CORS, ロギング, エラーハンドリング
- `backend/internal/interface/rest/response.go` - レスポンスヘルパー
- `backend/internal/interface/rest/validator.go` - リクエストバリデーション
- `backend/internal/interface/rest/event_handler.go` - Event API ハンドラー
- `backend/internal/interface/rest/business_day_handler.go` - EventBusinessDay API ハンドラー
- `backend/internal/interface/rest/shift_slot_handler.go` - ShiftSlot API ハンドラー
- `backend/internal/interface/rest/shift_assignment_handler.go` - ShiftAssignment API ハンドラー
- `backend/internal/interface/rest/member_handler.go` - Member API ハンドラー
- `backend/internal/interface/rest/availability_handler.go` - Availability API ハンドラー
- `backend/internal/interface/rest/notification_handler.go` - Notification API ハンドラー
- `backend/internal/interface/rest/audit_handler.go` - AuditLog API ハンドラー

#### インフラ層
- `backend/internal/infra/db/event_repository.go` - Event リポジトリ実装
- `backend/internal/infra/db/business_day_repository.go` - EventBusinessDay リポジトリ実装
- `backend/internal/infra/db/shift_repository.go` - Shift リポジトリ実装
- `backend/internal/infra/db/notification_repository.go` - Notification リポジトリ実装
- `backend/internal/infra/db/audit_repository.go` - AuditLog リポジトリ実装
- `backend/internal/infra/db/member_repository.go` - Member リポジトリ実装
- `backend/internal/infra/db/availability_repository.go` - Availability リポジトリ実装
- `backend/internal/infra/db/migrations/001_create_events_table.sql`
- `backend/internal/infra/db/migrations/002_create_event_business_days_table.sql`
- `backend/internal/infra/db/migrations/003_create_shift_slots_table.sql`
- `backend/internal/infra/db/migrations/004_create_shift_assignments_table.sql`
- `backend/internal/infra/db/migrations/005_create_notification_tables.sql`
- `backend/internal/infra/db/migrations/006_create_audit_logs_table.sql`
- `backend/internal/infra/db/migrations/007_create_members_table.sql`
- `backend/internal/infra/db/migrations/008_create_availabilities_table.sql`

#### API ドキュメント
- `backend/api/openapi.yaml` - OpenAPI 3.0 定義ファイル
- `backend/api/README.md` - API 利用ガイド

### Discord Bot（予定）

#### Backend API クライアント
- `bot/src/services/backendClient.ts` - HTTPClient 基盤クラス
- `bot/src/services/api/eventApi.ts` - Event API クライアント
- `bot/src/services/api/shiftApi.ts` - Shift API クライアント
- `bot/src/services/api/memberApi.ts` - Member API クライアント
- `bot/src/services/api/availabilityApi.ts` - Availability API クライアント

#### Discord コマンド（イベント管理）
- `bot/src/commands/event/create.ts` - `/event create` イベント作成
- `bot/src/commands/event/list.ts` - `/event list` イベント一覧
- `bot/src/commands/event/detail.ts` - `/event detail` イベント詳細
- `bot/src/commands/event/generateDays.ts` - `/event generate-days` 営業日生成

#### Discord コマンド（シフト管理）
- `bot/src/commands/shift/createSlot.ts` - `/shift create-slot` シフト枠作成
- `bot/src/commands/shift/view.ts` - `/shift view` シフト枠一覧
- `bot/src/commands/shift/confirm.ts` - `/shift confirm` シフト確定
- `bot/src/commands/shift/cancel.ts` - `/shift cancel` シフトキャンセル
- `bot/src/commands/shift/myShifts.ts` - `/shift my-shifts` 自分のシフト一覧

#### Discord コマンド（希望提出）
- `bot/src/commands/availability/submit.ts` - `/availability submit` 希望提出
- `bot/src/commands/availability/view.ts` - `/availability view` 提出済み希望確認
- `bot/src/commands/availability/summary.ts` - `/availability summary` 希望集計（管理者用）

#### ユーティリティ
- `bot/src/utils/embedBuilder.ts` - Discord Embed 生成ヘルパー
- `bot/src/utils/pagination.ts` - ページネーション UI ヘルパー
- `bot/src/utils/errorHandler.ts` - エラーハンドリング共通処理

---

## 次回のための改善メモ

### タスク計画時のレビューで得られた知見
1. **MVPマーカーの導入**: 大規模タスクでは `[MVP]` / `[v1.1]` / `[Nice-to-have]` マーカーで優先度を明示する
2. **親タスクの分割基準**: サブタスクが10個を超えたら、親タスクを3つ程度に分割する（3A/3B/3Cなど）
3. **集約境界の明文化**: エンティティ定義時に「どのエンティティが集約ルートか」「ID参照 vs 直接保持」を必ず明記
4. **同時実行制御の方針**: データ競合が発生しうる箇所は、実装前に排他戦略（DBロック/楽観ロック/ユニーク制約）を決める
5. **トランザクション境界の文書化**: 「どこまでを1トランザクションに含めるか」を Application Service 実装前に明記
6. **パフォーマンス対策の先行設計**: 頻発するクエリのインデックスは、マイグレーション作成時に必ず含める（後付けにしない）
7. **DDD レイヤ保護ルールの明示**: API ハンドラが Repository を直接触る問題を防ぐため、「状態変更は Service 経由必須」をルール化
8. **Multi-Tenant 前提の徹底**: 全 Repository メソッドで tenant_id を必須引数にし、将来的な本格認証への移行を容易にする

### 今後のタスクファイル作成時のチェックリスト
- [ ] MVPマーカーで優先度を明示したか？
- [ ] 親タスクが大きすぎないか？（サブタスク10個以上なら分割）
- [ ] 集約境界とトランザクション境界を明記したか？
- [ ] 同時実行制御が必要な箇所を洗い出したか？
- [ ] パフォーマンスに影響するインデックスを設計に含めたか？
- [ ] API の DDD レイヤ保護ルールを明記したか？
- [ ] Multi-Tenant 対応（tenant_id必須）を全箇所に適用したか？
- [ ] **日付・時刻・タイムゾーンのポリシーを明記したか？**
- [ ] **Soft Delete の戦略を決めたか？**
- [ ] **冪等性の扱いを定義したか？**

### レビューサイクル2で得られた追加の知見

9. **ドメインドキュメント調査の重要性**: レビュー指摘を受ける前に、必ず既存のドメインドキュメントを読み込む
   - 今回のケース: tenant_id の配置、Event/BusinessDay の関係は既にドキュメントに明記されていた
   - レビュアーの「うのみにせず、ドメインドキュメントで確認」という指摘は的確

10. **テーブル定義の具体化**: 「カラム名だけ」ではなく、型・制約・インデックスまで明記する
    - 特に Multi-Tenant 設計では、全テーブルの tenant_id と FK戦略を統一する
    - インデックスは「よく使うクエリ」から逆算して設計する（後付けにしない）

11. **集約境界の"言葉の曖昧さ"**: 「Event集約にBusinessDayを含む」という表現の危険性
    - 「含む」= 内部に保持 vs 「含む」= 管理下にある（ID参照）の2つの解釈がある
    - ドメインモデル図を見て、実際の関係性（1対多、FK）を確認する

12. **required_count 制御のようなビジネスルール**: DBレベルで保証できない制約は明記する
    - 「同じ枠に required_count を超えて割り当てない」は CHECK 制約では書けない
    - Application Service での排他制御とトランザクション設計が必須

13. **MVPスコープの"依存関係"**: 「希望収集なしでシフト確定」は可能
    - 最初から全機能を作らず、「管理者が直接割り当てる」という簡易版でMVPを通す
    - Member基本CRUDは必須だが、Availabilityは後回しでも動く

14. **タイムゾーン・Soft Delete・冪等性のような横断的関心事**: 実装前に共通ポリシーを決める
    - これらは後から「各テーブルで違う方針」になると収拾がつかなくなる
    - 最初に「データ設計の共通ポリシー」セクションとして明文化

*タスク進行中に気づいた追加の改善点をここに記録する*

---

## ✅ 設計決定事項（確定済み）

以下の設計方針で実装を進めます。

### ドメイン設計
1. **RecurringPattern の保存方法**: ✅ **専用テーブル + JSONB のハイブリッド**
   - `recurring_patterns` テーブルを新設
   - カラム: `pattern_id (ULID)`, `tenant_id`, `event_id`, `pattern_type (enum)`, `config (JSONB)`, `created_at`, `updated_at`
   - `config` にパターン内容（曜日リスト・日付リスト・例外日など）を JSONB で柔軟に持つ
   - Event とは 1:1 を基本とし、`UNIQUE(tenant_id, event_id)` を設定
   - **理由**: RecurringPattern はエンティティとして ID を持つ必要がある一方、パターンの中身は将来増減しやすいため、テーブル（ID管理）+ JSONB（柔軟性）のハイブリッドがバランス良い

2. **FrequencyControl の閾値**: ✅ **デフォルトは「10分以内に5件以上」**
   - デフォルト値: `WINDOW_MINUTES=10`, `MAX_NOTIFICATIONS=5`
   - 環境変数で上書き可能: `FREQ_CTRL_WINDOW_MINUTES`, `FREQ_CTRL_MAX_NOTIFICATIONS`
   - 将来的にテナント単位の設定テーブル（`notification_policies`）への拡張を見据える

3. **Notification 送信の実装タイミング**: ✅ **親タスク2は発火・ポリシー・ログまで、実送信は親タスク4**
   - **親タスク2**: NotificationEvent 発火、FrequencyControlPolicy チェック、NotificationLog 記録（`delivery_status = 'pending'`）
   - **親タスク4**: NotificationSender インターフェース実装、Discord への実送信、ステータス更新
   - Application 層に `NotificationSender interface { Send(ctx, evt) error }` を定義し、実装は DI

### API 設計
4. **認証・認可の実装範囲**: ✅ **v1 は簡易ヘッダー認証**
   - HTTP ヘッダーで `X-Tenant-ID: <ULID>`, `X-Member-ID: <ULID>` を受け取る
   - Repository / Service は必ず `tenant_id` を引数に持つ設計を徹底
   - 将来 JWT/OAuth2 を導入しても、「トークン → tenant_id/member_id 復元」層を差し替えるだけで対応可能
   - REST ハンドラでヘッダーをパースし、`context.Context` に埋め込んで Service に渡す

5. **API バージョニング**: ✅ **`/api/v1/` で固定**
   - v1 の間は変更しない
   - 将来大きな互換性ブレイクがある場合のみ `/api/v2/` を並列追加

6. **ページネーションのデフォルト値**: ✅ **デフォルト 20 / 最大 100 / 1-indexed**
   - クエリパラメータ: `page` (1から開始), `limit` (1〜100)
   - デフォルト: `limit=20`
   - REST ハンドラ共通処理で範囲チェック（`page<=0` → 1, `limit>100` → 100）

7. **レート制限（Rate Limiting）**: ✅ **v1 では実装しない**
   - 理由: 初期利用ユーザー数が少なく、FrequencyControl やビジネスルールで十分対応可能
   - DDD/ドメインの完成度を優先
   - v1.1 以降で Redis + middleware によるレート制限を検討

### テスト・開発環境
8. **テスト用の初期データ**: ✅ **Go コードベースのシード（+ 必要なら SQL）**
   - `cmd/seed` などの Go コマンドで「マイグレーション適用済みDBに対してシードを流す」
   - 必要なら `scripts/seed_test_data.sql` も併用可能だが、メインは Go に寄せる
   - **最低限のシード内容**: Tenant 1件、Member 3〜5件、Event 1件、RecurringPattern 1件、EventBusinessDay 3〜7日分、ShiftSlot 1営業日あたり 2〜3枠
   - **利点**: テストコードからシードロジックを使い回せる、条件分岐しやすい

9. **OpenAPI ドキュメント生成**: ✅ **v1 は手動で `openapi.yaml` を記述**
   - `backend/api/openapi.yaml` をソースオブトゥルースとする
   - Swagger UI 用エンドポイント: `GET /api/docs` (HTML), `GET /api/openapi.yaml` (YAML)
   - CI で `openapi-generator-cli validate` を実行し、構文ミスを早期検出
   - v1.1 以降で Go annotations → 自動生成への移行を検討

### 実装方針の確認（レビューフィードバック反映済み）

#### ドメインドキュメント調査結果に基づく設計決定

10. **tenant_id の配置戦略**: ✅ 全テーブルに tenant_id を追加（ドメインドキュメントで明示的に指定）
    - EventBusinessDay, ShiftSlot, ShiftAssignment, NotificationLog, AuditLog 全てに tenant_id カラムを配置
    - 理由: `docs/domain/10_tenant-and-event/ドメインモデル.md` 行290で「テナントIDの伝播: イベント営業日は tenant_id を直接保持」と明記
    - FK戦略: ULID（グローバルユニーク）+ tenant_id による二重チェック
    - 複合PK は採用しない（アプリケーション層での WHERE tenant_id 強制で対応）

11. **親タスク3の分割**: ✅ 3A（MVP: Event/BusinessDay）/ 3B（MVP: Shift）/ 3C（v1.1: Member/Availability/可視化）に分割済み

12. **Event / EventBusinessDay の集約境界**: ✅ 修正完了
    - Event 集約: Event（集約ルート） + RecurringPattern（エンティティ）
    - EventBusinessDay: 独立したエンティティ（Event に属するが、Event集約には含まれない）
    - 理由: 「Event は営業の定義、EventBusinessDay は生成されたインスタンス」（サブタスク 1.3.1 参照）

13. **ShiftAssignment の同時実行制御**: ✅ サブタスク 2.1.1, 2.1.2 に明記済み
    - `SELECT ... FOR UPDATE` で slot 単位でロック
    - 部分一意インデックス: `UNIQUE(slot_id, member_id, assignment_status) WHERE assignment_status = 'confirmed'`
    - required_count 制御は Application Service で実装（DBレベルでは保証不可）

14. **通知のトランザクション境界**: ✅ サブタスク 2.5.1 にトランザクションスコープと将来的なoutbox対応を明記済み

15. **FrequencyControl のインデックス**: ✅ サブタスク 2.2.3 で notification_logs に `INDEX(recipient_id, sent_at)` を明記

16. **API の DDD レイヤ保護**: ✅ サブタスク 3.1.1 に「状態変更は Application Service 経由必須」ルールを明記済み

17. **日付・時刻・タイムゾーン**: ✅ 「データ設計の共通ポリシー」セクションに追加
    - 営業日: DATE 型（JSTの日付）
    - シフト時刻: TIME WITHOUT TIME ZONE（深夜営業対応）
    - イベント発生日時: TIMESTAMP WITH TIME ZONE
    - テナントごとに timezone カラム（デフォルト: Asia/Tokyo）

18. **Soft Delete**: ✅ 「データ設計の共通ポリシー」セクションに追加
    - 全テーブルに `deleted_at` カラム
    - デフォルトで deleted_at IS NULL のレコードのみ返す
    - ShiftAssignment の特殊ケース: cancelled（履歴） vs deleted（完全削除）

19. **インデックス設計**: ✅ 各テーブル定義に必須インデックスを明記
    - events: (tenant_id, is_active), (tenant_id, event_name)
    - event_business_days: (tenant_id, target_date), (event_id, target_date)
    - shift_slots: (tenant_id, business_day_id), (business_day_id, start_time)
    - shift_assignments: (tenant_id, member_id, assignment_status), (slot_id, assignment_status)
    - notification_logs: (recipient_id, sent_at), (tenant_id, business_day_id, notification_type)
    - audit_logs: (tenant_id, entity_type, entity_id), (tenant_id, actor_id, timestamp)

20. **エラーと冪等性**: ✅ 「データ設計の共通ポリシー」セクションに追加
    - Idempotency-Key ヘッダーによる冪等性保証
    - 409 Conflict で既存リソースIDを返す
    - DELETE は既に削除済みでも 204 を返す

21. **MVP スコープと Bot の依存関係**: ✅ 「MVP スコープの明確化」セクションで整理
    - MVP: Member基本CRUD は含む、Availability（希望収集）は v1.1
    - Bot: 管理者による手動割り当てのみ MVP、自動割り当ては v1.1
    - 割り切り: 希望収集フローなしでシフト確定までを通す

---

## 📋 設計決定サマリ v1（コピペ用）

以下は実装時の参照用サマリです。必要に応じて `docs/architecture/決定事項_v1.md` として保存できます。

### RecurringPattern 保存方式
- `recurring_patterns` テーブルを新設する
- カラム例: pattern_id (ULID), tenant_id, event_id, pattern_type, config (JSONB), created_at, updated_at
- Event とは 1:1 を基本とし、UNIQUE(tenant_id, event_id) を張る
- EventBusinessDay.recurring_pattern_id はこのテーブルを参照

### FrequencyControl 閾値
- デフォルト値: 「10分以内に5件以上」でスパム判定
- 環境変数で上書き可能にする:
  - FREQ_CTRL_WINDOW_MINUTES
  - FREQ_CTRL_MAX_NOTIFICATIONS

### Notification 送信タイミング
- 親タスク2: NotificationEvent 発火 + FrequencyControl チェック + NotificationLog 記録まで
- 親タスク4: Discord 等への実送信を行う NotificationSender 実装を追加
- Application 層に NotificationSender インターフェースを定義し、実装は DI する

### 認証・認可（v1）
- 簡易ヘッダー認証を採用する:
  - X-Tenant-ID, X-Member-ID ヘッダーを使用
- Repository / Service は引数で tenant_id を必須とする
- 将来的に JWT/OAuth2 に置き換え可能な構成とする

### API バージョニング
- ベースパスは `/api/v1/` で固定
- 互換性ブレイク時のみ `/api/v2/` を並行提供する

### ページネーション
- クエリパラメータ: page (1-indexed), limit
- デフォルト: limit=20
- 上限: limit=100。超えたら 100 に丸める

### レート制限
- v1 ではアプリ側レート制限は導入しない
- 通知スパムは FrequencyControlPolicy で防ぐ
- v1.1 以降に Redis + rate limiter middleware を導入検討

### テスト用初期データ
- Go コードベースのシードコマンド（例: cmd/seed）を用意する
- 最低限のシード内容:
  - Tenant: 1件
  - Member: 数件
  - Event + RecurringPattern: 1セット
  - EventBusinessDay / ShiftSlot: 数日分 / 数枠

### OpenAPI ドキュメント
- v1 では `backend/api/openapi.yaml` を手動で記述する
- Swagger UI はこの YAML を読み込んで表示する
- CI で OpenAPI のバリデーションを行う


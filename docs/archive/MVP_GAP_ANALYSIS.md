# MVP ギャップ分析レポート

**作成日**: 2025-12-14
**最終更新**: 2025-12-14
**対象**: vrc-shift-scheduler

---

## 1. 仕様（MVP）サマリ

### 1.1 テナント/管理者
- テナント＝契約単位（店舗）
- ログインするのは「店長/副店長（管理者）」のみ
- 管理操作はテナント内にスコープ（他テナントは見えない）

### 1.2 メンバーマスタ
- 管理者がメンバーマスタを登録/更新する
- 将来的に一括登録（CSV等）を行う想定
- 公開回答は必ずメンバーマスタから選択（プルダウン）
- 公開回答ページでメンバー追加は不可

### 1.3 公開トークン
- 公開URLトークンは **UUID v4 固定**
- nanoid等の選択肢は採用しない
- tokenはUUID形式でバリデーション

### 1.4 出欠確認 / 日程調整
- 出欠確認（AttendanceCollection）と日程調整（DateSchedule）は公開回答ページを共通化
- 公開API（token）と管理API（認証）を分離
- 出欠回答は「同一 collection_id × member_id は上書き」を保証
- データは過去分も永続保持（削除しない運用）

---

## 2. 現在実装の実態

### 2.1 リポジトリ構成

```
vrc-shift-scheduler/
├── backend/              # Go製バックエンドAPI
│   ├── cmd/
│   │   ├── api/          # エントリポイント（古い）
│   │   ├── server/       # エントリポイント（新しい）
│   │   ├── migrate/      # マイグレーションツール
│   │   └── seed/         # シードデータ
│   └── internal/
│       ├── domain/       # ドメインモデル
│       ├── app/          # アプリケーションサービス
│       ├── infra/db/     # リポジトリ実装
│       └── interface/rest/  # RESTハンドラー
├── web-frontend/         # React + Vite フロントエンド
│   └── src/
│       ├── pages/        # ページコンポーネント
│       ├── components/   # 共通コンポーネント
│       └── lib/api/      # APIクライアント
├── bot/                  # Discord Bot (未使用)
└── docs/                 # ドキュメント
```

### 2.2 既存テーブル（マイグレーション実装済み）

| テーブル | ファイル | 状態 |
|---------|----------|------|
| tenants | 001_*.up.sql | ✅ 実装済み |
| events | 001_*.up.sql | ✅ 実装済み |
| recurring_patterns | 001_*.up.sql | ✅ 実装済み |
| event_business_days | 002_*.up.sql | ✅ 実装済み |
| members | 003_*.up.sql | ✅ 実装済み |
| positions | 003_*.up.sql | ✅ 実装済み |
| shift_slots | 003_*.up.sql | ✅ 実装済み |
| shift_plans | 004_*.up.sql | ✅ 実装済み |
| shift_assignments | 004_*.up.sql | ✅ 実装済み |
| notification_logs | 005_*.up.sql | ✅ 実装済み |
| notification_templates | 005_*.up.sql | ✅ 実装済み |
| audit_logs | 006_*.up.sql | ✅ 実装済み |
| admins | 007_*.up.sql | ✅ 実装済み |
| attendance_collections | 008_*.up.sql | ✅ 実装済み |
| attendance_responses | 008_*.up.sql | ✅ 実装済み |
| date_schedules | 009_*.up.sql | ✅ 実装済み |
| schedule_candidates | 009_*.up.sql | ✅ 実装済み |
| schedule_responses | 009_*.up.sql | ✅ 実装済み |
| invitations | 011_*.up.sql | ✅ 実装済み |

### 2.3 未実装テーブル（MVP必須）

| テーブル | 目的 | 状態 |
|---------|------|------|
| （なし） | - | ✅ すべて実装済み |

### 2.4 バックエンドAPI（実装済み）

| エンドポイント | ファイル | 状態 |
|---------------|----------|------|
| `POST /api/v1/events` | event_handler.go | ✅ 実装済み |
| `GET /api/v1/events` | event_handler.go | ✅ 実装済み |
| `GET /api/v1/events/{id}` | event_handler.go | ✅ 実装済み |
| `POST /api/v1/events/{id}/business-days` | business_day_handler.go | ✅ 実装済み |
| `GET /api/v1/events/{id}/business-days` | business_day_handler.go | ✅ 実装済み |
| `GET /api/v1/business-days/{id}` | business_day_handler.go | ✅ 実装済み |
| `POST /api/v1/business-days/{id}/shift-slots` | shift_slot_handler.go | ✅ 実装済み |
| `GET /api/v1/business-days/{id}/shift-slots` | shift_slot_handler.go | ✅ 実装済み |
| `POST /api/v1/members` | member_handler.go | ✅ 実装済み |
| `GET /api/v1/members` | member_handler.go | ✅ 実装済み |
| `GET /api/v1/members/{id}` | member_handler.go | ✅ 実装済み |
| `POST /api/v1/shift-assignments` | shift_assignment_handler.go | ✅ 実装済み |
| `GET /api/v1/shift-assignments` | shift_assignment_handler.go | ✅ 実装済み |
| `POST /api/v1/auth/login` | auth_handler.go | ✅ 実装済み |
| `POST /api/v1/attendance/collections` | attendance_handler.go | ✅ 実装済み |
| `GET /api/v1/attendance/collections/{id}` | attendance_handler.go | ✅ 実装済み |
| `POST /api/v1/attendance/collections/{id}/close` | attendance_handler.go | ✅ 実装済み |
| `GET /api/v1/attendance/collections/{id}/responses` | attendance_handler.go | ✅ 実装済み |
| `GET /api/v1/public/attendance/{token}` | attendance_handler.go | ✅ 実装済み |
| `POST /api/v1/public/attendance/{token}/responses` | attendance_handler.go | ✅ 実装済み |
| `POST /api/v1/schedules` | schedule_handler.go | ✅ 実装済み |
| `GET /api/v1/schedules/{id}` | schedule_handler.go | ✅ 実装済み |
| `POST /api/v1/schedules/{id}/decide` | schedule_handler.go | ✅ 実装済み |
| `POST /api/v1/schedules/{id}/close` | schedule_handler.go | ✅ 実装済み |
| `GET /api/v1/schedules/{id}/responses` | schedule_handler.go | ✅ 実装済み |
| `GET /api/v1/public/schedules/{token}` | schedule_handler.go | ✅ 実装済み |
| `POST /api/v1/public/schedules/{token}/responses` | schedule_handler.go | ✅ 実装済み |
| `POST /api/v1/invitations` | invitation_handler.go | ✅ 実装済み |
| `POST /api/v1/invitations/accept/{token}` | invitation_handler.go | ✅ 実装済み |

### 2.5 バックエンドAPI（未実装）

| エンドポイント | 目的 | 状態 |
|---------------|------|------|
| `POST /api/v1/setup` | 初回セットアップ | ❌ 未実装 |
| `PUT /api/v1/members/{id}` | メンバー更新 | ❌ 未実装 |
| `DELETE /api/v1/members/{id}` | メンバー削除/無効化 | ❌ 未実装 |
| `GET /api/v1/attendance/collections` | 出欠確認一覧 | ❌ 未実装 |
| `GET /api/v1/schedules` | 日程調整一覧 | ❌ 未実装 |

### 2.6 フロントエンド画面（実装済み）

| 画面 | ファイル | 状態 |
|------|----------|------|
| 管理者ログイン | AdminLogin.tsx | ✅ 実装済み（JWT認証） |
| イベント一覧 | EventList.tsx | ✅ 実装済み |
| 営業日一覧 | BusinessDayList.tsx | ✅ 実装済み |
| シフト枠一覧 | ShiftSlotList.tsx | ✅ 実装済み |
| シフト割り当て | AssignShift.tsx | ✅ 実装済み（プルダウン選択） |
| マイシフト | MyShifts.tsx | ✅ 実装済み |
| メンバー一覧 | Members.tsx | ✅ 実装済み |
| 出欠確認一覧 | AttendanceList.tsx | ✅ 実装済み |
| 日程調整一覧 | ScheduleList.tsx | ✅ 実装済み |
| 管理者招待 | AdminInvitation.tsx | ✅ 実装済み |
| 招待受理 | AcceptInvitation.tsx | ✅ 実装済み |
| 公開回答ページ（出欠） | public/AttendanceResponse.tsx | ✅ 実装済み |
| 公開回答ページ（日程） | public/ScheduleResponse.tsx | ✅ 実装済み |

### 2.7 フロントエンド画面（未実装）

| 画面 | 目的 | 状態 |
|------|------|------|
| メンバー編集/削除 | メンバーの更新・削除UI | ❌ 未実装（一覧のみ） |
| 出欠確認一覧取得 | 管理者向け一覧表示 | ⚠️ 画面はあるがAPI一覧取得が未実装 |
| 日程調整一覧取得 | 管理者向け一覧表示 | ⚠️ 画面はあるがAPI一覧取得が未実装 |

### 2.8 認証の現状

**現在の実装（middleware.go）**:
- ✅ JWT認証を実装済み（`Authorization: Bearer <token>`）
- ✅ JWTトークンに `admin_id`, `tenant_id`, `role` を含む
- ⚠️ フォールバックとして `X-Tenant-ID` ヘッダー認証も残存（段階移行用）
- ✅ 認証ミドルウェアでJWT検証を優先

**バックエンド認証**:
- ✅ `admins` テーブル実装済み（email + password_hash + role）
- ✅ `POST /api/v1/auth/login` でJWTトークン発行
- ✅ bcryptによるパスワードハッシュ化
- ✅ ロール: `owner`（店長） / `manager`（副店長）

**フロントエンドの認証（AdminLogin.tsx）**:
- ✅ メール+パスワードでログイン
- ✅ localStorage に `auth_token`（JWT）を保存
- ✅ トークン有効期限チェック実装済み
- ✅ ログイン状態に応じたルーティング保護

### 2.9 PublicToken の状態

**✅ 実装済み（`backend/internal/domain/common/id.go`）**:
- `PublicToken` 型定義
- UUID v4形式で生成（`uuid.New().String()`）
- バリデーション関数（`ValidatePublicToken`）
- ✅ 出欠確認・日程調整テーブルで使用中
- ✅ 公開API（`/api/v1/public/attendance/{token}`, `/api/v1/public/schedules/{token}`）で使用中

---

## 3. ギャップ一覧

| # | 項目 | MVP要件 | 現状 | 影響 | 修正方針 | 優先度 |
|---|------|---------|------|------|----------|--------|
| 1 | 管理者認証 | 店長/副店長のみログイン可能 | ✅ JWT認証実装済み | ✅ 完了 | - | ✅ 完了 |
| 2 | 出欠確認テーブル | attendance_collections, attendance_responses | ✅ 実装済み | ✅ 完了 | - | ✅ 完了 |
| 3 | 日程調整テーブル | date_schedules, schedule_candidates, schedule_responses | ✅ 実装済み | ✅ 完了 | - | ✅ 完了 |
| 4 | 公開API（出欠） | GET/POST /api/v1/public/attendance/{token} | ✅ 実装済み | ✅ 完了 | - | ✅ 完了 |
| 5 | 公開API（日程） | GET/POST /api/v1/public/schedules/{token} | ✅ 実装済み | ✅ 完了 | - | ✅ 完了 |
| 6 | 管理API（出欠） | POST/GET/PATCH /api/v1/attendance/collections | ✅ 実装済み（一覧取得のみ未実装） | 🟡 一覧取得が未実装 | GET一覧API追加 | P1 |
| 7 | 管理API（日程） | POST/GET/PATCH /api/v1/schedules | ✅ 実装済み（一覧取得のみ未実装） | 🟡 一覧取得が未実装 | GET一覧API追加 | P1 |
| 8 | 公開回答ページ | /p/attendance/{token}, /p/schedule/{token} | ✅ 実装済み | ✅ 完了 | - | ✅ 完了 |
| 9 | メンバー更新API | PUT /api/v1/members/{id} | ❌ 未実装 | 🟡 運用上必要 | handler追加 | P1 |
| 10 | メンバー削除/無効化API | DELETE /api/v1/members/{id} | ❌ 未実装 | 🟡 運用上必要 | handler追加 | P1 |
| 11 | メンバーマスタ管理画面 | 管理者がメンバーCRUD | ⚠️ 一覧のみ実装（編集/削除UIなし） | 🟡 運用上必要 | 編集/削除UI追加 | P1 |
| 12 | 出欠回答重複防止 | UNIQUE(collection_id, member_id)またはUPSERT | ✅ DB制約で実装済み | ✅ 完了 | - | ✅ 完了 |
| 13 | 日程調整回答重複防止 | 同一memberの回答が破綻しない設計 | ✅ UNIQUE制約で実装済み | ✅ 完了 | - | ✅ 完了 |
| 14 | CSV一括登録 | 将来対応想定 | ❌ 未実装 | 🟢 後回し可 | 別途実装 | P2 |
| 15 | 権限管理（店長/副店長） | ロールによる操作制限 | ⚠️ ロールは実装済みだが、操作制限は未実装 | 🟢 後回し可 | ミドルウェアでロールチェック追加 | P2 |
| 16 | 初回セットアップAPI | POST /api/v1/setup | ❌ 未実装 | 🟡 運用上必要 | セットアップAPI追加 | P1 |

---

## 4. 修正タスクの提案（実装順）

### Phase 1: 基盤整備（P0）✅ 完了

#### 1-1. 認証基盤の実装 ✅ 完了
- ✅ `admins` テーブル実装済み
- ✅ JWT認証実装済み
- ✅ 認証ミドルウェア実装済み

#### 1-2. 出欠確認のDB/ドメイン層 ✅ 完了
- ✅ `attendance_collections`, `attendance_responses` テーブル実装済み
- ✅ ドメインモデル・リポジトリ実装済み

#### 1-3. 出欠確認のAPI層 ✅ 完了
- ✅ 管理API実装済み（一覧取得のみ未実装）
- ✅ 公開API実装済み

#### 1-4. 日程調整のDB/ドメイン層 ✅ 完了
- ✅ `date_schedules`, `schedule_candidates`, `schedule_responses` テーブル実装済み
- ✅ ドメインモデル・リポジトリ実装済み

#### 1-5. 日程調整のAPI層 ✅ 完了
- ✅ 管理API実装済み（一覧取得のみ未実装）
- ✅ 公開API実装済み

#### 1-6. フロントエンド公開ページ ✅ 完了
- ✅ `AttendanceResponse.tsx` 実装済み
- ✅ `ScheduleResponse.tsx` 実装済み
- ✅ ルーティング設定済み

### Phase 2: 運用機能（P1）

#### 2-1. メンバーCRUD完成
- ❌ `PUT /api/v1/members/{id}` の実装（未実装）
- ❌ `DELETE /api/v1/members/{id}` の実装（論理削除、未実装）
- ⚠️ ドメイン層には更新・削除メソッドが実装済みだが、ハンドラーが未実装

#### 2-2. メンバーマスタ管理画面
- ✅ `web-frontend/src/pages/Members.tsx` 実装済み（一覧表示のみ）
- ❌ 編集・削除UIが未実装

#### 2-3. 出欠確認/日程調整管理画面
- ✅ `web-frontend/src/pages/AttendanceList.tsx` 実装済み
- ✅ `web-frontend/src/pages/ScheduleList.tsx` 実装済み
- ⚠️ 一覧取得APIが未実装のため、画面は動作しない可能性あり

#### 2-4. 一覧取得APIの実装
- ❌ `GET /api/v1/attendance/collections` の実装（未実装）
- ❌ `GET /api/v1/schedules` の実装（未実装）

### Phase 3: 拡張機能（P2）

#### 3-1. CSV一括登録
- インポートAPI
- フロントエンドUI

#### 3-2. 権限管理（店長/副店長）
- ロールベースのアクセス制御

---

## 5. 不明点・追加で決めるべき点

### 5.1 認証方式
- **選択肢**: メール+パスワード / Discord OAuth / ID+パスワード
- **推奨**: MVP では ID+パスワード（シンプル）。Discord OAuth は将来対応

### 5.2 店長/副店長の権限差
- **現在の想定**: 両者とも管理者として同等の権限
- **要確認**: 副店長に制限をかける操作があるか？

### 5.3 メンバーの識別子
- **現在**: display_name は重複可能（同一テナント内でも）
- **要確認**: 重複を許可するか？許可する場合、UIでどう区別するか？

### 5.4 公開トークンの有効期限
- **現在の想定**: コレクション/スケジュールのステータス（open/closed）で制御
- **要確認**: トークン自体に有効期限を持たせるか？

### 5.5 エラーレスポンスの統一
- **MVP仕様**: 不正トークンは 400/404 のどちらかに統一
- **推奨**: 404 Not Found（トークンが見つからない場合）

---

## 6. まとめ

### P0 ギャップ ✅ すべて完了

1. ✅ **管理者認証** - JWT認証実装済み
2. ✅ **出欠確認機能** - テーブル/API/画面すべて実装済み
3. ✅ **日程調整機能** - テーブル/API/画面すべて実装済み

### P1 残存ギャップ（運用上必要）

1. **メンバー更新・削除API**
   - 現状：ドメイン層は実装済みだが、ハンドラーが未実装
   - 影響：メンバー情報の修正ができない
   - 対応：`PUT /api/v1/members/{id}`, `DELETE /api/v1/members/{id}` の実装

2. **出欠確認・日程調整一覧取得API**
   - 現状：個別取得は実装済みだが、一覧取得が未実装
   - 影響：管理画面で一覧表示ができない
   - 対応：`GET /api/v1/attendance/collections`, `GET /api/v1/schedules` の実装

3. **初回セットアップAPI**
   - 現状：未実装
   - 影響：新規テナントの初期設定が手動
   - 対応：`POST /api/v1/setup` の実装

### 既存実装で活用できるもの

- ✅ テナントモデル・テーブル
- ✅ メンバーモデル・テーブル（CRUD の Read/Create まで、Update/Deleteはドメイン層のみ）
- ✅ PublicToken の型定義（UUID v4形式、バリデーション済み、使用中）
- ✅ レイヤードアーキテクチャの構成
- ✅ フロントエンドのプルダウン選択UI（公開回答ページで使用中）
- ✅ JWT認証基盤
- ✅ 出欠確認・日程調整の完全な実装

### 推定工数（残作業）

| Phase | 内容 | 推定工数 |
|-------|------|----------|
| Phase 1 | 基盤整備（認証・出欠・日程） | ✅ 完了 |
| Phase 2 | 運用機能（メンバー管理・一覧取得API） | 1-2日 |
| Phase 3 | 拡張機能（CSV・権限） | 2-3日 |

---

**次のアクション**: Phase 2（メンバーCRUD完成、一覧取得API）から着手することを推奨


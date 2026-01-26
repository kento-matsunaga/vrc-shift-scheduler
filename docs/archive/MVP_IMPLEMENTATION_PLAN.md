# MVP 実装計画書

**作成日**: 2025-12-14
**更新日**: 2025-12-14（DDD/レイヤード準拠版 v2）
**対象**: vrc-shift-scheduler

---

## 0. 結論（最短の実装順）

### P0 タスク実装順序（依存関係順）

```
1. 認証基盤（admins table + Domain + App + Infra/Security + REST）
   └─ 理由: 全ての管理APIの前提。これがないと"管理者のみ操作"が実現できない

2. 出欠確認 DB + Domain + App層
   └─ 理由: 公開API・管理APIの両方がこのテーブル＋ユースケースに依存

3. 出欠確認 API層（公開API + 管理API）
   └─ 理由: App層を呼ぶ薄いハンドラ。フロント実装の前にAPIが必要

4. 日程調整 DB + Domain + App層
   └─ 理由: 出欠確認と同じパターンで実装可能

5. 日程調整 API層（公開API + 管理API）
   └─ 理由: 出欠確認と同じパターン

6. フロントエンド公開ページ（/p/attendance, /p/schedule）
   └─ 理由: APIが揃ってから実装

7. フロントエンド管理画面（出欠/日程 作成・集計）
   └─ 理由: 認証＋APIが揃ってから実装
```

**なぜこの順序か**:
- 認証がないと管理APIのテナント境界が守れない
- DB層がないとDomain/App層が実装できない
- App層がないとREST層が実装できない（ハンドラはAppを呼ぶだけ）
- 出欠確認と日程調整は同じパターンなので、片方を先に完成させてパターンを確立する

---

## 0.1 アーキテクチャ方針（DDD/レイヤード準拠）

### レイヤー構成

```
backend/internal/
├── domain/           # ドメイン層：エンティティ、値オブジェクト、Repository IF、ドメインサービス
│   ├── auth/         #   Admin, Role, AdminID
│   ├── attendance/   #   AttendanceCollection (集約ルート), AttendanceResponse
│   └── schedule/     #   DateSchedule (集約ルート), CandidateDate, DateScheduleResponse
│
├── app/              # アプリケーション層：ユースケース（トランザクション境界、DTO変換、手続き）
│   ├── auth/         #   LoginUsecase
│   ├── attendance/   #   CreateCollectionUsecase, SubmitResponseUsecase, CloseCollectionUsecase
│   └── schedule/     #   CreateScheduleUsecase, SubmitResponseUsecase, DecideScheduleUsecase
│
├── infra/            # インフラ層：Repository実装、外部サービス実装
│   ├── db/           #   PostgreSQL Repository実装、TxManager
│   ├── security/     #   bcrypt, JWT実装
│   └── clock/        #   Clock実装（時刻取得の抽象化）
│
└── interface/        # インターフェース層：HTTP変換（薄いハンドラ）
    └── rest/         #   Handler（Request解析 → App呼び出し → Response変換）
```

### 依存の向き

```
Interface → App → Domain
              ↓
           Infra（Domain IF の実装）
```

### 各層の責務

| 層 | 責務 | やること | やらないこと |
|----|------|----------|--------------|
| **Domain** | ビジネスルール | 集約ルートの整合性、状態遷移、バリデーション | DB操作、HTTP、外部API、**time.Now()呼び出し** |
| **App** | ユースケース実行 | トランザクション管理、DTO変換、複数リポジトリ協調、**Clock経由で現在時刻取得** | HTTPリクエスト解析、SQLクエリ |
| **Infra** | 技術実装 | DB接続、SQL、外部API呼び出し、UPSERT実装 | ビジネスルール判定 |
| **Interface** | プロトコル変換 | HTTPリクエスト/レスポンス変換、エラーコード変換 | ビジネスルール、DB操作 |

### 0.1.1 DDD/レイヤード追加ルール

#### A. Domain層での time.Now() 禁止

**ルール**: Domain層のメソッドは `time.Now()` を呼ばない。現在時刻が必要な場合は引数で受け取る。

```go
// ❌ Bad: Domain層で time.Now() を呼ぶ
func (c *AttendanceCollection) Close() error {
    c.updatedAt = time.Now()  // 禁止
    ...
}

// ✅ Good: 引数で受け取る
func (c *AttendanceCollection) Close(now time.Time) error {
    c.updatedAt = now
    ...
}
```

**理由**: 
- テスト時に時刻を固定できる
- Domain層が外部依存を持たない

**App層での対応**: `Clock` インターフェースを導入し、App層で `clock.Now()` を呼んでDomainに渡す。

```go
// infra/clock/clock.go
type Clock interface {
    Now() time.Time
}

type RealClock struct{}
func (c *RealClock) Now() time.Time { return time.Now() }

// テスト用
type FixedClock struct { FixedTime time.Time }
func (c *FixedClock) Now() time.Time { return c.FixedTime }
```

#### B. 回答上書き（UPSERT）はRepository層で実装

**MVP方針**: 集約が `responses []` を内部に保持してメモリ上でUpsertするのではなく、**Repository層で `INSERT ... ON CONFLICT DO UPDATE` を実行**する。

```go
// ✅ MVP推奨: Repository層でUPSERT
type AttendanceRepository interface {
    FindByToken(ctx context.Context, token PublicToken) (*AttendanceCollection, error)
    Save(ctx context.Context, collection *AttendanceCollection) error
    UpsertResponse(ctx context.Context, response *AttendanceResponse) error  // ← DB側UPSERT
}

// App層
func (u *SubmitResponseUsecase) Execute(ctx context.Context, input SubmitResponseInput) error {
    collection, _ := u.collectionRepo.FindByToken(ctx, token)
    if err := collection.CanRespond(u.clock.Now()); err != nil {
        return err
    }
    response := attendance.NewAttendanceResponse(...)
    return u.collectionRepo.UpsertResponse(ctx, response)  // 全件ロードしない
}
```

**理由**:
- パフォーマンス: 全件ロード不要
- 整合性: DBの UNIQUE 制約 + ON CONFLICT で担保
- シンプル: 集約とDBの二重管理を避ける

**Domainの責務**: `CanRespond()` などの回答可能判定に集中。回答リスト管理はMVPでは集約に持たせない。

#### C. 管理APIでの tenant_id の扱い

**ルール**: 管理API（`/api/v1/*` 認証必要側）は、tenant_id を Body/Query で受け取らず、**JWT検証後の context から取得**する。

```go
// ✅ Good: Usecaseは ctxTenantID を受け取る
type CreateCollectionInput struct {
    TenantID    common.TenantID  // ← JWT/contextから取得した値
    Title       string
    TargetType  string
    TargetID    string
    Deadline    *time.Time
}

// REST Handler
func (h *AttendanceHandler) Create(w http.ResponseWriter, r *http.Request) {
    tenantID, _ := rest.GetTenantIDFromContext(r.Context())  // JWTから取得
    
    var req CreateCollectionRequest  // tenant_id フィールドなし
    json.NewDecoder(r.Body).Decode(&req)
    
    output, err := h.createUsecase.Execute(r.Context(), attendance.CreateCollectionInput{
        TenantID:   tenantID,  // contextから
        Title:      req.Title,
        TargetType: req.TargetType,
        ...
    })
}
```

**例外**: 
- ログインAPI（`/api/v1/auth/login`）は tenant_id を Body で受け取る（認証前なのでJWTがない）
- 公開API（`/api/v1/public/*`）は token から collection を引いて tenant_id を確定

#### D. トランザクション境界の方針

**方針**: App層でトランザクション境界を張れるようにする。

```go
// infra/db/tx.go
type TxManager interface {
    WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// 使用例（App層）
func (u *DecideScheduleUsecase) Execute(ctx context.Context, input DecideInput) error {
    return u.txManager.WithTx(ctx, func(txCtx context.Context) error {
        schedule, _ := u.scheduleRepo.FindByID(txCtx, input.ScheduleID)
        if err := schedule.Decide(input.CandidateID, u.clock.Now()); err != nil {
            return err
        }
        if err := u.scheduleRepo.Save(txCtx, schedule); err != nil {
            return err
        }
        if input.CreateBusinessDay {
            // 営業日作成も同一トランザクション
            return u.businessDayRepo.Create(txCtx, ...)
        }
        return nil
    })
}
```

**トランザクション必須のUsecase**:
- `SubmitResponseUsecase`（出欠/日程）: メンバー存在確認 + 回答登録
- `DecideScheduleUsecase`: 日程確定 + 営業日作成（オプション）
- `CloseCollectionUsecase` / `CloseScheduleUsecase`: ステータス更新

#### E. エンドポイント命名の統一

| 種別 | パス | 備考 |
|------|------|------|
| 公開API（出欠） | `/api/v1/public/attendance/{token}` | |
| 公開API（日程） | `/api/v1/public/date-schedules/{token}` | `date-schedule` ではなく `date-schedules` |
| 管理API（出欠） | `/api/v1/attendance-collections` | |
| 管理API（日程） | `/api/v1/date-schedules` | |
| フロントURL（出欠） | `/p/attendance/{token}` | |
| フロントURL（日程） | `/p/schedule/{token}` | APIと異なってOK |

**マイグレーションパス**: `backend/internal/infra/db/migrations/` に統一

#### F. 公開回答でのエラーメッセージ方針

| エラー種別 | HTTPステータス | メッセージ | 備考 |
|-----------|---------------|-----------|------|
| token invalid / not found | 404 | "Not found" | **詳細を出さない**（両方404に統一） |
| member_id 不正/存在しない | 400 | "Invalid request" | **理由を出さない**（攻撃者にヒントを与えない） |
| Collection closed / deadline | 403 | "Collection is closed" | 状況は伝えてOK |

---

## 1. MVP仕様（確定事項）

### 1.1 テナント・管理者
| 項目 | 仕様 |
|------|------|
| テナント | 契約単位（店舗）。既存 `tenants` テーブルを使用 |
| ログイン可能者 | 店長/副店長（管理者）のみ |
| 権限 | MVP では店長・副店長に権限差なし（同等の管理者権限） |
| テナント境界 | 管理操作は自テナント内のみ（JWTにtenant_idを含む） |

### 1.2 メンバーマスタ
| 項目 | 仕様 |
|------|------|
| 登録者 | 管理者のみ |
| 操作 | Create / Read / Update / Deactivate（論理削除） |
| CSV一括登録 | **MVPではやらない**（将来対応） |
| 公開回答ページでの追加 | **不可**（「管理者に依頼してください」メッセージ表示） |

### 1.3 公開トークン
| 項目 | 仕様 |
|------|------|
| 形式 | **UUID v4 固定**（nanoidは採用しない） |
| 既存実装 | `backend/internal/domain/common/id.go` の `PublicToken` 型を使用 |
| バリデーション | 不正なUUID形式は **404 Not Found** で統一 |

### 1.4 出欠確認・日程調整
| 項目 | 仕様 |
|------|------|
| 公開回答ページ | token URLでアクセス、認証不要 |
| 回答者選択 | **メンバーマスタからプルダウン必須**（自由入力不可） |
| 重複回答 | 同一 member_id は**上書き**（UNIQUE制約 + Repository側UPSERT） |
| データ保持 | **永続保持**（過去分も削除しない） |
| 公開API | `GET/POST /api/v1/public/attendance/{token}` 等 |
| 管理API | `POST/GET/PATCH /api/v1/attendance-collections` 等 |

### 1.5 MVPでやらないこと（スコープ外）
- [ ] CSV一括登録
- [ ] 店長/副店長の権限差
- [ ] 回答履歴の保持（上書きのみ）
- [ ] Discord BOT連携
- [ ] メール通知
- [ ] 匿名回答

---

## 2. 現状実装の"ズレ"まとめ

| # | 項目 | 現状 | 問題点（MVP仕様との乖離） | 重要度 |
|---|------|------|---------------------------|--------|
| Z1 | **Login.tsx** | 表示名入力 → `createMember` API → メンバー新規作成 | MVP仕様: 管理者のみログイン可能。現状は誰でもメンバー作成できる | 🔴 高 |
| Z2 | **認証ミドルウェア** | `X-Tenant-ID`/`X-Member-ID` ヘッダーのみ | パスワード認証なし。テナント境界がヘッダー詐称で破れる | 🔴 高 |
| Z3 | **管理者概念** | DB/API/UIに存在しない | members テーブルに role カラムなし。admins テーブルもない | 🔴 高 |
| Z4 | **出欠確認機能** | テーブル/API/画面すべて未実装 | MVPコア機能が動作しない | 🔴 高 |
| Z5 | **日程調整機能** | テーブル/API/画面すべて未実装 | MVPコア機能が動作しない | 🔴 高 |
| Z6 | **公開ページ** | `/p/` 系のルートが存在しない | キャストがtoken URLで回答できない | 🔴 高 |
| Z7 | **メンバー更新API** | `PUT /api/v1/members/{id}` 未実装 | 管理者がメンバー情報を修正できない | 🟡 中 |
| Z8 | **メンバー無効化API** | `DELETE /api/v1/members/{id}` 未実装 | 退職メンバーを無効化できない | 🟡 中 |
| Z9 | **App.tsx ログイン判定** | `localStorage.getItem('member_id')` の有無 | JWT検証なし。member_idを詐称可能 | 🔴 高 |

### 根拠ファイル

| ズレ | ファイルパス |
|------|-------------|
| Z1, Z9 | `web-frontend/src/pages/Login.tsx:40-46` |
| Z2 | `backend/internal/interface/rest/middleware.go:70-107` |
| Z3 | `backend/internal/infra/db/migrations/003_create_members_and_shift_slots.up.sql:7-21` |
| Z4, Z5 | `backend/internal/interface/rest/router.go` (attendance/schedule 系のルートなし) |
| Z6 | `web-frontend/src/App.tsx:14-31` (`/p/` 系のRouteなし) |

---

## 3. 修正タスク一覧（表）

| TaskID | 内容 | 優先度 | 受け入れ条件(DoD) | 対象候補ファイル | 備考 |
|--------|------|--------|-------------------|------------------|------|
| T1 | admins テーブル作成 | P0 | マイグレーション成功 & `\d admins` で確認 | `backend/internal/infra/db/migrations/007_*.up.sql` | tenant_id, email, password_hash, role |
| T2 | Auth Domain層 | P0 | `go build` 成功 | `domain/auth/admin.go`, `repository.go` | Admin エンティティ、Repository IF |
| T3 | Auth Infra/Security層 + Clock | P0 | 単体テスト通過 | `infra/security/bcrypt.go`, `jwt.go`, `infra/clock/clock.go` | bcrypt/JWT/Clock実装 |
| T4 | Auth App層（LoginUsecase） | P0 | 単体テスト通過 | `app/auth/login_usecase.go` | パスワード検証 → JWT発行 |
| T5 | Auth REST層 | P0 | `curl POST /api/v1/auth/login` で JWT 取得可能 | `rest/auth_handler.go`, `router.go` | ハンドラはUsecaseを呼ぶだけ |
| T6 | JWT認証ミドルウェア | P0 | Authorization: Bearer で認証通過、ctx に tenant_id 設定 | `rest/middleware.go` | 段階移行（X-Tenant-ID並行運用） |
| T7 | attendance テーブル群 | P0 | マイグレーション成功 | `backend/internal/infra/db/migrations/008_*.up.sql` | collections + responses（UNIQUE制約含む） |
| T8 | Attendance Domain層 | P0 | `go build` 成功、time.Now()なし | `domain/attendance/` | 集約ルート + CanRespond/Close(now) |
| T9 | Attendance App層 | P0 | 単体テスト通過、トランザクション対応 | `app/attendance/` | CreateUsecase, SubmitUsecase, CloseUsecase |
| T10 | Attendance Infra層 + TxManager | P0 | 単体テスト通過、UpsertResponse実装 | `infra/db/attendance_repository.go`, `tx.go` | Save/FindByToken/UpsertResponse（ON CONFLICT） |
| T11 | Attendance 公開API | P0 | `curl GET/POST /api/v1/public/attendance/{token}` で成功 | `rest/public_attendance_handler.go` | 認証不要、App層を呼ぶ |
| T12 | Attendance 管理API | P0 | `curl POST/GET/PATCH /api/v1/attendance-collections` で成功（tenant_idはJWTから） | `rest/attendance_handler.go` | 認証必要、ctx tenant_id使用 |
| T13 | schedule テーブル群 | P0 | マイグレーション成功 | `backend/internal/infra/db/migrations/009_*.up.sql` | schedules + candidates + responses |
| T14 | Schedule Domain層 | P0 | `go build` 成功、time.Now()なし | `domain/schedule/` | 集約ルート + CanRespond/Decide(now) |
| T15 | Schedule App層 | P0 | 単体テスト通過、トランザクション対応 | `app/schedule/` | CreateUsecase, SubmitUsecase, DecideUsecase |
| T16 | Schedule Infra層 | P0 | 単体テスト通過 | `infra/db/schedule_repository.go` | Save/FindByToken/UpsertResponse |
| T17 | Schedule 公開API | P0 | `curl GET/POST /api/v1/public/date-schedules/{token}` で成功 | `rest/public_schedule_handler.go` | 認証不要 |
| T18 | Schedule 管理API | P0 | `curl POST/GET/PATCH /api/v1/date-schedules` で成功 | `rest/schedule_handler.go` | 認証必要 |
| T19 | 公開回答ページ（出欠） | P0 | ブラウザで `/p/attendance/{token}` 表示 & 回答送信成功 | `pages/public/AttendanceResponse.tsx` | メンバープルダウン必須 |
| T20 | 公開回答ページ（日程） | P0 | ブラウザで `/p/schedule/{token}` 表示 & 回答送信成功 | `pages/public/ScheduleResponse.tsx` | メンバープルダウン必須 |
| T21 | 管理者ログイン画面 | P0 | ブラウザで `/login` → email/pw入力 → JWT取得 → 管理画面遷移 | `pages/AdminLogin.tsx` | 旧 Login.tsx を置換 |
| T22 | 出欠確認 管理画面 | P1 | ブラウザで作成・一覧・詳細・クローズ操作可能 | `pages/AttendanceManagement.tsx` | URLコピー機能 |
| T23 | 日程調整 管理画面 | P1 | ブラウザで作成・一覧・詳細・確定・クローズ操作可能 | `pages/ScheduleManagement.tsx` | URLコピー機能 |
| T24 | メンバー更新API | P1 | `curl PUT /api/v1/members/{id}` で成功 | `rest/member_handler.go` | display_name, email 等更新 |
| T25 | メンバー無効化API | P1 | `curl DELETE /api/v1/members/{id}` で論理削除 | `rest/member_handler.go` | is_active=false |
| T26 | メンバーマスタ管理画面 | P1 | ブラウザで作成・一覧・編集・無効化操作可能 | `pages/MemberManagement.tsx` | |
| T27 | セットアップAPI | P2 | `curl POST /api/v1/setup` でテナント+管理者作成 | `rest/setup_handler.go` | 初回セットアップ用 |

---

## 4. 具体的な実装ステップ（P0のみ詳細）

### Commit 1: 認証基盤 - DB + Domain（T1, T2）

**ファイル**:
```
backend/internal/
├── infra/db/migrations/
│   ├── 007_create_admins.up.sql       # 新規
│   └── 007_create_admins.down.sql     # 新規
└── domain/auth/
    ├── admin.go                        # 新規：Admin エンティティ
    ├── role.go                         # 新規：Role 値オブジェクト
    └── repository.go                   # 新規：AdminRepository IF
```

**admins テーブル設計**:
```sql
CREATE TABLE admins (
    admin_id CHAR(26) PRIMARY KEY,      -- ULID
    tenant_id CHAR(26) NOT NULL REFERENCES tenants(tenant_id),
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'manager',  -- 'owner' | 'manager'
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    CONSTRAINT uq_admins_tenant_email UNIQUE(tenant_id, email)
);

CREATE INDEX idx_admins_tenant ON admins(tenant_id) WHERE deleted_at IS NULL;
```

**domain/auth/admin.go（例）**:
```go
package auth

import (
    "time"
    "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// Admin は管理者（店長/副店長）を表すエンティティ
type Admin struct {
    adminID      AdminID
    tenantID     common.TenantID
    email        string
    passwordHash string  // ドメインはハッシュを保持するが、bcrypt処理はしない
    displayName  string
    role         Role
    isActive     bool
    createdAt    time.Time
    updatedAt    time.Time
}

// CanLogin は認証可能かを判定（ドメインルール）
func (a *Admin) CanLogin() bool {
    return a.isActive
}

// PasswordHash は認証処理用にハッシュを返す（App/Infra層でのみ使用）
func (a *Admin) PasswordHash() string {
    return a.passwordHash
}
```

**壊しやすいポイント**:
- Domain層に bcrypt.CompareHashAndPassword を書いてしまう → Infra/Security に分離
- password_hash を Admin の外に露出してしまう → getter を限定的に

---

### Commit 2: 認証基盤 - Infra/Security + Clock + App + REST（T3, T4, T5, T6）

**ファイル**:
```
backend/internal/
├── infra/
│   ├── db/
│   │   └── admin_repository.go         # 新規：AdminRepository 実装
│   ├── security/
│   │   ├── bcrypt.go                    # 新規：パスワードハッシュ化/検証
│   │   └── jwt.go                       # 新規：JWT発行/検証
│   └── clock/
│       └── clock.go                     # 新規：Clock インターフェース
├── app/auth/
│   ├── login_usecase.go                 # 新規：ログインユースケース
│   └── dto.go                           # 新規：LoginInput/LoginOutput
└── interface/rest/
    ├── auth_handler.go                  # 新規：認証API
    ├── middleware.go                    # 修正：JWT検証追加
    └── router.go                        # 修正：/auth/login 追加
```

**infra/clock/clock.go（例）**:
```go
package clock

import "time"

// Clock は現在時刻を取得するインターフェース
// App層で使用し、Domain層には now を引数で渡す
type Clock interface {
    Now() time.Time
}

// RealClock は本番用の実装
type RealClock struct{}

func NewRealClock() *RealClock {
    return &RealClock{}
}

func (c *RealClock) Now() time.Time {
    return time.Now()
}

// FixedClock はテスト用の実装
type FixedClock struct {
    FixedTime time.Time
}

func NewFixedClock(t time.Time) *FixedClock {
    return &FixedClock{FixedTime: t}
}

func (c *FixedClock) Now() time.Time {
    return c.FixedTime
}
```

**infra/security/bcrypt.go（例）**:
```go
package security

import "golang.org/x/crypto/bcrypt"

type PasswordHasher interface {
    Hash(password string) (string, error)
    Compare(hash, password string) error
}

type BcryptHasher struct {
    cost int
}

func NewBcryptHasher() *BcryptHasher {
    return &BcryptHasher{cost: 10}
}

func (h *BcryptHasher) Hash(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
    return string(bytes), err
}

func (h *BcryptHasher) Compare(hash, password string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
```

**app/auth/login_usecase.go（例）**:
```go
package auth

import (
    "context"
    "time"
    
    "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
    "github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/security"
)

type LoginUsecase struct {
    adminRepo      auth.AdminRepository
    passwordHasher security.PasswordHasher
    tokenIssuer    security.TokenIssuer
}

// LoginInput - ログインAPIは tenant_id を受け取る（認証前なのでJWTがないため）
type LoginInput struct {
    TenantID string  // ログイン時のみ Body で受け取る
    Email    string
    Password string
}

type LoginOutput struct {
    Token     string
    AdminID   string
    TenantID  string
    ExpiresAt time.Time
}

func (u *LoginUsecase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {
    // 1. Admin取得
    admin, err := u.adminRepo.FindByEmail(ctx, input.TenantID, input.Email)
    if err != nil {
        return nil, ErrInvalidCredentials  // 存在しない場合も同じエラー
    }
    
    // 2. ログイン可能かチェック（ドメインルール）
    if !admin.CanLogin() {
        return nil, ErrAccountDisabled
    }
    
    // 3. パスワード検証（Infra層に委譲）
    if err := u.passwordHasher.Compare(admin.PasswordHash(), input.Password); err != nil {
        return nil, ErrInvalidCredentials
    }
    
    // 4. JWT発行（Infra層に委譲）
    token, expiresAt, err := u.tokenIssuer.Issue(admin.AdminID(), admin.TenantID())
    if err != nil {
        return nil, err
    }
    
    return &LoginOutput{
        Token:     token,
        AdminID:   admin.AdminID().String(),
        TenantID:  admin.TenantID().String(),
        ExpiresAt: expiresAt,
    }, nil
}
```

**rest/auth_handler.go（例）**:
```go
package rest

import (
    "encoding/json"
    "errors"
    "net/http"
    
    "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/auth"
)

// LoginRequest - ログインAPIのみ tenant_id を Body で受け取る
type LoginRequest struct {
    TenantID string `json:"tenant_id"`  // ログイン時のみ
    Email    string `json:"email"`
    Password string `json:"password"`
}

// auth_handler は HTTP → DTO → Usecase → Response の薄い変換層
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // 1. リクエスト解析
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
        return
    }
    
    // 2. Usecase呼び出し（ビジネスロジックはここにない）
    output, err := h.loginUsecase.Execute(r.Context(), auth.LoginInput{
        TenantID: req.TenantID,
        Email:    req.Email,
        Password: req.Password,
    })
    if err != nil {
        // エラーコード変換
        switch {
        case errors.Is(err, auth.ErrInvalidCredentials):
            RespondError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Invalid email or password", nil)
        case errors.Is(err, auth.ErrAccountDisabled):
            RespondError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Account is disabled", nil)
        default:
            RespondInternalError(w)
        }
        return
    }
    
    // 3. レスポンス変換
    RespondJSON(w, http.StatusOK, LoginResponse{Token: output.Token})
}
```

**壊しやすいポイント**:
- JWT秘密鍵を環境変数から取得し損ねる → 起動時にpanicするようにする
- ハンドラにパスワード検証ロジックを書いてしまう → Usecaseに寄せる
- エラーメッセージで「メールが存在しない」「パスワードが違う」を区別してしまう → 攻撃者にヒントを与えない

---

### Commit 3: 出欠確認 - DB + Domain + App + Infra（T7, T8, T9, T10）

**ファイル**:
```
backend/internal/
├── infra/db/migrations/
│   ├── 008_create_attendance_tables.up.sql    # 新規
│   └── 008_create_attendance_tables.down.sql  # 新規
├── domain/attendance/
│   ├── collection.go                          # 新規：AttendanceCollection 集約ルート
│   ├── response.go                            # 新規：AttendanceResponse エンティティ
│   ├── status.go                              # 新規：Status 値オブジェクト
│   └── repository.go                          # 新規：Repository IF
├── app/attendance/
│   ├── create_collection_usecase.go           # 新規
│   ├── submit_response_usecase.go             # 新規
│   ├── close_collection_usecase.go            # 新規
│   ├── get_collection_usecase.go              # 新規
│   └── dto.go                                 # 新規
└── infra/db/
    ├── attendance_repository.go               # 新規
    ├── attendance_repository_test.go          # 新規
    └── tx.go                                  # 新規：TxManager
```

**domain/attendance/collection.go（例）**:
```go
package attendance

import (
    "time"
    "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// AttendanceCollection は出欠確認の集約ルート
// MVP方針: responses は集約内で保持しない（Repository側UPSERTで管理）
type AttendanceCollection struct {
    collectionID CollectionID
    tenantID     common.TenantID
    title        string
    description  string
    targetType   TargetType  // "event" | "business_day"
    targetID     string
    publicToken  common.PublicToken
    status       Status      // "open" | "closed"
    deadline     *time.Time
    createdAt    time.Time
    updatedAt    time.Time
}

// CanRespond は回答可能かを判定（ドメインルール）
// now は App層から Clock 経由で渡される
func (c *AttendanceCollection) CanRespond(now time.Time) error {
    if c.status != StatusOpen {
        return ErrCollectionClosed
    }
    if c.deadline != nil && now.After(*c.deadline) {
        return ErrDeadlinePassed
    }
    return nil
}

// Close はステータスをclosedに変更（ドメインルール）
// now は App層から Clock 経由で渡される
func (c *AttendanceCollection) Close(now time.Time) error {
    if c.status == StatusClosed {
        return ErrAlreadyClosed
    }
    c.status = StatusClosed
    c.updatedAt = now
    return nil
}
```

**domain/attendance/repository.go（例）**:
```go
package attendance

import (
    "context"
    "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// AttendanceCollectionRepository はコレクションの永続化インターフェース
type AttendanceCollectionRepository interface {
    // Save はコレクションを保存する
    Save(ctx context.Context, collection *AttendanceCollection) error
    
    // FindByID はIDでコレクションを取得する
    FindByID(ctx context.Context, tenantID common.TenantID, id CollectionID) (*AttendanceCollection, error)
    
    // FindByToken は公開トークンでコレクションを取得する
    FindByToken(ctx context.Context, token common.PublicToken) (*AttendanceCollection, error)
    
    // FindByTenantID はテナント内のコレクション一覧を取得する
    FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*AttendanceCollection, error)
    
    // UpsertResponse は回答を登録/更新する（ON CONFLICT DO UPDATE）
    // MVP方針: 回答の上書きはRepository層で行う
    UpsertResponse(ctx context.Context, response *AttendanceResponse) error
    
    // FindResponsesByCollectionID はコレクションの回答一覧を取得する
    FindResponsesByCollectionID(ctx context.Context, collectionID CollectionID) ([]*AttendanceResponse, error)
}
```

**infra/db/tx.go（例）**:
```go
package db

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
)

// TxManager はトランザクション管理インターフェース
type TxManager interface {
    WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type PgxTxManager struct {
    pool *pgxpool.Pool
}

func NewPgxTxManager(pool *pgxpool.Pool) *PgxTxManager {
    return &PgxTxManager{pool: pool}
}

func (m *PgxTxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
    tx, err := m.pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)
    
    // トランザクションを context に格納（Repository が取り出して使用）
    txCtx := context.WithValue(ctx, txKey, tx)
    
    if err := fn(txCtx); err != nil {
        return err
    }
    
    return tx.Commit(ctx)
}
```

**app/attendance/submit_response_usecase.go（例）**:
```go
package attendance

import (
    "context"
    
    "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
    "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
    "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
    "github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/clock"
    "github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
)

type SubmitResponseUsecase struct {
    collectionRepo attendance.AttendanceCollectionRepository
    memberRepo     member.MemberRepository
    txManager      db.TxManager
    clock          clock.Clock
}

type SubmitResponseInput struct {
    Token    string
    MemberID string
    Response string  // "attending" | "absent"
    Note     string
}

func (u *SubmitResponseUsecase) Execute(ctx context.Context, input SubmitResponseInput) error {
    // トランザクション内で実行
    return u.txManager.WithTx(ctx, func(txCtx context.Context) error {
        // 1. トークンからコレクション取得
        token, err := common.ParsePublicToken(input.Token)
        if err != nil {
            return ErrTokenInvalid  // → REST層で404に変換
        }
        
        collection, err := u.collectionRepo.FindByToken(txCtx, token)
        if err != nil {
            return ErrCollectionNotFound
        }
        
        // 2. 回答可能かチェック（ドメインルール）
        // ★ Clock経由で現在時刻を取得し、Domainに渡す
        now := u.clock.Now()
        if err := collection.CanRespond(now); err != nil {
            return err
        }
        
        // 3. メンバー存在確認（同一テナント）
        memberID, err := common.ParseMemberID(input.MemberID)
        if err != nil {
            return ErrMemberInvalid  // → REST層で400に変換（理由は出さない）
        }
        _, err = u.memberRepo.FindByID(txCtx, collection.TenantID(), memberID)
        if err != nil {
            return ErrMemberNotFound  // → REST層で400に変換（理由は出さない）
        }
        
        // 4. 回答作成
        responseType, err := attendance.ParseResponseType(input.Response)
        if err != nil {
            return ErrInvalidResponseType
        }
        response := attendance.NewAttendanceResponse(
            collection.CollectionID(),
            collection.TenantID(),
            memberID,
            responseType,
            input.Note,
            now,
        )
        
        // 5. 永続化（Repository側でUPSERT）
        // ★ MVP方針: 全件ロードせず、Repositoryが ON CONFLICT DO UPDATE を実行
        return u.collectionRepo.UpsertResponse(txCtx, response)
    })
}
```

**infra/db/attendance_repository.go（UPSERT部分の例）**:
```go
func (r *AttendanceRepository) UpsertResponse(ctx context.Context, response *attendance.AttendanceResponse) error {
    query := `
        INSERT INTO attendance_responses (
            response_id, tenant_id, collection_id, member_id, response, note, responded_at, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        ON CONFLICT (collection_id, member_id) DO UPDATE SET
            response = EXCLUDED.response,
            note = EXCLUDED.note,
            responded_at = EXCLUDED.responded_at,
            updated_at = EXCLUDED.updated_at
    `
    _, err := r.getConn(ctx).Exec(ctx, query,
        response.ResponseID().String(),
        response.TenantID().String(),
        response.CollectionID().String(),
        response.MemberID().String(),
        response.Response().String(),
        response.Note(),
        response.RespondedAt(),
        response.CreatedAt(),
        response.UpdatedAt(),
    )
    return err
}
```

**テーブル設計**:
```sql
CREATE TABLE attendance_collections (
    collection_id CHAR(26) PRIMARY KEY,
    tenant_id CHAR(26) NOT NULL REFERENCES tenants(tenant_id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    target_type VARCHAR(20) NOT NULL,
    target_id CHAR(26),
    public_token UUID NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    deadline TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_attendance_collections_tenant ON attendance_collections(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_attendance_collections_token ON attendance_collections(public_token);

CREATE TABLE attendance_responses (
    response_id CHAR(26) PRIMARY KEY,
    tenant_id CHAR(26) NOT NULL REFERENCES tenants(tenant_id),
    collection_id CHAR(26) NOT NULL REFERENCES attendance_collections(collection_id),
    member_id CHAR(26) NOT NULL REFERENCES members(member_id),
    response VARCHAR(20) NOT NULL,
    note TEXT,
    responded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- ★ UNIQUE制約: 同一コレクション×メンバーは1回答のみ（UPSERTで上書き）
    CONSTRAINT uq_attendance_response_member UNIQUE(collection_id, member_id)
);

CREATE INDEX idx_attendance_responses_collection ON attendance_responses(collection_id);
```

**壊しやすいポイント**:
- Domain層で `time.Now()` を呼んでしまう → App層で `clock.Now()` を呼び、Domainには引数で渡す
- 集約内で responses を管理してDBと二重管理になる → MVP では Repository 側 UPSERT に統一
- トランザクションを張り忘れる → `SubmitResponseUsecase` は必ず `WithTx` 内で実行

---

### Commit 4: 出欠確認 - REST層（T11, T12）

**ファイル**:
```
backend/internal/interface/rest/
├── public_attendance_handler.go   # 新規（認証不要）
├── attendance_handler.go          # 新規（認証必要）
└── router.go                      # 修正
```

**rest/public_attendance_handler.go（例）**:
```go
package rest

import (
    "encoding/json"
    "errors"
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/attendance"
)

// PublicAttendanceHandler は公開API用の薄いハンドラ
type PublicAttendanceHandler struct {
    getUsecase    *attendance.GetCollectionUsecase
    submitUsecase *attendance.SubmitResponseUsecase
}

func (h *PublicAttendanceHandler) GetCollection(w http.ResponseWriter, r *http.Request) {
    token := chi.URLParam(r, "token")
    
    // Usecase呼び出し
    output, err := h.getUsecase.Execute(r.Context(), attendance.GetCollectionInput{Token: token})
    if err != nil {
        // エラーコード変換（トークン系は全て404）
        switch {
        case errors.Is(err, attendance.ErrTokenInvalid),
             errors.Is(err, attendance.ErrCollectionNotFound):
            RespondError(w, http.StatusNotFound, "ERR_NOT_FOUND", "Not found", nil)
        default:
            RespondInternalError(w)
        }
        return
    }
    
    // レスポンス変換
    RespondJSON(w, http.StatusOK, toPublicCollectionResponse(output))
}

func (h *PublicAttendanceHandler) SubmitResponse(w http.ResponseWriter, r *http.Request) {
    token := chi.URLParam(r, "token")
    
    var req SubmitResponseRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request", nil)
        return
    }
    
    // Usecase呼び出し
    err := h.submitUsecase.Execute(r.Context(), attendance.SubmitResponseInput{
        Token:    token,
        MemberID: req.MemberID,
        Response: req.Response,
        Note:     req.Note,
    })
    if err != nil {
        switch {
        case errors.Is(err, attendance.ErrTokenInvalid),
             errors.Is(err, attendance.ErrCollectionNotFound):
            // ★ token系は全て404（詳細を出さない）
            RespondError(w, http.StatusNotFound, "ERR_NOT_FOUND", "Not found", nil)
        case errors.Is(err, attendance.ErrCollectionClosed),
             errors.Is(err, attendance.ErrDeadlinePassed):
            RespondError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Collection is closed", nil)
        case errors.Is(err, attendance.ErrMemberInvalid),
             errors.Is(err, attendance.ErrMemberNotFound):
            // ★ member系は400だが詳細を出さない（攻撃者にヒントを与えない）
            RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request", nil)
        default:
            RespondInternalError(w)
        }
        return
    }
    
    RespondJSON(w, http.StatusOK, map[string]string{"message": "Response submitted"})
}
```

**rest/attendance_handler.go（管理API、例）**:
```go
package rest

// CreateCollectionRequest - tenant_id は含まない（JWTから取得）
type CreateCollectionRequest struct {
    Title       string  `json:"title"`
    Description string  `json:"description,omitempty"`
    TargetType  string  `json:"target_type"`
    TargetID    string  `json:"target_id,omitempty"`
    Deadline    *string `json:"deadline,omitempty"`
}

func (h *AttendanceHandler) Create(w http.ResponseWriter, r *http.Request) {
    // ★ tenant_id は JWT/context から取得（Body からは受け取らない）
    tenantID, ok := GetTenantIDFromContext(r.Context())
    if !ok {
        RespondError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Unauthorized", nil)
        return
    }
    
    var req CreateCollectionRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
        return
    }
    
    // Usecase呼び出し
    output, err := h.createUsecase.Execute(r.Context(), attendance.CreateCollectionInput{
        TenantID:    tenantID,  // ★ context から
        Title:       req.Title,
        Description: req.Description,
        TargetType:  req.TargetType,
        TargetID:    req.TargetID,
        Deadline:    parseDeadline(req.Deadline),
    })
    if err != nil {
        // エラーコード変換
        RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error(), nil)
        return
    }
    
    RespondJSON(w, http.StatusCreated, toCollectionResponse(output))
}
```

**router.go 追加部分**:
```go
// 公開API（認証不要）
r.Route("/api/v1/public", func(r chi.Router) {
    // 認証ミドルウェアを適用しない
    publicAttendanceHandler := NewPublicAttendanceHandler(db)
    r.Get("/attendance/{token}", publicAttendanceHandler.GetCollection)
    r.Post("/attendance/{token}/responses", publicAttendanceHandler.SubmitResponse)
    
    publicScheduleHandler := NewPublicScheduleHandler(db)
    r.Get("/date-schedules/{token}", publicScheduleHandler.GetSchedule)      // ★ date-schedules に統一
    r.Post("/date-schedules/{token}/responses", publicScheduleHandler.SubmitResponse)
})

// 管理API（認証必要）- 既存の /api/v1 ルート内
r.Route("/attendance-collections", func(r chi.Router) {
    attendanceHandler := NewAttendanceHandler(db)
    r.Post("/", attendanceHandler.Create)
    r.Get("/", attendanceHandler.List)
    r.Get("/{collection_id}", attendanceHandler.GetDetail)
    r.Patch("/{collection_id}/close", attendanceHandler.Close)
})

r.Route("/date-schedules", func(r chi.Router) {  // ★ date-schedules に統一
    scheduleHandler := NewScheduleHandler(db)
    r.Post("/", scheduleHandler.Create)
    r.Get("/", scheduleHandler.List)
    r.Get("/{schedule_id}", scheduleHandler.GetDetail)
    r.Patch("/{schedule_id}/decide", scheduleHandler.Decide)
    r.Patch("/{schedule_id}/close", scheduleHandler.Close)
})
```

---

### Commit 5: 日程調整 - DB + Domain + App + Infra（T13, T14, T15, T16）

**ファイル**:
```
backend/internal/
├── infra/db/migrations/
│   ├── 009_create_schedule_tables.up.sql     # 新規
│   └── 009_create_schedule_tables.down.sql   # 新規
├── domain/schedule/
│   ├── schedule.go                           # 新規：DateSchedule 集約ルート
│   ├── candidate.go                          # 新規：CandidateDate エンティティ
│   ├── response.go                           # 新規：DateScheduleResponse エンティティ
│   ├── status.go                             # 新規：Status 値オブジェクト
│   └── repository.go                         # 新規：Repository IF
├── app/schedule/
│   ├── create_schedule_usecase.go            # 新規
│   ├── submit_response_usecase.go            # 新規
│   ├── decide_schedule_usecase.go            # 新規
│   ├── close_schedule_usecase.go             # 新規
│   └── dto.go                                # 新規
└── infra/db/
    ├── schedule_repository.go                # 新規
    └── schedule_repository_test.go           # 新規
```

**domain/schedule/schedule.go（例）**:
```go
package schedule

import (
    "time"
    "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// DateSchedule は日程調整の集約ルート
// MVP方針: responses は集約内で保持しない（Repository側UPSERTで管理）
type DateSchedule struct {
    scheduleID          ScheduleID
    tenantID            common.TenantID
    title               string
    description         string
    eventID             *common.EventID
    publicToken         common.PublicToken
    status              Status  // "open" | "closed" | "decided"
    deadline            *time.Time
    decidedCandidateID  *CandidateID
    candidates          []CandidateDate  // 候補日は集約内で保持（作成時に確定）
    createdAt           time.Time
    updatedAt           time.Time
}

// CanRespond は回答可能かを判定（ドメインルール）
// now は App層から Clock 経由で渡される
func (s *DateSchedule) CanRespond(now time.Time) error {
    if s.status != StatusOpen {
        return ErrScheduleClosed
    }
    if s.deadline != nil && now.After(*s.deadline) {
        return ErrDeadlinePassed
    }
    return nil
}

// Decide は開催日を決定する（ドメインルール）
// now は App層から Clock 経由で渡される
func (s *DateSchedule) Decide(candidateID CandidateID, now time.Time) error {
    if s.status == StatusDecided {
        return ErrAlreadyDecided
    }
    
    // 候補日が存在するかチェック
    found := false
    for _, c := range s.candidates {
        if c.CandidateID() == candidateID {
            found = true
            break
        }
    }
    if !found {
        return ErrCandidateNotFound
    }
    
    s.status = StatusDecided
    s.decidedCandidateID = &candidateID
    s.updatedAt = now
    return nil
}

// Close はステータスをclosedに変更（ドメインルール）
// now は App層から Clock 経由で渡される
func (s *DateSchedule) Close(now time.Time) error {
    if s.status == StatusClosed || s.status == StatusDecided {
        return ErrAlreadyClosed
    }
    s.status = StatusClosed
    s.updatedAt = now
    return nil
}
```

---

### Commit 6: 日程調整 - REST層（T17, T18）

（出欠確認と同じパターンなので省略）

---

### Commit 7: フロントエンド公開ページ（T19, T20）

**ファイル**:
```
web-frontend/src/
├── pages/public/
│   ├── AttendanceResponse.tsx   # 新規
│   └── ScheduleResponse.tsx     # 新規
├── lib/api/
│   └── publicApi.ts             # 新規
└── App.tsx                      # 修正（/p/... ルート追加）
```

---

### Commit 8: 管理者ログイン画面（T21）

**ファイル**:
```
web-frontend/src/
├── pages/
│   ├── AdminLogin.tsx           # 新規（旧 Login.tsx を置換）
│   └── Login.tsx                # 削除
├── lib/api/
│   └── authApi.ts               # 修正（JWT対応）
└── App.tsx                      # 修正
```

---

## 5. 既存コードをどう扱うか

### 5.1 Login.tsx の扱い

| 方針 | 内容 |
|------|------|
| **置換** | `AdminLogin.tsx` を新規作成し、旧 `Login.tsx` を削除 |
| **理由** | 現在の「表示名入力 → メンバー作成」は MVP 仕様と完全に逆。修正より作り直しが早い |

### 5.2 X-Tenant-ID / X-Member-ID の扱い

| 方針 | 内容 |
|------|------|
| **段階移行** | JWT認証を追加し、X-Tenant-ID は当面並行運用 |
| **理由** | 既存APIを壊さずに移行できる |
| **実装方針** | middleware.go で `Authorization: Bearer` があればJWT検証、なければ従来のヘッダー認証 |

### 5.3 既存の app/shift_assignment_service.go

| 方針 | 内容 |
|------|------|
| **そのまま残す** | 既存パターンとして参考になる。命名規則は `*Service` だが機能している |
| **将来的に** | `app/shift/` に移動して他のUsecaseと揃えてもよい |

---

## 6. 実装後の動作確認コマンド

### 6.1 起動

```bash
# コンテナ起動
docker compose up -d --build

# ログ確認
docker compose logs -f backend

# DB接続確認
docker compose exec db psql -U vrcshift -d vrcshift -c '\dt'
```

### 6.2 マイグレーション確認

```bash
# マイグレーション実行
docker compose exec backend go run ./cmd/migrate up

# テーブル確認
docker compose exec db psql -U vrcshift -d vrcshift -c '\d admins'
docker compose exec db psql -U vrcshift -d vrcshift -c '\d attendance_collections'
docker compose exec db psql -U vrcshift -d vrcshift -c '\d date_schedules'
```

### 6.3 認証API

```bash
# ヘルスチェック
curl http://localhost:8080/health

# ログイン（tenant_id を Body で指定）
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"tenant_id": "01HXXXXXX", "email": "admin@example.com", "password": "password123"}'

# 期待レスポンス: {"data": {"token": "eyJ..."}}
```

### 6.4 出欠確認 管理API

```bash
# JWT取得後
TOKEN="eyJ..."

# 出欠確認作成（tenant_id は JWT から取得されるため Body に含めない）
curl -X POST http://localhost:8080/api/v1/attendance-collections \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "1月7日営業 出欠確認",
    "target_type": "business_day",
    "target_id": "01HXXXXXX",
    "deadline": "2025-01-05T23:59:59+09:00"
  }'

# 期待レスポンス: {"data": {"collection_id": "...", "public_token": "550e8400-...", ...}}

# 出欠確認クローズ
curl -X PATCH http://localhost:8080/api/v1/attendance-collections/{collection_id}/close \
  -H "Authorization: Bearer $TOKEN"
```

### 6.5 出欠確認 公開API

```bash
# 公開ページデータ取得（認証不要）
curl http://localhost:8080/api/v1/public/attendance/550e8400-e29b-41d4-a716-446655440000

# 出欠回答登録（認証不要）
curl -X POST http://localhost:8080/api/v1/public/attendance/550e8400-e29b-41d4-a716-446655440000/responses \
  -H "Content-Type: application/json" \
  -d '{
    "member_id": "01HXXXXXX",
    "response": "attending",
    "note": "よろしくお願いします"
  }'

# 不正なトークン → 404（詳細なし）
curl http://localhost:8080/api/v1/public/attendance/invalid-token
# 期待: {"error": {"code": "ERR_NOT_FOUND", "message": "Not found"}}
```

### 6.6 日程調整 管理API

```bash
# 日程調整作成
curl -X POST http://localhost:8080/api/v1/date-schedules \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "2月特別イベント日程調整",
    "candidate_dates": [
      {"date": "2025-02-08", "start_time": "21:30", "end_time": "23:00"},
      {"date": "2025-02-15", "start_time": "21:30", "end_time": "23:00"}
    ],
    "deadline": "2025-01-31T23:59:59+09:00"
  }'
```

### 6.7 フロントエンド確認

```bash
# ブラウザで確認
# 管理者ログイン: http://localhost:5173/login
# 公開ページ: http://localhost:5173/p/attendance/{token}
# 公開ページ: http://localhost:5173/p/schedule/{token}
```

---

## 付録A: ファイル一覧（新規/修正）

### 新規ファイル

| パス | 目的 | 層 |
|------|------|-----|
| `backend/internal/infra/db/migrations/007_create_admins.up.sql` | admins テーブル | Infra |
| `backend/internal/infra/db/migrations/008_create_attendance_tables.up.sql` | 出欠確認テーブル | Infra |
| `backend/internal/infra/db/migrations/009_create_schedule_tables.up.sql` | 日程調整テーブル | Infra |
| `backend/internal/domain/auth/admin.go` | Admin エンティティ | Domain |
| `backend/internal/domain/auth/role.go` | Role 値オブジェクト | Domain |
| `backend/internal/domain/auth/repository.go` | AdminRepository IF | Domain |
| `backend/internal/domain/attendance/collection.go` | AttendanceCollection 集約ルート | Domain |
| `backend/internal/domain/attendance/response.go` | AttendanceResponse エンティティ | Domain |
| `backend/internal/domain/attendance/repository.go` | Repository IF | Domain |
| `backend/internal/domain/schedule/schedule.go` | DateSchedule 集約ルート | Domain |
| `backend/internal/domain/schedule/candidate.go` | CandidateDate エンティティ | Domain |
| `backend/internal/domain/schedule/response.go` | DateScheduleResponse エンティティ | Domain |
| `backend/internal/domain/schedule/repository.go` | Repository IF | Domain |
| `backend/internal/infra/security/bcrypt.go` | パスワードハッシュ化 | Infra |
| `backend/internal/infra/security/jwt.go` | JWT発行/検証 | Infra |
| `backend/internal/infra/clock/clock.go` | Clock インターフェース | Infra |
| `backend/internal/infra/db/tx.go` | TxManager | Infra |
| `backend/internal/infra/db/admin_repository.go` | Admin リポジトリ実装 | Infra |
| `backend/internal/infra/db/attendance_repository.go` | 出欠確認リポジトリ実装（UPSERT含む） | Infra |
| `backend/internal/infra/db/schedule_repository.go` | 日程調整リポジトリ実装 | Infra |
| `backend/internal/app/auth/login_usecase.go` | ログインユースケース | App |
| `backend/internal/app/attendance/create_collection_usecase.go` | 出欠確認作成 | App |
| `backend/internal/app/attendance/submit_response_usecase.go` | 出欠回答登録 | App |
| `backend/internal/app/attendance/close_collection_usecase.go` | 出欠確認クローズ | App |
| `backend/internal/app/schedule/create_schedule_usecase.go` | 日程調整作成 | App |
| `backend/internal/app/schedule/submit_response_usecase.go` | 日程調整回答登録 | App |
| `backend/internal/app/schedule/decide_schedule_usecase.go` | 日程調整確定 | App |
| `backend/internal/interface/rest/auth_handler.go` | 認証API | Interface |
| `backend/internal/interface/rest/attendance_handler.go` | 出欠確認 管理API | Interface |
| `backend/internal/interface/rest/public_attendance_handler.go` | 出欠確認 公開API | Interface |
| `backend/internal/interface/rest/schedule_handler.go` | 日程調整 管理API | Interface |
| `backend/internal/interface/rest/public_schedule_handler.go` | 日程調整 公開API | Interface |
| `web-frontend/src/pages/AdminLogin.tsx` | 管理者ログイン画面 | Frontend |
| `web-frontend/src/pages/public/AttendanceResponse.tsx` | 公開回答ページ（出欠） | Frontend |
| `web-frontend/src/pages/public/ScheduleResponse.tsx` | 公開回答ページ（日程） | Frontend |
| `web-frontend/src/lib/api/publicApi.ts` | 公開API クライアント | Frontend |

### 修正ファイル

| パス | 修正内容 |
|------|----------|
| `backend/internal/interface/rest/router.go` | 認証/出欠/日程 ルート追加 |
| `backend/internal/interface/rest/middleware.go` | JWT認証追加（段階移行） |
| `web-frontend/src/App.tsx` | `/p/...` ルート追加、ログイン画面差し替え |
| `web-frontend/src/lib/apiClient.ts` | JWT ヘッダー追加 |

### 削除候補ファイル

| パス | 理由 |
|------|------|
| `web-frontend/src/pages/Login.tsx` | AdminLogin.tsx で置換 |

---

## 付録B: DDD/レイヤード チェックリスト

各Commit時に確認すること：

### Domain層
- [ ] bcrypt/JWT/SQL などのインフラ技術が混入していないか
- [ ] `time.Now()` を呼んでいないか（引数で `now time.Time` を受け取っているか）
- [ ] 集約ルートに `CanXxx()`, `Close(now)`, `Decide(id, now)` などの状態遷移メソッドがあるか
- [ ] Repository IF がドメイン層にあるか

### App層
- [ ] ハンドラから直接Repositoryを呼んでいないか（App層経由か）
- [ ] `Clock` 経由で現在時刻を取得し、Domainに渡しているか
- [ ] トランザクション境界が `WithTx` で管理されているか（必要なUsecaseのみ）
- [ ] 管理APIのUsecaseは `ctxTenantID` を受け取っているか（Body/Queryから受け取っていないか）
- [ ] ドメインエラー（`ErrCollectionClosed` など）を定義しているか

### Infra層
- [ ] Repository で `UpsertResponse` が `ON CONFLICT DO UPDATE` で実装されているか
- [ ] TxManager が context 経由でトランザクションを管理しているか

### Interface層
- [ ] ハンドラに if 文でビジネスルールを書いていないか
- [ ] エラーコード変換（Domain Error → HTTP Status）がハンドラにあるか
- [ ] 管理APIのハンドラは tenant_id を Body/Query から受け取っていないか（JWT/contextから取得しているか）
- [ ] 公開APIのエラーメッセージが詳細を出しすぎていないか（token系→404、member系→400で詳細なし）

---

## 変更履歴

### 2025-12-14（DDD/レイヤード準拠版 v2）

- **A. Domain層から time.Now() を排除**
  - `Close()`, `Decide()`, `CanRespond()` の例コードを `now time.Time` 引数を受け取る形に修正
  - `infra/clock/clock.go` に `Clock` インターフェース導入の方針を追記
  - App層で `clock.Now()` を呼んでDomainに渡すパターンを明記

- **B. 回答上書き（UPSERT）の責務を整理**
  - 集約が `responses []` を保持するパターンから、Repository側で `ON CONFLICT DO UPDATE` を実行する方針に変更
  - `AttendanceCollection` の例コードから `UpsertResponse` メソッドを削除し、「MVPでは集約内で responses を保持しない」と明記
  - `AttendanceCollectionRepository.UpsertResponse()` の例コードを追加

- **C. 管理APIで tenant_id をリクエストから受け取らない**
  - ルールを0.1.1セクションに明記
  - `CreateCollectionRequest` の例コードから `tenant_id` フィールドを削除
  - ハンドラの例コードで `GetTenantIDFromContext()` を使用
  - ログインAPIのみ例外として `tenant_id` を Body で受け取ることを明記

- **D. トランザクション境界の方針を追加**
  - `infra/db/tx.go` に `TxManager` / `WithTx` の方針を追記
  - `SubmitResponseUsecase` の例コードを `WithTx` 内で実行する形に修正
  - トランザクション必須のUsecaseを DoD に追記（T9, T10, T15）

- **E. エンドポイント命名の統一**
  - 公開API: `/api/v1/public/date-schedule/{token}` → `/api/v1/public/date-schedules/{token}` に統一
  - マイグレーションパス: `backend/internal/infra/db/migrations/` に統一
  - 0.1.1セクションに命名表を追加

- **F. 公開回答でのエラーメッセージ方針を明記**
  - token invalid / not found → 404 "Not found"（詳細なし）
  - member_id 不正/存在しない → 400 "Invalid request"（詳細なし）
  - ハンドラの例コードを修正

---

**次のアクション**: Commit 1（認証基盤 - DB + Domain）から着手

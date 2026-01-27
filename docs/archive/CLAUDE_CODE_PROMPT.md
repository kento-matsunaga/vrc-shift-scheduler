# Claude Code 実装指示プロンプト

このファイルは Claude Code に MVP 実装を依頼する際のプロンプトテンプレートです。

---

## 基本プロンプト（全体指示）

```
あなたは vrc-shift-scheduler の開発アシスタントです。

# 必読ドキュメント
実装を始める前に、以下のドキュメントを必ず読んでください：
- docs/MVP_IMPLEMENTATION_PLAN.md（MVP実装計画書 - 全体設計とCommit単位のタスク）
- docs/MVP_GAP_ANALYSIS.md（現状分析 - 既存実装の確認）

# 実装ルール（厳守）

## DDD/レイヤード
1. Domain層で `time.Now()` を呼ばない。現在時刻は App層で `clock.Now()` を呼び、引数でDomainに渡す
2. 管理APIの tenant_id は JWT/context から取得する。Body/Query で受け取らない
3. 回答の上書き（UPSERT）は Repository層で `ON CONFLICT DO UPDATE` を実行する
4. トランザクションが必要な Usecase は `txManager.WithTx()` 内で実行する

## エラーハンドリング（公開API）
- token invalid / not found → 404 "Not found"（詳細を出さない）
- member_id 不正 → 400 "Invalid request"（詳細を出さない）

## 命名
- 日程調整APIは `date-schedules`（複数形）に統一
- マイグレーションは `backend/internal/infra/db/migrations/` に配置

# 作業の進め方
1. 計画書の「Commit X」単位で実装を進める
2. 各Commit完了時に、計画書の「DDD/レイヤード チェックリスト」を確認
3. 実装後は動作確認コマンド（計画書 6章）で検証

# 質問があれば
計画書に記載のない判断が必要な場合は、実装前に確認してください。
```

---

## Commit 別プロンプト

### Commit 1: 認証基盤 - DB + Domain

```
docs/MVP_IMPLEMENTATION_PLAN.md の「Commit 1: 認証基盤 - DB + Domain（T1, T2）」を実装してください。

# 対象ファイル
- backend/internal/infra/db/migrations/007_create_admins.up.sql
- backend/internal/infra/db/migrations/007_create_admins.down.sql
- backend/internal/domain/auth/admin.go
- backend/internal/domain/auth/role.go
- backend/internal/domain/auth/repository.go

# 確認ポイント
- [ ] admins テーブルに UNIQUE(tenant_id, email) 制約があるか
- [ ] Admin エンティティに CanLogin() メソッドがあるか
- [ ] Domain層に bcrypt/JWT の import がないか
- [ ] AdminRepository インターフェースが domain/auth に定義されているか

# 完了条件
- マイグレーション成功（docker compose exec backend go run ./cmd/migrate up）
- go build 成功
```

### Commit 2: 認証基盤 - Infra/Security + Clock + App + REST

```
docs/MVP_IMPLEMENTATION_PLAN.md の「Commit 2: 認証基盤 - Infra/Security + Clock + App + REST」を実装してください。

# 対象ファイル
- backend/internal/infra/db/admin_repository.go
- backend/internal/infra/security/bcrypt.go
- backend/internal/infra/security/jwt.go
- backend/internal/infra/clock/clock.go
- backend/internal/app/auth/login_usecase.go
- backend/internal/app/auth/dto.go
- backend/internal/interface/rest/auth_handler.go
- backend/internal/interface/rest/router.go（修正）
- backend/internal/interface/rest/middleware.go（修正）

# 確認ポイント
- [ ] Clock インターフェースが infra/clock に定義されているか
- [ ] LoginUsecase が passwordHasher.Compare() を使っているか（直接bcrypt呼び出しでない）
- [ ] LoginInput の tenant_id は Body から受け取る（ログインのみ例外）
- [ ] JWT検証失敗時のエラーメッセージが詳細を漏らしていないか
- [ ] middleware.go で Authorization: Bearer があればJWT検証、なければ従来のヘッダー認証（段階移行）

# 完了条件
- curl POST /api/v1/auth/login で JWT 取得可能
- 単体テスト通過
```

### Commit 3: 出欠確認 - DB + Domain + App + Infra

```
docs/MVP_IMPLEMENTATION_PLAN.md の「Commit 3: 出欠確認 - DB + Domain + App + Infra」を実装してください。

# 対象ファイル
- backend/internal/infra/db/migrations/008_create_attendance_tables.up.sql
- backend/internal/infra/db/migrations/008_create_attendance_tables.down.sql
- backend/internal/domain/attendance/collection.go
- backend/internal/domain/attendance/response.go
- backend/internal/domain/attendance/status.go
- backend/internal/domain/attendance/repository.go
- backend/internal/app/attendance/create_collection_usecase.go
- backend/internal/app/attendance/submit_response_usecase.go
- backend/internal/app/attendance/close_collection_usecase.go
- backend/internal/app/attendance/get_collection_usecase.go
- backend/internal/app/attendance/dto.go
- backend/internal/infra/db/attendance_repository.go
- backend/internal/infra/db/tx.go

# 確認ポイント
- [ ] attendance_responses に UNIQUE(collection_id, member_id) 制約があるか
- [ ] AttendanceCollection の CanRespond(now), Close(now) が time.Time を引数で受け取っているか
- [ ] Domain層で time.Now() を呼んでいないか
- [ ] SubmitResponseUsecase が txManager.WithTx() 内で実行されているか
- [ ] UpsertResponse が ON CONFLICT DO UPDATE で実装されているか

# 完了条件
- マイグレーション成功
- go build 成功
- 単体テスト通過
```

### Commit 4: 出欠確認 - REST層

```
docs/MVP_IMPLEMENTATION_PLAN.md の「Commit 4: 出欠確認 - REST層」を実装してください。

# 対象ファイル
- backend/internal/interface/rest/public_attendance_handler.go
- backend/internal/interface/rest/attendance_handler.go
- backend/internal/interface/rest/router.go（修正）

# 確認ポイント
- [ ] 公開API（/api/v1/public/attendance/{token}）が認証ミドルウェアを通らないか
- [ ] 管理APIのハンドラが tenant_id を Body から受け取っていないか（GetTenantIDFromContext使用）
- [ ] token invalid / collection not found が両方 404 になっているか
- [ ] member_id 不正のエラーメッセージが "Invalid request" のみか（詳細なし）

# 完了条件
- curl GET/POST /api/v1/public/attendance/{token} 成功
- curl POST/GET/PATCH /api/v1/attendance-collections 成功（JWT必要）
```

### Commit 5: 日程調整 - DB + Domain + App + Infra

```
docs/MVP_IMPLEMENTATION_PLAN.md の「Commit 5: 日程調整 - DB + Domain + App + Infra」を実装してください。

出欠確認（Commit 3）と同じパターンで実装してください。

# 確認ポイント
- [ ] DateSchedule の CanRespond(now), Decide(candidateID, now), Close(now) が time.Time を引数で受け取っているか
- [ ] Domain層で time.Now() を呼んでいないか
- [ ] DecideScheduleUsecase が txManager.WithTx() 内で実行されているか

# 完了条件
- マイグレーション成功
- go build 成功
```

### Commit 6: 日程調整 - REST層

```
docs/MVP_IMPLEMENTATION_PLAN.md の「Commit 6: 日程調整 - REST層」を実装してください。

# 確認ポイント
- [ ] 公開APIのパスが /api/v1/public/date-schedules/{token}（複数形）になっているか
- [ ] 管理APIのパスが /api/v1/date-schedules（複数形）になっているか

# 完了条件
- curl GET/POST /api/v1/public/date-schedules/{token} 成功
- curl POST/GET/PATCH /api/v1/date-schedules 成功
```

### Commit 7: フロントエンド公開ページ

```
docs/MVP_IMPLEMENTATION_PLAN.md の「Commit 7: フロントエンド公開ページ」を実装してください。

# 対象ファイル
- web-frontend/src/pages/public/AttendanceResponse.tsx
- web-frontend/src/pages/public/ScheduleResponse.tsx
- web-frontend/src/lib/api/publicApi.ts
- web-frontend/src/App.tsx（修正）

# 確認ポイント
- [ ] 回答者選択がメンバーマスタからのプルダウンになっているか（自由入力不可）
- [ ] メンバーがいない場合「管理者に依頼してください」メッセージが表示されるか
- [ ] /p/attendance/{token}, /p/schedule/{token} でアクセス可能か

# 完了条件
- ブラウザで公開ページが表示される
- 回答送信が成功する
```

### Commit 8: 管理者ログイン画面

```
docs/MVP_IMPLEMENTATION_PLAN.md の「Commit 8: 管理者ログイン画面」を実装してください。

# 対象ファイル
- web-frontend/src/pages/AdminLogin.tsx（新規）
- web-frontend/src/pages/Login.tsx（削除）
- web-frontend/src/lib/api/authApi.ts（修正）
- web-frontend/src/App.tsx（修正）

# 確認ポイント
- [ ] email + password でログインする形式になっているか
- [ ] JWT を localStorage に保存しているか
- [ ] 旧 Login.tsx（表示名入力でmember作成）が削除されているか
- [ ] apiClient が Authorization: Bearer ヘッダーを付与しているか

# 完了条件
- ブラウザで /login からログイン可能
- ログイン後に管理画面に遷移
```

---

## トラブルシューティング用プロンプト

### ビルドエラー時

```
以下のビルドエラーを解決してください。
計画書（docs/MVP_IMPLEMENTATION_PLAN.md）の設計方針に従って修正してください。

エラー内容：
[エラーメッセージを貼り付け]
```

### DDD違反の修正

```
以下のコードが DDD/レイヤードの方針に違反しています。
計画書（docs/MVP_IMPLEMENTATION_PLAN.md）の「0.1.1 DDD/レイヤード追加ルール」を参照して修正してください。

対象ファイル：[ファイルパス]
問題：[例: Domain層で time.Now() を呼んでいる]
```

---

## 注意事項

1. **計画書を常に参照**: 実装中に判断に迷ったら計画書を確認
2. **Commit単位で完結**: 各Commitは独立して動作確認可能な状態にする
3. **テスト駆動**: 可能な限り単体テストを先に書く
4. **段階移行**: 既存のX-Tenant-IDヘッダー認証は当面残す（JWT優先だが並行運用）


# API エンドポイント一覧

VRC Shift Scheduler の REST API エンドポイント一覧です。

## 認証

すべてのエンドポイント（Public APIを除く）はJWT認証が必要です。

```
Authorization: Bearer <token>
```

## レスポンス形式

### 成功時

```json
{
  "data": { ... }
}
```

### エラー時

```json
{
  "error": {
    "code": "ERR_XXX",
    "message": "エラーメッセージ"
  }
}
```

## エンドポイント一覧

### 認証 API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| POST | `/api/v1/auth/login` | 不要 | ログイン |
| POST | `/api/v1/setup` | 不要 | 初回セットアップ |
| POST | `/api/v1/auth/register-by-invite` | 不要 | 招待URL経由メンバー登録 |
| GET | `/api/v1/auth/password-reset-status` | 不要 | パスワードリセット状態確認 |
| POST | `/api/v1/auth/reset-password` | 不要 | パスワードリセット |

### 管理者 API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| POST | `/api/v1/admins/me/change-password` | 必要 | 自分のパスワード変更 |
| POST | `/api/v1/admins/{id}/allow-password-reset` | 必要 | 他管理者のパスワードリセット許可（Owner） |

### テナント API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| GET | `/api/v1/tenants/me` | 必要 | テナント情報取得 |
| PUT | `/api/v1/tenants/me` | 必要 | テナント情報更新 |
| GET | `/api/v1/settings/manager-permissions` | 必要 | マネージャー権限取得 |
| PUT | `/api/v1/settings/manager-permissions` | 必要 | マネージャー権限更新（Owner） |

### 招待 API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| POST | `/api/v1/invitations` | 必要 | 管理者招待 |
| POST | `/api/v1/invitations/accept/{token}` | 不要 | 招待受理 |

### メンバー API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| POST | `/api/v1/members` | 必要 | メンバー作成 |
| GET | `/api/v1/members` | 必要 | メンバー一覧取得 |
| GET | `/api/v1/members/{id}` | 必要 | メンバー詳細取得 |
| PUT | `/api/v1/members/{id}` | 必要 | メンバー更新 |
| DELETE | `/api/v1/members/{id}` | 必要 | メンバー削除 |
| GET | `/api/v1/members/me` | 必要 | 自分の情報取得 |
| GET | `/api/v1/members/recent-attendance` | 必要 | 直近出欠状況取得 |
| POST | `/api/v1/members/bulk-import` | 必要 | メンバー一括登録 |
| POST | `/api/v1/members/bulk-update-roles` | 必要 | ロール一括更新 |

### ロール API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| GET | `/api/v1/roles` | 必要 | ロール一覧取得 |
| POST | `/api/v1/roles` | 必要 | ロール作成 |
| GET | `/api/v1/roles/{id}` | 必要 | ロール詳細取得 |
| PUT | `/api/v1/roles/{id}` | 必要 | ロール更新 |
| DELETE | `/api/v1/roles/{id}` | 必要 | ロール削除 |

### イベント API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| POST | `/api/v1/events` | 必要 | イベント作成 |
| GET | `/api/v1/events` | 必要 | イベント一覧取得 |
| GET | `/api/v1/events/{id}` | 必要 | イベント詳細取得 |
| PUT | `/api/v1/events/{id}` | 必要 | イベント更新 |
| DELETE | `/api/v1/events/{id}` | 必要 | イベント削除 |
| POST | `/api/v1/events/{id}/generate-business-days` | 必要 | 営業日自動生成 |
| GET | `/api/v1/events/{id}/groups` | 必要 | グループ割り当て取得 |
| PUT | `/api/v1/events/{id}/groups` | 必要 | グループ割り当て更新 |
| GET | `/api/v1/events/{id}/business-days` | 必要 | 営業日一覧取得 |

### 営業日 API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| POST | `/api/v1/events/{eventId}/business-days` | 必要 | 営業日作成 |
| GET | `/api/v1/business-days/{id}` | 必要 | 営業日詳細取得 |
| PATCH | `/api/v1/business-days/{id}` | 必要 | 営業日更新 |
| POST | `/api/v1/business-days/{id}/apply-template` | 必要 | テンプレート適用 |
| POST | `/api/v1/business-days/{id}/save-as-template` | 必要 | テンプレートとして保存 |

### シフト枠 API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| POST | `/api/v1/business-days/{id}/shift-slots` | 必要 | シフト枠作成 |
| GET | `/api/v1/business-days/{id}/shift-slots` | 必要 | シフト枠一覧取得 |
| GET | `/api/v1/shift-slots/{id}` | 必要 | シフト枠詳細取得 |
| PUT | `/api/v1/shift-slots/{id}` | 必要 | シフト枠更新 |
| DELETE | `/api/v1/shift-slots/{id}` | 必要 | シフト枠削除 |

#### シフト枠のバリデーション

- `priority`: 1以上の整数（0や負の値は不可）
- `capacity`: 1以上の整数

### シフト割り当て API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| POST | `/api/v1/shift-assignments` | 必要 | シフト確定 |
| GET | `/api/v1/shift-assignments` | 必要 | 割り当て一覧取得 |
| GET | `/api/v1/shift-assignments/{id}` | 必要 | 割り当て詳細取得 |
| PATCH | `/api/v1/shift-assignments/{id}/status` | 必要 | ステータス変更 |
| DELETE | `/api/v1/shift-assignments/{id}` | 必要 | キャンセル |

### テンプレート API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| GET | `/api/v1/events/{eventId}/templates` | 必要 | テンプレート一覧取得 |
| GET | `/api/v1/events/{eventId}/templates/{id}` | 必要 | テンプレート詳細取得 |
| POST | `/api/v1/events/{eventId}/templates` | 必要 | テンプレート作成 |
| PUT | `/api/v1/events/{eventId}/templates/{id}` | 必要 | テンプレート更新 |
| DELETE | `/api/v1/events/{eventId}/templates/{id}` | 必要 | テンプレート削除 |

### インスタンス API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| GET | `/api/v1/events/{eventId}/instances` | 必要 | インスタンス一覧取得 |
| POST | `/api/v1/events/{eventId}/instances` | 必要 | インスタンス作成 |
| GET | `/api/v1/instances/{id}` | 必要 | インスタンス詳細取得 |
| PUT | `/api/v1/instances/{id}` | 必要 | インスタンス更新 |
| DELETE | `/api/v1/instances/{id}` | 必要 | インスタンス削除 |

### メンバーグループ API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| POST | `/api/v1/member-groups` | 必要 | グループ作成 |
| GET | `/api/v1/member-groups` | 必要 | グループ一覧取得 |
| GET | `/api/v1/member-groups/{id}` | 必要 | グループ詳細取得 |
| PUT | `/api/v1/member-groups/{id}` | 必要 | グループ更新 |
| DELETE | `/api/v1/member-groups/{id}` | 必要 | グループ削除 |
| PUT | `/api/v1/member-groups/{id}/members` | 必要 | メンバー割り当て |

### ロールグループ API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| POST | `/api/v1/role-groups` | 必要 | グループ作成 |
| GET | `/api/v1/role-groups` | 必要 | グループ一覧取得 |
| GET | `/api/v1/role-groups/{id}` | 必要 | グループ詳細取得 |
| PUT | `/api/v1/role-groups/{id}` | 必要 | グループ更新 |
| DELETE | `/api/v1/role-groups/{id}` | 必要 | グループ削除 |
| PUT | `/api/v1/role-groups/{id}/roles` | 必要 | ロール割り当て |

### 出欠収集 API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| GET | `/api/v1/attendance/collections` | 必要 | 出欠収集一覧取得 |
| POST | `/api/v1/attendance/collections` | 必要 | 出欠収集作成 |
| GET | `/api/v1/attendance/collections/{id}` | 必要 | 出欠収集詳細取得 |
| DELETE | `/api/v1/attendance/collections/{id}` | 必要 | 出欠収集削除 |
| POST | `/api/v1/attendance/collections/{id}/close` | 必要 | 締め切り |
| GET | `/api/v1/attendance/collections/{id}/responses` | 必要 | 回答一覧取得 |
| PUT | `/api/v1/attendance/collections/{id}/responses` | 必要 | 回答更新（管理者） |

#### 管理者による回答更新 API

**PUT** `/api/v1/attendance/collections/{id}/responses`

管理者が締め切り後でも出欠回答を更新できるAPI。

##### リクエスト

```json
{
  "member_id": "01HXXXXXXXXXXXXXXX",
  "target_date_id": "01HXXXXXXXXXXXXXXX",
  "response": "available",
  "note": "備考（オプション）",
  "available_from": "0001-01-01T21:00:00Z",
  "available_to": "0001-01-01T23:00:00Z"
}
```

##### レスポンス

```json
{
  "data": {
    "response_id": "01HXXXXXXXXXXXXXXX",
    "collection_id": "01HXXXXXXXXXXXXXXX",
    "member_id": "01HXXXXXXXXXXXXXXX",
    "target_date_id": "01HXXXXXXXXXXXXXXX",
    "response": "available",
    "note": "備考",
    "available_from": "0001-01-01T21:00:00Z",
    "available_to": "0001-01-01T23:00:00Z",
    "responded_at": "2026-01-17T12:00:00Z"
  }
}
```

##### response の値

| 値 | 説明 |
|----|------|
| `available` | 参加可能 |
| `unavailable` | 参加不可 |
| `maybe` | 未定 |

### 日程調整 API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| POST | `/api/v1/schedules` | 必要 | 日程調整作成 |
| GET | `/api/v1/schedules` | 必要 | 日程調整一覧取得 |
| GET | `/api/v1/schedules/{id}` | 必要 | 日程調整詳細取得 |
| DELETE | `/api/v1/schedules/{id}` | 必要 | 日程調整削除 |
| POST | `/api/v1/schedules/{id}/decide` | 必要 | 日程決定 |
| POST | `/api/v1/schedules/{id}/close` | 必要 | 締め切り |
| GET | `/api/v1/schedules/{id}/responses` | 必要 | 回答一覧取得 |

### インポート API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| POST | `/api/v1/imports/members` | 必要 | CSVインポート |
| GET | `/api/v1/imports` | 必要 | ジョブ一覧取得 |
| GET | `/api/v1/imports/{id}/status` | 必要 | ステータス取得 |
| GET | `/api/v1/imports/{id}/result` | 必要 | 結果詳細取得 |

### お知らせ API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| GET | `/api/v1/announcements` | 必要 | お知らせ一覧取得 |
| GET | `/api/v1/announcements/unread-count` | 必要 | 未読件数取得 |
| POST | `/api/v1/announcements/{id}/read` | 必要 | 既読にする |
| POST | `/api/v1/announcements/read-all` | 必要 | 全て既読にする |

### チュートリアル API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| GET | `/api/v1/tutorials` | 必要 | チュートリアル一覧取得 |
| GET | `/api/v1/tutorials/{id}` | 必要 | チュートリアル詳細取得 |

### 実績出欠 API

| メソッド | エンドポイント | 認証 | 説明 |
|---------|---------------|------|------|
| GET | `/api/v1/actual-attendance` | 必要 | 実績出欠データ取得 |
| POST | `/api/v1/actual-attendance` | 必要 | 実績出欠作成/更新 |

### 公開 API（認証不要、トークンベース）

| メソッド | エンドポイント | 説明 |
|---------|---------------|------|
| GET | `/api/v1/public/attendance/{token}` | 出欠収集取得 |
| POST | `/api/v1/public/attendance/{token}/responses` | 出欠回答送信 |
| GET | `/api/v1/public/attendance/{token}/responses` | 全回答一覧取得 |
| GET | `/api/v1/public/attendance/{token}/members/{memberId}/responses` | メンバー回答取得 |
| GET | `/api/v1/public/members` | メンバー一覧取得 |
| GET | `/api/v1/public/schedules/{token}` | 日程調整取得 |
| POST | `/api/v1/public/schedules/{token}/responses` | 日程回答送信 |
| GET | `/api/v1/public/schedules/{token}/responses` | 全回答一覧取得 |
| POST | `/api/v1/public/license/claim` | ライセンスクレーム |

## HTTPステータスコード

| コード | 説明 |
|--------|------|
| 200 | 成功 |
| 201 | 作成成功 |
| 204 | 成功（レスポンスボディなし） |
| 400 | バリデーションエラー |
| 401 | 認証エラー |
| 403 | 権限エラー |
| 404 | リソースが見つからない |
| 405 | メソッド不許可 |
| 409 | 競合（重複など） |
| 500 | サーバーエラー |

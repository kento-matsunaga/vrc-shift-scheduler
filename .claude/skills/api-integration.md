---
description: REST API仕様、エンドポイント一覧、レスポンス形式
---

# API Integration

VRC Shift Scheduler の REST API 仕様。API開発・統合テスト時の参照。

---

## 認証

すべてのエンドポイント（Public APIを除く）はJWT認証が必要。

```
Authorization: Bearer <token>
```

---

## レスポンス形式

### 成功時
```json
{"data": { ... }}
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

---

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
| 409 | 競合（重複など） |
| 500 | サーバーエラー |

---

## エンドポイント一覧

### 認証 API（認証不要）

| メソッド | エンドポイント | 説明 |
|---------|---------------|------|
| POST | `/api/v1/auth/login` | ログイン |
| POST | `/api/v1/setup` | 初回セットアップ |
| POST | `/api/v1/auth/register-by-invite` | 招待URL経由メンバー登録 |
| GET | `/api/v1/auth/password-reset-status` | パスワードリセット状態確認 |
| POST | `/api/v1/auth/reset-password` | パスワードリセット |

### テナント API

| メソッド | エンドポイント | 説明 |
|---------|---------------|------|
| GET | `/api/v1/tenants/me` | テナント情報取得 |
| PUT | `/api/v1/tenants/me` | テナント情報更新 |
| GET | `/api/v1/settings/manager-permissions` | マネージャー権限取得 |
| PUT | `/api/v1/settings/manager-permissions` | マネージャー権限更新（Owner） |

### メンバー API

| メソッド | エンドポイント | 説明 |
|---------|---------------|------|
| POST | `/api/v1/members` | メンバー作成 |
| GET | `/api/v1/members` | メンバー一覧取得 |
| GET | `/api/v1/members/{id}` | メンバー詳細取得 |
| PUT | `/api/v1/members/{id}` | メンバー更新 |
| DELETE | `/api/v1/members/{id}` | メンバー削除 |
| POST | `/api/v1/members/bulk-import` | メンバー一括登録 |
| POST | `/api/v1/members/bulk-update-roles` | ロール一括更新 |

### ロール API

| メソッド | エンドポイント | 説明 |
|---------|---------------|------|
| GET | `/api/v1/roles` | ロール一覧取得 |
| POST | `/api/v1/roles` | ロール作成 |
| GET | `/api/v1/roles/{id}` | ロール詳細取得 |
| PUT | `/api/v1/roles/{id}` | ロール更新 |
| DELETE | `/api/v1/roles/{id}` | ロール削除 |

### イベント API

| メソッド | エンドポイント | 説明 |
|---------|---------------|------|
| POST | `/api/v1/events` | イベント作成 |
| GET | `/api/v1/events` | イベント一覧取得 |
| GET | `/api/v1/events/{id}` | イベント詳細取得 |
| PUT | `/api/v1/events/{id}` | イベント更新 |
| DELETE | `/api/v1/events/{id}` | イベント削除 |
| POST | `/api/v1/events/{id}/generate-business-days` | 営業日自動生成 |
| GET | `/api/v1/events/{id}/business-days` | 営業日一覧取得 |

### 営業日 API

| メソッド | エンドポイント | 説明 |
|---------|---------------|------|
| POST | `/api/v1/events/{eventId}/business-days` | 営業日作成 |
| GET | `/api/v1/business-days/{id}` | 営業日詳細取得 |
| PATCH | `/api/v1/business-days/{id}` | 営業日更新 |
| POST | `/api/v1/business-days/{id}/apply-template` | テンプレート適用 |
| POST | `/api/v1/business-days/{id}/save-as-template` | テンプレートとして保存 |

### シフト枠 API

| メソッド | エンドポイント | 説明 |
|---------|---------------|------|
| POST | `/api/v1/business-days/{id}/shift-slots` | シフト枠作成 |
| GET | `/api/v1/business-days/{id}/shift-slots` | シフト枠一覧取得 |
| GET | `/api/v1/shift-slots/{id}` | シフト枠詳細取得 |
| PUT | `/api/v1/shift-slots/{id}` | シフト枠更新 |
| DELETE | `/api/v1/shift-slots/{id}` | シフト枠削除 |

**バリデーション**:
- `priority`: 1以上の整数
- `capacity`: 1以上の整数

### シフト割り当て API

| メソッド | エンドポイント | 説明 |
|---------|---------------|------|
| POST | `/api/v1/shift-assignments` | シフト確定 |
| GET | `/api/v1/shift-assignments` | 割り当て一覧取得 |
| GET | `/api/v1/shift-assignments/{id}` | 割り当て詳細取得 |
| PATCH | `/api/v1/shift-assignments/{id}/status` | ステータス変更 |
| DELETE | `/api/v1/shift-assignments/{id}` | キャンセル |

### テンプレート API

| メソッド | エンドポイント | 説明 |
|---------|---------------|------|
| GET | `/api/v1/events/{eventId}/templates` | テンプレート一覧取得 |
| GET | `/api/v1/events/{eventId}/templates/{id}` | テンプレート詳細取得 |
| POST | `/api/v1/events/{eventId}/templates` | テンプレート作成 |
| PUT | `/api/v1/events/{eventId}/templates/{id}` | テンプレート更新 |
| DELETE | `/api/v1/events/{eventId}/templates/{id}` | テンプレート削除 |

### インスタンス API

| メソッド | エンドポイント | 説明 |
|---------|---------------|------|
| GET | `/api/v1/events/{eventId}/instances` | インスタンス一覧取得 |
| POST | `/api/v1/events/{eventId}/instances` | インスタンス作成 |
| GET | `/api/v1/instances/{id}` | インスタンス詳細取得 |
| PUT | `/api/v1/instances/{id}` | インスタンス更新 |
| DELETE | `/api/v1/instances/{id}` | インスタンス削除 |

### 出欠収集 API

| メソッド | エンドポイント | 説明 |
|---------|---------------|------|
| GET | `/api/v1/attendance/collections` | 出欠収集一覧取得 |
| POST | `/api/v1/attendance/collections` | 出欠収集作成 |
| GET | `/api/v1/attendance/collections/{id}` | 出欠収集詳細取得 |
| DELETE | `/api/v1/attendance/collections/{id}` | 出欠収集削除 |
| POST | `/api/v1/attendance/collections/{id}/close` | 締め切り |
| GET | `/api/v1/attendance/collections/{id}/responses` | 回答一覧取得 |
| PUT | `/api/v1/attendance/collections/{id}/responses` | 回答更新（管理者） |

**出欠確認の回答値**: `attending`（出席）/ `absent`（欠席）/ `maybe`（未定）

### 日程調整 API

| メソッド | エンドポイント | 説明 |
|---------|---------------|------|
| POST | `/api/v1/schedules` | 日程調整作成 |
| GET | `/api/v1/schedules` | 日程調整一覧取得 |
| GET | `/api/v1/schedules/{id}` | 日程調整詳細取得 |
| DELETE | `/api/v1/schedules/{id}` | 日程調整削除 |
| POST | `/api/v1/schedules/{id}/decide` | 日程決定 |
| POST | `/api/v1/schedules/{id}/close` | 締め切り |
| GET | `/api/v1/schedules/{id}/responses` | 回答一覧取得 |

**日程調整の回答値**: `available`（参加可能）/ `unavailable`（参加不可）/ `maybe`（未定）

### メンバーグループ API

| メソッド | エンドポイント | 説明 |
|---------|---------------|------|
| POST | `/api/v1/member-groups` | グループ作成 |
| GET | `/api/v1/member-groups` | グループ一覧取得 |
| GET | `/api/v1/member-groups/{id}` | グループ詳細取得 |
| PUT | `/api/v1/member-groups/{id}` | グループ更新 |
| DELETE | `/api/v1/member-groups/{id}` | グループ削除 |
| PUT | `/api/v1/member-groups/{id}/members` | メンバー割り当て |

### ロールグループ API

| メソッド | エンドポイント | 説明 |
|---------|---------------|------|
| POST | `/api/v1/role-groups` | グループ作成 |
| GET | `/api/v1/role-groups` | グループ一覧取得 |
| GET | `/api/v1/role-groups/{id}` | グループ詳細取得 |
| PUT | `/api/v1/role-groups/{id}` | グループ更新 |
| DELETE | `/api/v1/role-groups/{id}` | グループ削除 |
| PUT | `/api/v1/role-groups/{id}/roles` | ロール割り当て |

---

## 公開 API（認証不要・トークンベース）

| メソッド | エンドポイント | 説明 |
|---------|---------------|------|
| GET | `/api/v1/public/attendance/{token}` | 出欠収集取得 |
| POST | `/api/v1/public/attendance/{token}/responses` | 出欠回答送信 |
| GET | `/api/v1/public/attendance/{token}/responses` | 全回答一覧取得 |
| GET | `/api/v1/public/members` | メンバー一覧取得 |
| GET | `/api/v1/public/schedules/{token}` | 日程調整取得 |
| POST | `/api/v1/public/schedules/{token}/responses` | 日程回答送信 |
| GET | `/api/v1/public/schedules/{token}/responses` | 全回答一覧取得 |
| POST | `/api/v1/public/license/claim` | ライセンスクレーム |

---

## テスト用認証

```bash
# ログイン
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin1@example.com", "password": "password123"}'
```

---

## 詳細ドキュメント

- `docs/api-endpoints.md` - 完全なAPI仕様
- `docs/verification/API_CONTRACT_MATRIX.md` - API契約マトリックス

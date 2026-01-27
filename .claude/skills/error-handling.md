---
description: エラーハンドリングパターン、ドメインエラー、バリデーションエラー
---

# Error Handling

VRC Shift Scheduler のエラーハンドリングパターン。

---

## エラー階層

```
common.DomainError
├── ValidationError (INVALID_INPUT)
├── NotFoundError (NOT_FOUND)
├── ForbiddenError (FORBIDDEN)
└── ConflictError (CONFLICT)
```

---

## エラー種別

### 1. バリデーションエラー

入力値の検証失敗時に使用:

```go
func NewValidationError(message string, cause error) error {
    return &DomainError{
        Code:    "INVALID_INPUT",
        Message: message,
        Cause:   cause,
    }
}

// 使用例
if t.tenantName == "" {
    return common.NewValidationError("tenant_name is required", nil)
}

if len(t.tenantName) > 255 {
    return common.NewValidationError("tenant_name must be less than 255 characters", nil)
}
```

### 2. 存在しないリソース

リソースが見つからない時に使用:

```go
func NewDomainError(code ErrorCode, message string) error {
    return &DomainError{
        Code:    code,
        Message: message,
    }
}

// 使用例
if tenant == nil {
    return common.NewDomainError(common.ErrNotFound, "Tenant not found")
}
```

### 3. 権限エラー

アクセス権限がない時に使用:

```go
// 使用例
if claims.TenantID != tenantID {
    return common.NewDomainError(common.ErrForbidden, "Access denied")
}

if claims.Role != "owner" {
    return common.NewDomainError(common.ErrForbidden, "Owner role required")
}
```

### 4. 競合エラー

状態の競合時に使用:

```go
// 使用例
if !status.CanTransitionTo(newStatus) {
    return common.NewValidationError(
        fmt.Sprintf("invalid status transition from %s to %s", status, newStatus),
        nil,
    )
}
```

---

## レイヤー別エラーハンドリング

### Domain層

ビジネスルール違反を検出:

```go
func (t *Tenant) SetStatusActive(now time.Time) error {
    if !t.status.CanTransitionTo(TenantStatusActive) {
        return common.NewValidationError(
            fmt.Sprintf("invalid status transition from %s to active", t.status),
            nil,
        )
    }
    t.status = TenantStatusActive
    t.updatedAt = now
    return nil
}
```

### Application層（Usecase）

エラーをラップして追加コンテキストを付与:

```go
func (uc *TenantUsecase) GetTenant(ctx context.Context, id TenantID) (*Tenant, error) {
    tenant, err := uc.tenantRepo.FindByID(ctx, id)
    if err != nil {
        if common.IsNotFoundError(err) {
            return nil, common.NewDomainError(common.ErrNotFound, "Tenant not found")
        }
        return nil, fmt.Errorf("failed to find tenant: %w", err)
    }
    return tenant, nil
}
```

### Infrastructure層

データベースエラーをドメインエラーに変換:

```go
func (r *tenantRepositoryImpl) FindByID(ctx context.Context, id TenantID) (*Tenant, error) {
    row := r.pool.QueryRow(ctx, query, id)

    var t tenantRow
    err := row.Scan(&t.TenantID, ...)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, common.NewDomainError(common.ErrNotFound, "tenant not found")
        }
        return nil, fmt.Errorf("failed to scan tenant: %w", err)
    }
    // ...
}
```

### Interface層（Handler）

エラーをHTTPレスポンスに変換:

```go
func (h *Handler) handleError(w http.ResponseWriter, err error) {
    var domainErr *common.DomainError
    if errors.As(err, &domainErr) {
        switch domainErr.Code {
        case common.ErrNotFound:
            writeError(w, http.StatusNotFound, domainErr.Code, domainErr.Message)
        case common.ErrForbidden:
            writeError(w, http.StatusForbidden, domainErr.Code, domainErr.Message)
        case "INVALID_INPUT":
            writeError(w, http.StatusBadRequest, domainErr.Code, domainErr.Message)
        default:
            writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
        }
        return
    }

    // 予期しないエラー
    log.Printf("Unexpected error: %v", err)
    writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
}
```

---

## APIレスポンス形式

### 成功

```json
{
  "data": {
    "tenant_id": "...",
    "tenant_name": "..."
  }
}
```

### エラー

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Tenant not found"
  }
}
```

---

## エラーコード一覧

| コード | HTTPステータス | 説明 |
|-------|---------------|------|
| `INVALID_INPUT` | 400 | バリデーションエラー |
| `UNAUTHORIZED` | 401 | 認証エラー |
| `FORBIDDEN` | 403 | 権限エラー |
| `NOT_FOUND` | 404 | リソースが見つからない |
| `CONFLICT` | 409 | 状態の競合 |
| `INTERNAL_ERROR` | 500 | 内部エラー |

---

## エラーチェックユーティリティ

```go
// NotFoundエラーの判定
func IsNotFoundError(err error) bool {
    var domainErr *DomainError
    if errors.As(err, &domainErr) {
        return domainErr.Code == ErrNotFound
    }
    return false
}

// 使用例
if common.IsNotFoundError(err) {
    return nil, common.NewDomainError(common.ErrNotFound, "Tenant not found")
}
```

---

## ベストプラクティス

1. **エラーは必ず処理する** - `_` で無視しない
2. **コンテキストを付与してラップ** - `fmt.Errorf("failed to xxx: %w", err)`
3. **ドメインエラーはDomain/App層で作成** - Infra層では変換のみ
4. **ユーザー向けメッセージは英語** - 多言語対応はフロントエンドで
5. **ログには詳細を、レスポンスには概要を** - 内部情報を漏らさない

---

## 関連ファイル

- `backend/internal/domain/common/errors.go` - エラー定義
- `backend/internal/interface/rest/error_handler.go` - HTTPエラーハンドリング

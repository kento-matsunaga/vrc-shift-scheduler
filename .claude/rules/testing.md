# テストルール

## テストファイルの配置

ソースファイルと隣接して配置:

```
internal/domain/tenant/
  tenant.go
  tenant_test.go
  repository.go
```

## テスト関数の命名

Testプレフィックスで説明的な名前を使用:

```go
func TestNewTenant_Success(t *testing.T) {}
func TestNewTenant_ErrorWhenNameEmpty(t *testing.T) {}
func TestNewTenant_ErrorWhenNameTooLong(t *testing.T) {}
func TestTenant_UpdateTenantName_Success(t *testing.T) {}
func TestTenant_StatusTransitions(t *testing.T) {}
```

## テーブル駆動テスト

類似のテストケースが複数ある場合に使用:

```go
func TestTenantStatus_IsValid(t *testing.T) {
    tests := []struct {
        status   tenant.TenantStatus
        expected bool
    }{
        {tenant.TenantStatusActive, true},
        {tenant.TenantStatusGrace, true},
        {tenant.TenantStatusSuspended, true},
        {tenant.TenantStatus("invalid"), false},
        {tenant.TenantStatus(""), false},
    }

    for _, tt := range tests {
        t.Run(string(tt.status), func(t *testing.T) {
            result := tt.status.IsValid()
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## テストの独立性

各テストは独立して実行可能にする:

```go
func TestTenant_Delete(t *testing.T) {
    // セットアップ: このテスト用に新規エンティティを作成
    now := time.Now()
    ten, _ := tenant.NewTenant(now, "Test Organization", "Asia/Tokyo")

    // 初期状態のアサーション
    if ten.IsDeleted() {
        t.Error("New tenant should not be deleted")
    }

    // アクション
    ten.Delete(now)

    // アサーション
    if !ten.IsDeleted() {
        t.Error("Tenant should be deleted after Delete()")
    }
}
```

## 外部パッケージテストパターン

外部からテストしてパブリックAPIのみを検証:

```go
package tenant_test  // 注: _test サフィックス

import (
    "testing"
    "github.com/.../internal/domain/tenant"
)

// パブリックAPIのみをテスト
func TestNewTenant_Success(t *testing.T) {
    ten, err := tenant.NewTenant(...)
}
```

## モック

テスト容易性のためインターフェースを使用:

```go
// テストファイル内でモックを定義
type mockTenantRepository struct {
    findByIDFunc func(ctx context.Context, id common.TenantID) (*tenant.Tenant, error)
}

func (m *mockTenantRepository) FindByID(ctx context.Context, id common.TenantID) (*tenant.Tenant, error) {
    return m.findByIDFunc(ctx, id)
}

// テストで使用
func TestUsecase_GetTenant(t *testing.T) {
    mockRepo := &mockTenantRepository{
        findByIDFunc: func(ctx context.Context, id common.TenantID) (*tenant.Tenant, error) {
            return expectedTenant, nil
        },
    }

    uc := NewTenantUsecase(mockRepo)
    result, err := uc.GetTenant(ctx, tenantID)
    // アサーション
}
```

## テストの実行

```bash
# 全バックエンドテストを実行
cd backend && go test ./...

# 特定パッケージのテストを実行
go test ./internal/domain/tenant/...

# 詳細出力で実行
go test -v ./...

# カバレッジ付きで実行
go test -cover ./...

# 単一テストを実行
go test -run TestNewTenant_Success ./internal/domain/tenant/
```

## テストカバレッジ目安

- Domainエンティティ: 80%以上
- Usecase: 70%以上
- Handler: 60%以上

## テスト対象

### Domain層
- エンティティ作成（成功とバリデーション失敗）
- ビジネスメソッドの動作
- ステート遷移
- 値オブジェクトのバリデーション

### Application層
- Usecaseの正常系パス
- エラーハンドリングパス
- トランザクション動作

### Infrastructure層
- リポジトリのCRUD操作（統合テスト）
- 外部サービス連携

## テストチェックリスト

PR提出前に確認:
- [ ] 新規コードにテストがある
- [ ] 正常系パスをカバー
- [ ] エラー系パスをカバー
- [ ] エッジケースをテスト
- [ ] フレーキーテストがない（時間依存など）
- [ ] テスト名が説明的
- [ ] `go test ./...` がパスする

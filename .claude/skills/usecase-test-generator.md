---
description: DDD/クリーンアーキテクチャのユースケース層テストを効率的に作成するスキル
---

# usecase-test-generator

ユースケース層のユニットテストを効率的に作成するためのガイド。

## 前提条件

- Go 1.21+
- testify パッケージ（assert, require, mock）
- DDD/クリーンアーキテクチャ構成

## 1. モックリポジトリの作成

### 基本構造（関数フィールド方式）

```go
type mockXxxRepository struct {
    createFunc   func(ctx context.Context, entity *domain.Xxx) error
    findByIDFunc func(ctx context.Context, tenantID common.TenantID, id common.XxxID) (*domain.Xxx, error)
    updateFunc   func(ctx context.Context, entity *domain.Xxx) error
    deleteFunc   func(ctx context.Context, tenantID common.TenantID, id common.XxxID) error
}

func (m *mockXxxRepository) Create(ctx context.Context, entity *domain.Xxx) error {
    if m.createFunc != nil {
        return m.createFunc(ctx, entity)
    }
    return nil
}

func (m *mockXxxRepository) FindByID(ctx context.Context, tenantID common.TenantID, id common.XxxID) (*domain.Xxx, error) {
    if m.findByIDFunc != nil {
        return m.findByIDFunc(ctx, tenantID, id)
    }
    return nil, nil
}
```

### ポイント

- 関数フィールド方式でテストケースごとに振る舞いをカスタマイズ
- `xxxFunc != nil` チェックでデフォルト動作を提供
- 戻り値が nil の可能性がある場合は nil チェック

## 2. テストヘルパーの作成

```go
func createTestTenantID(t *testing.T) common.TenantID {
    t.Helper()
    return common.NewTenantIDWithTime(time.Now())
}

func createTestEntity(t *testing.T, tenantID common.TenantID) *domain.Xxx {
    t.Helper()
    now := time.Now()
    entity, err := domain.NewXxx(now, tenantID, "Test Name", "Test Description")
    if err != nil {
        t.Fatalf("failed to create test entity: %v", err)
    }
    return entity
}
```

### ポイント

- `t.Helper()` を必ず呼び出す（エラー時のスタックトレース改善）
- 生成に失敗した場合は `t.Fatalf` で即座に終了

## 3. テストケースの構造

### 正常系テスト

```go
func TestXxxUsecase_Success(t *testing.T) {
    // Arrange
    tenantID := createTestTenantID(t)
    testEntity := createTestEntity(t, tenantID)

    mockRepo := &mockXxxRepository{
        createFunc: func(ctx context.Context, entity *domain.Xxx) error {
            return nil
        },
    }

    uc := NewXxxUsecase(mockRepo)

    input := XxxInput{
        TenantID: tenantID.String(),
        Name:     "New Entity",
    }

    // Act
    result, err := uc.Execute(context.Background(), input)

    // Assert
    if err != nil {
        t.Errorf("expected no error, got %v", err)
    }
    if result == nil {
        t.Fatal("expected result, got nil")
    }
    if result.Name != "New Entity" {
        t.Errorf("expected name 'New Entity', got '%s'", result.Name)
    }
}
```

### 異常系テスト

```go
func TestXxxUsecase_ErrorWhenNotFound(t *testing.T) {
    tenantID := createTestTenantID(t)
    entityID := createTestEntityID(t)

    mockRepo := &mockXxxRepository{
        findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.XxxID) (*domain.Xxx, error) {
            return nil, common.NewNotFoundError("entity", eid.String())
        },
    }

    uc := NewXxxUsecase(mockRepo)

    input := XxxInput{
        TenantID: tenantID.String(),
        EntityID: entityID.String(),
    }

    result, err := uc.Execute(context.Background(), input)

    if err == nil {
        t.Error("expected error, got nil")
    }
    if result != nil {
        t.Errorf("expected nil result, got %v", result)
    }
    if !common.IsNotFoundError(err) {
        t.Errorf("expected not found error, got %v", err)
    }
}
```

## 4. 副作用の検証

```go
func TestXxxUsecase_SideEffect(t *testing.T) {
    tenantID := createTestTenantID(t)

    deleteCalled := false
    mockRepo := &mockXxxRepository{
        deleteFunc: func(ctx context.Context, tid common.TenantID, eid common.XxxID) error {
            deleteCalled = true
            return nil
        },
    }

    uc := NewDeleteXxxUsecase(mockRepo)
    _ = uc.Execute(context.Background(), input)

    if !deleteCalled {
        t.Error("expected delete to be called")
    }
}
```

## 5. 状態変更の検証

```go
func TestXxxUsecase_StateChange(t *testing.T) {
    var updatedEntity *domain.Xxx

    mockRepo := &mockXxxRepository{
        findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.XxxID) (*domain.Xxx, error) {
            return testEntity, nil
        },
        updateFunc: func(ctx context.Context, entity *domain.Xxx) error {
            updatedEntity = entity
            return nil
        },
    }

    uc := NewUpdateXxxUsecase(mockRepo)
    _, _ = uc.Execute(context.Background(), input)

    if updatedEntity == nil {
        t.Fatal("expected entity to be updated")
    }
    if !updatedEntity.IsActive() {
        t.Error("expected entity to be active")
    }
}
```

## 6. context.Context の扱い

- テストでは `context.Background()` を使用
- タイムアウトテストが必要な場合は `context.WithTimeout` を使用

## 7. エラーケースの網羅

以下のエラーケースを必ずテストせよ：

| カテゴリ | テストケース |
|----------|--------------|
| 入力検証 | 必須フィールド欠落、長さ制限超過、不正なID形式 |
| 存在確認 | ID不正、リソース未存在 |
| 権限 | テナント分離違反 |
| 依存 | リポジトリエラー、外部サービスエラー |

## 8. テストファイルの命名規則

- ファイル名: `usecase_test.go`（同一パッケージ）
- パッケージ名: 外部パッケージテストの場合は `package xxx_test`
- 関数名: `TestXxxUsecase_シナリオ`

## 9. テスト構成の整理

```go
// =============================================================================
// Mock Repositories
// =============================================================================

// ... モック定義 ...

// =============================================================================
// Test Helpers
// =============================================================================

// ... ヘルパー関数 ...

// =============================================================================
// CreateXxxUsecase Tests
// =============================================================================

// ... テストケース ...
```

## 適用例

このスキルは以下のユースケーステストに適用済み：

- calendar/usecase_test.go（16テスト）
- attendance/usecase_test.go

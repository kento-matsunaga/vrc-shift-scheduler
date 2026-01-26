# DDD & クリーンアーキテクチャ ルール

## レイヤー分離（必須）

レイヤー境界を絶対に侵害しない:

```
interface/rest  →  app (usecase)  →  domain  ←  infra/db
     ↓                ↓                           ↓
  Handlers        Usecases          Entities    Repositories
```

### 禁止パターン

```go
// NG: UsecaseがInfraをインポート
import "github.com/.../internal/infra/db"

// NG: DomainがAppをインポート
import "github.com/.../internal/app/tenant"

// NG: HandlerがDBに直接アクセス
func Handler() {
    db.Query("SELECT * FROM tenants")  // 禁止！
}
```

### 正しいパターン

```go
// Domainがインターフェースを定義
type TenantRepository interface {
    FindByID(ctx context.Context, id TenantID) (*Tenant, error)
    Save(ctx context.Context, tenant *Tenant) error
}

// Infraがインターフェースを実装
type tenantRepositoryImpl struct {
    pool *pgxpool.Pool
}

// Usecaseがインターフェースを使用
type TenantUsecase struct {
    tenantRepo tenant.TenantRepository
}
```

## エンティティ設計

### プライベートフィールドとゲッター

```go
type Tenant struct {
    tenantID   common.TenantID  // プライベート
    tenantName string           // プライベート
}

func (t *Tenant) TenantID() common.TenantID {
    return t.tenantID
}
```

### ファクトリ関数

```go
// NewXxx: 新規エンティティ作成用
func NewTenant(now time.Time, name, timezone string) (*Tenant, error) {
    tenant := &Tenant{
        tenantID:  common.NewTenantID(),
        createdAt: now,
    }
    if err := tenant.validate(); err != nil {
        return nil, err
    }
    return tenant, nil
}

// ReconstructXxx: 永続化からの復元用
func ReconstructTenant(id TenantID, name string, ...) (*Tenant, error) {
    tenant := &Tenant{tenantID: id, tenantName: name, ...}
    if err := tenant.validate(); err != nil {
        return nil, err
    }
    return tenant, nil
}
```

### ドメインバリデーション

バリデーションはドメインエンティティ内で行う:

```go
func (t *Tenant) validate() error {
    if t.tenantName == "" {
        return common.NewValidationError("tenant_name is required", nil)
    }
    if len(t.tenantName) > 255 {
        return common.NewValidationError("tenant_name must be less than 255 characters", nil)
    }
    return nil
}
```

## リポジトリパターン

### インターフェースはDomainに定義

```go
// internal/domain/tenant/repository.go
type TenantRepository interface {
    FindByID(ctx context.Context, id common.TenantID) (*Tenant, error)
    FindAll(ctx context.Context) ([]*Tenant, error)
    Save(ctx context.Context, tenant *Tenant) error
    Delete(ctx context.Context, id common.TenantID) error
}
```

### 実装はInfraに配置

```go
// internal/infra/db/tenant_repository.go
type tenantRepositoryImpl struct {
    pool *pgxpool.Pool
}

func (r *tenantRepositoryImpl) FindByID(ctx context.Context, id common.TenantID) (*Tenant, error) {
    // SQL実装
}
```

## トランザクション管理

複数テーブル操作時はTxManagerを使用:

```go
return uc.txManager.WithTx(ctx, func(txCtx context.Context) error {
    if err := uc.repo1.Save(txCtx, entity1); err != nil {
        return err
    }
    if err := uc.repo2.Save(txCtx, entity2); err != nil {
        return err
    }
    return nil
})
```

## テナント分離（必須）

全てのクエリにtenant_idフィルタリングを含める:

```sql
-- OK
SELECT * FROM members WHERE tenant_id = $1 AND member_id = $2

-- NG: テナント分離がない！
SELECT * FROM members WHERE member_id = $1
```

## ソフトデリート

`deleted_at` カラムを使用:

```go
func (t *Tenant) Delete(now time.Time) {
    t.deletedAt = &now
    t.updatedAt = now
}

func (t *Tenant) IsDeleted() bool {
    return t.deletedAt != nil
}
```

---
description: PostgreSQL/pgx データベースパターンとベストプラクティス
---

# Database Patterns

VRC Shift Scheduler の PostgreSQL/pgx データベース操作パターン。

---

## 技術スタック

- **PostgreSQL 16**
- **pgx** (Go PostgreSQL driver)
- **pgxpool** (コネクションプール)

---

## マイグレーション

### ファイル命名規則

```
NNN_description.up.sql    # 適用用
NNN_description.down.sql  # ロールバック用（任意）
```

例:
- `001_create_tenants.up.sql`
- `039_migrate_instance_data.up.sql`

### マイグレーション実行

```bash
# 状態確認
docker exec vrc-shift-backend /app/migrate -action=status

# 適用
docker exec vrc-shift-backend /app/migrate -action=up

# ロールバック（1つ戻す）
docker exec vrc-shift-backend /app/migrate -action=down -steps=1
```

### マイグレーションSQL例

```sql
-- 039_migrate_instance_data.up.sql
BEGIN;

-- テーブル作成
CREATE TABLE IF NOT EXISTS shift_instances (
    instance_id VARCHAR(26) PRIMARY KEY,
    tenant_id VARCHAR(26) NOT NULL REFERENCES tenants(tenant_id),
    template_id VARCHAR(26) NOT NULL REFERENCES shift_templates(template_id),
    target_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- インデックス
CREATE INDEX idx_shift_instances_tenant_id ON shift_instances(tenant_id);
CREATE INDEX idx_shift_instances_template_id ON shift_instances(template_id);
CREATE INDEX idx_shift_instances_target_date ON shift_instances(target_date);

COMMIT;
```

---

## テナント分離（必須）

### 全クエリにtenant_idを含める

```go
// OK: テナント分離あり
query := `
    SELECT member_id, display_name, email
    FROM members
    WHERE tenant_id = $1 AND deleted_at IS NULL
    ORDER BY display_name
`
rows, err := r.pool.Query(ctx, query, tenantID)

// NG: テナント分離なし（禁止）
query := `SELECT * FROM members WHERE member_id = $1`
```

### 複合インデックス設計

```sql
-- tenant_id を先頭に配置
CREATE INDEX idx_members_tenant_active
ON members(tenant_id, deleted_at)
WHERE deleted_at IS NULL;

CREATE INDEX idx_shift_slots_tenant_business_day
ON shift_slots(tenant_id, business_day_id);
```

---

## リポジトリパターン

### インターフェース（Domain層）

```go
// internal/domain/member/repository.go
type MemberRepository interface {
    FindByID(ctx context.Context, tenantID common.TenantID, memberID MemberID) (*Member, error)
    FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*Member, error)
    Save(ctx context.Context, member *Member) error
    Delete(ctx context.Context, tenantID common.TenantID, memberID MemberID) error
}
```

### 実装（Infra層）

```go
// internal/infra/db/member_repository.go
type memberRepositoryImpl struct {
    pool *pgxpool.Pool
}

func NewMemberRepository(pool *pgxpool.Pool) member.MemberRepository {
    return &memberRepositoryImpl{pool: pool}
}

func (r *memberRepositoryImpl) FindByID(
    ctx context.Context,
    tenantID common.TenantID,
    memberID member.MemberID,
) (*member.Member, error) {
    query := `
        SELECT member_id, tenant_id, display_name, email, is_active,
               created_at, updated_at, deleted_at
        FROM members
        WHERE tenant_id = $1 AND member_id = $2 AND deleted_at IS NULL
    `

    row := r.pool.QueryRow(ctx, query, tenantID, memberID)

    var m memberRow
    err := row.Scan(
        &m.MemberID, &m.TenantID, &m.DisplayName, &m.Email, &m.IsActive,
        &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
    )
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, common.NewDomainError(common.ErrNotFound, "member not found")
        }
        return nil, fmt.Errorf("failed to scan member: %w", err)
    }

    return m.toEntity()
}
```

---

## トランザクション管理

### TxManager

```go
// 複数テーブル操作時はトランザクション必須
err := uc.txManager.WithTx(ctx, func(txCtx context.Context) error {
    // 1. メンバー作成
    if err := uc.memberRepo.Save(txCtx, member); err != nil {
        return err
    }

    // 2. ロール割り当て
    if err := uc.roleRepo.AssignRole(txCtx, member.MemberID(), roleID); err != nil {
        return err
    }

    // 3. 監査ログ記録
    if err := uc.auditRepo.Log(txCtx, auditLog); err != nil {
        return err
    }

    return nil
})
```

### 実装

```go
// internal/infra/db/tx_manager.go
type TxManager struct {
    pool *pgxpool.Pool
}

func (tm *TxManager) WithTx(ctx context.Context, fn func(context.Context) error) error {
    tx, err := tm.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }

    // コンテキストにトランザクションを格納
    txCtx := context.WithValue(ctx, txKey, tx)

    if err := fn(txCtx); err != nil {
        if rbErr := tx.Rollback(ctx); rbErr != nil {
            return fmt.Errorf("rollback failed: %v (original: %w)", rbErr, err)
        }
        return err
    }

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("failed to commit: %w", err)
    }

    return nil
}
```

---

## ソフトデリート

### テーブル設計

```sql
CREATE TABLE members (
    member_id VARCHAR(26) PRIMARY KEY,
    tenant_id VARCHAR(26) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    -- ... その他カラム
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE  -- NULLなら有効、値ありなら削除済み
);
```

### クエリパターン

```go
// 有効なレコードのみ取得
query := `SELECT * FROM members WHERE tenant_id = $1 AND deleted_at IS NULL`

// 削除済みを含めて取得
query := `SELECT * FROM members WHERE tenant_id = $1`

// 論理削除
query := `UPDATE members SET deleted_at = $1, updated_at = $1 WHERE member_id = $2`
```

---

## N+1クエリ対策

### 問題のあるコード

```go
// NG: N+1クエリ
events, _ := eventRepo.FindByTenantID(ctx, tenantID)
for _, event := range events {
    // 各イベントに対してクエリ発行 → N回のクエリ
    businessDays, _ := businessDayRepo.FindByEventID(ctx, event.EventID())
}
```

### 解決策: JOINまたはIN句

```go
// OK: JOIN
query := `
    SELECT e.event_id, e.event_name, bd.business_day_id, bd.target_date
    FROM events e
    LEFT JOIN event_business_days bd ON e.event_id = bd.event_id
    WHERE e.tenant_id = $1 AND e.deleted_at IS NULL
    ORDER BY e.event_id, bd.target_date
`

// OK: IN句で一括取得
eventIDs := extractEventIDs(events)
query := `
    SELECT business_day_id, event_id, target_date
    FROM event_business_days
    WHERE event_id = ANY($1) AND deleted_at IS NULL
`
rows, _ := pool.Query(ctx, query, eventIDs)
```

---

## インデックス戦略

### 基本ルール

1. **外部キーにインデックス** - JOIN性能向上
2. **検索条件カラムにインデックス** - WHERE句で使用
3. **tenant_idを先頭に** - マルチテナント分離

### 部分インデックス

```sql
-- 有効なレコードのみ対象
CREATE INDEX idx_members_tenant_active
ON members(tenant_id, display_name)
WHERE deleted_at IS NULL;

-- 特定ステータスのみ対象
CREATE INDEX idx_shift_assignments_confirmed
ON shift_assignments(slot_id)
WHERE assignment_status = 'confirmed';
```

### 複合インデックス

```sql
-- 頻繁な検索パターンに合わせる
CREATE INDEX idx_shift_slots_business_day
ON shift_slots(tenant_id, business_day_id, deleted_at);
```

---

## 接続プール設定

```go
// internal/config/database.go
func NewDBPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(databaseURL)
    if err != nil {
        return nil, err
    }

    // 接続プール設定
    config.MaxConns = 25
    config.MinConns = 5
    config.MaxConnLifetime = time.Hour
    config.MaxConnIdleTime = 30 * time.Minute
    config.HealthCheckPeriod = time.Minute

    pool, err := pgxpool.NewWithConfig(ctx, config)
    if err != nil {
        return nil, err
    }

    // 接続テスト
    if err := pool.Ping(ctx); err != nil {
        return nil, err
    }

    return pool, nil
}
```

---

## バッチ処理

### バッチINSERT

```go
// pgx.CopyFrom を使用した高速バッチINSERT
rows := [][]interface{}{
    {id1, tenantID, name1, createdAt},
    {id2, tenantID, name2, createdAt},
    // ...
}

copyCount, err := pool.CopyFrom(
    ctx,
    pgx.Identifier{"members"},
    []string{"member_id", "tenant_id", "display_name", "created_at"},
    pgx.CopyFromRows(rows),
)
```

### バッチUPDATE

```go
// 一括更新
query := `
    UPDATE shift_assignments
    SET assignment_status = 'cancelled', updated_at = $1
    WHERE assignment_id = ANY($2)
`
_, err := pool.Exec(ctx, query, now, assignmentIDs)
```

---

## 監視・デバッグ

### クエリログ

```go
// 開発環境でのクエリログ
config.ConnConfig.Tracer = &QueryTracer{}

type QueryTracer struct{}

func (t *QueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
    log.Printf("SQL: %s, Args: %v", data.SQL, data.Args)
    return ctx
}
```

### スロークエリ検出

```sql
-- PostgreSQL設定
ALTER SYSTEM SET log_min_duration_statement = 1000;  -- 1秒以上
SELECT pg_reload_conf();
```

---

## チェックリスト

作業前に確認:
- [ ] 全クエリにtenant_idフィルタリングがある
- [ ] 外部キーにインデックスがある
- [ ] 複数テーブル操作時はトランザクションを使用
- [ ] N+1クエリがない
- [ ] ソフトデリート（deleted_at）を考慮している
- [ ] SQLインジェクション対策（パラメータ化クエリ）

---

## 関連ファイル

- `backend/internal/infra/db/` - リポジトリ実装
- `backend/internal/infra/db/migrations/` - マイグレーションファイル
- `backend/cmd/migrate/` - マイグレーションツール

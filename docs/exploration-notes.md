# コードベース探索結果

探索日時: 2025-12-19

## 1. プロジェクト構造

### ディレクトリ構成
```
backend/
├── cmd/
│   ├── server/        # メインエントリーポイント
│   ├── migrate/       # マイグレーションCLI
│   └── seed/          # シードデータ投入
├── internal/
│   ├── domain/        # ドメイン層（エンティティ、リポジトリインターフェース）
│   ├── application/   # アプリケーション層（ユースケース）
│   ├── app/           # 追加ユースケース
│   ├── infra/         # インフラ層
│   │   ├── db/        # リポジトリ実装、マイグレーション
│   │   └── security/  # JWT, bcrypt
│   └── interface/
│       └── rest/      # HTTPハンドラー、ルーター
```

## 2. データベース層

### マイグレーション方式
- **場所**: `internal/infra/db/migrations/`
- **命名規則**: `{NNN}_description.{up|down}.sql` (例: `001_create_tenants.up.sql`)
- **現在**: 17個のマイグレーションファイル
- **管理**: `cmd/migrate/main.go` による独自実装、`schema_migrations` テーブルで追跡

### リポジトリパターン
- インターフェース: `internal/domain/{feature}/repository.go`
- 実装: `internal/infra/db/{entity}_repository.go`
- `*pgxpool.Pool` を注入
- ソフトデリート: `deleted_at IS NULL` でフィルタリング

### トランザクション管理
- `internal/infra/db/tx.go` に `TxManager` 実装
- `WithTx(ctx, fn)` でトランザクション制御
- コンテキストに tx を埋め込み、`GetTx(ctx, pool)` で取得

## 3. 既存テナントテーブル構造

```sql
CREATE TABLE tenants (
    tenant_id CHAR(26) PRIMARY KEY,        -- ULID
    tenant_name VARCHAR(255) NOT NULL,
    timezone VARCHAR(50) DEFAULT 'Asia/Tokyo',
    is_active BOOLEAN DEFAULT true,        -- ★ status カラムは未存在
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);
```

**重要**: `status` カラムは存在しない。`is_active` のみ。

## 4. 認証・認可

### JWT実装
- `internal/infra/security/jwt.go`
- HS256、24時間有効
- Claims: `admin_id`, `tenant_id`, `role`

### ミドルウェア
- `internal/interface/rest/middleware.go`
- `Auth(tokenVerifier)`: JWT認証 + X-Tenant-ID フォールバック
- コンテキストキー: `tenant_id`, `admin_id`, `role`

### 管理者テーブル
```sql
CREATE TABLE admins (
    admin_id CHAR(26) PRIMARY KEY,
    tenant_id CHAR(26) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    display_name VARCHAR(255),
    role VARCHAR(20) DEFAULT 'manager',  -- 'owner' | 'manager'
    ...
);
```

## 5. APIパターン

### レスポンス形式
```json
// 成功
{"data": {...}}

// エラー
{"error": {"code": "ERR_XXX", "message": "..."}}
```

### ルーティング構造
- ルーター: chi
- 認証不要: `/health`, `/api/v1/auth/login`, `/api/v1/public/*`
- 認証必要: `/api/v1/*`

### ハンドラーパターン
1. `GetTenantID(ctx)` でテナントID取得
2. リクエストボディをパース
3. ユースケース実行
4. `RespondSuccess()` / `RespondError()` でレスポンス

## 6. ID生成

- **ULID**: 26文字、ソート可能
- `common.NewULID()` で生成
- 各ドメインに専用型: `TenantID`, `AdminID`, `EventID` など

## 7. 既存機能で不足しているもの

| 機能 | 状態 |
|------|------|
| tenants.status カラム | 未存在（is_active のみ） |
| レート制限 | 未実装 |
| メール送信 | 未実装 |
| 管理者専用API (/admin) | 未実装 |
| 監査ログテーブル | 存在するが詳細要確認 |

## 8. 実装方針

マネタイズ基盤実装にあたり、以下の方針で進める：

1. **マイグレーション**: `018_xxx.up.sql` から開始
2. **エンティティ**: `internal/domain/billing/` 等に新規作成
3. **リポジトリ**: 既存パターンに従い `internal/infra/db/` に実装
4. **ハンドラー**: `internal/interface/rest/` に追加
5. **ルーティング**: `router.go` に追加

## 9. 要確認事項

探索で判明した疑問点（ユーザーに確認が必要）：

1. `is_active` と `status` の関係
2. Claim で作成するのは `admin` か別の `user` か
3. レート制限の実装方法
4. 管理者API の認証方式
5. メール送信の即時実装 vs プレースホルダー

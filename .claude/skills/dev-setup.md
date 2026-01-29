---
description: 開発環境構築、環境変数、テストアカウント
---

# Development Setup

VRC Shift Scheduler の開発環境セットアップ手順。

---

## 必要な環境

| 項目 | バージョン |
|------|-----------|
| Go | 1.24+ |
| Node.js | 20+ |
| PostgreSQL | 16 |
| Docker & Docker Compose | 最新 |

---

## クイックスタート（Docker）

```bash
# プロジェクトルートで実行
docker compose up -d

# バックエンド起動確認
curl http://localhost:8080/health
# -> {"status":"ok"}

# シードデータの投入
cd backend
DATABASE_URL="postgres://vrcshift:vrcshift@localhost:5432/vrcshift?sslmode=disable" \
  go run cmd/seed/main.go

# フロントエンド起動
cd web-frontend
npm install
npm run dev
# -> http://localhost:5173
```

---

## テストアカウント

| メールアドレス | パスワード | 権限 |
|---------------|-----------|------|
| admin1@example.com | password123 | owner |

**注意**: 開発/テスト環境専用。本番環境では使用不可。

---

## 環境変数

### Backend（必須）

| 変数名 | 説明 | 例 |
|--------|------|-----|
| `DATABASE_URL` | PostgreSQL接続文字列 | `postgresql://vrcshift:vrcshift@localhost:5432/vrcshift?sslmode=disable` |
| `JWT_SECRET` | JWT署名用シークレット | `your_development_secret_key` |

### Backend（オプション）

| 変数名 | 説明 | デフォルト |
|--------|------|-----------|
| `PORT` | サーバーポート | `8080` |
| `ENVIRONMENT` | 実行環境 | `development` |
| `LOG_LEVEL` | ログレベル | `info` |
| `ALLOWED_ORIGINS` | CORS許可オリジン | `*` |

### Frontend

| 変数名 | 説明 | 例 |
|--------|------|-----|
| `VITE_API_BASE_URL` | バックエンドAPIのベースURL | `http://localhost:8080` |

### PostgreSQL（Docker用）

| 変数名 | 説明 | デフォルト |
|--------|------|-----------|
| `POSTGRES_DB` | データベース名 | `vrcshift` |
| `POSTGRES_USER` | ユーザー名 | `vrcshift` |
| `POSTGRES_PASSWORD` | パスワード | `vrcshift` |

---

## 開発環境設定例

### backend/.env

```env
DATABASE_URL=postgresql://vrcshift:vrcshift@localhost:5432/vrcshift?sslmode=disable
JWT_SECRET=your_development_secret_key
PORT=8080
ENVIRONMENT=development
LOG_LEVEL=debug
ALLOWED_ORIGINS=http://localhost:5173
```

### web-frontend/.env

```env
VITE_API_BASE_URL=http://localhost:8080
```

---

## 開発コマンド

### バックエンド

```bash
cd backend

# ローカル起動
DATABASE_URL="postgres://vrcshift:vrcshift@localhost:5432/vrcshift?sslmode=disable" \
JWT_SECRET=test_secret_key \
go run cmd/server/main.go

# テスト実行
JWT_SECRET=test_secret_key go test ./...

# ビルド
go build -o bin/server cmd/server/main.go

# フォーマット
go fmt ./...

# Lint
golangci-lint run
```

### フロントエンド

```bash
cd web-frontend

# 開発サーバー起動
npm run dev

# ビルド
npm run build

# プレビュー（ビルド後）
npm run preview

# Lint
npm run lint
```

### マイグレーション

```bash
# Docker環境
docker exec vrc-shift-backend /app/migrate -action=status
docker exec vrc-shift-backend /app/migrate -action=up

# ローカル環境
cd backend
go run cmd/migrate/main.go
```

---

## データベース

### 接続情報（開発環境）

```
Host: localhost
Port: 5432
Database: vrcshift
User: vrcshift
Password: vrcshift
```

### リセット

```bash
# すべてのコンテナとボリュームを削除
docker compose down -v

# 再起動
docker compose up -d

# シードデータ投入
cd backend
DATABASE_URL="postgres://vrcshift:vrcshift@localhost:5432/vrcshift?sslmode=disable" \
  go run cmd/seed/main.go
```

### 直接クエリ

```bash
# 管理者一覧
docker exec vrc-shift-scheduler-db-1 psql -U vrcshift -d vrcshift \
  -c "SELECT admin_id, email, display_name, role FROM admins WHERE deleted_at IS NULL;"
```

---

## トラブルシューティング

### バックエンドが起動しない

```bash
# コンテナのログを確認
docker logs vrc-shift-scheduler-backend-1

# データベース接続確認
docker exec vrc-shift-scheduler-db-1 psql -U vrcshift -c '\l'
```

### JWTトークンの有効期限切れ

```bash
# 新しいトークンを取得
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin1@example.com","password":"password123"}'
```

トークンのデフォルト有効期限: **24時間**

### CORSエラー

Backend の `ALLOWED_ORIGINS` にフロントエンドのURLを追加：

```env
ALLOWED_ORIGINS=http://localhost:5173
```

### npm install エラー

```bash
# キャッシュをクリア
npm cache clean --force
npm install
```

---

## URL一覧

### 認証関連
- 管理者ログイン: http://localhost:5173/admin/login
- 管理者招待: http://localhost:5173/admin/invite
- 招待受理: http://localhost:5173/invite/{token}

### 管理画面
- イベント一覧: http://localhost:5173/events
- メンバー一覧: http://localhost:5173/members

### 公開ページ（認証不要）
- 出欠確認: http://localhost:5173/p/attendance/{token}
- 日程調整: http://localhost:5173/p/schedule/{token}

---

## 関連ドキュメント

- `docs/development/DEVELOPMENT.md` - 開発ガイド
- `docs/setup/ENVIRONMENT_VARIABLES.md` - 環境変数詳細
- `docs/setup/SETUP.md` - セットアップ手順

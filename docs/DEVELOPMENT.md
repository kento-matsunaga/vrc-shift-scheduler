# 開発環境セットアップガイド

## 目次
- [環境構築](#環境構築)
- [開発用アカウント情報](#開発用アカウント情報)
- [API エンドポイント](#api-エンドポイント)
- [トラブルシューティング](#トラブルシューティング)

---

## 環境構築

### 1. 必要な環境
- Docker & Docker Compose
- Node.js 20+ (フロントエンド開発時)
- Go 1.24+ (バックエンド開発時)
- PostgreSQL 16 (ローカル開発時)

### 2. Docker で起動

```bash
# プロジェクトルートで実行
docker compose up -d

# バックエンド起動確認
curl http://localhost:8080/health
# -> {"status":"ok"}
```

### 3. シードデータの投入

```bash
cd backend
DATABASE_URL="postgres://vrcshift:vrcshift@localhost:5432/vrcshift?sslmode=disable" \
  go run cmd/seed/main.go
```

### 4. フロントエンド起動

```bash
cd web-frontend
npm install
npm run dev
# -> http://localhost:5173
```

---

## 開発用アカウント情報

### 管理者ログイン (Admin Login)

フロントエンド: [http://localhost:5173/admin/login](http://localhost:5173/admin/login)

#### テストアカウント

| メールアドレス | パスワード | 権限 | 説明 |
|---|---|---|---|
| `admin1@example.com` | `password123` | owner | シードデータで作成されるアカウント |

**注意**: このアカウントは開発/テスト環境専用です。本番環境では使用しないでください。シードデータを投入して使用してください。

---

## API エンドポイント

### ベースURL
- ローカル開発: `http://localhost:8080`
- Docker: `http://localhost:8080`

### 認証 API

#### ログイン
```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "admin1@example.com",
  "password": "password123"
}

# レスポンス例
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "admin_id": "01KBHMYWYKRV8PK8EVYGF1SHV0",
    "tenant_id": "01KBHMYWYKRV8PK8EVYGF1SHV0",
    "role": "owner",
    "expires_at": "2025-12-17T19:58:51+09:00"
  }
}
```

#### トークンを使用したAPI呼び出し

```bash
# JWTトークンを取得
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin1@example.com","password":"password123"}' \
  | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

# 認証が必要なAPIを呼び出し
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/events
```

### 主要エンドポイント

#### イベント管理
- `GET /api/v1/events` - イベント一覧取得
- `POST /api/v1/events` - イベント作成
- `GET /api/v1/events/:event_id` - イベント詳細取得
- `PUT /api/v1/events/:event_id` - イベント更新
- `DELETE /api/v1/events/:event_id` - イベント削除

#### メンバー管理
- `GET /api/v1/members` - メンバー一覧取得
- `POST /api/v1/members` - メンバー作成
- `GET /api/v1/members/:member_id` - メンバー詳細取得

#### 出欠確認 (Attendance)
- `POST /api/v1/attendance` - 出欠確認作成
- `GET /api/v1/attendance/:attendance_id` - 出欠確認詳細取得
- `GET /api/v1/public/attendance/:token` - 公開URL経由でアクセス（認証不要）
- `POST /api/v1/public/attendance/:token/respond` - 出欠回答送信（認証不要）

#### 日程調整 (Schedule)
- `POST /api/v1/schedules` - 日程調整作成
- `GET /api/v1/schedules/:schedule_id` - 日程調整詳細取得
- `GET /api/v1/public/schedules/:token` - 公開URL経由でアクセス（認証不要）
- `POST /api/v1/public/schedules/:token/respond` - 日程回答送信（認証不要）

#### 管理者招待
- `POST /api/v1/invitations` - 管理者招待作成
- `GET /api/v1/invitations/:token` - 招待情報取得
- `POST /api/v1/invitations/:token/accept` - 招待受理

---

## トラブルシューティング

### バックエンドが起動しない

```bash
# コンテナのログを確認
docker logs vrc-shift-scheduler-backend-1

# データベース接続確認
docker exec vrc-shift-scheduler-db-1 psql -U vrcshift -c '\l'
```

### データベースをリセットしたい

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

### JWT トークンの有効期限が切れた

```bash
# 新しいトークンを取得
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin1@example.com","password":"password123"}'
```

トークンのデフォルト有効期限は **24時間** です。

### フロントエンドでCORSエラーが出る

`.env.local` ファイルでAPIのベースURLが正しく設定されているか確認してください：

```bash
# web-frontend/.env.local
VITE_API_BASE_URL=http://localhost:8080
```

---

## その他の情報

### データベース接続情報

```
Host: localhost
Port: 5432
Database: vrcshift
User: vrcshift
Password: vrcshift
```

### pgAdmin で接続

```bash
# pgAdminコンテナを起動する場合
docker run -d \
  --name pgadmin \
  --network vrc-shift-scheduler_default \
  -p 5050:80 \
  -e PGADMIN_DEFAULT_EMAIL=admin@example.com \
  -e PGADMIN_DEFAULT_PASSWORD=admin \
  dpage/pgadmin4

# http://localhost:5050 にアクセス
# Host: vrc-shift-scheduler-db-1
```

### 環境変数

バックエンドで使用される主な環境変数：

- `DATABASE_URL` - PostgreSQL接続文字列
- `JWT_SECRET` - JWT署名用のシークレットキー (開発環境: `test_secret_key`)
- `PORT` - サーバーポート (デフォルト: `8080`)

---

## 開発ワークフロー

### バックエンド開発

```bash
cd backend

# ローカルで起動
DATABASE_URL="postgres://vrcshift:vrcshift@localhost:5432/vrcshift?sslmode=disable" \
JWT_SECRET=test_secret_key \
go run cmd/server/main.go

# テスト実行
JWT_SECRET=test_secret_key go test ./...

# ビルド
go build -o bin/server cmd/server/main.go
```

### フロントエンド開発

```bash
cd web-frontend

# 開発サーバー起動
npm run dev

# ビルド
npm run build

# プレビュー
npm run preview
```

---

**最終更新**: 2025-12-19

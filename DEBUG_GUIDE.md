# VRC Shift Scheduler - デバッグガイド

## 🚀 サーバー起動状況

### ✅ バックエンドサーバー
- **URL**: http://localhost:8080
- **Health Check**: http://localhost:8080/health
- **Status**: 🟢 起動中

### ✅ フロントエンドサーバー
- **URL**: http://localhost:5173
- **Status**: 🟢 起動中

### ✅ データベース
- **Host**: localhost:5432
- **Database**: vrcshift
- **User**: vrcshift
- **Password**: vrcshift
- **Status**: 🟢 起動中

---

## 🔐 テストアカウント

### 管理者アカウント #1
```
Email: admin@test.com
Password: password123
Tenant ID: 01KBHMYWYKRV8PK8EVYGF1SHV0
Role: owner
```

### 管理者アカウント #2
```
Email: admin1@example.com
Password: password123
Tenant ID: 01KCGJ95CK7YB8WFPQ78NJ5C4S
Role: owner
```

---

## 🔗 主要なエンドポイント・URL

### 認証関連
- **管理者ログイン**: http://localhost:5173/admin/login
- **管理者招待（要認証）**: http://localhost:5173/admin/invite
- **招待受理（認証不要）**: http://localhost:5173/invite/{token}

### 管理画面（要認証）
- **イベント一覧**: http://localhost:5173/events
- **メンバー一覧**: http://localhost:5173/members
- **自分のシフト**: http://localhost:5173/my-shifts

### 公開ページ（認証不要）
- **出欠確認**: http://localhost:5173/p/attendance/{token}
- **日程調整**: http://localhost:5173/p/schedule/{token}

---

## 📡 バックエンド API エンドポイント

### 認証API
```bash
# ログイン
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json
{
  "email": "admin@test.com",
  "password": "password123"
}

# 管理者招待（要JWT認証）
POST http://localhost:8080/api/v1/invitations
Authorization: Bearer {JWT_TOKEN}
Content-Type: application/json
{
  "email": "newadmin@example.com",
  "role": "admin"
}

# 招待受理（認証不要）
POST http://localhost:8080/api/v1/invitations/accept/{token}
Content-Type: application/json
{
  "display_name": "新管理者",
  "password": "password123"
}
```

---

## 🧪 テスト手順

### 1. ログインテスト
1. ブラウザで http://localhost:5173/admin/login にアクセス
2. 以下の情報でログイン:
   - Email: `admin@test.com`
   - Password: `password123`
3. ログイン成功後、イベント一覧画面に遷移することを確認

### 2. 管理者招待機能テスト
1. ログイン後、ナビゲーションバーの「管理者招待」をクリック
2. 以下の情報を入力:
   - Email: `newadmin@example.com`
   - Role: `admin` または `manager`
3. 「招待を送信」ボタンをクリック
4. 招待URLが表示されることを確認
5. 「URLをコピー」ボタンでクリップボードにコピーされることを確認

### 3. 招待受理機能テスト
1. 生成された招待URLをコピー（例: http://localhost:5173/invite/abc123...）
2. 新しいブラウザ（シークレットモード推奨）で招待URLにアクセス
3. 以下の情報を入力:
   - 表示名: `新管理者`
   - パスワード: `password123`
   - パスワード（確認）: `password123`
4. 「登録」ボタンをクリック
5. 登録完了後、ログイン画面にリダイレクトされることを確認
6. 登録したメールアドレスとパスワードでログインできることを確認

---

## 🛠️ デバッグコマンド

### サーバーログの確認
```bash
# バックエンドログ
tail -f /tmp/backend.log

# フロントエンドログ
tail -f /tmp/frontend.log
```

### データベース直接確認
```bash
# 管理者一覧を確認
docker exec vrc-shift-scheduler-db-1 psql -U vrcshift -d vrcshift \
  -c "SELECT admin_id, email, display_name, role, tenant_id FROM admins WHERE deleted_at IS NULL;"

# 招待一覧を確認
docker exec vrc-shift-scheduler-db-1 psql -U vrcshift -d vrcshift \
  -c "SELECT invitation_id, email, role, token, expires_at, accepted_at FROM invitations ORDER BY created_at DESC LIMIT 10;"

# テナント一覧を確認
docker exec vrc-shift-scheduler-db-1 psql -U vrcshift -d vrcshift \
  -c "SELECT tenant_id, tenant_name, timezone FROM tenants;"
```

### API動作確認（curl）
```bash
# ログインAPIテスト
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"password123"}'

# 招待APIテスト（JWTトークンが必要）
TOKEN="YOUR_JWT_TOKEN_HERE"
curl -X POST http://localhost:8080/api/v1/invitations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"email":"test@example.com","role":"admin"}'
```

---

## ⚠️ トラブルシューティング

### サーバーが起動しない場合
```bash
# プロセスを確認
ps aux | grep "go run"

# ポートを確認
lsof -i :8080
lsof -i :5173

# サーバーを再起動
pkill -f "go run"
cd /home/erenoa6621/dev/vrc-shift-scheduler/backend
JWT_SECRET=test_secret_key DATABASE_URL="postgres://vrcshift:vrcshift@localhost:5432/vrcshift?sslmode=disable" go run cmd/server/main.go
```

### データベース接続エラー
```bash
# PostgreSQLコンテナの状態を確認
docker ps | grep postgres

# コンテナを再起動
docker restart vrc-shift-scheduler-db-1

# 接続テスト
docker exec vrc-shift-scheduler-db-1 psql -U vrcshift -d vrcshift -c "SELECT 1;"
```

### マイグレーションエラー
```bash
# マイグレーション状態を確認
docker exec vrc-shift-scheduler-db-1 psql -U vrcshift -d vrcshift \
  -c "SELECT migration_id, applied_at FROM schema_migrations ORDER BY migration_id;"

# マイグレーションを再実行
cd /home/erenoa6621/dev/vrc-shift-scheduler/backend
DATABASE_URL="postgres://vrcshift:vrcshift@localhost:5432/vrcshift?sslmode=disable" go run cmd/migrate/main.go
```

---

## 📝 実装済み機能一覧

### ✅ 認証機能
- [x] 管理者ログイン（email + password のみ、tenant_id 不要）
- [x] JWT認証（Bearer Token）
- [x] ログアウト

### ✅ 管理者招待機能
- [x] 招待作成（POST /api/v1/invitations）
- [x] 招待URL生成
- [x] 招待URLコピー機能
- [x] 招待受理（POST /api/v1/invitations/accept/{token}）
- [x] 招待トークン有効期限チェック（7日間）
- [x] Email重複チェック

### ✅ フロントエンド
- [x] AdminLoginページ（Tailwind + Glass morphism デザイン）
- [x] AdminInvitationページ（管理者招待画面）
- [x] AcceptInvitationページ（招待受理画面）
- [x] ナビゲーションバーに「管理者招待」リンク追加
- [x] ルーティング設定

### ✅ データベース
- [x] Migration 010: admins.email グローバル一意制約
- [x] Migration 011: invitations テーブル作成

---

## 🎨 デザイン仕様

### カラーパレット
- **背景グラデーション**: `from-slate-900 via-purple-900 to-slate-900`
- **ガラスモーフィズム**: `bg-white/10 backdrop-blur-lg`
- **ボーダー**: `border-white/20`
- **アクセントカラー**: Purple 600/700
- **成功カラー**: Green 500/600
- **エラーカラー**: Red 500/600

### コンポーネント
- 統一されたフォームスタイル
- ホバー効果とトランジションアニメーション
- レスポンシブデザイン

---

## 📊 現在のデータベース構造

### テナント
- 2つのテナントが存在
- 各テナントに1人の管理者（owner）が存在

### 管理者
- admin@test.com (tenant: 01KBHMYWYKRV8PK8EVYGF1SHV0)
- admin1@example.com (tenant: 01KCGJ95CK7YB8WFPQ78NJ5C4S)

### 招待
- 招待データは動的に生成されます

---

**Last Updated**: 2025-12-15
**Version**: Alpha

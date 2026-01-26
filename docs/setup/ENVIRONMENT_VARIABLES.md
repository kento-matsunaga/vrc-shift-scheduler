# 環境変数一覧

VRC Shift Scheduler で使用する環境変数の詳細説明です。

## 📋 目次

- [Backend 環境変数](#backend-環境変数)
- [Frontend 環境変数](#frontend-環境変数)
- [PostgreSQL 環境変数](#postgresql-環境変数)
- [環境別の設定例](#環境別の設定例)

---

## Backend 環境変数

### 必須

| 変数名 | 説明 | デフォルト値 | 例 |
|--------|------|--------------|-----|
| `DATABASE_URL` | PostgreSQL 接続文字列 | なし | `postgresql://vrcshift:vrcshift@localhost:5432/vrcshift?sslmode=disable` |
| `JWT_SECRET` | JWT署名用シークレットキー | なし | `your_secure_secret_key_here` |

### オプション

| 変数名 | 説明 | デフォルト値 | 例 |
|--------|------|--------------|-----|
| `PORT` | サーバーのポート番号 | `8080` | `8080` |
| `HOST` | バインドするホスト | `0.0.0.0` | `0.0.0.0`, `localhost` |
| `ENVIRONMENT` | 実行環境 | `development` | `development`, `staging`, `production` |
| `LOG_LEVEL` | ログレベル | `info` | `debug`, `info`, `warn`, `error` |
| `ALLOWED_ORIGINS` | CORS許可オリジン（カンマ区切り） | `*` | `https://example.com,http://localhost:3000` |
| `ENABLE_AUDIT_LOG` | 監査ログの有効化 | `false` | `true`, `false` |
| `ENABLE_NOTIFICATION` | 通知機能の有効化 | `false` | `true`, `false` |
| `DISCORD_WEBHOOK_URL` | Discord Webhook URL（将来実装用） | なし | `https://discord.com/api/webhooks/...` |

### Billing/課金関連

| 変数名 | 説明 | デフォルト値 | 例 |
|--------|------|--------------|-----|
| `STRIPE_SECRET_KEY` | Stripe シークレットキー | なし | `sk_live_...` |
| `STRIPE_WEBHOOK_SECRET` | Stripe Webhook署名シークレット | なし | `whsec_...` |
| `STRIPE_PRICE_ID` | Stripe 価格ID（月額プラン） | なし | `price_...` |
| `GRACE_PERIOD_DAYS` | 支払い失敗後の猶予期間（日数） | `14` | `14` |
| `LICENSE_KEY_PREFIX` | ライセンスキーのプレフィックス | `VRCSS-` | `VRCSS-` |

### 設定例

#### 開発環境（`backend/.env`）

```env
DATABASE_URL=postgresql://vrcshift:vrcshift@localhost:5432/vrcshift?sslmode=disable
JWT_SECRET=your_development_secret_key
PORT=8080
ENVIRONMENT=development
LOG_LEVEL=debug
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
```

#### 本番環境（Render等）

```env
DATABASE_URL=<Renderが自動生成>
JWT_SECRET=<安全なランダム文字列を設定>
PORT=8080
ENVIRONMENT=production
LOG_LEVEL=info
ALLOWED_ORIGINS=https://your-frontend-domain.com
```

---

## Frontend 環境変数

### 必須

| 変数名 | 説明 | デフォルト値 | 例 |
|--------|------|--------------|-----|
| `VITE_API_BASE_URL` | バックエンドAPIのベースURL | なし | `http://localhost:8080`, `https://api.example.com` |

### オプション

| 変数名 | 説明 | デフォルト値 | 例 |
|--------|------|--------------|-----|
| `VITE_TENANT_ID` | 固定テナントID（テスト用） | なし | `01234567890123456789012345` |
| `VITE_ENABLE_DEBUG` | デバッグモードの有効化 | `false` | `true`, `false` |

### 設定例

#### 開発環境（`web-frontend/.env`）

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_TENANT_ID=01H7XXXXXXXXXXXXXXXXXX
VITE_ENABLE_DEBUG=true
```

#### 本番環境（Render等）

```env
VITE_API_BASE_URL=https://your-backend-domain.com
VITE_TENANT_ID=01H7XXXXXXXXXXXXXXXXXX
VITE_ENABLE_DEBUG=false
```

---

## PostgreSQL 環境変数

Docker Compose 使用時のみ。

| 変数名 | 説明 | デフォルト値 | 例 |
|--------|------|--------------|-----|
| `POSTGRES_DB` | データベース名 | `vrcshift` | `vrcshift` |
| `POSTGRES_USER` | ユーザー名 | `vrcshift` | `vrcshift` |
| `POSTGRES_PASSWORD` | パスワード | なし（必須） | `vrcshift` |
| `POSTGRES_PORT` | ポート番号 | `5432` | `5432` |

---

## 環境別の設定例

### ローカル開発環境

```bash
# backend/.env
DATABASE_URL=postgresql://vrcshift:vrcshift@localhost:5432/vrcshift?sslmode=disable
JWT_SECRET=your_development_secret_key
PORT=8080
ENVIRONMENT=development
LOG_LEVEL=debug
ALLOWED_ORIGINS=http://localhost:5173

# web-frontend/.env
VITE_API_BASE_URL=http://localhost:8080
VITE_TENANT_ID=01H7XXXXXXXXXXXXXXXXXX
VITE_ENABLE_DEBUG=true
```

### Docker Compose（ステージング）

```bash
# .env.prod
POSTGRES_PASSWORD=secure_password_here
BACKEND_PORT=8080
FRONTEND_PORT=80
ENVIRONMENT=staging
LOG_LEVEL=info
ALLOWED_ORIGINS=http://localhost:80
VITE_API_BASE_URL=http://localhost:8080
VITE_TENANT_ID=01H7XXXXXXXXXXXXXXXXXX
```

起動：

```bash
docker-compose -f docker-compose.prod.yml up -d
```

### Render（本番環境）

#### Backend Service

Render の環境変数設定画面で設定：

```
DATABASE_URL=<Renderが自動生成>
JWT_SECRET=<安全なランダム文字列を設定>
PORT=8080
ENVIRONMENT=production
LOG_LEVEL=info
ALLOWED_ORIGINS=https://your-frontend.onrender.com
ENABLE_AUDIT_LOG=false
ENABLE_NOTIFICATION=false
```

#### Frontend Static Site

ビルドコマンド：

```bash
npm install && npm run build
```

環境変数（Build時）：

```
VITE_API_BASE_URL=https://your-backend.onrender.com
VITE_TENANT_ID=01H7XXXXXXXXXXXXXXXXXX
VITE_ENABLE_DEBUG=false
```

---

## トラブルシューティング

### Backend が起動しない

1. `DATABASE_URL` が正しいか確認
2. PostgreSQL が起動しているか確認：`docker ps` or `systemctl status postgresql`
3. ログを確認：`docker logs vrc-shift-backend`

### Frontend がバックエンドに接続できない

1. `VITE_API_BASE_URL` が正しいか確認
2. CORS エラーの場合：Backend の `ALLOWED_ORIGINS` にフロントエンドのURLを追加
3. ブラウザの開発者ツール（Network タブ）でリクエストを確認

### Docker Compose が起動しない

1. `.env.prod` ファイルが存在するか確認
2. `POSTGRES_PASSWORD` が設定されているか確認
3. ポートが空いているか確認：`lsof -i :8080` or `netstat -tuln | grep 8080`

---

## 参考リンク

- [Backend 環境変数サンプル](../backend/.env.example)
- [Frontend 環境変数サンプル](../web-frontend/.env.example)
- [Docker Compose 環境変数サンプル](../.env.prod.example)
- [セットアップガイド](../SETUP.md)


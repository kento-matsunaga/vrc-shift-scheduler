# ポートと環境変数の整合性チェック

> 最終更新: 2025-12-19

## 概要

このドキュメントは、フロントエンドとバックエンド間のポート・環境変数の整合性を確認した結果です。

---

## 1. 推奨設定

### 1.1 バックエンド

**必須環境変数**:

| 変数名 | 説明 | デフォルト | 必須 |
|--------|------|-----------|------|
| `DATABASE_URL` | PostgreSQL接続文字列 | なし | ✅ |
| `JWT_SECRET` | JWT署名シークレット | なし | ✅ |
| `PORT` | サーバーポート | `8080` | - |

**設定例**:

```bash
DATABASE_URL=postgresql://vrcshift:vrcshift@localhost:5432/vrcshift?sslmode=disable
JWT_SECRET=your_secret_key_here
PORT=8080
```

### 1.2 フロントエンド

**環境変数**:

| 変数名 | 説明 | デフォルト |
|--------|------|-----------|
| `VITE_API_BASE_URL` | バックエンドAPIのURL | `http://localhost:8080` |

**設定例** (`web-frontend/.env.local`):

```bash
VITE_API_BASE_URL=http://localhost:8080
```

---

## 2. 開発環境の構成

### 2.1 ローカル開発（Docker Compose不使用）

```
┌──────────────────┐      ┌──────────────────┐      ┌──────────────────┐
│   Frontend       │      │   Backend        │      │   PostgreSQL     │
│   :5173          │─────▶│   :8080          │─────▶│   :5432          │
│   (Vite)         │      │   (Go)           │      │                  │
└──────────────────┘      └──────────────────┘      └──────────────────┘
```

**起動手順**:

```bash
# 1. データベース起動
docker-compose up -d db

# 2. バックエンド起動
cd backend
DATABASE_URL="postgres://vrcshift:vrcshift@localhost:5432/vrcshift?sslmode=disable" \
JWT_SECRET="your_secret" \
go run ./cmd/server

# 3. フロントエンド起動
cd web-frontend
npm run dev
```

### 2.2 Docker Compose

```bash
# 全サービス起動
docker-compose up -d

# ポート確認
docker-compose ps
```

**ポートマッピング**:

| サービス | コンテナポート | ホストポート |
|---------|---------------|-------------|
| backend | 8080 | 8080 |
| frontend | 5173 | 5173 |
| db | 5432 | 5432 |

---

## 3. 動作確認

### 3.1 バックエンド

```bash
# ヘルスチェック
curl http://localhost:8080/health
# 期待: {"status":"ok"}
```

### 3.2 フロントエンド

```bash
# ブラウザでアクセス
open http://localhost:5173
```

### 3.3 API接続確認

ブラウザの開発者ツール（Network タブ）で、APIリクエストが `http://localhost:8080` に送信されていることを確認。

---

## 4. トラブルシューティング

### 問題: フロントエンドからバックエンドに接続できない

**確認事項**:

1. バックエンドが起動しているか確認: `curl http://localhost:8080/health`
2. 環境変数 `VITE_API_BASE_URL` が正しく設定されているか確認
3. CORSエラーの場合: バックエンドの `ALLOWED_ORIGINS` を確認

### 問題: データベースに接続できない

**確認事項**:

1. PostgreSQLが起動しているか: `docker ps`
2. `DATABASE_URL` が正しい形式か確認
3. 認証情報が正しいか確認

### 問題: JWT認証エラー

**確認事項**:

1. `JWT_SECRET` が設定されているか確認
2. トークンの有効期限が切れていないか確認

---

**作成日**: 2025-12-19
**更新者**: ドキュメント検証

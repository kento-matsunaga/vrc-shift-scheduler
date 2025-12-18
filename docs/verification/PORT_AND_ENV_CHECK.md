# ポートと環境変数の整合性チェック

## 概要

このドキュメントは、フロントエンドとバックエンド間のポート・環境変数の整合性を確認し、正しく通信できる設定を整理したものです。

---

## 1. 現在の設定状況

### 1.1 バックエンド側のポート設定

**ファイル**: `backend/cmd/server/main.go`

- **デフォルトポート**: `8080`
- **環境変数**: `PORT` で上書き可能
- **実装箇所**: 24-27行目
  ```go
  port := os.Getenv("PORT")
  if port == "" {
      port = "8080"
  }
  ```

### 1.2 Docker Compose のポート設定

#### docker-compose.yml (開発環境)

**ファイル**: `docker-compose.yml`

- **バックエンドサービス**:
  - コンテナ内ポート: `8080`
  - ホストマッピング: `8090:8080` (29行目)
  - 環境変数: `API_PORT: "8080"` (27行目)
- **データベースサービス**:
  - ホストマッピング: `5432:5432` (9行目)

#### docker-compose.prod.yml (本番環境)

**ファイル**: `docker-compose.prod.yml`

- **バックエンドサービス**:
  - コンテナ内ポート: `8080`
  - ホストマッピング: `${BACKEND_PORT:-8080}:8080` (47行目)
  - 環境変数: `PORT: ${BACKEND_PORT:-8080}` (39行目)
- **フロントエンドサービス**:
  - ビルド時引数: `VITE_API_BASE_URL: ${VITE_API_BASE_URL:-http://localhost:8080}` (63行目)
  - 環境変数: `VITE_API_BASE_URL: ${VITE_API_BASE_URL:-http://localhost:8080}` (70行目)
  - ホストマッピング: `${FRONTEND_PORT:-80}:80` (74行目)

### 1.3 フロントエンド側のAPIベースURL設定

**ファイル**: `web-frontend/src/lib/apiClient.ts`

- **デフォルト値**: `http://localhost:8080` (10行目)
- **環境変数**: `VITE_API_BASE_URL` で上書き可能
- **実装箇所**: 10行目
  ```typescript
  constructor(baseURL: string = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080') {
      this.baseURL = baseURL;
  }
  ```

---

## 2. 問題点の特定

### 2.1 開発環境でのポート不整合

**問題**: `docker-compose.yml` では、バックエンドがホストの `8090` ポートにマッピングされているが、フロントエンドのデフォルトは `8080` を参照している。

- **バックエンド**: ホストの `8090` ポートでリッスン
- **フロントエンド**: `http://localhost:8080` をデフォルトで参照

**影響**: Docker Compose で起動した場合、フロントエンドからバックエンドに接続できない。

### 2.2 環境変数の未設定

**問題**: フロントエンド側で `.env` ファイルが存在しないため、環境変数による設定ができない。

**影響**: 開発環境ごとに異なるポート設定に対応できない。

---

## 3. 推奨設定

### 3.1 ローカル開発モード（Docker Compose 不使用）

**前提**: バックエンドとフロントエンドを別々のターミナルで起動する場合

#### バックエンド起動コマンド

```bash
cd backend
go run ./cmd/server/main.go
# または
PORT=8080 go run ./cmd/server/main.go
```

- **リッスンポート**: `:8080`

#### フロントエンド起動コマンド

```bash
cd web-frontend
npm run dev -- --port 5173
```

- **開発サーバーポート**: `:5173`
- **APIベースURL**: `http://localhost:8080` (デフォルト)

**推奨**: フロントエンド側で `.env.local` を作成して明示的に設定

```bash
# web-frontend/.env.local
VITE_API_BASE_URL=http://localhost:8080
```

### 3.2 Docker Compose 開発環境

**前提**: `docker-compose.yml` を使用して起動する場合

#### 推奨設定変更

**オプション1: docker-compose.yml のポートマッピングを変更**

`docker-compose.yml` の29行目を以下のように変更:

```yaml
ports:
  - "8080:8080"  # 8090:8080 から変更
```

**オプション2: フロントエンドの環境変数を設定**

`web-frontend/.env.local` を作成:

```bash
# web-frontend/.env.local
VITE_API_BASE_URL=http://localhost:8090
```

**推奨**: オプション1を採用（ポート番号を統一）

### 3.3 本番環境（docker-compose.prod.yml）

**前提**: 本番環境では、フロントエンドとバックエンドが同じネットワーク内で通信する

#### 推奨設定

**環境変数ファイル** (`.env.production` など):

```bash
BACKEND_PORT=8080
FRONTEND_PORT=80
VITE_API_BASE_URL=http://backend:8080  # コンテナ間通信
# または
VITE_API_BASE_URL=http://localhost:8080  # 外部からアクセスする場合
```

**注意**: ビルド時に `VITE_API_BASE_URL` が埋め込まれるため、ビルド前に正しい値を設定する必要がある。

---

## 4. 修正手順

### 4.1 開発環境の修正（推奨）

#### 手順1: docker-compose.yml のポートマッピングを変更

**ファイル**: `docker-compose.yml`

```yaml
# 変更前
ports:
  - "8090:8080"

# 変更後
ports:
  - "8080:8080"
```

#### 手順2: フロントエンドの環境変数ファイルを作成（オプション）

**ファイル**: `web-frontend/.env.local`

```bash
VITE_API_BASE_URL=http://localhost:8080
VITE_TENANT_ID=your_tenant_id_here
```

**注意**: `.env.local` は Git にコミットしない（`.gitignore` に追加推奨）

### 4.2 本番環境の修正

#### 手順1: 環境変数ファイルの作成

**ファイル**: `.env.production`

```bash
BACKEND_PORT=8080
FRONTEND_PORT=80
VITE_API_BASE_URL=http://your-domain.com:8080
VITE_TENANT_ID=your_tenant_id_here
```

#### 手順2: docker-compose.prod.yml のビルド引数を確認

**ファイル**: `docker-compose.prod.yml`

63行目の `args` セクションで、`VITE_API_BASE_URL` が正しく設定されていることを確認。

---

## 5. 動作確認手順

### 5.1 ローカル開発環境

1. **バックエンド起動**
   ```bash
   cd backend
   go run ./cmd/server/main.go
   ```
   - ログに `Starting server on port 8080...` が表示されることを確認

2. **フロントエンド起動**
   ```bash
   cd web-frontend
   npm run dev -- --port 5173
   ```
   - ブラウザで `http://localhost:5173` にアクセス

3. **API接続確認**
   - ブラウザの開発者ツールで Network タブを開く
   - ログイン画面で API リクエストが `http://localhost:8080` に送信されていることを確認

### 5.2 Docker Compose 環境

1. **サービス起動**
   ```bash
   docker-compose up -d
   ```

2. **ポート確認**
   ```bash
   docker-compose ps
   ```
   - バックエンドが `0.0.0.0:8080->8080/tcp` でマッピングされていることを確認

3. **API接続確認**
   ```bash
   curl http://localhost:8080/health
   ```
   - `{"status":"ok"}` が返ることを確認

---

## 6. まとめ

### 現在の状態

- **開発環境**: ポート不整合あり（バックエンド: 8090、フロントエンド: 8080）
- **本番環境**: 設定は正しいが、環境変数の明示的な設定が必要

### 推奨される修正

1. **docker-compose.yml**: ポートマッピングを `8090:8080` から `8080:8080` に変更
2. **web-frontend/.env.local**: 開発環境用の環境変数ファイルを作成（オプション）
3. **本番環境**: `.env.production` を作成して環境変数を明示的に設定

### 修正後の期待動作

- **ローカル開発**: バックエンド `:8080`、フロントエンド `:5173`、API接続 `http://localhost:8080`
- **Docker Compose**: バックエンド `:8080`、フロントエンド `:5173`（または `:80`）、API接続 `http://localhost:8080`
- **本番環境**: バックエンド `:8080`、フロントエンド `:80`、API接続は環境変数で設定

---

**作成日**: 2025-01-XX  
**作成者**: 検証専用アシスタント（Auto）





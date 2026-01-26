# 開発環境セットアップ

VRC Shift Scheduler の開発環境セットアップ手順です。

## 必要な環境

- **Go**: 1.24 以上
- **Node.js**: 20.x 以上
- **PostgreSQL**: 16
- **Git**

## 1. Go のインストール

### 自動インストール（推奨）

プロジェクトルートで以下のスクリプトを実行してください：

```bash
./scripts/install-go-local.sh
```

このスクリプトは：
- Go 1.24 を `$HOME/.local/go` にインストール
- `$HOME/.bashrc` に PATH を自動追加

インストール後、以下のコマンドで PATH を反映：

```bash
source ~/.bashrc
# または
export PATH=$HOME/.local/go/bin:$HOME/go/bin:$PATH
```

### 手動インストール

1. [Go 公式サイト](https://go.dev/dl/)から Go 1.24 以上をダウンロード
2. インストール先を選択（例: `/usr/local/go` または `$HOME/.local/go`）
3. PATH を設定：
   ```bash
   export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
   ```

## 2. Node.js のインストール

### nvm を使用（推奨）

```bash
# nvm のインストール（未インストールの場合）
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash

# Node.js 20.x をインストール
nvm install 20
nvm use 20
```

### 手動インストール

[Node.js 公式サイト](https://nodejs.org/)から LTS 版をダウンロードしてインストール。

## 3. PostgreSQL のインストール

### Ubuntu/Debian

```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

### macOS

```bash
brew install postgresql@16
brew services start postgresql@16
```

### Docker（開発環境）

```bash
docker run --name vrc-shift-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=vrc_shift_scheduler \
  -p 5432:5432 \
  -d postgres:16
```

## 4. データベースのセットアップ

### データベースの作成

```bash
# PostgreSQL にログイン
psql -U postgres

# データベース作成（Docker 以外の場合）
CREATE DATABASE vrc_shift_scheduler;
CREATE USER vrc_shift_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE vrc_shift_scheduler TO vrc_shift_user;
\q
```

### 環境変数の設定

backend/.env ファイルを作成：

```bash
cd backend
cp .env.example .env
```

`.env` を編集：

```env
DATABASE_URL=postgresql://vrc_shift_user:your_password@localhost:5432/vrc_shift_scheduler?sslmode=disable
PORT=8080
```

### マイグレーションの実行

```bash
cd backend
go run ./cmd/migrate/main.go
```

## 5. フロントエンドのセットアップ

```bash
cd web-frontend

# 依存関係のインストール
npm install

# 環境変数の設定
cp .env.example .env
```

`.env` を編集：

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_TENANT_ID=your-tenant-ulid
```

## 6. 動作確認

### バックエンドの起動

```bash
cd backend
go run ./cmd/server/main.go
```

サーバーが起動したら、http://localhost:8080/health でヘルスチェック。

### フロントエンドの起動

```bash
cd web-frontend
npm run dev
```

ブラウザで http://localhost:5173 にアクセス。

## 7. ビルド

### バックエンド

```bash
cd backend
go build -o bin/server ./cmd/server/main.go
./bin/server
```

### フロントエンド

```bash
cd web-frontend
npm run build
# dist/ フォルダに静的ファイルが生成されます
```

## トラブルシューティング

### Go のバージョンが古い

```bash
go version  # 1.24 以上であることを確認
```

1.24 未満の場合：

```bash
# プロジェクトのインストールスクリプトを実行
./scripts/install-go-local.sh

# PATH を再読み込み
source ~/.bashrc
```

### PostgreSQL 接続エラー

```bash
# PostgreSQL が起動しているか確認
sudo systemctl status postgresql

# 接続テスト
psql -h localhost -U vrc_shift_user -d vrc_shift_scheduler
```

### npm install エラー

```bash
# Node.js のバージョン確認
node -v  # 18.x 以上

# キャッシュをクリア
npm cache clean --force
npm install
```

## 開発用コマンド

### バックエンド

```bash
# テスト実行
go test ./...

# 統合テスト（DB が必要）
go test -tags=integration ./internal/infra/db/...

# フォーマット
go fmt ./...

# Lint
golangci-lint run
```

### フロントエンド

```bash
# 開発サーバー起動
npm run dev

# ビルド
npm run build

# プレビュー（ビルド後）
npm run preview

# Lint
npm run lint
```

## 次のステップ

- [開発ガイド](../development/DEVELOPMENT.md) で開発ガイド・API情報を確認
- [環境変数](./ENVIRONMENT_VARIABLES.md) で環境変数を確認


#!/bin/bash
#
# VRC Shift Scheduler ブートストラップスクリプト
# プロジェクトの開発環境を一括でセットアップします
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  VRC Shift Scheduler - Bootstrap"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# ==================== Go のチェック ====================
echo "📦 [1/5] Go のバージョンチェック..."

REQUIRED_GO_VERSION="1.23"
CURRENT_GO_VERSION=""

if command -v go &> /dev/null; then
    CURRENT_GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo "   現在の Go バージョン: ${CURRENT_GO_VERSION}"
    
    # バージョン比較（簡易版）
    MAJOR_MINOR=$(echo "$CURRENT_GO_VERSION" | cut -d. -f1,2)
    if [[ "$(printf '%s\n' "$REQUIRED_GO_VERSION" "$MAJOR_MINOR" | sort -V | head -n1)" != "$REQUIRED_GO_VERSION" ]]; then
        echo "   ⚠️  Go のバージョンが古いです（必要: ${REQUIRED_GO_VERSION} 以上）"
        read -p "   Go をアップグレードしますか？ [y/N]: " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            bash "$SCRIPT_DIR/install-go-local.sh"
            export PATH=$HOME/.local/go/bin:$PATH
            echo "   ✅ Go をアップグレードしました"
        else
            echo "   ⚠️  古い Go バージョンではビルドに失敗する可能性があります"
        fi
    else
        echo "   ✅ Go のバージョンは要件を満たしています"
    fi
else
    echo "   ❌ Go がインストールされていません"
    read -p "   Go をインストールしますか？ [y/N]: " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        bash "$SCRIPT_DIR/install-go-local.sh"
        export PATH=$HOME/.local/go/bin:$PATH
        echo "   ✅ Go をインストールしました"
    else
        echo "   ❌ Go が必要です。インストール後に再実行してください"
        exit 1
    fi
fi

echo ""

# ==================== Node.js のチェック ====================
echo "📦 [2/5] Node.js のバージョンチェック..."

if command -v node &> /dev/null; then
    NODE_VERSION=$(node -v | sed 's/v//')
    echo "   現在の Node.js バージョン: ${NODE_VERSION}"
    
    MAJOR_VERSION=$(echo "$NODE_VERSION" | cut -d. -f1)
    if [ "$MAJOR_VERSION" -lt 18 ]; then
        echo "   ⚠️  Node.js のバージョンが古いです（推奨: 18.x 以上）"
    else
        echo "   ✅ Node.js のバージョンは要件を満たしています"
    fi
else
    echo "   ⚠️  Node.js がインストールされていません"
    echo "   フロントエンド開発には Node.js 18.x 以上が必要です"
    echo "   インストール手順: https://nodejs.org/ または nvm を使用"
fi

echo ""

# ==================== PostgreSQL のチェック ====================
echo "📦 [3/5] PostgreSQL のチェック..."

if command -v psql &> /dev/null; then
    PG_VERSION=$(psql --version | awk '{print $3}')
    echo "   PostgreSQL バージョン: ${PG_VERSION}"
    echo "   ✅ PostgreSQL がインストールされています"
else
    echo "   ⚠️  PostgreSQL がインストールされていません"
    echo "   Docker で起動する場合:"
    echo "   docker run --name vrc-shift-postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres:14"
fi

echo ""

# ==================== 環境変数ファイルのチェック ====================
echo "📝 [4/5] 環境変数ファイルのチェック..."

# backend/.env
if [ ! -f "$PROJECT_ROOT/backend/.env" ]; then
    if [ -f "$PROJECT_ROOT/backend/.env.example" ]; then
        cp "$PROJECT_ROOT/backend/.env.example" "$PROJECT_ROOT/backend/.env"
        echo "   ✅ backend/.env を作成しました（.env.example からコピー）"
        echo "   ⚠️  DATABASE_URL を編集してください"
    else
        echo "   ⚠️  backend/.env.example が見つかりません"
    fi
else
    echo "   ✅ backend/.env は既に存在します"
fi

# web-frontend/.env
if [ ! -f "$PROJECT_ROOT/web-frontend/.env" ]; then
    if [ -f "$PROJECT_ROOT/web-frontend/.env.example" ]; then
        cp "$PROJECT_ROOT/web-frontend/.env.example" "$PROJECT_ROOT/web-frontend/.env"
        echo "   ✅ web-frontend/.env を作成しました（.env.example からコピー）"
        echo "   ⚠️  VITE_TENANT_ID を設定してください"
    else
        echo "   ⚠️  web-frontend/.env.example が見つかりません"
    fi
else
    echo "   ✅ web-frontend/.env は既に存在します"
fi

echo ""

# ==================== 依存関係のインストール ====================
echo "📥 [5/5] 依存関係のインストール..."

# Go の依存関係
echo "   → Go モジュールの依存関係を整理中..."
cd "$PROJECT_ROOT/backend"
go mod tidy
echo "   ✅ Go の依存関係を整理しました"

# Node.js の依存関係
if command -v npm &> /dev/null && [ -d "$PROJECT_ROOT/web-frontend" ]; then
    echo "   → npm パッケージをインストール中..."
    cd "$PROJECT_ROOT/web-frontend"
    if [ -f "package.json" ]; then
        npm install
        echo "   ✅ npm パッケージをインストールしました"
    else
        echo "   ⚠️  web-frontend/package.json が見つかりません"
    fi
else
    if ! command -v npm &> /dev/null; then
        echo "   ⚠️  npm がインストールされていません（Node.js が必要）"
    else
        echo "   ⚠️  web-frontend ディレクトリが見つかりません"
    fi
fi

cd "$PROJECT_ROOT"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✅ ブートストラップ完了！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📝 次のステップ:"
echo ""
echo "1. データベースのセットアップ:"
echo "   cd backend && go run ./cmd/migrate/main.go"
echo ""
echo "2. バックエンドの起動:"
echo "   cd backend && go run ./cmd/server/main.go"
echo ""
echo "3. フロントエンドの起動:"
echo "   cd web-frontend && npm run dev"
echo ""
echo "詳細は SETUP.md を参照してください。"
echo ""


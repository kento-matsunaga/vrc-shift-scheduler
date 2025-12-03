#!/bin/bash
#
# デプロイ前チェックスクリプト
# ビルドとヘルスチェックを確認
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  デプロイ前チェック"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

cd "$PROJECT_ROOT"

# ==================== 1. Backend ビルドチェック ====================
echo "📦 [1/4] Backend ビルドチェック..."
cd backend
export PATH=$HOME/.local/go/bin:$PATH

if go build -o /tmp/server-test ./cmd/server/main.go; then
    echo "   ✅ Backend ビルド成功"
    rm -f /tmp/server-test
else
    echo "   ❌ Backend ビルド失敗"
    exit 1
fi

cd "$PROJECT_ROOT"
echo ""

# ==================== 2. Frontend ビルドチェック ====================
echo "📦 [2/4] Frontend ビルドチェック..."
cd web-frontend

if [ ! -d "node_modules" ]; then
    echo "   ⚠️  node_modules がありません。npm install を実行します..."
    npm install
fi

if npm run build; then
    echo "   ✅ Frontend ビルド成功"
else
    echo "   ❌ Frontend ビルド失敗"
    exit 1
fi

cd "$PROJECT_ROOT"
echo ""

# ==================== 3. Docker イメージビルドチェック ====================
echo "🐳 [3/4] Docker イメージビルドチェック..."

echo "   → Backend イメージビルド..."
if docker build -t vrc-shift-backend:test -f backend/Dockerfile backend/; then
    echo "   ✅ Backend イメージビルド成功"
else
    echo "   ❌ Backend イメージビルド失敗"
    exit 1
fi

echo "   → Frontend イメージビルド..."
if docker build -t vrc-shift-frontend:test \
    --build-arg VITE_API_BASE_URL=http://localhost:8080 \
    --build-arg VITE_TENANT_ID=test \
    -f web-frontend/Dockerfile web-frontend/; then
    echo "   ✅ Frontend イメージビルド成功"
else
    echo "   ❌ Frontend イメージビルド失敗"
    exit 1
fi

echo ""

# ==================== 4. 環境変数ファイルチェック ====================
echo "📝 [4/4] 環境変数ファイルチェック..."

if [ ! -f ".env.prod" ]; then
    echo "   ⚠️  .env.prod が見つかりません"
    echo "   .env.prod.example をコピーして設定してください："
    echo "   cp .env.prod.example .env.prod"
else
    echo "   ✅ .env.prod が存在します"
    
    # 必須の環境変数をチェック
    missing_vars=()
    required_vars=("POSTGRES_PASSWORD" "VITE_API_BASE_URL")
    
    for var in "${required_vars[@]}"; do
        if ! grep -q "^${var}=" .env.prod || grep -q "^${var}=$" .env.prod || grep -q "^${var}=CHANGE_ME" .env.prod; then
            missing_vars+=("$var")
        fi
    done
    
    if [ ${#missing_vars[@]} -gt 0 ]; then
        echo "   ⚠️  以下の環境変数が未設定です："
        for var in "${missing_vars[@]}"; do
            echo "      - $var"
        done
    else
        echo "   ✅ 必須の環境変数が設定されています"
    fi
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✅ デプロイ前チェック完了"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📝 次のステップ:"
echo "1. .env.prod を確認・編集"
echo "2. docker-compose -f docker-compose.prod.yml up -d"
echo "3. ヘルスチェック確認"
echo "   curl http://localhost:8080/health"
echo "   curl http://localhost/"
echo ""


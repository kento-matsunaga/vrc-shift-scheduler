#!/usr/bin/env bash
set -e

cd "$(dirname "$0")/.."

echo "🚀 VRC Shift Scheduler - Development Bootstrap"
echo "=============================================="

# Create .env if it doesn't exist
if [ ! -f .env ]; then
  cp .env.example .env
  echo "✅ .env を作成しました（中身は適宜編集してください）"
else
  echo "📝 .env は既に存在します"
fi

# Generate go.sum if it doesn't exist
if [ ! -f backend/go.sum ]; then
  echo ""
  echo "📦 backend/go.sum を生成中..."
  docker run --rm -v "$(pwd)/backend:/app" -w /app golang:1.22 go mod tidy
  echo "✅ go.sum を生成しました"
fi

# Build Docker images
echo ""
echo "🔨 Docker イメージをビルド中..."
docker compose build

# Start database only
echo ""
echo "🐘 PostgreSQL を起動中..."
docker compose up -d db

# Wait for database to be ready
echo ""
echo "⏳ PostgreSQL の起動を待機中..."
sleep 3

# Check if database is ready
until docker compose exec -T db pg_isready -U vrcshift -d vrcshift > /dev/null 2>&1; do
  echo "  PostgreSQL is not ready yet, waiting..."
  sleep 2
done

echo "✅ PostgreSQL が起動しました"

echo ""
echo "=============================================="
echo "🎉 セットアップ完了！"
echo ""
echo "次のステップ:"
echo "  1. .env ファイルを編集して Discord Bot トークンを設定"
echo "  2. Backend を起動:"
echo "     - Docker: docker compose up backend"
echo "     - ローカル: cd backend && go run ./cmd/api"
echo "  3. Bot を起動:"
echo "     - Docker: docker compose up bot"
echo "     - ローカル: cd bot && pnpm install && pnpm dev"
echo ""
echo "全サービス一括起動: docker compose up"
echo "=============================================="


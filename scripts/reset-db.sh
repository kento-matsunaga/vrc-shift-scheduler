#!/bin/bash
#
# データベースリセットスクリプト
# 全テーブルをTRUNCATEしてシードデータを投入
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 環境の選択
ENVIRONMENT="${1:-development}"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  データベースリセット"
echo "  Environment: ${ENVIRONMENT}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 本番環境では実行不可
if [ "$ENVIRONMENT" = "production" ]; then
    echo "❌ 本番環境でのリセットは禁止されています"
    exit 1
fi

# 確認プロンプト
read -p "本当にデータベースをリセットしますか？ [y/N]: " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "キャンセルしました"
    exit 0
fi

# 環境変数の読み込み
if [ -f "$PROJECT_ROOT/backend/.env" ]; then
    export $(grep -v '^#' "$PROJECT_ROOT/backend/.env" | xargs)
elif [ -f "$PROJECT_ROOT/.env.prod" ]; then
    export $(grep -v '^#' "$PROJECT_ROOT/.env.prod" | xargs)
else
    echo "❌ .env ファイルが見つかりません"
    exit 1
fi

# DATABASE_URLの確認
if [ -z "$DATABASE_URL" ]; then
    echo "❌ DATABASE_URL が設定されていません"
    exit 1
fi

echo "📊 データベース接続確認..."

# PostgreSQL接続情報を抽出
DB_HOST=$(echo $DATABASE_URL | sed -n 's/.*@\([^:]*\):.*/\1/p')
DB_PORT=$(echo $DATABASE_URL | sed -n 's/.*:\([0-9]*\)\/.*/\1/p')
DB_NAME=$(echo $DATABASE_URL | sed -n 's/.*\/\([^?]*\).*/\1/p')
DB_USER=$(echo $DATABASE_URL | sed -n 's/.*\/\/\([^:]*\):.*/\1/p')
DB_PASS=$(echo $DATABASE_URL | sed -n 's/.*:\/\/[^:]*:\([^@]*\)@.*/\1/p')

export PGPASSWORD="$DB_PASS"

echo "   Host: $DB_HOST"
echo "   Port: $DB_PORT"
echo "   Database: $DB_NAME"
echo "   User: $DB_USER"
echo ""

echo "🗑️  テーブルをTRUNCATEしています..."

# テーブル一覧を取得してTRUNCATE
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" <<EOF
-- 外部キー制約を一時的に無効化
SET session_replication_role = replica;

-- 全テーブルをTRUNCATE（シーケンスもリセット）
TRUNCATE TABLE
    shift_assignments,
    shift_slots,
    event_business_days,
    events,
    members
RESTART IDENTITY CASCADE;

-- 外部キー制約を再度有効化
SET session_replication_role = DEFAULT;

-- 確認
SELECT 'shift_assignments' as table_name, COUNT(*) as count FROM shift_assignments
UNION ALL
SELECT 'shift_slots', COUNT(*) FROM shift_slots
UNION ALL
SELECT 'event_business_days', COUNT(*) FROM event_business_days
UNION ALL
SELECT 'events', COUNT(*) FROM events
UNION ALL
SELECT 'members', COUNT(*) FROM members;
EOF

if [ $? -eq 0 ]; then
    echo "✅ TRUNCATE 完了"
else
    echo "❌ TRUNCATE 失敗"
    exit 1
fi

echo ""
echo "🌱 シードデータを投入しています..."

cd "$PROJECT_ROOT/backend"

# PATH を設定（Go 1.24.11を使用）
export PATH=$HOME/.local/go/bin:$PATH

# シードコマンドを実行
if go run ./cmd/seed/main.go --env="$ENVIRONMENT" --tenants=1; then
    echo "✅ シード投入完了"
else
    echo "❌ シード投入失敗"
    exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✅ データベースリセット完了"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📊 現在のデータ："
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
SELECT 'events' as table_name, COUNT(*) as count FROM events
UNION ALL
SELECT 'business_days', COUNT(*) FROM event_business_days
UNION ALL
SELECT 'shift_slots', COUNT(*) FROM shift_slots
UNION ALL
SELECT 'members', COUNT(*) FROM members
UNION ALL
SELECT 'assignments', COUNT(*) FROM shift_assignments;"
echo ""


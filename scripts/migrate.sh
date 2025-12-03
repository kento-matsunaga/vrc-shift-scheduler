#!/usr/bin/env bash
set -e

cd "$(dirname "$0")/.."

# TODO: 将来 DB マイグレーションツールを入れる
# 候補:
#   - golang-migrate/migrate
#   - pressly/goose
#   - atlas

echo "⚠️  マイグレーションツールは未実装です"
echo "将来的に golang-migrate または goose を導入予定"


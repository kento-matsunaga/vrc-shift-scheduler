#!/bin/bash
# テスト結果をチェックし、失敗時に警告を表示
# PostToolUse: go test 実行後に呼び出される

read -r input

if echo "$input" | grep -q "FAIL"; then
    echo "[Hook] テスト失敗を検出" >&2
fi

echo "$input"

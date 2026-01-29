#!/bin/bash
# PR作成後にURLを表示
# PostToolUse: gh pr create 実行後に呼び出される

read -r input

if url=$(echo "$input" | grep -oP "https://github.com/[^/]+/[^/]+/pull/\d+"); then
    echo "[Hook] PR作成完了: $url" >&2
fi

echo "$input"

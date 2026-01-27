#!/bin/bash
# Goファイルにfmt.Printlnが残っていないかチェック
# 使用方法:
#   PostToolUse: echo "$input" | bash check-fmt-println.sh --file
#   Stop: bash check-fmt-println.sh --git

set -e

MODE="${1:---git}"

case "$MODE" in
    --file)
        # PostToolUse: 編集されたファイルをチェック
        read -r input
        file=$(echo "$input" | jq -r ".tool_input.file_path // empty" 2>/dev/null)

        if [ -n "$file" ] && [ -f "$file" ]; then
            if grep -n "fmt.Println" "$file" 2>/dev/null; then
                echo "[Hook] WARNING: fmt.Printlnが残っています" >&2
            fi
        fi

        echo "$input"
        ;;
    --git)
        # Stop: Git差分のGoファイルをチェック
        if git rev-parse --git-dir > /dev/null 2>&1; then
            files=$(git diff --name-only HEAD 2>/dev/null | grep "\.go$" || true)

            if [ -n "$files" ]; then
                for f in $files; do
                    if [ -f "$f" ] && grep -l "fmt.Println" "$f" 2>/dev/null; then
                        echo "[Hook] WARNING: $f にfmt.Printlnが残っています" >&2
                    fi
                done
            fi
        fi
        ;;
    *)
        echo "Usage: $0 [--file|--git]" >&2
        exit 1
        ;;
esac

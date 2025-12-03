#!/bin/bash
#
# Go インストール・アップグレードスクリプト（sudo 不要版）
# ホームディレクトリにインストールします
#
# Usage: ./scripts/install-go-local.sh [version]
#

set -e

# デフォルトバージョン
DEFAULT_VERSION="1.23.4"
GO_VERSION="${1:-$DEFAULT_VERSION}"

# インストール先
INSTALL_DIR="$HOME/.local"
GO_ROOT="$INSTALL_DIR/go"

echo "🚀 Go ${GO_VERSION} をインストールします..."
echo "   インストール先: ${GO_ROOT}"

# アーキテクチャの検出
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "❌ サポートされていないアーキテクチャ: $ARCH"
        exit 1
        ;;
esac

# ダウンロード URL
GO_TAR="go${GO_VERSION}.${OS}-${ARCH}.tar.gz"
GO_URL="https://go.dev/dl/${GO_TAR}"

echo "📥 ダウンロード中: ${GO_URL}"

# 一時ディレクトリでダウンロード
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

if ! curl -LO "$GO_URL"; then
    echo "❌ ダウンロードに失敗しました"
    rm -rf "$TMP_DIR"
    exit 1
fi

# インストールディレクトリ作成
mkdir -p "$INSTALL_DIR"

# 既存の Go を削除
if [ -d "$GO_ROOT" ]; then
    echo "🗑️  既存の Go を削除中..."
    rm -rf "$GO_ROOT"
fi

# 新しい Go を展開
echo "📦 Go ${GO_VERSION} をインストール中..."
tar -C "$INSTALL_DIR" -xzf "$GO_TAR"

# クリーンアップ
cd - > /dev/null
rm -rf "$TMP_DIR"

# PATH 設定の確認・追加
SHELL_RC=""
if [ -f "$HOME/.bashrc" ]; then
    SHELL_RC="$HOME/.bashrc"
elif [ -f "$HOME/.zshrc" ]; then
    SHELL_RC="$HOME/.zshrc"
elif [ -f "$HOME/.profile" ]; then
    SHELL_RC="$HOME/.profile"
fi

if [ -n "$SHELL_RC" ]; then
    # Go の PATH 設定がなければ追加
    if ! grep -q "export PATH=.*:$HOME/.local/go/bin" "$SHELL_RC"; then
        echo "" >> "$SHELL_RC"
        echo "# Go (installed by vrc-shift-scheduler setup)" >> "$SHELL_RC"
        echo "export PATH=\$PATH:\$HOME/.local/go/bin" >> "$SHELL_RC"
        echo "export PATH=\$PATH:\$HOME/go/bin" >> "$SHELL_RC"
        echo "✅ PATH を ${SHELL_RC} に追加しました"
    else
        echo "✅ PATH は既に設定済みです"
    fi
fi

# 現在のシェルにも PATH を適用
export PATH=$PATH:$HOME/.local/go/bin
export PATH=$PATH:$HOME/go/bin

# インストール確認
INSTALLED_VERSION=$($HOME/.local/go/bin/go version | awk '{print $3}')
echo ""
echo "🎉 Go のインストールが完了しました！"
echo "   インストールバージョン: ${INSTALLED_VERSION}"
echo "   インストール先: ${GO_ROOT}"
echo ""
echo "📝 次のコマンドを実行してください："
echo "   source ${SHELL_RC}"
echo ""
echo "または、今すぐ使うには："
echo "   export PATH=\$PATH:\$HOME/.local/go/bin:\$HOME/go/bin"
echo ""


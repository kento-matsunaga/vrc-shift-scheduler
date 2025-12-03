#!/bin/bash
#
# Go インストール・アップグレードスクリプト
# Usage: ./scripts/install-go.sh [version]
#
# 例:
#   ./scripts/install-go.sh          # 最新の 1.23 系をインストール
#   ./scripts/install-go.sh 1.23.4   # 特定バージョンをインストール
#

set -e

# デフォルトバージョン
DEFAULT_VERSION="1.23.4"
GO_VERSION="${1:-$DEFAULT_VERSION}"

echo "🚀 Go ${GO_VERSION} をインストールします..."

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

# 既存の Go を削除（/usr/local/go）
if [ -d "/usr/local/go" ]; then
    echo "🗑️  既存の Go を削除中..."
    sudo rm -rf /usr/local/go
fi

# 新しい Go を展開
echo "📦 Go ${GO_VERSION} をインストール中..."
sudo tar -C /usr/local -xzf "$GO_TAR"

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
    if ! grep -q "export PATH=.*:/usr/local/go/bin" "$SHELL_RC"; then
        echo "" >> "$SHELL_RC"
        echo "# Go (installed by vrc-shift-scheduler setup)" >> "$SHELL_RC"
        echo 'export PATH=$PATH:/usr/local/go/bin' >> "$SHELL_RC"
        echo 'export PATH=$PATH:$HOME/go/bin' >> "$SHELL_RC"
        echo "✅ PATH を ${SHELL_RC} に追加しました"
    else
        echo "✅ PATH は既に設定済みです"
    fi
fi

# 現在のシェルにも PATH を適用
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:$HOME/go/bin

# インストール確認
INSTALLED_VERSION=$(/usr/local/go/bin/go version | awk '{print $3}')
echo ""
echo "🎉 Go のインストールが完了しました！"
echo "   インストールバージョン: ${INSTALLED_VERSION}"
echo ""
echo "📝 次のコマンドを実行してください："
echo "   source ${SHELL_RC}"
echo "   または新しいターミナルを開いてください"
echo ""


#!/bin/sh
set -e

# 環境変数を JavaScript ファイルとして出力
# フロントエンドから window.__ENV__ でアクセス可能
cat > /usr/share/nginx/html/env-config.js <<EOF
window.__ENV__ = {
  VITE_API_BASE_URL: "${VITE_API_BASE_URL:-http://localhost:8080}",
  VITE_TENANT_ID: "${VITE_TENANT_ID:-}",
  VITE_ENABLE_DEBUG: "${VITE_ENABLE_DEBUG:-false}"
};
EOF

# index.html に env-config.js を自動挿入
if [ -f /usr/share/nginx/html/index.html ]; then
    sed -i 's|</head>|<script src="/env-config.js"></script></head>|' /usr/share/nginx/html/index.html
fi

echo "✅ Environment configuration applied:"
cat /usr/share/nginx/html/env-config.js

# Nginx を起動
exec "$@"


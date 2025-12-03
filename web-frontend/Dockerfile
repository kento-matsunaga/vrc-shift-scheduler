# VRC Shift Scheduler - Frontend Dockerfile
# Multi-stage build for production

# ========== Build Stage ==========
FROM node:20-alpine AS builder

# 作業ディレクトリを設定
WORKDIR /app

# package.json と package-lock.json をコピー（キャッシュ効率化）
COPY package*.json ./

# 依存関係をインストール
RUN npm ci

# ソースコードをコピー
COPY . .

# ビルド（環境変数は実行時に設定）
ARG VITE_API_BASE_URL
ARG VITE_TENANT_ID
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
ENV VITE_TENANT_ID=$VITE_TENANT_ID

RUN npm run build

# ========== Runtime Stage ==========
FROM nginx:alpine

# Nginx 設定をコピー
COPY nginx.conf /etc/nginx/nginx.conf

# ビルド成果物をコピー
COPY --from=builder /app/dist /usr/share/nginx/html

# 環境変数を読み込むスクリプトをコピー
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# ヘルスチェック
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

# ポートを公開
EXPOSE 80

# エントリポイントを設定
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]


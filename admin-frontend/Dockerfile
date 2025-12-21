# VRC Shift Scheduler - Admin Frontend Dockerfile
# Multi-stage build for production

# ========== Build Stage ==========
FROM node:20-alpine AS builder

WORKDIR /app

# package.json と package-lock.json をコピー
COPY package*.json ./

# 依存関係をインストール
RUN npm ci

# ソースコードをコピー
COPY . .

# ビルド
ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

RUN npm run build

# ========== Runtime Stage ==========
FROM nginx:alpine

# Nginx 設定をコピー
COPY nginx.conf /etc/nginx/nginx.conf

# ビルド成果物をコピー
COPY --from=builder /app/dist /usr/share/nginx/html

# ヘルスチェック
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

# ポートを公開
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]

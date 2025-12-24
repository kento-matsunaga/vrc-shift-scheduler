# VRC Shift Scheduler - Backend Dockerfile
# Multi-stage build for production

# ========== Build Stage ==========
FROM golang:1.24-alpine AS builder

# 必要なツールをインストール
RUN apk add --no-cache git

# 作業ディレクトリを設定
WORKDIR /app

# Go モジュールのダウンロード（キャッシュ効率化）
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# バイナリをビルド（静的リンク）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/server \
    ./cmd/server/main.go

# マイグレーションコマンドもビルド
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/migrate \
    ./cmd/migrate/main.go

# シードコマンドもビルド
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/seed \
    ./cmd/seed/main.go

# ========== Runtime Stage ==========
FROM alpine:latest

# 必要なパッケージをインストール
RUN apk --no-cache add ca-certificates tzdata

# 作業ディレクトリを設定
WORKDIR /app

# ビルド成果物をコピー
COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/bin/migrate /app/migrate
COPY --from=builder /app/bin/seed /app/seed

# 環境変数のデフォルト値を設定
ENV PORT=8080
ENV HOST=0.0.0.0
ENV ENVIRONMENT=production
ENV LOG_LEVEL=info
ENV TZ=Asia/Tokyo

# ヘルスチェック
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/health || exit 1

# ポートを公開
EXPOSE 8080

# サーバーを起動
CMD ["/app/server"]

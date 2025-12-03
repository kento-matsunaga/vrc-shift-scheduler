# Backend API

VRC Shift Scheduler の Go 製バックエンド API サーバです。

## 技術スタック

- Go 1.22+
- HTTP Router: [chi](https://github.com/go-chi/chi)
- Database: PostgreSQL (pgx)
- Configuration: envconfig

## ディレクトリ構成

```
backend/
├── cmd/
│   └── api/
│       └── main.go          # エントリポイント
├── internal/
│   ├── config/
│   │   └── config.go        # 環境変数設定
│   ├── http/
│   │   └── router.go        # HTTPルーター
│   ├── domain/              # ドメイン層（エンティティ、VO、リポジトリIF）
│   │   ├── tenant/
│   │   ├── event/
│   │   ├── member/
│   │   ├── shift/
│   │   ├── availability/
│   │   └── notification/
│   ├── app/                 # アプリケーション層（ユースケース）
│   └── infra/               # インフラ層（DB、外部サービス）
│       └── db/
│           └── db.go
├── go.mod
├── go.sum
├── Dockerfile
└── README.md
```

## 環境変数

| 変数名         | 説明                     | デフォルト値  |
| -------------- | ------------------------ | ------------- |
| `APP_ENV`      | 実行環境                 | `development` |
| `API_PORT`     | APIサーバのポート        | `8080`        |
| `DATABASE_URL` | PostgreSQL接続文字列     | (必須)        |

## ローカル開発

```bash
# 依存関係のダウンロード
go mod download

# サーバ起動
go run ./cmd/api
```

## API エンドポイント

| Method | Path      | Description    |
| ------ | --------- | -------------- |
| GET    | `/health` | ヘルスチェック |


# vrc-shift-scheduler

VRChat イベント／コンカフェ向け「シフト管理 + Discord 連携」サービス

## 概要

VRChat のコンカフェ／イベント向けに、Discord 連携でシフト希望の収集・確定・通知を行う小さな Web サービスです。

将来的にはマルチテナント SaaS（1 Discord サーバ = 1 テナント）として運用することを視野に入れています。

## 使用技術

| レイヤー        | 技術                      |
| --------------- | ------------------------- |
| Backend API     | Go 1.22+ (chi, pgx)       |
| Discord Bot     | Node.js / TypeScript (discord.js) |
| Database        | PostgreSQL 16             |
| Infrastructure  | Docker / Docker Compose   |

## ディレクトリ構成

```
vrc-shift-scheduler/
├── backend/          # Go製 API サーバ（DDD寄せ）
├── bot/              # Discord Bot (Node/TS)
├── web/              # API Sandbox（バニラHTML+JS）
├── docs/
│   └── domain/       # 業務知識・データモデル・ドメインモデル設計
├── scripts/          # 開発用スクリプト
├── docker-compose.yml
├── .env.example
└── README.md
```

## 開発環境のセットアップ

### 前提条件

- Docker & Docker Compose
- Go 1.22+（ローカル開発時）
- Node.js 22+ & pnpm（ローカル開発時）

### 1. 環境変数の設定

```bash
cp .env.example .env
# .env を編集して DISCORD_BOT_TOKEN などを設定
```

### 2. ブートストラップスクリプトの実行

```bash
./scripts/bootstrap-dev.sh
```

これにより以下が実行されます：
- `.env` が存在しない場合、`.env.example` からコピー
- `backend/go.sum` が存在しない場合、Docker で生成
- Docker イメージのビルド
- PostgreSQL コンテナの起動

### 3. Backend の起動

**Docker で起動する場合:**

```bash
docker compose up backend
```

**ローカルで起動する場合:**

```bash
cd backend
go run ./cmd/api
```

### 4. Bot の起動

**Docker で起動する場合:**

```bash
docker compose up bot
```

**ローカルで起動する場合:**

```bash
cd bot
pnpm install
pnpm dev
```

### 全サービスを一括起動

```bash
docker compose up
```

## ドキュメント

`docs/domain/` 配下に業務知識・データモデル・ドメインモデル設計を蓄積していきます。

- `00_project-overview/` - プロジェクト全体の概要
- `10_tenant-and-event/` - テナント・イベント関連のドメイン知識

各ディレクトリには以下の3ファイルを配置します：
- `業務知識.md` - ビジネスルールや業務フローの説明
- `データモデル.mdc` - テーブル設計やデータシナリオ
- `ドメインモデル設計.mdc` - エンティティ・VO・集約の設計

## API 動作確認

### API Sandbox（推奨）

ブラウザで簡単にAPIをテストできるツールを用意しています：

```bash
# Backendを起動後、ブラウザで以下を開く
# file:///path/to/vrc-shift-scheduler/web/index.html

# または VSCode Live Server / npx serve を使用
cd web
npx serve .
```

詳細は [`web/README.md`](./web/README.md) を参照してください。

### 実装済み API エンドポイント

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | ヘルスチェック |
| POST | `/api/v1/events` | Event 作成 |
| GET | `/api/v1/events` | Event 一覧取得 |
| GET | `/api/v1/events/:event_id` | Event 詳細取得 |
| POST | `/api/v1/events/:event_id/business-days` | 営業日作成 |
| GET | `/api/v1/events/:event_id/business-days` | 営業日一覧取得 |
| GET | `/api/v1/business-days/:business_day_id` | 営業日詳細取得 |

**認証**: 現在は簡易ヘッダー認証（`X-Tenant-ID`, `X-Member-ID`）を使用

## ライセンス

Private


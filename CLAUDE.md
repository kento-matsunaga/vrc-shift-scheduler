# CLAUDE.md - プロジェクト固有の指示

このファイルは Claude Code がプロジェクトを理解するためのコンテキストを提供します。

## プロジェクト概要

VRChat イベント向けのシフト管理システム（マルチテナント SaaS）

## 技術スタック

- **Backend**: Go 1.21+, chi router, PostgreSQL, pgx
- **Frontend**: React + TypeScript + Vite + Tailwind CSS
- **認証**: JWT（admins テーブル）
- **ID生成**: ULID (`common.NewULID()`)
- **アーキテクチャ**: DDD + Clean Architecture

## ディレクトリ構成

```
backend/
├── cmd/server/          # エントリーポイント
├── internal/
│   ├── domain/          # ドメイン層（エンティティ、リポジトリインターフェース）
│   ├── application/     # アプリケーション層（ユースケース）
│   ├── infra/           # インフラ層（DB実装、セキュリティ）
│   │   ├── db/          # リポジトリ実装、マイグレーション
│   │   └── security/    # JWT, bcrypt
│   └── interface/       # インターフェース層
│       └── rest/        # HTTPハンドラー、ルーター
web-frontend/
├── src/
│   ├── components/      # UIコンポーネント
│   ├── pages/           # ページコンポーネント
│   └── lib/api/         # API クライアント
```

## 命名規則

- ファイル: `snake_case.go`, `PascalCase.tsx`
- 変数/関数: `camelCase`
- 型/構造体: `PascalCase`
- DB カラム: `snake_case`
- マイグレーション: `YYYYMMDDHHMMSS_description.up.sql`

## API レスポンス形式

```json
// 成功
{"data": {...}}

// エラー
{"error": {"code": "ERR_XXX", "message": "..."}}
```

## 重要な規則

1. **リポジトリパターン厳守**: Usecase から DB 直接アクセス禁止
2. **テナント分離**: 全 API で `tenant_id` スコープ必須
3. **ソフトデリート**: `deleted_at` カラム使用
4. **トランザクション**: 複数テーブル更新時は必須

## 現在のタスク

マネタイズ基盤の実装（詳細は `docs/monetization-impl-prompt.md` 参照）

### フェーズ

1. **Phase 1: 基盤** - DB マイグレーション、認可ガード
2. **Phase 2: 決済連携** - BOOTH Claim, Stripe Webhook
3. **Phase 3: 管理機能** - 管理 API, 管理 UI
4. **Phase 4: 運用** - バッチ処理、手順書

### 主要な追加テーブル

- `plans` - プラン定義
- `entitlements` - テナントの権利
- `subscriptions` - Stripe サブスク情報
- `license_keys` - BOOTH キー（ハッシュ保存）
- `webhook_events` - 冪等性用
- `audit_logs` - 監査ログ

### 状態管理

- `tenants.status`: `active` / `grace` / `suspended`
- `entitlements.revoked_at`: 即時停止フラグ
- `entitlements.plan_code`: `LIFETIME` / `SUB_980`

## 開発コマンド

```bash
# Backend
cd backend
go build -o server ./cmd/server/
DATABASE_URL="postgres://..." JWT_SECRET="..." ./server

# Frontend
cd web-frontend
npm run dev

# Docker
docker compose up -d
```

## テスト用認証

```bash
# ログイン
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin1@example.com", "password": "password123"}'
```

# CLAUDE.md - プロジェクト固有の指示

このファイルは Claude Code がプロジェクトを理解するためのコンテキストを提供します。

## プロジェクト概要

VRChat イベント向けのシフト管理システム（マルチテナント SaaS）

## 技術スタック

- **Backend**: Go 1.24+, chi router, PostgreSQL 16, pgx
- **Frontend**: React 19 + TypeScript 5.9 + Vite 7 + Tailwind CSS 4
- **認証**: JWT（admins テーブル）
- **ID生成**: ULID (`common.NewULID()`)
- **アーキテクチャ**: DDD + Clean Architecture

## ディレクトリ構成

```
backend/
├── cmd/
│   ├── server/       # メインサーバー
│   ├── api/          # API サーバー（別エントリーポイント）
│   ├── migrate/      # マイグレーションツール
│   ├── seed/         # シードデータ投入
│   └── batch/        # バッチ処理（grace-expiry, webhook-cleanup）
├── internal/
│   ├── domain/       # ドメイン層（エンティティ、リポジトリIF）
│   ├── app/          # アプリケーション層（ユースケース）
│   ├── infra/        # インフラ層（DB実装、セキュリティ）
│   │   ├── db/       # リポジトリ実装、マイグレーション
│   │   └── security/ # JWT, bcrypt
│   └── interface/
│       └── rest/     # HTTPハンドラー、ルーター
web-frontend/         # テナント向けフロントエンド
admin-frontend/       # 運営管理コンソール
```

## ドメイン一覧（16領域）

| ドメイン | 説明 |
|---------|------|
| `announcement` | お知らせ機能 |
| `attendance` | 出欠確認 |
| `auth` | 認証・認可 |
| `availability` | 空き状況 |
| `billing` | 課金・ライセンス管理 |
| `common` | 共通（ULID、エラー） |
| `event` | イベント・営業日 |
| `import` | CSVインポート |
| `member` | メンバー・グループ |
| `notification` | 通知 |
| `role` | 役職・ロールグループ |
| `schedule` | 日程調整 |
| `services` | ドメインサービス |
| `shift` | シフト枠・テンプレート・インスタンス |
| `tenant` | テナント管理 |
| `tutorial` | チュートリアル |

## フロントエンド構成

**web-frontend（テナント向け）**
- イベント管理、シフト管理、メンバー管理、出欠確認、日程調整

**admin-frontend（運営向け）**
- テナント管理、ライセンスキー発行、監査ログ、お知らせ管理

## 命名規則

- ファイル: `snake_case.go`, `PascalCase.tsx`
- 変数/関数: `camelCase`
- 型/構造体: `PascalCase`
- DB カラム: `snake_case`
- マイグレーション: `NNN_description.up.sql`（最新: 039_migrate_instance_data）

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

## 開発コマンド

```bash
# Docker起動（推奨）
docker compose up -d

# バックエンドテスト
cd backend && go test ./...

# フロントエンドテスト
cd web-frontend && npm test

# マイグレーション状態確認
docker exec vrc-shift-backend /app/migrate -action=status

# マイグレーション実行
docker exec vrc-shift-backend /app/migrate -action=up
```

## テスト用認証

```bash
# ログイン
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin1@example.com", "password": "password123"}'
```

## 本番環境

- **サーバー**: ConoHa VPS (163.44.103.76)
- **デプロイパス**: /opt/vrcshift
- **デプロイ方法**: tarball作成 → SCP → 展開 → `docker-compose.prod.yml` で起動
- **重要**: 必ず `docker-compose.prod.yml` を使用（開発用 `docker-compose.yml` 禁止）

詳細は `LOCAL_DEPLOY_GUIDE.md`、`docs/PRODUCTION_DEPLOYMENT.md` 参照

## 主要ドキュメント

| ドキュメント | 内容 |
|-------------|------|
| `docs/BEGINNER_GUIDE.md` | 初心者向けセットアップ |
| `docs/DEVELOPMENT.md` | 開発環境・テストアカウント |
| `docs/UBIQUITOUS_LANGUAGE.md` | ドメイン用語辞書（853行） |
| `docs/api-endpoints.md` | API一覧 |
| `docs/BILLING_OPERATIONS.md` | 課金運用手順 |
| `docs/BRANCH_STRATEGY.md` | ブランチ運用ガイド |

## 現在のバージョン

- **本番**: v1.7.1
- **マイグレーション**: 039（migrate_instance_data）

---

## Claude Code 設定

### 利用可能なコマンド

| コマンド | 説明 |
|---------|------|
| `/review` | コード品質・セキュリティレビュー |
| `/test` | バックエンドテスト実行 |
| `/migrate` | DBマイグレーション管理 |
| `/deploy` | 本番デプロイ手順 |

### エージェント

| エージェント | 用途 |
|-------------|------|
| `code-reviewer` | コード品質・セキュリティレビュー |
| `ddd-reviewer` | DDD/クリーンアーキテクチャ準拠確認 |
| `security-reviewer` | セキュリティ脆弱性分析 |
| `planner` | 機能実装計画の立案 |

### ルール（常時適用）

| ルール | 内容 |
|-------|------|
| `go-coding-style` | Goコーディング規約 |
| `ddd-patterns` | DDD/クリーンアーキテクチャルール |
| `security` | セキュリティルール |
| `testing` | テストルール |

### スキル

| スキル | 内容 |
|-------|------|
| `domain-knowledge` | ドメイン知識（エンティティ、制約） |
| `git-workflow` | Git/PRワークフロー |
| `billing-operations` | 課金運用手順 |
| `production-deploy` | 本番デプロイ手順 |
| `api-integration` | API統合パターン |
| `stripe-integration` | Stripe決済連携パターン |
| `error-handling` | エラーハンドリングパターン |
| `frontend-patterns` | React/TypeScript/Tailwind CSSパターン |
| `database-patterns` | PostgreSQL/pgxパターン |
| `incident-response` | インシデント対応手順 |

### Hooks（自動実行）

| トリガー | 動作 |
|---------|------|
| `PreToolUse` | force push警告、本番docker-compose警告 |
| `PostToolUse` | テスト失敗検出、PR URL表示、fmt.Printlnチェック |
| `Stop` | Goファイルのfmt.Println残留チェック |

詳細は `.claude/` ディレクトリを参照

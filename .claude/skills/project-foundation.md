---
description: プロジェクトの全体像、技術スタック、コーディング規則、ドメイン用語
---

# Project Foundation

VRC Shift Scheduler プロジェクトの基盤知識。全ての開発作業の前提となる情報。

---

## プロジェクト概要

VRChat イベント向けのシフト管理システム（マルチテナント SaaS）

- **対象**: VRChat内で活動する団体・店舗・イベント運営チーム
- **主要機能**: イベント管理、シフト管理、メンバー管理、出欠確認、日程調整

---

## 技術スタック

| レイヤー | 技術 |
|---------|------|
| Backend | Go 1.24+, chi router, PostgreSQL 16, pgx |
| Frontend | React 19 + TypeScript 5.9 + Vite 7 + Tailwind CSS 4 |
| 認証 | JWT（admins テーブル） |
| ID生成 | ULID (`common.NewULID()`) |
| アーキテクチャ | DDD + Clean Architecture |

---

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

---

## 命名規則

| 対象 | 規則 | 例 |
|------|------|-----|
| Goファイル | snake_case | `shift_slot.go` |
| TSXファイル | PascalCase | `ShiftSlot.tsx` |
| 変数/関数 | camelCase | `getShiftSlot` |
| 型/構造体 | PascalCase | `ShiftSlot` |
| DBカラム | snake_case | `shift_slot_id` |
| マイグレーション | `NNN_description.up.sql` | `039_migrate_instance_data.up.sql` |

---

## 重要なルール（必守）

1. **リポジトリパターン厳守**: Usecase から DB 直接アクセス禁止
2. **テナント分離**: 全 API で `tenant_id` スコープ必須
3. **ソフトデリート**: `deleted_at` カラム使用
4. **トランザクション**: 複数テーブル更新時は必須

---

## API レスポンス形式

```json
// 成功
{"data": {...}}

// エラー
{"error": {"code": "ERR_XXX", "message": "..."}}
```

---

## ドメイン一覧（16領域）

| ドメイン | 説明 |
|---------|------|
| `tenant` | テナント管理 |
| `event` | イベント・営業日 |
| `member` | メンバー・グループ |
| `role` | 役職・ロールグループ |
| `shift` | シフト枠・テンプレート・インスタンス |
| `attendance` | 出欠確認 |
| `schedule` | 日程調整 |
| `billing` | 課金・ライセンス管理 |
| `auth` | 認証・認可 |
| `notification` | 通知 |
| `announcement` | お知らせ機能 |
| `availability` | 空き状況 |
| `import` | CSVインポート |
| `services` | ドメインサービス |
| `tutorial` | チュートリアル |
| `common` | 共通（ULID、エラー） |

---

## ユビキタス言語（主要用語）

### コア概念

| 日本語 | 英語（コード） | 説明 |
|--------|---------------|------|
| テナント | `tenant` | 団体・店舗単位の最上位境界 |
| イベント | `event` | 営業・イベント単位 |
| 営業日 | `business_day` | 1回分の営業日 |
| メンバー | `member` | テナントに所属する人物 |
| ロール | `role` | メンバーの役割（キャスト、スタッフ等） |

### シフト関連

| 日本語 | 英語（コード） | 説明 |
|--------|---------------|------|
| シフト枠 | `shift_slot` | 時間帯×インスタンス×ポジションの1人分の席 |
| ポジション | `position` | 営業時の役割（カウンター、テーブル等） |
| インスタンス | `instance` | VRChatの部屋単位 |
| シフト割り当て | `shift_assignment` | 枠への人員配置 |
| シフト確定 | `shift_plan` | 最終配置計画 |

### 出欠・日程調整

| 日本語 | 英語（コード） | 説明 |
|--------|---------------|------|
| 出欠確認 | `attendance_collection` | 営業日への出欠収集 |
| 日程調整 | `date_schedule` | 候補日から開催日を決める |
| 公開トークン | `public_token` | 認証不要アクセス用UUID |

### 状態値

| 概念 | 値 | 説明 |
|------|-----|------|
| イベント種別 | `normal` / `special` | 通常営業 / 特別営業 |
| 管理者ロール | `owner` / `manager` | オーナー / マネージャー |
| シフト計画状態 | `draft` / `published` / `finalized` | 下書き / 公開 / 確定 |
| 出欠回答 | `attending` / `absent` / `maybe` | 出席 / 欠席 / 未定 |
| 日程可否 | `available` / `unavailable` / `maybe` | ○ / × / △ |

---

## 共通パターン

### ID生成
```go
id := common.NewULID()
```

### テナント境界
全てのデータは必ず1つのテナントに属し、テナント間でのデータ参照・変更は禁止。

### ソフトデリート
```sql
WHERE deleted_at IS NULL
```

### 深夜営業（日跨ぎ）
終了時刻 < 開始時刻 の場合、翌日扱い（例: 21:30〜25:00）

---

## 関連ドキュメント

- `docs/domain/UBIQUITOUS_LANGUAGE.md` - 完全な用語辞書（853行）
- `docs/domain/` - 各ドメインの詳細設計
- `docs/api/api-endpoints.md` - API仕様

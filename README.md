# VRC Shift Scheduler

VRChat コミュニティ向けシフト管理システム

## 🚀 Quick Start

> 🪟 **Windows 11 の方へ**：まずは **[docs/setup-windows.md](docs/setup-windows.md)** を実施してください（Windows Terminal / WSL2 Ubuntu / Docker Desktop の準備と起動方法）。

### Docker Compose で起動（推奨）

```bash
# プロジェクトをクローン
git clone git@github.com:kento-matsunaga/vrc-shift-scheduler.git
cd vrc-shift-scheduler

# 開発環境を起動（PostgreSQL + Backend + Frontend）
docker compose up -d --build

# マイグレーション
docker compose exec backend go run ./cmd/migrate/main.go

# シード（任意：テスト用データ投入）
docker compose exec backend go run ./cmd/seed/main.go
```

- バックエンド： http://localhost:8080/health
- フロントエンド： http://localhost:5173

### ローカル起動（Docker なし）

```bash
# ブートストラップスクリプトを実行
./scripts/bootstrap.sh

# バックエンド起動
cd backend && go run ./cmd/server/main.go

# フロントエンド起動（別ターミナル）
cd web-frontend && npm run dev
```

---

## 📖 ドキュメント

| ドキュメント | 説明 |
|-------------|------|
| [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) | **開発ガイド（テストアカウント・API情報）** |
| [docs/setup-windows.md](docs/setup-windows.md) | Windows 11 セットアップ（WSL2 + Docker Desktop） |
| [SETUP.md](SETUP.md) | 詳細なセットアップ手順（macOS / Linux） |
| [docs/ENVIRONMENT_VARIABLES.md](docs/ENVIRONMENT_VARIABLES.md) | 環境変数の説明 |
| [docs/DEPLOYMENT_SERVER_REQUIREMENTS.md](docs/DEPLOYMENT_SERVER_REQUIREMENTS.md) | **サーバー選定・デプロイメント要件まとめ** |

---

## 🛠️ 技術スタック

### バックエンド

- **Go 1.23+**
- **go-chi/chi v5** - HTTP ルーター
- **pgx v5** - PostgreSQL ドライバー
- **PostgreSQL 16**

### フロントエンド

- **React 19** + **TypeScript**
- **Vite 7**
- **Tailwind CSS 4**
- **React Router**

---

## 📁 プロジェクト構成

```
vrc-shift-scheduler/
├── backend/
│   ├── cmd/
│   │   ├── server/       # HTTP サーバー
│   │   ├── migrate/      # DB マイグレーション
│   │   └── seed/         # データシード
│   ├── internal/
│   │   ├── domain/       # ドメインモデル
│   │   ├── app/          # アプリケーションサービス
│   │   ├── infra/        # インフラ層（DB リポジトリ）
│   │   └── interface/    # REST API ハンドラー
│   └── Makefile          # 開発用コマンド
├── web-frontend/
│   └── src/
│       ├── components/   # React コンポーネント
│       ├── pages/        # ページコンポーネント
│       └── lib/          # API クライアント
├── bot/                  # Discord Bot（オプション）
├── docs/                 # ドキュメント
└── docker-compose.yml    # 開発環境定義
```

---

## 🧪 テスト

```bash
# バックエンドテスト（Docker内）
docker compose exec backend go test ./...

# または Makefile を使用
docker compose exec backend make test
```

---

## 📝 開発ワークフロー

### ブランチ運用

| ブランチ | 用途 |
|----------|------|
| `main` | 本番用。直接 push 禁止。PR 経由でマージ |
| `feature/xxx` | 新機能開発用 |
| `fix/xxx` | バグ修正用 |

### 開発フロー

1. `main` から作業ブランチを作成
2. コードを実装・テスト
3. コミット & プッシュ
4. Pull Request を作成
5. レビュー後、マージ

---

## 🤝 コントリビューション

プロジェクトへの貢献を歓迎します！

1. このリポジトリを Fork（または招待を受ける）
2. Feature ブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'feat: 変更内容'`)
4. ブランチを Push (`git push origin feature/amazing-feature`)
5. Pull Request を作成

---

## 📧 お問い合わせ

- **Issue Tracker**: [GitHub Issues](https://github.com/kento-matsunaga/vrc-shift-scheduler/issues)

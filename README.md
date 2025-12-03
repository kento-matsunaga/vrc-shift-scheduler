# VRC Shift Scheduler

VRChat コミュニティ向けシフト管理システム

## 🚀 Quick Start

### ブートストラップ（初回セットアップ）

```bash
# プロジェクトをクローン
git clone <repository-url>
cd vrc-shift-scheduler

# ブートストラップスクリプトを実行
./scripts/bootstrap.sh
```

このスクリプトは以下を自動的に実行します：
- Go 1.23+ のバージョンチェック・インストール
- Node.js 18+ のバージョンチェック
- PostgreSQL のチェック
- 環境変数ファイル（.env）の作成
- 依存関係のインストール

### データベースのセットアップ

```bash
cd backend
go run ./cmd/migrate/main.go
```

### 開発サーバーの起動

**バックエンド:**

```bash
cd backend
go run ./cmd/server/main.go
# http://localhost:8080
```

**フロントエンド:**

```bash
cd web-frontend
npm run dev
# http://localhost:5173
```

## 📖 ドキュメント

- **[SETUP.md](SETUP.md)** - 詳細なセットアップ手順
- **[backend/TASKS_PUBLIC_ALPHA_RELEASE.md](backend/TASKS_PUBLIC_ALPHA_RELEASE.md)** - Public Alpha リリースタスク
- **[backend/docs/ARCHITECTURE.md](backend/docs/ARCHITECTURE.md)** - システムアーキテクチャ
- **[backend/docs/API.md](backend/docs/API.md)** - API ドキュメント

## 🛠️ 技術スタック

### バックエンド
- **Go 1.23+**
- **go-chi/chi v5** - HTTP ルーター
- **pgx v5** - PostgreSQL ドライバー
- **PostgreSQL 14+**

### フロントエンド
- **React 18**
- **TypeScript**
- **Vite**
- **Tailwind CSS**
- **React Router**
- **Axios**

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
│   └── migrations/       # SQL マイグレーションファイル
├── web-frontend/
│   ├── src/
│   │   ├── components/   # React コンポーネント
│   │   ├── pages/        # ページコンポーネント
│   │   ├── lib/          # API クライアント
│   │   └── types/        # TypeScript 型定義
│   └── public/           # 静的ファイル
└── scripts/
    ├── bootstrap.sh      # 初回セットアップスクリプト
    ├── install-go.sh     # Go インストール（sudo 版）
    └── install-go-local.sh # Go インストール（ローカル版）
```

## 🧪 テスト

```bash
# バックエンドテスト
cd backend
go test ./...

# 統合テスト（DB が必要）
go test -tags=integration ./internal/infra/db/...

# フロントエンドテスト
cd web-frontend
npm test
```

## 🐳 Docker（開発環境）

PostgreSQL を Docker で起動：

```bash
docker run --name vrc-shift-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=vrc_shift_scheduler \
  -p 5432:5432 \
  -d postgres:14
```

## 📝 開発ワークフロー

1. **Issue を作成** - 実装する機能やバグ修正の Issue を作成
2. **ブランチを作成** - `feature/xxx` または `fix/xxx` ブランチを作成
3. **実装 & テスト** - コードを実装し、テストを追加
4. **コミット** - 意味のあるコミットメッセージで commit
5. **PR を作成** - main ブランチへの Pull Request を作成
6. **レビュー & マージ** - コードレビュー後、マージ

## 🤝 コントリビューション

プロジェクトへの貢献を歓迎します！

1. このリポジトリを Fork
2. Feature ブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add some amazing feature'`)
4. ブランチを Push (`git push origin feature/amazing-feature`)
5. Pull Request を作成

## 📄 ライセンス

[MIT License](LICENSE)

## 📧 お問い合わせ

- **Issue Tracker**: [GitHub Issues](https://github.com/your-org/vrc-shift-scheduler/issues)
- **Discord**: [招待リンク]

---

**Note**: このプロジェクトは Public Alpha テスト準備中です。詳細は [TASKS_PUBLIC_ALPHA_RELEASE.md](backend/TASKS_PUBLIC_ALPHA_RELEASE.md) を参照してください。

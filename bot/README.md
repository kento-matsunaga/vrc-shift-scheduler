# Discord Bot

VRC Shift Scheduler の Discord Bot です。

## 技術スタック

- Node.js 22+
- TypeScript 5.x
- discord.js 14.x

## ディレクトリ構成

```
bot/
├── src/
│   └── index.ts      # エントリポイント
├── package.json
├── tsconfig.json
├── Dockerfile
└── README.md
```

## 環境変数

| 変数名              | 説明                        | 必須 |
| ------------------- | --------------------------- | ---- |
| `DISCORD_BOT_TOKEN` | Discord Bot トークン        | ✅   |
| `DISCORD_APP_ID`    | Discord Application ID      | ✅   |
| `DISCORD_GUILD_ID`  | 開発用 Guild ID（省略可）   | ❌   |
| `BACKEND_BASE_URL`  | Backend API の URL          | ❌   |

## ローカル開発

```bash
# 依存関係のインストール
pnpm install

# 開発サーバ起動（ホットリロード）
pnpm dev

# ビルド
pnpm build

# 本番起動
pnpm start
```

## スラッシュコマンド

| コマンド | 説明           |
| -------- | -------------- |
| `/ping`  | Pong! を返す   |

## Discord Developer Portal での設定

1. [Discord Developer Portal](https://discord.com/developers/applications) でアプリケーションを作成
2. Bot を有効化し、トークンを取得
3. OAuth2 > URL Generator で `bot` と `applications.commands` スコープを選択
4. 生成された URL で Bot をサーバに招待


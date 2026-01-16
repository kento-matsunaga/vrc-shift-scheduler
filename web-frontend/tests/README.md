# API統合テスト

VRC Shift SchedulerのAPI統合テストドキュメント。

## 概要

Playwrightを使用したAPI統合テストスイートです。全APIエンドポイントの正常系・異常系をテストします。

## テスト統計

- **総テスト数**: 461+
- **テストファイル数**: 21
- **カバレッジ**: 全APIエンドポイント

## 環境設定

### 前提条件

- Node.js 18+
- Docker & Docker Compose
- バックエンドAPIが起動していること

### セットアップ

```bash
# 依存関係のインストール
cd web-frontend
npm install

# Playwrightのインストール
npx playwright install
```

### 環境変数

| 変数名 | デフォルト値 | 説明 |
|--------|-------------|------|
| `API_BASE_URL` | `http://backend:8080` | APIのベースURL |

## テスト実行

### Docker環境（推奨）

```bash
# 全テスト実行
docker compose exec web-frontend npm run test:api

# 特定ファイルのみ
docker compose exec web-frontend npx playwright test tests/api/auth.spec.ts

# ワーカー数指定（デフォルト: 1）
docker compose exec web-frontend npx playwright test tests/api --workers=1
```

### ローカル環境

```bash
# 環境変数を設定してから実行
API_BASE_URL=http://localhost:8080 npx playwright test tests/api
```

### テスト結果の確認

```bash
# HTMLレポート表示
npx playwright show-report
```

## テストファイル構成

```
tests/
├── api/                          # API統合テスト
│   ├── auth.spec.ts              # 認証API
│   ├── admin.spec.ts             # 管理者API
│   ├── tenant.spec.ts            # テナントAPI
│   ├── invitation.spec.ts        # 招待API
│   ├── member.spec.ts            # メンバーAPI
│   ├── role.spec.ts              # ロールAPI
│   ├── event.spec.ts             # イベントAPI
│   ├── business-day.spec.ts      # 営業日API
│   ├── shift-slot.spec.ts        # シフト枠API
│   ├── shift-assignment.spec.ts  # シフト割当API
│   ├── template.spec.ts          # テンプレートAPI
│   ├── instance.spec.ts          # インスタンスAPI
│   ├── member-group.spec.ts      # メンバーグループAPI
│   ├── role-group.spec.ts        # ロールグループAPI
│   ├── attendance.spec.ts        # 出欠収集API
│   ├── schedule.spec.ts          # スケジュールAPI
│   ├── import.spec.ts            # インポートAPI
│   ├── announcement.spec.ts      # お知らせAPI
│   ├── tutorial.spec.ts          # チュートリアルAPI
│   ├── public.spec.ts            # 公開API
│   └── actual-attendance.spec.ts # 実績出欠API
├── e2e/                          # E2Eテスト
│   └── login.spec.ts             # ログインE2E
└── utils/                        # ユーティリティ
    ├── api-client.ts             # APIクライアント
    ├── auth.ts                   # 認証ヘルパー
    └── index.ts                  # エクスポート
```

## テストパターン

### 正常系テスト

- 有効な認証での正常リクエスト
- レスポンス形式の検証
- CRUDフローの確認

### 異常系テスト

| パターン | 期待ステータス |
|----------|----------------|
| 認証なし | 401 |
| 無効なトークン | 401 |
| 必須パラメータなし | 400 |
| 無効なパラメータ形式 | 400 |
| 存在しないリソース | 404 |
| SQLインジェクション試行 | 400+ |

## APIクライアントの使用方法

### 基本的な使い方

```typescript
import { loginAsAdmin, getUnauthenticatedClient } from '../utils/auth';
import { ENDPOINTS, ApiClient } from '../utils/api-client';

// 認証済みクライアント
const { client, loginData } = await loginAsAdmin(request);

// 認証なしクライアント
const unauthClient = getUnauthenticatedClient(request);

// カスタムトークン
const customClient = new ApiClient(request);
customClient.setToken('your-token');
```

### リクエストメソッド

```typescript
// GET（自動エラーハンドリング）
const data = await client.get<ResponseType>('/api/v1/endpoint');

// POST（自動エラーハンドリング）
const result = await client.post<ResponseType>('/api/v1/endpoint', { data });

// Raw（ステータスコード確認用）
const response = await client.raw('GET', '/api/v1/endpoint');
expect(response.status()).toBe(200);
```

### エラーハンドリング

```typescript
import { ApiClientError } from '../utils/api-client';

try {
  await client.get('/api/v1/endpoint');
} catch (error) {
  if (error instanceof ApiClientError) {
    console.log(error.status);    // HTTPステータスコード
    console.log(error.response);  // エラーレスポンス
  }
}
```

## テスト認証情報

```typescript
// テスト用の認証情報（シードデータ）
const TEST_CREDENTIALS = {
  email: 'admin1@example.com',
  password: 'password123',
};
```

## ベストプラクティス

### 1. テストの独立性

各テストは独立して実行可能であるべきです。前のテストの結果に依存しないでください。

### 2. クリーンアップ

データを変更するテストは、必ず`try-finally`でクリーンアップを行ってください。

```typescript
test('パスワード変更テスト', async ({ request }) => {
  let passwordChanged = false;
  try {
    // パスワード変更
    passwordChanged = true;
    // テスト...
  } finally {
    if (passwordChanged) {
      // パスワードを元に戻す
    }
  }
});
```

### 3. シリアル実行

状態を共有するテストは`test.describe.configure({ mode: 'serial' })`を使用してください。

### 4. ステータスコードの許容

APIの実装によって返るステータスコードが異なる場合があります：

```typescript
// 複数のステータスコードを許容
expect([200, 404]).toContain(response.status());

// 400以上を許容
expect(response.status()).toBeGreaterThanOrEqual(400);
```

## トラブルシューティング

### テストがタイムアウトする

```bash
# タイムアウトを延長
npx playwright test --timeout=60000
```

### 認証エラーが発生する

1. バックエンドAPIが起動しているか確認
2. シードデータが投入されているか確認
3. `API_BASE_URL`が正しいか確認

### 並列実行で失敗する

状態を共有するテストは`--workers=1`で実行してください。

```bash
npx playwright test tests/api --workers=1
```

## 関連ドキュメント

- [Playwright公式ドキュメント](https://playwright.dev/)
- [APIエンドポイント一覧](../../docs/api-endpoints.md)

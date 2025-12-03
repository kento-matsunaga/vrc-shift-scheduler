# VRC Shift Scheduler API Sandbox

バニラHTML+JSで作った、REST API動作確認用のミニフロントエンドです。

## 🚀 使い方

### 1. Backend を起動

```bash
cd backend
DATABASE_URL="postgres://vrcshift:vrcshift@localhost:5432/vrcshift?sslmode=disable" \
PORT=8082 \
./bin/server

# または docker-compose で起動している場合はそちらを使用
```

### 2. ブラウザで `index.html` を開く

#### 方法A: VSCode Live Server を使う
1. VSCode で `web/index.html` を開く
2. 右クリック → "Open with Live Server"

#### 方法B: npx serve を使う
```bash
cd web
npx serve .
# http://localhost:3000 で開く
```

#### 方法C: 直接開く
```bash
# Linuxの場合
xdg-open web/index.html

# macOSの場合
open web/index.html

# Windowsの場合
start web/index.html
```

### 3. API を叩く

1. **接続設定**
   - `API Base URL`: `http://localhost:8082`（デフォルト）
   - `Tenant ID`: データベースに存在する有効なテナントID
   
2. **クイックアクション**
   - 「GET /health」ボタンで疎通確認
   - 「GET /api/v1/events」でEvent一覧取得
   - 「POST /api/v1/events」でEvent作成（サンプルJSON自動入力）

3. **手動でリクエスト作成**
   - Method と Path を入力
   - 必要に応じて Request Body を JSON で記述
   - 「送信」ボタンでリクエスト実行

## 📋 実装済みエンドポイント

| Method | Path | 説明 |
|--------|------|------|
| GET | `/health` | ヘルスチェック |
| POST | `/api/v1/events` | Event 作成 |
| GET | `/api/v1/events` | Event 一覧取得 |
| GET | `/api/v1/events/:event_id` | Event 詳細取得 |
| POST | `/api/v1/events/:event_id/business-days` | 営業日作成 |
| GET | `/api/v1/events/:event_id/business-days` | 営業日一覧取得 |
| GET | `/api/v1/business-days/:business_day_id` | 営業日詳細取得 |

## 💡 使い方のコツ

### Event 作成からBusinessDay作成までの流れ

1. **Event 作成**
   ```
   POST /api/v1/events
   Body: {
     "event_name": "VRC定期イベント",
     "event_type": "normal",
     "description": "毎週開催"
   }
   ```

2. **レスポンスから `event_id` をコピー**
   ```json
   {
     "data": {
       "event_id": "01KBHN7M8QT9HS5XSJ6QR30WRQ",
       ...
     }
   }
   ```

3. **営業日作成**
   ```
   POST /api/v1/events/01KBHN7M8QT9HS5XSJ6QR30WRQ/business-days
   Body: {
     "target_date": "2025-12-10",
     "start_time": "20:00",
     "end_time": "23:00",
     "occurrence_type": "special"
   }
   ```

4. **営業日一覧取得**
   ```
   GET /api/v1/events/01KBHN7M8QT9HS5XSJ6QR30WRQ/business-days
   ```

### 日付範囲フィルタ

営業日一覧取得時にクエリパラメータで日付範囲を指定できます：

```
GET /api/v1/events/:event_id/business-days?start_date=2025-12-01&end_date=2025-12-31
```

## 🔧 認証

現在は簡易ヘッダー認証を使用しています：

- `X-Tenant-ID`: テナントID（必須）
- `X-Member-ID`: メンバーID（一部エンドポイントで必要）

サンドボックスの「接続設定」セクションで設定できます。

## 🎨 カスタマイズ

### プリセットの追加

`main.js` の末尾にボタンイベントリスナーを追加：

```javascript
const presetMyCustom = document.getElementById("presetMyCustom");
presetMyCustom.addEventListener("click", () => {
  methodSelect.value = "GET";
  pathInput.value = "/api/v1/my-custom-endpoint";
  bodyInput.value = "";
});
```

`index.html` にボタンを追加：

```html
<button id="presetMyCustom" class="btn-secondary">
  My Custom Action
</button>
```

## 📝 次のステップ

このサンドボックスで基本的なAPI動作を確認したら：

1. 必要な機能を洗い出す
2. React/Next.js/Svelte などで本格的なフロントエンドを構築
3. カレンダービュー、ドラッグ&ドロップなどのリッチなUIを実装

## 🐛 トラブルシューティング

### CORS エラーが出る

Backend の CORS 設定を確認してください。現在の実装では `Access-Control-Allow-Origin: *` が設定されています。

### Tenant ID が不正

データベースに存在する有効なテナントIDを使用してください：

```bash
docker exec vrc-shift-scheduler-db-1 psql -U vrcshift -d vrcshift -c "SELECT tenant_id FROM tenants;"
```

### API が 404 を返す

Backend が正しいポート（8082）で起動しているか確認してください。


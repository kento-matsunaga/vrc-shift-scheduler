# 出欠確認・日程調整 機能ガイド

## ✅ 実装完了

キャストに出欠や日程の回答を求める機能を実装しました！

---

## 🎯 使い方

### 1. 管理者側：出欠確認を作成

1. ログイン後、ナビゲーションバーの **「出欠確認」** をクリック
   ```
   http://localhost:5173/attendance
   ```

2. フォームに以下を入力:
   - **タイトル**: 例「12月のシフト出欠確認」
   - **説明**: 詳細や注意事項（任意）
   - **回答締切**: 締切日時（任意）

3. **「出欠確認を作成」** ボタンをクリック

4. 公開URLが表示されます:
   ```
   http://localhost:5173/p/attendance/{トークン}
   ```

5. **「URLをコピー」** でURLをコピーし、メンバーに送信

### 2. メンバー側：出欠を回答

1. 管理者から送られたURLにアクセス

2. フォームに以下を入力:
   - **お名前**: ドロップダウンから選択
   - **出欠**: 「参加」または「不参加」を選択
   - **備考**: 補足情報（任意）

3. **「回答を送信」** ボタンをクリック

4. 送信完了！

---

## 📅 使い方（日程調整）

### 1. 管理者側：日程調整を作成

1. ログイン後、ナビゲーションバーの **「日程調整」** をクリック
   ```
   http://localhost:5173/schedules
   ```

2. フォームに以下を入力:
   - **タイトル**: 例「忘年会の日程調整」
   - **説明**: 詳細や注意事項（任意）
   - **候補日**: 複数の候補日時を入力
     - 「+ 候補日を追加」で候補を増やせます
   - **回答締切**: 締切日時（任意）

3. **「日程調整を作成」** ボタンをクリック

4. 公開URLが表示されます:
   ```
   http://localhost:5173/p/schedule/{トークン}
   ```

5. **「URLをコピー」** でURLをコピーし、メンバーに送信

### 2. メンバー側：参加可能な日程を回答

1. 管理者から送られたURLにアクセス

2. フォームに以下を入力:
   - **お名前**: ドロップダウンから選択
   - **参加可能な日程**: 候補日から複数選択可能
   - **備考**: 補足情報（任意）

3. **「回答を送信」** ボタンをクリック

4. 送信完了！

---

## 🔗 主要URL

| 機能 | 管理画面 | 公開ページ |
|------|---------|-----------|
| 出欠確認 | http://localhost:5173/attendance | http://localhost:5173/p/attendance/{token} |
| 日程調整 | http://localhost:5173/schedules | http://localhost:5173/p/schedule/{token} |

---

## 🧪 テスト手順

### 出欠確認のテスト

1. **ログイン**:
   ```
   Email: admin@test.com
   Password: password123
   ```

2. **出欠確認を作成**:
   - 「出欠確認」→ タイトル「テスト出欠確認」を入力 → 作成
   - 公開URLをコピー

3. **新しいブラウザ（シークレットモード）**:
   - コピーしたURLにアクセス
   - メンバーを選択（例: 田中太郎）
   - 出欠を選択（例: 参加）
   - 回答を送信

4. **確認**: 送信完了画面が表示される

### 日程調整のテスト

1. **ログイン**:
   ```
   Email: admin@test.com
   Password: password123
   ```

2. **日程調整を作成**:
   - 「日程調整」→ タイトル「テスト日程調整」を入力
   - 候補日を3つ入力（例: 明日、明後日、3日後）
   - 作成
   - 公開URLをコピー

3. **新しいブラウザ（シークレットモード）**:
   - コピーしたURLにアクセス
   - メンバーを選択（例: 佐藤花子）
   - 参加可能な日程を選択（複数選択可）
   - 回答を送信

4. **確認**: 送信完了画面が表示される

---

## 📊 実装内容

### 新規作成ファイル

1. **API Client層**:
   - `src/lib/api/attendanceApi.ts` - 出欠確認API
   - `src/lib/api/scheduleApi.ts` - 日程調整API

2. **管理画面**:
   - `src/pages/AttendanceList.tsx` - 出欠確認作成画面
   - `src/pages/ScheduleList.tsx` - 日程調整作成画面

3. **ルーティング**:
   - `src/App.tsx` - `/attendance`, `/schedules` ルート追加
   - `src/components/Layout.tsx` - ナビゲーションリンク追加

### 既存ファイル（実装済み）

- `src/pages/public/AttendanceResponse.tsx` - 出欠回答ページ
- `src/pages/public/ScheduleResponse.tsx` - 日程回答ページ

---

## 🎨 画面仕様

### 出欠確認作成画面

- シンプルなフォーム形式
- タイトル、説明、締切を入力
- 作成後に公開URLを表示
- URLコピー機能
- プレビュー機能

### 日程調整作成画面

- 複数の候補日を入力可能
- 候補日の追加・削除が可能
- 作成後に公開URLを表示
- URLコピー機能
- プレビュー機能

### デザイン

- クリーンでシンプルなUI
- Tailwind CSS使用
- レスポンシブデザイン
- 成功時のグリーンハイライト
- エラー表示

---

## 📝 バックエンドAPI

### 出欠確認API（管理用）

```bash
# 作成
POST /api/v1/attendance/collections
Authorization: Bearer {JWT}
{
  "title": "12月のシフト出欠確認",
  "description": "詳細",
  "target_type": "event",
  "deadline": "2025-12-20T23:59:59Z"
}

# 取得
GET /api/v1/attendance/collections/{collection_id}
Authorization: Bearer {JWT}

# 締切
POST /api/v1/attendance/collections/{collection_id}/close
Authorization: Bearer {JWT}

# 回答一覧
GET /api/v1/attendance/collections/{collection_id}/responses
Authorization: Bearer {JWT}
```

### 出欠確認API（公開）

```bash
# 出欠確認情報取得
GET /api/v1/public/attendance/{token}

# 回答送信
POST /api/v1/public/attendance/{token}/responses
{
  "member_id": "01XXX...",
  "response": "attending",
  "note": "備考"
}
```

### 日程調整API（管理用）

```bash
# 作成
POST /api/v1/schedules
Authorization: Bearer {JWT}
{
  "title": "忘年会の日程調整",
  "description": "詳細",
  "candidate_dates": [
    "2025-12-20T19:00:00Z",
    "2025-12-21T19:00:00Z",
    "2025-12-22T19:00:00Z"
  ],
  "deadline": "2025-12-15T23:59:59Z"
}

# 取得
GET /api/v1/schedules/{schedule_id}
Authorization: Bearer {JWT}

# 日程決定
POST /api/v1/schedules/{schedule_id}/decide
Authorization: Bearer {JWT}
{
  "decided_date": "2025-12-20T19:00:00Z"
}

# 締切
POST /api/v1/schedules/{schedule_id}/close
Authorization: Bearer {JWT}

# 回答一覧
GET /api/v1/schedules/{schedule_id}/responses
Authorization: Bearer {JWT}
```

### 日程調整API（公開）

```bash
# 日程調整情報取得
GET /api/v1/public/schedules/{token}

# 回答送信
POST /api/v1/public/schedules/{token}/responses
{
  "member_id": "01XXX...",
  "available_dates": [
    "2025-12-20T19:00:00Z",
    "2025-12-21T19:00:00Z"
  ],
  "note": "備考"
}
```

---

## 🚀 導線フロー

```
管理者
  ↓
ログイン → ナビゲーション「出欠確認」or「日程調整」
  ↓
作成フォーム入力 → 作成ボタン
  ↓
公開URL発行 → URLをコピー
  ↓
メンバーにURL送信（Discord、メールなど）

メンバー
  ↓
公開URLにアクセス
  ↓
名前選択 → 回答入力 → 送信
  ↓
送信完了
```

---

**実装完了！すぐにテストできます 🎉**

ブラウザで http://localhost:5173 にアクセスして、ナビゲーションバーの「出欠確認」または「日程調整」をクリックしてください。

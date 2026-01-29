---
description: テストシナリオ、テストコマンド、動作確認チェックリスト
---

# Testing

VRC Shift Scheduler のテスト実施ガイド。

---

## テストコマンド

### バックエンド

```bash
cd backend

# 全テスト実行
JWT_SECRET=test_secret_key go test ./...

# 統合テスト（DB必要）
go test -tags=integration ./internal/infra/db/...

# 特定パッケージのテスト
go test ./internal/domain/shift/...

# カバレッジ付き
go test -cover ./...

# 詳細出力
go test -v ./...
```

### フロントエンド

```bash
cd web-frontend

# テスト実行
npm test

# Lint
npm run lint

# 型チェック
npm run type-check
```

### E2Eテスト（Playwright）

```bash
cd web-frontend

# E2Eテスト実行
npx playwright test

# UIモードで実行
npx playwright test --ui

# 特定のテストファイル
npx playwright test tests/auth.spec.ts
```

---

## テストシナリオ

### シナリオA: 基本フロー

1. **ログイン**
   - admin1@example.com / password123 でログイン
   - ダッシュボードに遷移することを確認

2. **イベント作成**
   - イベント名「テストイベント」を入力
   - 一覧に表示されることを確認

3. **営業日追加**
   - 日付、時刻を入力
   - 営業日一覧に追加されることを確認

4. **シフト枠作成**
   - ポジション選択、必要人数「2」を入力
   - 割り当て状況が「0/2」と表示されることを確認

5. **メンバー登録**
   - 表示名を入力
   - メンバー一覧に登録されることを確認

6. **シフト割り当て**
   - メンバーをシフト枠に割り当て
   - 割り当て状況が「1/2」に更新されることを確認

---

### シナリオB: 満員チェック

1. シフト枠（必要人数2）に2人割り当て
2. 3人目を割り当てようとする
3. **期待**: 「このシフト枠は既に満員です」エラー

---

### シナリオC: 深夜営業（日またぎ）

1. 営業日を作成（21:00-02:00）
2. シフト枠を作成（23:00-01:00）
3. **確認**: 時刻が正しく表示される
4. **確認**: `is_overnight` フラグが `true`

---

### シナリオD: マルチテナント境界

1. テナントAでイベント・メンバーを作成
2. テナントBでログイン
3. **確認**: テナントAのデータが表示されない
4. **API確認**: テナントBからテナントAのリソースにアクセス → `403` or `404`

---

### シナリオE: 出欠確認

1. 出欠確認を作成
2. 公開URLをコピー
3. シークレットウィンドウで公開URLにアクセス
4. **確認**: 認証なしで回答画面が表示される
5. 回答を送信
6. 管理画面で回答が表示されることを確認

---

### シナリオF: 日程調整

1. 日程調整を作成（複数候補日）
2. 公開URLで回答
3. 管理画面で集計結果を確認

---

## ロール機能テスト

### 手順

1. **ロール作成**
   - ロール名、説明、色、表示順序を入力
   - 一覧に表示されることを確認

2. **メンバーにロール割り当て**
   - メンバー編集でロールにチェック
   - 複数ロール選択可能

3. **ロールフィルター**
   - 「ロールでフィルター」ドロップダウンで選択
   - 該当ロールを持つメンバーのみ表示

4. **ロール削除**
   - ロールを削除
   - メンバーからもロールバッジが消えることを確認

---

## バリデーションテスト

| 入力 | 期待結果 |
|-----|---------|
| イベント名空白 | 「イベント名は必須です」エラー |
| 時刻「25:00」 | 「時刻の形式が不正です」エラー |
| 開始時刻 > 終了時刻 | 「開始時刻は終了時刻より前」エラー |
| 必要人数「0」 | 「1以上である必要があります」エラー |

---

## データベース確認コマンド

```bash
# PostgreSQLコンテナに接続
docker exec -it vrc-shift-scheduler-db-1 psql -U vrcshift -d vrcshift

# 管理者一覧
SELECT admin_id, email, display_name, role FROM admins WHERE deleted_at IS NULL;

# イベント一覧
SELECT event_id, event_name, event_type FROM events WHERE deleted_at IS NULL;

# シフト割り当て確認
SELECT sa.assignment_id, m.display_name, ss.slot_name
FROM shift_assignments sa
JOIN members m ON sa.member_id = m.member_id
JOIN shift_slots ss ON sa.slot_id = ss.slot_id
WHERE sa.deleted_at IS NULL;

# ロールとメンバーの関連
SELECT m.display_name, r.name
FROM members m
JOIN member_roles mr ON m.member_id = mr.member_id
JOIN roles r ON mr.role_id = r.role_id
WHERE r.deleted_at IS NULL;

# 終了
\q
```

---

## テスト完了チェックリスト

### 基本機能
- [ ] ログイン/ログアウト
- [ ] イベント CRUD
- [ ] 営業日 CRUD
- [ ] シフト枠 CRUD
- [ ] メンバー CRUD
- [ ] シフト割り当て

### 高度な機能
- [ ] 満員チェック
- [ ] 深夜営業
- [ ] マルチテナント境界

### 公開ページ
- [ ] 出欠確認（公開URL）
- [ ] 日程調整（公開URL）

### ロール機能
- [ ] ロール CRUD
- [ ] メンバーへのロール割り当て
- [ ] ロールフィルター

### その他
- [ ] ブラウザ互換性（Chrome, Firefox, Safari）
- [ ] モバイル表示

---

## トラブルシューティング

### テストが失敗する

```bash
# コンテナログ確認
docker logs vrc-shift-scheduler-backend-1

# DB接続確認
docker exec vrc-shift-scheduler-db-1 psql -U vrcshift -c '\l'
```

### フロントエンドテストが失敗する

```bash
# 依存関係を再インストール
rm -rf node_modules
npm install

# キャッシュクリア
npm cache clean --force
```

---

## 関連ドキュメント

- `docs/tester/test-scenarios.md` - 詳細なテストシナリオ
- `docs/tester/quickstart.md` - テスター向けクイックスタート
- `backend/TESTING_GUIDE.md` - ロール機能テスト手順

# 🚀 クイックスタートガイド

## サーバー起動状況

✅ **すべてのサーバーが起動済みです！**

```
PostgreSQL: ✅ localhost:5432
Backend:    ✅ http://localhost:8080
Frontend:   ✅ http://localhost:5173
```

---

## 🔐 ログインしてテスト開始

### ステップ 1: ログイン
1. ブラウザを開いて以下にアクセス:
   ```
   http://localhost:5173/admin/login
   ```

2. 以下の情報でログイン:
   ```
   Email:    admin@test.com
   Password: password123
   ```

3. ログイン成功！イベント一覧画面が表示されます

---

## 👥 管理者招待機能のテスト

### ステップ 2: 管理者を招待
1. ナビゲーションバーの **「管理者招待」** をクリック

2. 以下の情報を入力:
   ```
   Email: test-newadmin@example.com
   Role:  admin (管理者)
   ```

3. **「招待を送信」** ボタンをクリック

4. 招待URLが表示されます（例: `http://localhost:5173/invite/abc123...`）

5. **「URLをコピー」** ボタンをクリックしてURLをコピー

### ステップ 3: 招待を受理
1. **新しいブラウザウィンドウ（シークレットモード推奨）** を開く

2. コピーした招待URLをアドレスバーに貼り付けてアクセス

3. 以下の情報を入力:
   ```
   表示名:           新管理者テスト
   パスワード:       mypassword123
   パスワード（確認）: mypassword123
   ```

4. **「登録」** ボタンをクリック

5. 登録完了のアラートが表示され、ログイン画面にリダイレクトされます

6. 新しく登録した情報でログイン:
   ```
   Email:    test-newadmin@example.com
   Password: mypassword123
   ```

7. ログイン成功！新しい管理者アカウントで管理画面にアクセスできます

---

## 🎯 主要な画面URL

| 画面 | URL |
|------|-----|
| 管理者ログイン | http://localhost:5173/admin/login |
| 管理者招待 | http://localhost:5173/admin/invite |
| イベント一覧 | http://localhost:5173/events |
| メンバー一覧 | http://localhost:5173/members |

---

## 🔧 トラブルシューティング

### ログインできない場合
- メールアドレスとパスワードを再確認してください
- ブラウザのキャッシュをクリアしてください
- ブラウザのコンソール（F12）でエラーを確認してください

### 招待URLが無効な場合
- 招待URLは **7日間** のみ有効です
- 期限切れの場合は、再度招待を作成してください

### サーバーが応答しない場合
- サーバーログを確認してください:
  ```bash
  tail -f /tmp/backend.log
  tail -f /tmp/frontend.log
  ```

---

## 📖 詳細なドキュメント

より詳細な情報は以下を参照してください:
- **DEBUG_GUIDE.md** - 完全なデバッグガイドとAPIリファレンス

---

**Happy Testing! 🎉**

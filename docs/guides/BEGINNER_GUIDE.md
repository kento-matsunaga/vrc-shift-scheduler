# 初心者向け開発ガイド

このガイドは、プログラミング初心者の方がこのプロジェクトで開発を始めるための手順書です。

---

## 目次

1. [必要なソフトウェアのインストール](#1-必要なソフトウェアのインストール)
2. [プロジェクトの取得](#2-プロジェクトの取得)
3. [開発環境の起動](#3-開発環境の起動)
4. [動作確認](#4-動作確認)
5. [コードの編集](#5-コードの編集)
6. [変更をGitHubに反映する](#6-変更をgithubに反映する)
7. [プルリクエストの作成](#7-プルリクエストの作成)
8. [よくあるエラーと解決方法](#8-よくあるエラーと解決方法)

---

## 1. 必要なソフトウェアのインストール

開発を始める前に、以下のソフトウェアをインストールしてください。

### Windows の場合

#### 1-1. WSL2 (Windows Subsystem for Linux)

Windows で Linux コマンドを使えるようにします。

1. **PowerShell を管理者として開く**
   - スタートメニューで「PowerShell」と検索
   - 右クリック → 「管理者として実行」

2. **以下のコマンドを実行**
   ```powershell
   wsl --install
   ```

3. **パソコンを再起動**

4. **Ubuntu のセットアップ**
   - 再起動後、自動的に Ubuntu が起動します
   - ユーザー名とパスワードを設定してください（忘れないようにメモ）

#### 1-2. Docker Desktop

1. https://www.docker.com/products/docker-desktop/ にアクセス
2. 「Download for Windows」をクリック
3. ダウンロードしたファイルを実行してインストール
4. パソコンを再起動
5. Docker Desktop を起動
6. 設定 → Resources → WSL Integration で「Ubuntu」にチェックを入れる

#### 1-3. Visual Studio Code (VSCode)

1. https://code.visualstudio.com/ にアクセス
2. 「Download for Windows」をクリック
3. ダウンロードしたファイルを実行してインストール
4. VSCode を起動して以下の拡張機能をインストール：
   - 「WSL」（Microsoft 製）
   - 「Japanese Language Pack」（日本語化したい場合）

#### 1-4. Git

WSL の Ubuntu 内で以下を実行：

```bash
# Ubuntu ターミナルを開く（スタートメニューで「Ubuntu」と検索）
sudo apt update
sudo apt install git -y

# Git の初期設定（自分の名前とメールアドレスに変更してください）
git config --global user.name "あなたの名前"
git config --global user.email "your-email@example.com"
```

### Mac の場合

#### 1-1. Homebrew

ターミナルを開いて以下を実行：

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

#### 1-2. Docker Desktop

```bash
brew install --cask docker
```

インストール後、Docker Desktop を起動してください。

#### 1-3. Git

```bash
brew install git

# Git の初期設定（自分の名前とメールアドレスに変更してください）
git config --global user.name "あなたの名前"
git config --global user.email "your-email@example.com"
```

#### 1-4. Visual Studio Code

```bash
brew install --cask visual-studio-code
```

---

## 2. プロジェクトの取得

### 2-1. GitHub アカウントの準備

1. https://github.com にアクセス
2. アカウントを作成（持っていない場合）
3. プロジェクトオーナーに GitHub ユーザー名を伝えて、リポジトリへのアクセス権をもらう

### 2-2. SSH キーの設定

GitHub にコードをアップロードするために必要です。

```bash
# SSH キーを生成（Enter を押してデフォルト設定で OK）
ssh-keygen -t ed25519 -C "your-email@example.com"

# 生成されたキーを表示
cat ~/.ssh/id_ed25519.pub
```

表示された文字列（`ssh-ed25519 AAAA...` で始まる長い文字列）をコピーして：

1. GitHub にログイン
2. 右上のアイコン → Settings
3. 左メニューの「SSH and GPG keys」
4. 「New SSH key」をクリック
5. Title に「My PC」など分かりやすい名前を入力
6. Key にコピーした文字列を貼り付け
7. 「Add SSH key」をクリック

### 2-3. プロジェクトをダウンロード

```bash
# 作業用フォルダを作成して移動
mkdir -p ~/dev
cd ~/dev

# プロジェクトをダウンロード（クローン）
git clone git@github.com:kento-matsunaga/vrc-shift-scheduler.git

# プロジェクトフォルダに移動
cd vrc-shift-scheduler
```

---

## 3. 開発環境の起動

### 3-1. Docker で起動

```bash
# プロジェクトフォルダにいることを確認
cd ~/dev/vrc-shift-scheduler

# Docker コンテナを起動（初回は数分かかります）
docker compose up -d
```

**このコマンドで以下が起動します：**
- PostgreSQL（データベース）: ポート 5432
- Backend（API サーバー）: ポート 8080
- Web Frontend（ユーザー画面）: ポート 5173

### 3-2. 起動状態の確認

```bash
# 起動中のコンテナを確認
docker compose ps
```

以下のような表示が出れば OK：

```
NAME                    STATUS
vrc-shift-scheduler-db-1          Up (healthy)
vrc-shift-scheduler-backend-1     Up
vrc-shift-scheduler-web-frontend-1  Up
```

### 3-3. 停止する場合

```bash
# 開発を終えるとき
docker compose down
```

---

## 4. 動作確認

### 4-1. バックエンド（API）の確認

ブラウザで http://localhost:8080/health にアクセス

以下が表示されれば OK：
```json
{"status":"ok"}
```

### 4-2. フロントエンドの確認

ブラウザで http://localhost:5173 にアクセス

ログイン画面が表示されれば OK です。

### 4-3. テストアカウントでログイン

初回はシードデータ（テスト用データ）を投入する必要があります：

```bash
# シードデータを投入
docker compose exec backend go run ./cmd/seed/main.go
```

その後、以下のアカウントでログインできます：

| メールアドレス | パスワード |
|---------------|-----------|
| admin1@example.com | password123 |

---

## 5. コードの編集

### 5-1. VSCode でプロジェクトを開く

**Windows (WSL) の場合：**

```bash
# WSL ターミナルで実行
cd ~/dev/vrc-shift-scheduler
code .
```

**Mac の場合：**

```bash
cd ~/dev/vrc-shift-scheduler
code .
```

### 5-2. プロジェクト構成

```
vrc-shift-scheduler/
├── backend/          # バックエンド（Go 言語）
│   ├── cmd/          # メインプログラム
│   ├── internal/     # 内部ロジック
│   └── Dockerfile    # Docker 設定
│
├── web-frontend/     # ユーザー向け画面（React）
│   ├── src/          # ソースコード
│   │   ├── components/  # 部品（ボタンなど）
│   │   ├── pages/       # ページ
│   │   └── lib/         # 共通処理
│   └── package.json  # 依存ライブラリ
│
├── admin-frontend/   # 管理画面（React）
│   └── src/          # ソースコード
│
├── docker-compose.yml  # Docker 設定
└── docs/               # ドキュメント
```

### 5-3. フロントエンドの編集

`web-frontend/src/` 以下のファイルを編集すると、ブラウザに自動で反映されます（ホットリロード）。

**例：ページタイトルを変更する**

1. `web-frontend/src/pages/EventList.tsx` を開く
2. 編集して保存
3. ブラウザが自動で更新される

### 5-4. バックエンドの編集

バックエンドを編集した場合は、再起動が必要です：

```bash
# バックエンドを再起動
docker compose restart backend
```

---

## 6. 変更をGitHubに反映する

### 6-1. Git の基本用語

| 用語 | 意味 |
|-----|-----|
| **リポジトリ** | プロジェクト全体のこと |
| **ブランチ** | 作業の分岐。本番に影響を与えずに開発できる |
| **コミット** | 変更を記録すること |
| **プッシュ** | ローカルの変更を GitHub に送ること |
| **プルリクエスト (PR)** | 変更を取り込んでもらう依頼 |

### 6-2. 作業の流れ

```
[develop] ─────────────────────────────────────────
     │
     └── [feature/xxx] ──●──●──●── (自分の作業)
                         ↓  ↓  ↓
                       commit
                              ↓
                            push
                              ↓
                         Pull Request
                              ↓
                           merge
                              ↓
[develop] ────────────────────●────────────────────
```

### 6-3. 新しい作業を始める

**必ず `develop` ブランチから始めてください！**

```bash
# 1. develop ブランチに移動
git checkout develop

# 2. 最新の状態を取得
git pull origin develop

# 3. 作業用ブランチを作成（名前は作業内容がわかるように）
git checkout -b feature/add-logout-button
```

**ブランチ名の例：**
- `feature/add-logout-button` - 新機能追加
- `fix/login-error` - バグ修正
- `docs/update-readme` - ドキュメント更新

### 6-4. 変更をコミットする

```bash
# 1. 変更したファイルを確認
git status

# 2. 変更をステージング（コミット対象に追加）
git add .

# 3. コミット（変更を記録）
git commit -m "ログアウトボタンを追加"
```

**コミットメッセージの書き方：**
- 何をしたかを簡潔に書く
- 日本語 OK
- 例：「ログアウトボタンを追加」「ログインエラーを修正」

### 6-5. GitHub にプッシュする

```bash
# GitHub に送信
git push origin feature/add-logout-button
```

---

## 7. プルリクエストの作成

### 7-1. GitHub でプルリクエストを作成

1. GitHub のリポジトリページにアクセス
   https://github.com/kento-matsunaga/vrc-shift-scheduler

2. 「Compare & pull request」ボタンをクリック
   （プッシュ直後は上部に表示されます）

3. 以下を設定：
   - **base**: `develop` ← 重要！`main` にしないこと
   - **compare**: `feature/add-logout-button`（自分のブランチ）

4. タイトルと説明を入力：
   ```
   タイトル: ログアウトボタンを追加

   説明:
   ## 変更内容
   - ヘッダーにログアウトボタンを追加しました
   - クリックするとログイン画面に戻ります

   ## 確認方法
   1. ログインする
   2. 右上のログアウトボタンをクリック
   3. ログイン画面に戻ることを確認
   ```

5. 「Create pull request」をクリック

### 7-2. レビュー後のマージ

1. レビュアーがコードを確認
2. 修正依頼があれば対応してプッシュ
3. 承認されたら「Merge pull request」をクリック
4. マージ完了！

### 7-3. 作業完了後のクリーンアップ

```bash
# develop に戻る
git checkout develop

# 最新を取得
git pull origin develop

# 使い終わったブランチを削除（任意）
git branch -d feature/add-logout-button
```

---

## 8. よくあるエラーと解決方法

### エラー: Docker が起動しない

**症状：**
```
Cannot connect to the Docker daemon
```

**解決方法：**
1. Docker Desktop が起動しているか確認
2. 起動していなければ、Docker Desktop を起動
3. Windows の場合は、WSL Integration が有効か確認

### エラー: ポートが使用中

**症状：**
```
Bind for 0.0.0.0:5173 failed: port is already allocated
```

**解決方法：**
```bash
# 既存のコンテナを停止
docker compose down

# 再起動
docker compose up -d
```

### エラー: git push が拒否される

**症状：**
```
error: failed to push some refs
```

**解決方法：**
```bash
# 最新の変更を取り込む
git pull origin develop

# もう一度プッシュ
git push origin feature/xxx
```

### エラー: 変更が反映されない（フロントエンド）

**解決方法：**
1. ブラウザのキャッシュをクリア（Ctrl + Shift + R）
2. Docker コンテナを再起動：
   ```bash
   docker compose restart web-frontend
   ```

### エラー: データベース接続エラー

**症状：**
```
connection refused
```

**解決方法：**
```bash
# データベースが起動しているか確認
docker compose ps

# 起動していない場合
docker compose up -d db

# 少し待ってからバックエンドを再起動
docker compose restart backend
```

---

## 困ったときは

1. **エラーメッセージを検索する**
   - Google で「エラーメッセージ」を検索
   - Stack Overflow で同様の問題を探す

2. **チームメンバーに聞く**
   - エラーメッセージのスクリーンショットを共有
   - 何をしようとして、何が起きたかを説明

3. **最初からやり直す**
   ```bash
   # すべてのコンテナとデータを削除して再構築
   docker compose down -v
   docker compose up -d --build
   ```

---

## 補足：よく使う Git コマンド一覧

| コマンド | 説明 |
|---------|-----|
| `git status` | 変更状態を確認 |
| `git add .` | すべての変更をステージング |
| `git commit -m "メッセージ"` | コミット |
| `git push origin ブランチ名` | プッシュ |
| `git pull origin develop` | 最新を取得 |
| `git checkout ブランチ名` | ブランチを切り替え |
| `git checkout -b 新ブランチ名` | 新しいブランチを作成して切り替え |
| `git branch` | ブランチ一覧を表示 |
| `git log --oneline -5` | 最近のコミット履歴を表示 |

---

## 次のステップ

- [開発ガイド（API情報など）](./DEVELOPMENT.md)
- [Windows セットアップ詳細](./setup-windows.md)

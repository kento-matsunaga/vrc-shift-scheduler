# 🪟 Windows 開発環境セットアップガイド

VRC Shift Scheduler を Windows で開発するための完全ガイドです。  
**対象**: Windows 11 / プログラミング初心者 / Git未経験者OK

---

## 📋 目次

0. [コマンドを打つ場所](#0-コマンドを打つ場所powershellでもvscodeでもok)
1. [事前準備チェック](#1-事前準備チェック)
2. [WSL2 のインストール](#2-wsl2-のインストール)
3. [Docker Desktop のインストール](#3-docker-desktop-のインストール)
4. [Git のセットアップ](#4-git-のセットアップ)
5. [エディタのインストール（任意）](#5-エディタのインストール任意)
6. [プロジェクトのクローン](#6-プロジェクトのクローン)
7. [開発環境の起動](#7-開発環境の起動)
8. [動作確認](#8-動作確認)
9. [開発の始め方（Git ワークフロー）](#9-開発の始め方git-ワークフロー)
10. [よくあるトラブルと対処法](#10-よくあるトラブルと対処法)

---

## 0. コマンドを打つ場所（PowerShellでもVSCodeでもOK）

このガイドの `powershell` コマンドは、次のどれで実行してもOKです。

### ✅ おすすめ：VSCode の統合ターミナル

1. VSCode でリポジトリを開く（File → Open Folder → `vrc-shift-scheduler`）
2. ターミナルを開く：`` Ctrl + ` ``（バッククォート）
3. 右上の `+`（新しいターミナル）→ **PowerShell** を選ぶ
4. 以後、このターミナルにコマンドをコピペでOK

### ✅ もちろんOK：Windows Terminal / PowerShell

- Windows Terminal（PowerShell）
- スタートメニューから PowerShell

### ⚠️「管理者として実行」が必要な作業だけは注意

WSL2 のインストールなど **管理者権限が必要な手順** は、VSCode ではなく  
スタートメニューから **「PowerShell（管理者として実行）」** を使ってください。

> 💡 このガイドで「⚠️ 管理者権限が必要」と書いてある箇所だけ注意すればOKです。

---

## 1. 事前準備チェック

### 必要なもの

- [x] Windows 11 PC（個人PC、管理者権限あり）
- [x] インターネット接続
- [x] GitHub アカウント（持っていない場合は [github.com](https://github.com) で作成）

### GitHub の権限について

| やりたいこと | 必要な権限 |
|-------------|-----------|
| clone（ダウンロード）だけ | **不要**（Public リポジトリなので誰でもOK） |
| push（コード反映）したい | **招待が必要**（オーナーから Write 権限をもらう） |
| 招待されていない場合 | Fork → 自分のリポジトリで作業 → PR を送る |

### PC の空き容量確認

Docker は約 **10GB 以上** の空き容量が必要です。

1. エクスプローラーを開く
2. 「PC」を選択
3. C ドライブの空き容量を確認（20GB 以上推奨）

---

## 2. WSL2 のインストール

WSL2（Windows Subsystem for Linux 2）は、Windows 上で Linux を動かすための機能です。  
Docker Desktop が内部で使用します。

### 手順

1. **⚠️ PowerShell を管理者として開く**（この手順だけ管理者権限が必要）
   - スタートメニューで「PowerShell」を検索
   - **「管理者として実行」** をクリック

2. **以下のコマンドをコピー＆ペーストして Enter**

```powershell
wsl --install
```

3. **PC を再起動**

4. **再起動後、Ubuntu のセットアップ**
   - 自動的に Ubuntu のウィンドウが開きます
   - ユーザー名を入力（半角英数字、小文字推奨。例: `tanaka`）
   - パスワードを入力（入力しても画面に表示されませんが、入力されています）
   - パスワード確認のため、もう一度入力

5. **インストール確認**

ターミナルで以下を実行（VSCode でもOK）：

```powershell
wsl --version
```

バージョン情報が表示されればOK！

---

## 3. Docker Desktop のインストール

### 手順

1. **Docker Desktop をダウンロード**
   - [https://www.docker.com/products/docker-desktop/](https://www.docker.com/products/docker-desktop/) にアクセス
   - 「Download for Windows」をクリック

2. **インストーラーを実行**
   - ダウンロードした `Docker Desktop Installer.exe` をダブルクリック
   - 「Use WSL 2 instead of Hyper-V」にチェックが入っていることを確認
   - 「OK」をクリックしてインストール

3. **PC を再起動**（インストーラーに促されたら）

4. **Docker Desktop を起動**
   - スタートメニューから「Docker Desktop」を検索して起動
   - 利用規約に同意（Accept）
   - チュートリアルはスキップしてOK

5. **WSL2 統合を有効化**
   - Docker Desktop の右上の歯車アイコン（Settings）をクリック
   - 左メニューの「Resources」→「WSL integration」を選択
   - 「Enable integration with my default WSL distro」がオンになっていることを確認
   - 「Ubuntu」のトグルもオンにする
   - 「Apply & Restart」をクリック

6. **インストール確認**

ターミナルで以下を実行（VSCode でもOK）：

```powershell
docker --version
docker compose version
```

バージョン情報が表示されればOK！

---

## 4. Git のセットアップ

### 4-1. Git for Windows のインストール

1. **Git をダウンロード**
   - [https://git-scm.com/download/win](https://git-scm.com/download/win) にアクセス
   - 自動でダウンロードが始まります

2. **インストーラーを実行**
   - ダウンロードした `Git-x.xx.x-64-bit.exe` をダブルクリック
   - 基本的にすべて「Next」でOK（デフォルト設定のまま）
   - 最後に「Install」→「Finish」

3. **インストール確認**

ターミナルで以下を実行（VSCode でもOK）：

```powershell
git --version
```

バージョンが表示されればOK！

### 4-2. Git の初期設定

ターミナルで以下を実行（`あなたの名前` と `メールアドレス` は自分のものに変更）：

```powershell
git config --global user.name "あなたの名前"
git config --global user.email "your-email@example.com"
```

> 💡 GitHub に登録したメールアドレスと同じものを使ってください

### 4-3. GitHub への SSH 接続設定（推奨）

毎回パスワードを入力しなくて済むようになります。

1. **SSH キーを生成**

ターミナルで以下を実行：

```powershell
ssh-keygen -t ed25519 -C "your-email@example.com"
```

- 「Enter file in which to save the key」→ そのまま Enter
- 「Enter passphrase」→ パスフレーズを入力（省略可、セキュリティ上は設定推奨）
- 「Enter same passphrase again」→ 同じパスフレーズを入力

2. **公開鍵をコピー**

```powershell
cat ~/.ssh/id_ed25519.pub
```

表示された文字列（`ssh-ed25519 AAAA...` から始まる1行）をすべてコピー

3. **GitHub に登録**
   - [https://github.com/settings/keys](https://github.com/settings/keys) にアクセス
   - 「New SSH key」をクリック
   - Title: 任意（例: `My Windows PC`）
   - Key: コピーした公開鍵を貼り付け
   - 「Add SSH key」をクリック

4. **接続テスト**

```powershell
ssh -T git@github.com
```

「Hi ユーザー名! You've successfully authenticated...」と表示されればOK！

---

## 5. エディタのインストール（任意）

お好みのエディタを使ってください。どれでもOKです。

### おすすめ: Visual Studio Code

1. [https://code.visualstudio.com/](https://code.visualstudio.com/) からダウンロード
2. インストーラーを実行

### おすすめ拡張機能

VSCode を開いて、左側の拡張機能アイコン（四角が4つ）から以下をインストール：

| 拡張機能名 | 説明 |
|------------|------|
| Docker | Docker ファイルのサポート |
| ESLint | JavaScript/TypeScript の Lint |
| Prettier | コードフォーマッター |
| Go | Go 言語サポート |
| Tailwind CSS IntelliSense | Tailwind の補完 |

---

## 6. プロジェクトのクローン

### 6-1. 作業フォルダの作成

ターミナルで以下を実行：

```powershell
mkdir ~/dev
cd ~/dev
```

### 6-2. リポジトリをクローン

```powershell
git clone git@github.com:kento-matsunaga/vrc-shift-scheduler.git
```

> 💡 SSH 接続設定をしていない場合は以下を使用：
> ```powershell
> git clone https://github.com/kento-matsunaga/vrc-shift-scheduler.git
> ```

### 6-3. ディレクトリに移動

```powershell
cd vrc-shift-scheduler
```

### 6-4. VSCode で開く（推奨）

```powershell
code .
```

以後は **VSCode の統合ターミナル** でコマンドを実行できます。  
（`` Ctrl + ` `` でターミナルを開く）

---

## 7. 開発環境の起動

### 7-1. Docker Desktop が起動していることを確認

タスクバー右下のシステムトレイに Docker のクジラアイコンがあればOK。  
なければスタートメニューから「Docker Desktop」を起動。

### 7-2. 開発環境を起動

ターミナルで以下を実行：

```powershell
cd ~/dev/vrc-shift-scheduler
docker compose up -d --build
```

> 💡 `--build` を付けると、Dockerfile の変更も反映されます（初回は必ず付けましょう）

初回は Docker イメージのダウンロードとビルドで **5〜10分** かかります。  
☕ コーヒーでも飲んで待ちましょう。

### 7-3. 起動状態の確認

```powershell
docker compose ps
```

以下のように `running` と表示されればOK：

```
NAME                           STATUS
vrc-shift-scheduler-db-1       running
vrc-shift-scheduler-backend-1  running
vrc-shift-scheduler-web-frontend-1  running
```

### 7-4. マイグレーションの実行

データベースのテーブルを作成します。

```powershell
docker compose exec backend go run ./cmd/migrate/main.go
```

> 💡 または Makefile を使用:
> ```powershell
> docker compose exec backend make migrate
> ```

### 7-5. シードデータの投入（任意）

テスト用の初期データを投入します（開発を始める前に実行推奨）。

```powershell
docker compose exec backend go run ./cmd/seed/main.go
```

> 💡 または Makefile を使用:
> ```powershell
> docker compose exec backend make seed
> ```

---

## 8. 動作確認

### 8-1. バックエンドの確認

ブラウザで以下にアクセス：

👉 **http://localhost:8080/health**

`{"status":"ok"}` のような応答があればOK！

### 8-2. フロントエンドの確認

ブラウザで以下にアクセス：

👉 **http://localhost:5173**

画面が表示されればOK！

### 8-3. テストの実行

```powershell
# バックエンドのテスト
docker compose exec backend go test ./...

# または Makefile を使用
docker compose exec backend make test
```

テストが通れば（`PASS` と表示されれば）OK！

---

## 9. 開発の始め方（Git ワークフロー）

### ブランチ運用ルール

| ブランチ | 用途 |
|----------|------|
| `main` | 本番用。直接 push 禁止。PR 経由でマージ |
| `feature/xxx` | 新機能開発用 |
| `fix/xxx` | バグ修正用 |

### 開発フロー

#### 1. 最新の main を取得

```powershell
git checkout main
git pull origin main
```

#### 2. 作業ブランチを作成

```powershell
# 新機能の場合
git checkout -b feature/add-login-page

# バグ修正の場合
git checkout -b fix/header-layout
```

#### 3. コードを編集

エディタでコードを編集します。

#### 4. 変更を確認

```powershell
git status
git diff
```

#### 5. 変更をコミット

```powershell
git add .
git commit -m "feat: ログインページを追加"
```

> 💡 コミットメッセージは日本語でもOKです

#### 6. リモートにプッシュ

```powershell
git push origin feature/add-login-page
```

#### 7. Pull Request を作成

1. GitHub のリポジトリページにアクセス
   - https://github.com/kento-matsunaga/vrc-shift-scheduler
2. 「Compare & pull request」ボタンをクリック
3. PR のタイトルと説明を記入
4. 「Create pull request」をクリック

#### 8. レビュー後にマージ

オーナーのレビューが通ったら、マージされます。

---

## 10. よくあるトラブルと対処法

### ❌ ポートが既に使われている（5432, 8080, 5173）

**エラー例:**
```
Error: port 5432 is already in use
```

**対処法:**

1. 使用中のポートを確認

```powershell
netstat -ano | findstr :5432
```

2. プロセスを終了

```powershell
# PID が 12345 の場合
taskkill /F /PID 12345
```

3. または Docker を再起動

```powershell
docker compose down
docker compose up -d
```

---

### ❌ Docker のメモリ不足

**症状:**
- コンテナが頻繁に停止する
- `OOMKilled` エラーが出る

**対処法:**

1. Docker Desktop の Settings を開く
2. 「Resources」→「Advanced」を選択
3. 「Memory」を 4GB 以上に設定（推奨: 6GB）
4. 「Apply & Restart」をクリック

---

### ❌ Go のバージョン違いでビルドエラー

**エラー例:**
```
go: go.mod requires go >= 1.23
```

**対処法:**

Docker 内で実行すれば問題ありません（Docker イメージに正しいバージョンが入っています）。

```powershell
# ローカルではなく Docker 内で実行
docker compose exec backend go version
```

---

### ❌ npm install で node-gyp エラー

**エラー例:**
```
gyp ERR! build error
```

**対処法:**

Docker 内で実行すれば問題ありません。

```powershell
# node_modules を削除して再起動
docker compose down
docker volume rm vrc-shift-scheduler_frontend_node_modules
docker compose up -d
```

---

### ❌ WSL2 とファイル共有が遅い

**症状:**
- ファイルの変更が反映されるのが遅い
- Docker が全体的に遅い

**対処法:**

Windows 側のファイルを WSL 内にコピーして開発する方法があります。

```bash
# WSL のターミナルで
cd ~
git clone git@github.com:kento-matsunaga/vrc-shift-scheduler.git
cd vrc-shift-scheduler
docker compose up -d
```

---

### ❌ docker compose up でコンテナがすぐ終了する

**対処法:**

ログを確認：

```powershell
docker compose logs backend
docker compose logs web-frontend
```

よくある原因：
- 環境変数の設定ミス
- ポートの競合
- Dockerfile のエラー

---

### ❌ Permission denied エラー（SSH 接続時）

**エラー例:**
```
Permission denied (publickey)
```

**対処法:**

1. SSH キーが正しく生成されているか確認

```powershell
ls ~/.ssh/
```

`id_ed25519` と `id_ed25519.pub` があればOK。

2. SSH agent を起動

```powershell
Start-Service ssh-agent
ssh-add ~/.ssh/id_ed25519
```

3. GitHub に公開鍵が登録されているか確認
   - https://github.com/settings/keys

---

## 🎉 セットアップ完了！

以上でセットアップは完了です。

### 困ったときは

1. このドキュメントの「よくあるトラブルと対処法」を確認
2. エラーメッセージをコピーして Google 検索
3. チームメンバーに Discord で相談

### 次のステップ

- [ ] http://localhost:5173 で画面が表示されることを確認
- [ ] `docker compose exec backend make test` でテストが通ることを確認
- [ ] テスト用のブランチを作成して PR を作成してみる

---

**Happy Coding! 🚀**


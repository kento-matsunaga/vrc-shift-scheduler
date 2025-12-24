# 🪟 Windows 開発環境セットアップガイド（完成版）

VRC Shift Scheduler を Windows 11 で開発するための完全ガイドです。  
**対象**: Windows 11 / 初心者OK / Git未経験OK  
**方針**: コマンドは **Windows Terminal の Ubuntu（WSL2）** で実行します（確実・高速・迷いにくい）

---

## 📋 目次

0. [このガイドのやり方（結論）](#0-このガイドのやり方結論)
1. [事前準備チェック](#1-事前準備チェック)
2. [Windows Terminal のインストール](#2-windows-terminal-のインストール)
3. [WSL2 + Ubuntu のインストール](#3-wsl2--ubuntu-のインストール)
4. [Docker Desktop のインストールと設定](#4-docker-desktop-のインストールと設定)
5. [Git（Ubuntu側）のセットアップ](#5-gitubuntu側のセットアップ)
6. [エディタ（任意：VSCode / Cursor）](#6-エディタ任意vscode--cursor)
7. [プロジェクトのクローン（Ubuntuで）](#7-プロジェクトのクローンubuntuで)
8. [環境変数（.env）の作り方](#8-環境変数envの作り方)
9. [開発環境の起動（Docker Compose）](#9-開発環境の起動docker-compose)
10. [DBマイグレーション・シード投入](#10-dbマイグレーションシード投入)
11. [動作確認（ブラウザ）](#11-動作確認ブラウザ)
12. [テスト実行](#12-テスト実行)
13. [開発の始め方（Git ワークフロー）](#13-開発の始め方git-ワークフロー)
14. [停止・初期化（困った時に戻す）](#14-停止初期化困った時に戻す)
15. [よくあるトラブルと対処法](#15-よくあるトラブルと対処法)

---

## 0. このガイドのやり方（結論）

- 以後のコマンドは基本 **Windows Terminal → Ubuntu** で実行します
- 例外：**WSL2 を有効化する最初の1回だけ**「PowerShell（管理者）」を使います

---

## 1. 事前準備チェック

### 必要なもの

- Windows 11（個人PC / 管理者権限あり）
- インターネット接続
- ディスク空き 20GB 以上推奨（Dockerが使うため）

### GitHubの権限について

| やりたいこと | 必要な権限 |
|---|---|
| clone（ダウンロード） | 不要（PublicリポジトリなのでOK） |
| 同じリポジトリへ push | オーナーから招待（Write権限） |
| 招待されていない | Forkして作業 → PRを送る |

---

## 2. Windows Terminal のインストール

Windows Terminal を入れて、そこから Ubuntu を開いて作業します。

### インストール方法（どちらか）

**A. Microsoft Store から入れる（おすすめ）**

- Microsoft Store で「Windows Terminal」を検索してインストール

**B. winget で入れる（使える人向け）**

- PowerShell で：

```powershell
winget install --id Microsoft.WindowsTerminal -e
```

インストール後、スタートメニューから **Windows Terminal** を起動してください。

---

## 3. WSL2 + Ubuntu のインストール

### 3-1. WSL2 を有効化（⚠️管理者権限が必要）

スタートメニューで「PowerShell」を検索 → **管理者として実行** → 下を実行：

```powershell
wsl --install
```

終わったら **再起動** します。

---

### 3-2. Ubuntu をインストール（必要なら）

`wsl --install` で Ubuntu が入ることが多いですが、入っていない場合は次で入れます。

**Microsoft Store から「Ubuntu」を入れる**

- Microsoft Store で「Ubuntu」を検索 → インストール
  - どれを選べばいいか迷う場合：`Ubuntu`（公式）でOK

---

### 3-3. Windows Terminal から Ubuntu を開く

Windows Terminal を開いて、タブから **Ubuntu** を選びます。

初回はユーザー名/パスワード作成が走ります。

---

### 3-4. WSL 動作確認（Ubuntuで実行）

Ubuntuターミナルで：

```bash
wsl.exe -l -v
```

`Ubuntu` が `Running` / `Stopped` で表示され、`VERSION 2` ならOKです。

---

## 4. Docker Desktop のインストールと設定

### 4-1. Docker Desktop をインストール

- [https://www.docker.com/products/docker-desktop/](https://www.docker.com/products/docker-desktop/) からダウンロード
- インストール中に **「Use WSL 2 instead of Hyper-V」** にチェックがあることを確認

---

### 4-2. Docker Desktop の初期設定（重要）

Docker Desktop を起動してから：

1. 右上の歯車（Settings）
2. **General**
   - ✅ Use the WSL 2 based engine（有効）
3. **Resources → WSL Integration**
   - ✅ Enable integration with my default WSL distro（有効）
   - ✅ Ubuntu を ON
4. Apply & Restart（再起動）

---

### 4-3. Docker 動作確認（Ubuntuで実行）

Ubuntuターミナルで：

```bash
docker --version
docker compose version
```

バージョンが出たらOK。

> もし `Cannot connect to the Docker daemon` が出たら
> → Docker Desktop が起動していない可能性が高いです（後述のトラブル参照）

---

## 5. Git（Ubuntu側）のセットアップ

Windows側のGitではなく、**Ubuntu側のGit**で統一すると事故が減ります。

### 5-1. Git をインストール（Ubuntu）

```bash
sudo apt update
sudo apt install -y git
git --version
```

### 5-2. Git 初期設定

```bash
git config --global user.name "あなたの名前"
git config --global user.email "your-email@example.com"
```

---

### 5-3. SSH 接続（推奨）

#### SSHキー生成（Ubuntu）

```bash
ssh-keygen -t ed25519 -C "your-email@example.com"
```

基本 Enter 連打でOK（passphraseは任意）

#### 公開鍵を表示してコピー

```bash
cat ~/.ssh/id_ed25519.pub
```

#### GitHub に登録

- GitHub → Settings → SSH and GPG keys → New SSH key
- Key に貼り付けて追加

#### 接続テスト

```bash
ssh -T git@github.com
```

`Hi ...` が出ればOK。

---

## 6. エディタ（任意：VSCode / Cursor）

### おすすめ（WSL連携で確実）

- **VSCode**（Remote - WSL拡張を入れるとUbuntu内のコード編集が快適）
- **Cursor** でも同様に WSL 内フォルダを開けます

> 重要：このガイドは「コマンド実行は Ubuntu」で統一します。
> エディタは好きなものを使ってOKです。

---

## 7. プロジェクトのクローン（Ubuntuで）

Ubuntuターミナルで：

```bash
mkdir -p ~/dev
cd ~/dev
git clone git@github.com:kento-matsunaga/vrc-shift-scheduler.git
cd vrc-shift-scheduler
```

SSHが未設定ならHTTPSでもOK：

```bash
git clone https://github.com/kento-matsunaga/vrc-shift-scheduler.git
```

---

## 8. 環境変数（.env）の作り方

**基本：起動だけなら不要な場合もあります。**

ただし、チーム開発は `.env` を作る運用にしておくと説明が簡単で事故が減ります。

```bash
# 例：プロジェクト直下
cp .env.example .env 2>/dev/null || true
```

Bot を動かす場合は、`.env` に以下を入れる（例）：

```env
DISCORD_BOT_TOKEN=xxxxx
DISCORD_APP_ID=yyyyy
VITE_TENANT_ID=dev-tenant
```

---

## 9. 開発環境の起動（Docker Compose）

### 9-1. 起動（Botなし：基本これ）

Ubuntuターミナルでリポジトリ直下にいることを確認して：

```bash
docker compose up -d --build
```

### 9-2. Botも起動したい場合（オプション）

```bash
docker compose --profile bot up -d --build
```

### 9-3. 起動状態の確認

```bash
docker compose ps
```

---

## 10. DBマイグレーション・シード投入

### 10-1. マイグレーション

```bash
docker compose exec backend /app/migrate
```

### 10-2. シード（任意）

```bash
docker compose exec backend /app/seed
```

シード投入後のログイン情報：
- **Email**: `admin1@example.com`
- **Password**: `password123`

> **注意**: 上記コマンドが動かない場合は、先に `docker compose build backend` でコンテナを再ビルドしてください。

---

## 11. 動作確認（ブラウザ）

- バックエンド： [http://localhost:8080/health](http://localhost:8080/health)
- フロントエンド： [http://localhost:5173](http://localhost:5173)

---

## 12. テスト実行

### バックエンド

バックエンドのテストはローカルでGoをインストールして実行します：

```bash
# Goのインストール（Ubuntu）
sudo apt update && sudo apt install -y golang-go

# backendディレクトリでテスト実行
cd backend
DATABASE_URL="postgres://vrcshift:vrcshift@localhost:5432/vrcshift?sslmode=disable" JWT_SECRET=test go test ./...
```

### フロントエンド（用意されている場合）

```bash
docker compose exec web-frontend npm test
```

---

## 13. 開発の始め方（Git ワークフロー）

### ブランチ運用

- `main`：直接push禁止（PRでマージ）
- `feature/xxx`：新機能
- `fix/xxx`：修正

### 作業手順（Ubuntuで）

```bash
git checkout main
git pull origin main
git checkout -b feature/my-task

# 編集…

git status
git add .
git commit -m "feat: 変更内容"
git push -u origin feature/my-task
```

あとは GitHub で Pull Request を作成します。

> 招待されていない（pushできない）場合は Fork → PR の流れでOK

---

## 14. 停止・初期化（困った時に戻す）

### 停止（DBデータは残る）

```bash
docker compose down
```

### 完全初期化（DBも消える：注意）

```bash
docker compose down -v
```

---

## 15. よくあるトラブルと対処法

### ❌ Docker が動かない / daemon に接続できない

- Docker Desktop が起動しているか確認（タスクバーのクジラ）
- Docker Desktop → Settings → WSL Integration で Ubuntu が ON か確認

---

### ❌ ポート競合（5432 / 8080 / 5173）

Ubuntuで確認（8080例）：

```bash
sudo ss -ltnp | grep ':8080' || true
```

手早く戻す：

```bash
docker compose down
docker compose up -d --build
```

---

### ❌ 反映が遅い / ホットリロードしない

- Windows側のフォルダではなく **Ubuntu側（~/dev/..）で作業しているか**確認
  → このガイド通りなら基本大丈夫です

---

### ❌ SSH Permission denied (publickey)

Ubuntuで：

```bash
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_ed25519
ssh -T git@github.com
```

---

### ❌ Botが落ちる

- `.env` の `DISCORD_BOT_TOKEN / DISCORD_APP_ID` が入っているか確認
- bot起動は必要になってからでOK（まずは本体だけ動けばOK）

---

## 🎉 セットアップ完了！

次のチェック：

- [ ] `docker compose up -d --build` が成功
- [ ] [http://localhost:5173](http://localhost:5173) が表示される
- [ ] [http://localhost:8080/health](http://localhost:8080/health) が応答する
- [ ] `docker compose exec backend /app/migrate` でマイグレーションが通る
- [ ] ブランチ作ってPR作成までできる

---

**Happy Coding! 🚀**

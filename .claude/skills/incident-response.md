---
description: インシデント対応手順と過去の教訓
---

# Incident Response

VRC Shift Scheduler のインシデント対応手順と過去の教訓。

---

## インシデント対応フロー

```
1. 検知・報告
     ↓
2. 影響範囲の特定
     ↓
3. 一時対応（必要に応じてロールバック）
     ↓
4. 根本原因の調査
     ↓
5. 恒久対応
     ↓
6. ポストモーテム作成
```

---

## ロールバック手順

### バックエンドのロールバック

```bash
# 非対話的環境からの場合（Claude Code等）
DISPLAY=:0 SSH_ASKPASS=/tmp/askpass.sh SSH_ASKPASS_REQUIRE=force \
  setsid -w ssh -o StrictHostKeyChecking=no root@163.44.103.76 "
    cd /opt/vrcshift && \
    docker-compose -f docker-compose.prod.yml --env-file .env.prod down && \
    git checkout v1.8.0 && \
    docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d
  "

# 対話的環境からの場合
ssh root@163.44.103.76

# サーバー上で実行
cd /opt/vrcshift

# 利用可能なタグを確認
git tag --list 'v*' --sort=-v:refname | head -10

# 戻したいバージョンにチェックアウト
git fetch origin
git checkout v1.8.0

# サービス再起動（docker-composeを使用）
docker-compose -f docker-compose.prod.yml --env-file .env.prod down
docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d

# ログ確認
docker logs vrc-shift-backend --tail=100 -f
```

### データベースのロールバック

```bash
# マイグレーションを1つ戻す
docker exec vrc-shift-backend /app/migrate -action=down -steps=1

# 状態確認
docker exec vrc-shift-backend /app/migrate -action=status
```

---

## 過去のインシデント

### 2026-01-17: v1.7.0 デプロイ失敗

**概要**: develop → main のスカッシュマージにより、次回マージ時にコンフリクト発生

**根本原因**:
- GitHub の「Squash and merge」を使用
- スカッシュマージは親コミット情報を失う
- Git が「どこまでマージ済みか」を判断できなくなる

**影響**:
- main と develop の履歴が乖離
- 全ファイルにコンフリクト発生
- デプロイ作業が数時間停滞

**解決策**:
```bash
# mainをdevelopの状態に強制リセット
git checkout main
git reset --hard origin/develop
git push --force origin main
```

**恒久対応**:
- develop → main は **必ず「Create a merge commit」** を使用
- BRANCH_STRATEGY.md に明記
- スカッシュマージは feature → develop のみ許可

**教訓**:
- Git の merge 戦略は事前にチームで統一
- ブランチ運用ルールはドキュメント化必須
- 本番デプロイ前にローカルでマージテスト

### 2026-01-27: v1.9.0 デプロイ時の学び

**概要**: デプロイ自体は成功したが、複数の技術的問題に遭遇

**遭遇した問題と解決策**:

1. **SSH_ASKPASSが動作しない**
   - 原因: `setsid`を使用していなかった
   - 解決: `setsid -w ssh ...`で実行

2. **docker compose vs docker-compose**
   - 原因: サーバーには古い`docker-compose`（ハイフン付き）がインストール
   - 解決: `docker-compose`を使用

3. **環境変数が反映されない**
   - 原因: `restart`ではなく`rm`→`up`が必要
   - 解決: コンテナを再作成

4. **batchバイナリがない**
   - 原因: Dockerfileにbatchのビルドが含まれていなかった
   - 解決: Dockerfileを修正

**教訓**:
- 非対話的環境でSSH_ASKPASSを使う場合は`setsid`必須
- サーバーのDockerバージョンを事前確認
- 環境変数変更時はコンテナ再作成
- 新しいバイナリ追加時はDockerfileも更新

---

## トラブルシューティング

### バックエンドが起動しない

```bash
# 1. ログ確認
docker-compose -f docker-compose.prod.yml logs backend --tail=100

# 2. コンテナ状態確認
docker ps -a | grep backend

# 3. 環境変数確認
docker exec vrc-shift-backend env | grep -E 'DATABASE|JWT'

# 4. DB接続確認
docker exec vrc-shift-db psql -U vrcshift -d vrcshift -c "SELECT 1"
```

### データベース接続エラー

```bash
# 1. DBコンテナ確認
docker ps | grep db

# 2. ポート確認
docker exec vrc-shift-db pg_isready -U vrcshift

# 3. 接続プール状態
docker exec vrc-shift-db psql -U vrcshift -d vrcshift -c \
  "SELECT count(*) FROM pg_stat_activity WHERE datname = 'vrcshift'"
```

### マイグレーションエラー

```bash
# 1. 現在の状態確認
docker exec vrc-shift-backend /app/migrate -action=status

# 2. dirty状態のリセット（要注意）
docker exec vrc-shift-db psql -U vrcshift -d vrcshift -c \
  "UPDATE schema_migrations SET dirty = false"

# 3. 再実行
docker exec vrc-shift-backend /app/migrate -action=up
```

---

## 緊急連絡フロー

```
1. 本番障害検知
     ↓
2. Slackで即時報告（#incidents チャンネル）
     ↓
3. 影響範囲を簡潔に記載
     ↓
4. 対応状況を随時更新
     ↓
5. 解決後、ポストモーテム作成
```

---

## ポストモーテムテンプレート

```markdown
# インシデント報告: [タイトル]

**日時**: YYYY-MM-DD HH:MM - HH:MM
**影響範囲**: [影響を受けたサービス/ユーザー数]
**重大度**: Critical / High / Medium / Low

## タイムライン

- HH:MM - 障害検知
- HH:MM - 調査開始
- HH:MM - 原因特定
- HH:MM - 対応完了

## 根本原因

[原因の詳細説明]

## 対応内容

[実施した対応の詳細]

## 再発防止策

- [ ] アクションアイテム1
- [ ] アクションアイテム2

## 教訓

[今後に活かすべき学び]
```

---

## 監視項目

### アプリケーション

- APIレスポンスタイム
- エラーレート（5xx）
- リクエスト数/分

### データベース

- 接続数
- スロークエリ数
- ディスク使用率

### インフラ

- CPU使用率
- メモリ使用率
- ディスクI/O

---

## 関連ドキュメント

- `docs/deployment/PRODUCTION_DEPLOYMENT.md` - デプロイ手順
- `docs/development/BRANCH_STRATEGY.md` - ブランチ運用
- `docs/incidents/` - 過去のインシデント報告

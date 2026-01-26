# 本番デプロイ

本番環境へのデプロイ手順:

## 前提条件確認

1. developブランチが最新か確認
2. 全テストがパス
3. コードレビュー完了
4. SSHコンフィグが設定済み（下記参照）

## SSH設定（初回のみ）

`~/.ssh/config` に以下を追加:

```
Host vrcshift-prod
    HostName <本番サーバーIP>
    User vrcshift
    IdentityFile ~/.ssh/id_rsa
```

**注意**: 本番サーバーのIPアドレスは `docs/PRODUCTION_DEPLOYMENT.md` を参照

## デプロイ手順

### 1. tarball作成

```bash
./scripts/build-prod-tarball.sh
```

### 2. サーバーへ転送

```bash
scp vrcshift-*.tar.gz vrcshift-prod:/opt/vrcshift/
```

### 3. サーバーで展開

```bash
ssh vrcshift-prod
cd /opt/vrcshift
tar -xzf vrcshift-*.tar.gz
```

### 4. サービス再起動

```bash
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml up -d
```

### 5. 動作確認

```bash
docker-compose -f docker-compose.prod.yml logs -f
```

## 重要事項

- **必ず `docker-compose.prod.yml` を使用**（開発用 `docker-compose.yml` は禁止）
- デプロイ前にバックアップを確認
- 問題発生時は即座にロールバック

## ロールバック手順

1. 前バージョンのtarballを展開
2. docker-compose再起動

## チェックリスト

- [ ] テスト全パス
- [ ] コードレビュー完了
- [ ] バックアップ確認
- [ ] tarball作成
- [ ] サーバー転送
- [ ] 展開・再起動
- [ ] 動作確認

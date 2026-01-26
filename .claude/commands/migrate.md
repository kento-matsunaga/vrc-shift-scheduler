# マイグレーション

データベースマイグレーションの管理:

## 状態確認

```bash
docker exec vrc-shift-backend /app/migrate -action=status
```

## マイグレーション実行

```bash
# 全て適用
docker exec vrc-shift-backend /app/migrate -action=up

# 1つだけ適用
docker exec vrc-shift-backend /app/migrate -action=up -steps=1
```

## ロールバック

```bash
# 1つ戻す
docker exec vrc-shift-backend /app/migrate -action=down -steps=1
```

## 新規マイグレーション作成

ファイル命名規則: `NNN_description.up.sql`

場所: `backend/internal/infra/db/migrations/`

例:
- `040_add_new_column.up.sql`
- `040_add_new_column.down.sql`

## チェックリスト

- [ ] up と down の両方を作成
- [ ] down が正しくロールバックできる
- [ ] 本番データに影響がないか確認
- [ ] インデックスが適切に設定されている

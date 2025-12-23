# Claude Code Review Instructions

`@claude review` コマンドで実行されるレビュー指示です。

## レビュー観点

### 1. Backend: DDD 原則の遵守
- レイヤー分離: Domain -> Application -> Infrastructure -> Interface
- リポジトリパターン: Usecase から DB 直接アクセス禁止
- ドメインロジック: ビジネスロジックが Domain 層に配置されているか

### 2. Frontend: デザイン整合性
- 既存の UI パターンに従っているか
- Tailwind CSS クラスの一貫性
- レスポンシブ対応

### 3. Frontend-Backend 疎通
- API エンドポイントの正確性
- TypeScript 型と JSON 構造の一致
- エラーハンドリング
- 認証ヘッダー (JWT, X-Tenant-ID)
- 例外的なワークアラウンドがないか

### 4. 一時的実装の検出
- TODO/FIXME コメント
- ハードコードされたテストデータ
- console.log デバッグ出力
- any 型の使用
- モックデータを返す関数

## 出力形式

```markdown
## Code Review Summary

### DDD 原則: ✅/❌
### デザイン整合性: ✅/❌
### API 疎通: ✅/❌
### 一時的実装: ✅ なし / ❌ あり

### 詳細
（具体的な問題と改善案）

### 総合評価
Approve / Request Changes / Comment
```

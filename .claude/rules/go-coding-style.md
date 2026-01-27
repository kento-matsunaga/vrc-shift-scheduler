# Go コーディングスタイル

## エラーハンドリング（必須）

エラーは必ず明示的に処理する:

```go
// NG: エラーを無視
result, _ := riskyOperation()

// OK: エラーを処理
result, err := riskyOperation()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

## 構造体の定義順序

一貫した順序で記述:
1. プライベートフィールド
2. パブリックフィールド（必要な場合）
3. コンストラクタ関数
4. ゲッター
5. ビジネスメソッド
6. プライベートヘルパーメソッド

## ファイル構成

- エンティティ/集約ごとに1ファイル
- テストファイルは隣接: `xxx.go` と `xxx_test.go`
- 500行以内を目安
- ユーティリティは別パッケージに抽出

## 命名規則

- ファイル: `snake_case.go`
- 変数/関数: `camelCase`
- 型/構造体: `PascalCase`
- DBカラム: `snake_case`

## エラーメッセージ

ドメインバリデーションは英語メッセージを使用:

```go
// OK
return common.NewValidationError("tenant_name is required", nil)

// ユーザー向けエラー
return common.NewDomainError(common.ErrNotFound, "Tenant not found")
```

## インポートの整理

以下の順序でグループ化:
1. 標準ライブラリ
2. サードパーティパッケージ
3. 内部パッケージ

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/go-chi/chi/v5"

    "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)
```

## コード品質チェックリスト

作業完了前に確認:
- [ ] 全てのエラーが明示的に処理されている
- [ ] 未使用の変数やインポートがない
- [ ] 関数は小さい（50行以内が目安）
- [ ] ファイルは焦点が絞られている（500行以内）
- [ ] エラーにコンテキストを付与してラップ
- [ ] 新規コードにテストを記述
- [ ] ハードコードされた値がない（定数/設定を使用）

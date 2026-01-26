# セキュリティルール

## シークレット（必須）

シークレットを絶対にハードコードしない:

```go
// NG: ハードコードされたシークレット
const apiKey = "sk-live-abc123..."

// OK: 環境変数から取得
apiKey := os.Getenv("STRIPE_API_KEY")
```

## SQLインジェクション防止（必須）

必ずパラメータ化クエリを使用:

```go
// NG: 文字列結合
query := fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", userID)

// OK: パラメータ化クエリ
query := "SELECT * FROM users WHERE id = $1"
rows, err := pool.Query(ctx, query, userID)
```

## 入力バリデーション（重要）

境界で全てのユーザー入力を検証:

```go
// Handlerで入力を検証
func (h *Handler) CreateTenant(w http.ResponseWriter, r *http.Request) {
    var req CreateTenantRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        // エラー処理
    }

    // バリデーション
    if req.TenantName == "" {
        // バリデーションエラーを返す
    }
}

// Domainでも検証（多層防御）
func NewTenant(now time.Time, name string) (*Tenant, error) {
    if name == "" {
        return nil, common.NewValidationError("tenant_name is required", nil)
    }
    // ...
}
```

## 認証（必須）

JWTトークンを必ず検証:

```go
// ミドルウェアでJWTを抽出・検証
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := extractToken(r)
        claims, err := validateJWT(token)
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        ctx := context.WithValue(r.Context(), "claims", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## 認可（必須）

操作前に権限を確認:

```go
func (uc *TenantUsecase) Delete(ctx context.Context, tenantID string) error {
    claims := ctx.Value("claims").(*Claims)

    // テナントアクセスを確認
    if claims.TenantID != tenantID {
        return common.NewDomainError(common.ErrForbidden, "Access denied")
    }

    // ロールを確認
    if claims.Role != "owner" {
        return common.NewDomainError(common.ErrForbidden, "Owner role required")
    }

    // 削除処理を続行
}
```

## 機密データの取り扱い

### パスワードの保存

```go
// 必ずパスワードをハッシュ化
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// 平文での保存は禁止
admin.Password = password  // 禁止！
```

### ログ内のPII

```go
// NG: 機密データをログ出力
log.Printf("User login: email=%s, password=%s", email, password)

// OK: 機密データをマスク
log.Printf("User login: email=%s", maskEmail(email))
```

## Webhookセキュリティ

Webhook署名を検証:

```go
func (h *WebhookHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
    payload, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    // Stripe署名を検証
    event, err := webhook.ConstructEvent(
        payload,
        r.Header.Get("Stripe-Signature"),
        webhookSecret,
    )
    if err != nil {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }

    // 検証済みイベントを処理
}
```

## セキュリティチェックリスト

デプロイ前に確認:
- [ ] コードにハードコードされたシークレットがない
- [ ] 全てのSQLクエリがパラメータ化されている
- [ ] 境界で入力バリデーションを実施
- [ ] 保護されたルートに認証ミドルウェアを適用
- [ ] Usecaseで認可チェックを実施
- [ ] パスワードをbcryptでハッシュ化
- [ ] ログに機密データがない
- [ ] Webhook署名を検証
- [ ] 本番環境でHTTPSを強制
- [ ] CORSが適切に設定されている

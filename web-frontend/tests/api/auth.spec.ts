import { test, expect } from '@playwright/test';
import {
  ENDPOINTS,
  TEST_CREDENTIALS,
  ApiResponse,
  ApiClient,
} from '../utils/api-client';
import { LoginResponse, loginAsAdmin, getUnauthenticatedClient } from '../utils/auth';

/**
 * Auth API Tests
 *
 * Endpoints:
 * 1. POST /api/v1/auth/login - ログイン
 * 2. POST /api/v1/setup - 初回セットアップ
 * 3. POST /api/v1/auth/register-by-invite - 招待登録
 * 4. GET /api/v1/auth/password-reset-status - パスワードリセット状態確認
 * 5. POST /api/v1/auth/reset-password - パスワードリセット
 */

test.describe('Auth API', () => {
  // ============================================================
  // 1. POST /api/v1/auth/login - ログイン
  // ============================================================
  test.describe('POST /api/v1/auth/login', () => {
    test.describe('正常系', () => {
      test('有効な認証情報でログイン成功', async ({ request }) => {
        const response = await request.post(ENDPOINTS.login, {
          data: TEST_CREDENTIALS,
        });

        expect(response.status()).toBe(200);

        const body = (await response.json()) as ApiResponse<LoginResponse>;
        expect(body.data).toBeDefined();
        expect(body.data.token).toBeDefined();
        expect(body.data.token).not.toBe('');
        expect(body.data.admin_id).toBeDefined();
        expect(body.data.tenant_id).toBeDefined();
        expect(body.data.role).toBeDefined();
        expect(body.data.expires_at).toBeDefined();
      });

      test('トークンの有効期限が設定されている', async ({ request }) => {
        const response = await request.post(ENDPOINTS.login, {
          data: TEST_CREDENTIALS,
        });

        expect(response.status()).toBe(200);

        const body = (await response.json()) as ApiResponse<LoginResponse>;
        const expiresAt = new Date(body.data.expires_at);
        const now = new Date();

        // 有効期限が現在時刻より後であること
        expect(expiresAt.getTime()).toBeGreaterThan(now.getTime());
      });
    });

    test.describe('異常系', () => {
      test('存在しないメールアドレスで401エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.login, {
          data: {
            email: 'nonexistent@example.com',
            password: TEST_CREDENTIALS.password,
          },
        });

        expect(response.status()).toBe(401);

        const body = await response.json();
        expect(body.error).toBeDefined();
      });

      test('間違ったパスワードで401エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.login, {
          data: {
            email: TEST_CREDENTIALS.email,
            password: 'wrongpassword123',
          },
        });

        expect(response.status()).toBe(401);

        const body = await response.json();
        expect(body.error).toBeDefined();
      });

      test('空のメールアドレスで400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.login, {
          data: {
            email: '',
            password: TEST_CREDENTIALS.password,
          },
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('空のパスワードで400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.login, {
          data: {
            email: TEST_CREDENTIALS.email,
            password: '',
          },
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('認証情報なしで400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.login, {
          data: {},
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('不正なメールアドレス形式で400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.login, {
          data: {
            email: 'not-an-email',
            password: TEST_CREDENTIALS.password,
          },
        });

        // 400 (validation error) or 401 (auth failed)
        expect([400, 401]).toContain(response.status());
      });

      test('SQLインジェクション試行は安全に処理される', async ({ request }) => {
        const response = await request.post(ENDPOINTS.login, {
          data: {
            email: "admin@example.com' OR '1'='1",
            password: "' OR '1'='1",
          },
        });

        // Should not succeed, should return 400 or 401
        expect(response.status()).toBeGreaterThanOrEqual(400);
        expect(response.status()).not.toBe(200);
      });
    });
  });

  // ============================================================
  // 2. POST /api/v1/setup - 初回セットアップ
  // ============================================================
  test.describe('POST /api/v1/setup', () => {
    test.describe('異常系', () => {
      // Note: 正常系はテナントが存在しない状態でのみ実行可能
      // シードデータが存在する環境では失敗する

      test('既にテナントが存在する場合はエラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.setup, {
          data: {
            organization_name: 'Test Organization',
            admin_name: 'Test Admin',
            password: 'testpassword123',
          },
        });

        // 既存テナントがある場合は 400 or 409 (Conflict)
        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('組織名なしで400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.setup, {
          data: {
            admin_name: 'Test Admin',
            password: 'testpassword123',
          },
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('管理者名なしで400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.setup, {
          data: {
            organization_name: 'Test Organization',
            password: 'testpassword123',
          },
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('パスワードなしで400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.setup, {
          data: {
            organization_name: 'Test Organization',
            admin_name: 'Test Admin',
          },
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('空のリクエストボディで400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.setup, {
          data: {},
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('短すぎるパスワードで400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.setup, {
          data: {
            organization_name: 'Test Organization',
            admin_name: 'Test Admin',
            password: '123', // Too short
          },
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });
    });
  });

  // ============================================================
  // 3. POST /api/v1/auth/register-by-invite - 招待登録
  // ============================================================
  test.describe('POST /api/v1/auth/register-by-invite', () => {
    test.describe('異常系', () => {
      test('無効な招待トークンで400/404エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.registerByInvite, {
          data: {
            invite_token: 'invalid-token-12345',
            display_name: 'Test User',
            password: 'testpassword123',
          },
        });

        // Invalid token: 400 or 404
        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('招待トークンなしで400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.registerByInvite, {
          data: {
            display_name: 'Test User',
            password: 'testpassword123',
          },
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('表示名なしで400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.registerByInvite, {
          data: {
            invite_token: 'some-token',
            password: 'testpassword123',
          },
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('パスワードなしで400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.registerByInvite, {
          data: {
            invite_token: 'some-token',
            display_name: 'Test User',
          },
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('空のリクエストボディで400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.registerByInvite, {
          data: {},
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('期限切れの招待トークンで400/403エラー', async ({ request }) => {
        // Note: This would require an expired token in the database
        const response = await request.post(ENDPOINTS.registerByInvite, {
          data: {
            invite_token: 'expired-token-12345',
            display_name: 'Test User',
            password: 'testpassword123',
          },
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });
    });
  });

  // ============================================================
  // 4. GET /api/v1/auth/password-reset-status - パスワードリセット状態確認
  // ============================================================
  test.describe('GET /api/v1/auth/password-reset-status', () => {
    test.describe('正常系', () => {
      test('存在するメールアドレスのリセット状態を取得', async ({ request }) => {
        const response = await request.get(ENDPOINTS.passwordResetStatus, {
          params: {
            email: TEST_CREDENTIALS.email,
          },
        });

        // 200 (success) or 429 (rate limited)
        expect([200, 429]).toContain(response.status());

        if (response.status() === 200) {
          const body = await response.json();
          expect(body.data).toBeDefined();
          expect(typeof body.data.allowed).toBe('boolean');
        }
      });

      test('レスポンスに必要なフィールドが含まれる', async ({ request }) => {
        const response = await request.get(ENDPOINTS.passwordResetStatus, {
          params: {
            email: TEST_CREDENTIALS.email,
          },
        });

        // 200 (success) or 429 (rate limited)
        expect([200, 429]).toContain(response.status());

        if (response.status() === 200) {
          const body = await response.json();
          expect(body.data).toHaveProperty('allowed');
          // allowed=trueの場合は expires_at と tenant_id も含まれる可能性
        }
      });
    });

    test.describe('異常系', () => {
      test('存在しないメールアドレスの場合', async ({ request }) => {
        const response = await request.get(ENDPOINTS.passwordResetStatus, {
          params: {
            email: 'nonexistent@example.com',
          },
        });

        // Security: don't reveal if email exists
        // Should return 200 with allowed=false, 404, or 429 (rate limited)
        expect([200, 404, 429]).toContain(response.status());
      });

      test('emailパラメータなしで400エラー', async ({ request }) => {
        const response = await request.get(ENDPOINTS.passwordResetStatus);

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('空のemailパラメータで400エラー', async ({ request }) => {
        const response = await request.get(ENDPOINTS.passwordResetStatus, {
          params: {
            email: '',
          },
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('不正なメールアドレス形式', async ({ request }) => {
        const response = await request.get(ENDPOINTS.passwordResetStatus, {
          params: {
            email: 'not-an-email',
          },
        });

        // 400 (validation), 200/404 (processed but not found), or 429 (rate limited)
        expect([200, 400, 404, 429]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 5. POST /api/v1/auth/reset-password - パスワードリセット
  // ============================================================
  test.describe('POST /api/v1/auth/reset-password', () => {
    test.describe('異常系', () => {
      // Note: 正常系はパスワードリセットが許可された状態でのみ実行可能

      test('無効なライセンスキーで400/401エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.resetPassword, {
          data: {
            email: TEST_CREDENTIALS.email,
            license_key: 'invalid-license-key',
            new_password: 'newpassword123',
            confirm_new_password: 'newpassword123',
          },
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('パスワード不一致で400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.resetPassword, {
          data: {
            email: TEST_CREDENTIALS.email,
            license_key: 'some-key',
            new_password: 'newpassword123',
            confirm_new_password: 'differentpassword123',
          },
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('emailなしで400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.resetPassword, {
          data: {
            license_key: 'some-key',
            new_password: 'newpassword123',
            confirm_new_password: 'newpassword123',
          },
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('ライセンスキーなしで400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.resetPassword, {
          data: {
            email: TEST_CREDENTIALS.email,
            new_password: 'newpassword123',
            confirm_new_password: 'newpassword123',
          },
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('新パスワードなしで400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.resetPassword, {
          data: {
            email: TEST_CREDENTIALS.email,
            license_key: 'some-key',
            confirm_new_password: 'newpassword123',
          },
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('確認パスワードなしで400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.resetPassword, {
          data: {
            email: TEST_CREDENTIALS.email,
            license_key: 'some-key',
            new_password: 'newpassword123',
          },
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('空のリクエストボディで400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.resetPassword, {
          data: {},
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('短すぎるパスワードで400エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.resetPassword, {
          data: {
            email: TEST_CREDENTIALS.email,
            license_key: 'some-key',
            new_password: '123',
            confirm_new_password: '123',
          },
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('リセットが許可されていないユーザーで403エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.resetPassword, {
          data: {
            email: TEST_CREDENTIALS.email,
            license_key: 'valid-but-not-allowed',
            new_password: 'newpassword123',
            confirm_new_password: 'newpassword123',
          },
        });

        // Should fail because reset is not allowed or key is invalid
        expect(response.status()).toBeGreaterThanOrEqual(400);
      });
    });
  });

  // ============================================================
  // Health Check
  // ============================================================
  test.describe('Health Check', () => {
    test('ヘルスチェックが200を返す', async ({ request }) => {
      const response = await request.get(ENDPOINTS.health);

      expect(response.status()).toBe(200);
    });
  });

  // ============================================================
  // Token Validation (認証トークンの検証)
  // ============================================================
  test.describe('Token Validation', () => {
    test('有効なトークンでテナント情報にアクセス可能', async ({ request }) => {
      const { client } = await loginAsAdmin(request);

      const response = await client.raw('GET', ENDPOINTS.tenant);

      expect(response.status()).toBe(200);
      const body = await response.json();
      expect(body.data).toBeDefined();
    });

    test('トークンなしで認証エラー', async ({ request }) => {
      const client = getUnauthenticatedClient(request);

      const response = await client.raw('GET', ENDPOINTS.tenant);

      // 400 (missing header) or 401 (unauthorized)
      expect([400, 401]).toContain(response.status());
    });

    test('無効なトークンで401エラー', async ({ request }) => {
      const invalidClient = new ApiClient(request);
      invalidClient.setToken('invalid-token-12345');

      const response = await invalidClient.raw('GET', ENDPOINTS.tenant);

      expect(response.status()).toBe(401);
    });

    test('改ざんされたトークンで401エラー', async ({ request }) => {
      const { loginData } = await loginAsAdmin(request);

      // Tamper with the token
      const tamperedToken = loginData.token.slice(0, -5) + 'XXXXX';

      const tamperedClient = new ApiClient(request);
      tamperedClient.setToken(tamperedToken);

      const response = await tamperedClient.raw('GET', ENDPOINTS.tenant);

      expect(response.status()).toBe(401);
    });

    test('空のトークンで認証エラー', async ({ request }) => {
      const emptyTokenClient = new ApiClient(request);
      emptyTokenClient.setToken('');

      const response = await emptyTokenClient.raw('GET', ENDPOINTS.tenant);

      expect([400, 401]).toContain(response.status());
    });
  });
});

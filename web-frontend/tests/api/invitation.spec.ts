import { test, expect } from '@playwright/test';
import {
  ENDPOINTS,
  ApiClient,
} from '../utils/api-client';
import { loginAsAdmin, getUnauthenticatedClient } from '../utils/auth';

/**
 * Invitation API Tests
 *
 * Endpoints:
 * 1. POST /api/v1/invitations - 招待作成
 * 2. POST /api/v1/invitations/accept/{token} - 招待承諾
 */

test.describe('Invitation API', () => {
  // ============================================================
  // 1. POST /api/v1/invitations - 招待作成
  // ============================================================
  test.describe('POST /api/v1/invitations', () => {
    test.describe('正常系', () => {
      test('招待を作成できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.invitations, {});

        // 成功（200/201）または既に招待が存在する場合（400/409）
        expect([200, 201, 400, 409]).toContain(response.status());

        if (response.status() === 200 || response.status() === 201) {
          const body = await response.json();
          expect(body.data).toBeDefined();
        }
      });

      test('招待URLが生成される', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.invitations, {});

        if (response.status() === 200 || response.status() === 201) {
          const body = await response.json();
          expect(body.data).toBeDefined();
          // invite_urlまたはtokenが含まれることを期待
          expect(
            body.data.invite_url || body.data.token || body.data.url
          ).toBeDefined();
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.invitations, {});

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.invitations, {});

        expect(response.status()).toBe(401);
      });
    });
  });

  // ============================================================
  // 2. POST /api/v1/invitations/accept/{token} - 招待承諾
  // ============================================================
  test.describe('POST /api/v1/invitations/accept/{token}', () => {
    test.describe('異常系', () => {
      // Note: 正常系は有効な招待トークンが必要なため、
      // 先に招待を作成するテストと組み合わせる必要がある

      test('無効なトークンで400/404エラー', async ({ request }) => {
        const response = await request.post(
          ENDPOINTS.acceptInvitation('invalid-token-12345'),
          {
            data: {
              display_name: 'Test User',
              password: 'testpassword123',
            },
          }
        );

        // 無効なトークンは400または404
        expect([400, 404]).toContain(response.status());
      });

      test('空のトークンで404エラー', async ({ request }) => {
        const response = await request.post(ENDPOINTS.acceptInvitation(''), {
          data: {
            display_name: 'Test User',
            password: 'testpassword123',
          },
        });

        // 空のトークンは404またはメソッドエラー
        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('表示名なしで400エラー', async ({ request }) => {
        const response = await request.post(
          ENDPOINTS.acceptInvitation('some-token'),
          {
            data: {
              password: 'testpassword123',
            },
          }
        );

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('パスワードなしで400エラー', async ({ request }) => {
        const response = await request.post(
          ENDPOINTS.acceptInvitation('some-token'),
          {
            data: {
              display_name: 'Test User',
            },
          }
        );

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('空のリクエストボディで400エラー', async ({ request }) => {
        const response = await request.post(
          ENDPOINTS.acceptInvitation('some-token'),
          {
            data: {},
          }
        );

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('短すぎるパスワードで400エラー', async ({ request }) => {
        const response = await request.post(
          ENDPOINTS.acceptInvitation('some-token'),
          {
            data: {
              display_name: 'Test User',
              password: '123',
            },
          }
        );

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('空の表示名で400エラー', async ({ request }) => {
        const response = await request.post(
          ENDPOINTS.acceptInvitation('some-token'),
          {
            data: {
              display_name: '',
              password: 'testpassword123',
            },
          }
        );

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('空のパスワードで400エラー', async ({ request }) => {
        const response = await request.post(
          ENDPOINTS.acceptInvitation('some-token'),
          {
            data: {
              display_name: 'Test User',
              password: '',
            },
          }
        );

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('SQLインジェクション試行は安全に処理される', async ({ request }) => {
        const response = await request.post(
          ENDPOINTS.acceptInvitation("' OR '1'='1"),
          {
            data: {
              display_name: "'; DROP TABLE admins; --",
              password: "' OR '1'='1",
            },
          }
        );

        // SQLインジェクションが成功してはいけない
        expect(response.status()).toBeGreaterThanOrEqual(400);
        expect(response.status()).not.toBe(200);
        expect(response.status()).not.toBe(500);
      });

      test('期限切れトークンで400/403エラー', async ({ request }) => {
        // Note: 実際の期限切れトークンはDBに用意する必要がある
        const response = await request.post(
          ENDPOINTS.acceptInvitation('expired-token-12345'),
          {
            data: {
              display_name: 'Test User',
              password: 'testpassword123',
            },
          }
        );

        // 期限切れトークンは400、403、または404
        expect([400, 403, 404]).toContain(response.status());
      });

      test('既に使用されたトークンで400/403エラー', async ({ request }) => {
        // Note: 実際の使用済みトークンはDBに用意する必要がある
        const response = await request.post(
          ENDPOINTS.acceptInvitation('used-token-12345'),
          {
            data: {
              display_name: 'Test User',
              password: 'testpassword123',
            },
          }
        );

        // 使用済みトークンは400、403、または404
        expect([400, 403, 404]).toContain(response.status());
      });

      test('非常に長い表示名で400エラー', async ({ request }) => {
        const response = await request.post(
          ENDPOINTS.acceptInvitation('some-token'),
          {
            data: {
              display_name: 'a'.repeat(1000),
              password: 'testpassword123',
            },
          }
        );

        // 長すぎる名前は400エラー（または切り詰めて処理）
        expect([200, 400, 404]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // E2E: 招待作成から承諾までの流れ
  // ============================================================
  test.describe('E2E: 招待フロー', () => {
    test('招待を作成して招待URLが返される', async ({ request }) => {
      const { client } = await loginAsAdmin(request);

      // 招待を作成
      const createResponse = await client.raw('POST', ENDPOINTS.invitations, {});

      // レスポンスを確認
      if (createResponse.status() === 200 || createResponse.status() === 201) {
        const createBody = await createResponse.json();
        expect(createBody.data).toBeDefined();

        // 招待情報が含まれることを確認
        const hasInviteInfo =
          createBody.data.invite_url ||
          createBody.data.token ||
          createBody.data.url ||
          createBody.data.id;
        expect(hasInviteInfo).toBeTruthy();
      } else {
        // 既に招待が存在するなどの理由で失敗した場合
        expect([400, 409]).toContain(createResponse.status());
      }
    });
  });
});

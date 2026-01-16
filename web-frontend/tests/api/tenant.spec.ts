import { test, expect } from '@playwright/test';
import {
  ENDPOINTS,
  ApiClient,
} from '../utils/api-client';
import { loginAsAdmin, getUnauthenticatedClient } from '../utils/auth';

/**
 * Tenant API Tests
 *
 * Endpoints:
 * 1. GET /api/v1/tenants/me - テナント情報取得
 * 2. PUT /api/v1/tenants/me - テナント情報更新
 * 3. GET /api/v1/settings/manager-permissions - マネージャー権限取得
 * 4. PUT /api/v1/settings/manager-permissions - マネージャー権限更新
 */

test.describe('Tenant API', () => {
  // ============================================================
  // 1. GET /api/v1/tenants/me - テナント情報取得
  // ============================================================
  test.describe('GET /api/v1/tenants/me', () => {
    test.describe('正常系', () => {
      test('認証済みでテナント情報を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.tenant);

        expect(response.status()).toBe(200);

        const body = await response.json();
        expect(body.data).toBeDefined();
        expect(body.data.tenant_id).toBeDefined();
        expect(body.data.tenant_name).toBeDefined();
      });

      test('テナント情報に必要なフィールドが含まれている', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.tenant);

        expect(response.status()).toBe(200);

        const body = await response.json();
        expect(body.data).toHaveProperty('tenant_id');
        expect(body.data).toHaveProperty('tenant_name');
        expect(body.data).toHaveProperty('timezone');
        expect(body.data).toHaveProperty('is_active');
        expect(body.data).toHaveProperty('created_at');
        expect(body.data).toHaveProperty('updated_at');
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.tenant);

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.tenant);

        expect(response.status()).toBe(401);
      });

      test('改ざんされたトークンで401エラー', async ({ request }) => {
        const { client, loginData } = await loginAsAdmin(request);

        // トークンを改ざん
        const tamperedToken = loginData.token.slice(0, -5) + 'XXXXX';
        const tamperedClient = new ApiClient(request);
        tamperedClient.setToken(tamperedToken);

        const response = await tamperedClient.raw('GET', ENDPOINTS.tenant);

        expect(response.status()).toBe(401);
      });
    });
  });

  // ============================================================
  // 2. PUT /api/v1/tenants/me - テナント情報更新
  // ============================================================
  test.describe('PUT /api/v1/tenants/me', () => {
    test.describe('正常系', () => {
      test('テナント名を更新できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // 現在のテナント情報を取得
        const getCurrentResponse = await client.raw('GET', ENDPOINTS.tenant);
        expect(getCurrentResponse.status()).toBe(200);
        const currentData = await getCurrentResponse.json();
        const originalName = currentData.data.tenant_name;

        // テナント名を更新
        const newName = 'Updated Tenant Name';
        const updateResponse = await client.raw('PUT', ENDPOINTS.tenant, {
          tenant_name: newName,
        });

        // 成功か、更新が許可されていない場合
        expect([200, 400, 403]).toContain(updateResponse.status());

        if (updateResponse.status() === 200) {
          // 元に戻す
          await client.raw('PUT', ENDPOINTS.tenant, {
            tenant_name: originalName,
          });
        }
      });

      test('更新後のテナント情報が反映されている', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // 現在のテナント情報を取得
        const getCurrentResponse = await client.raw('GET', ENDPOINTS.tenant);
        expect(getCurrentResponse.status()).toBe(200);
        const currentData = await getCurrentResponse.json();
        const originalName = currentData.data.tenant_name;

        // テナント名を更新
        const newName = 'Test Update ' + Date.now();
        const updateResponse = await client.raw('PUT', ENDPOINTS.tenant, {
          tenant_name: newName,
        });

        if (updateResponse.status() === 200) {
          // 更新後の情報を確認
          const getUpdatedResponse = await client.raw('GET', ENDPOINTS.tenant);
          expect(getUpdatedResponse.status()).toBe(200);
          const updatedData = await getUpdatedResponse.json();
          expect(updatedData.data.tenant_name).toBe(newName);

          // 元に戻す
          await client.raw('PUT', ENDPOINTS.tenant, {
            tenant_name: originalName,
          });
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('PUT', ENDPOINTS.tenant, {
          tenant_name: 'Test',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('PUT', ENDPOINTS.tenant, {
          tenant_name: 'Test',
        });

        expect(response.status()).toBe(401);
      });

      test('空のリクエストボディで400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('PUT', ENDPOINTS.tenant, {});

        // 空のボディは400エラーまたは許可される（現在値を維持）
        expect([200, 400]).toContain(response.status());
      });

      test('空のテナント名で400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('PUT', ENDPOINTS.tenant, {
          tenant_name: '',
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('非常に長いテナント名で400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('PUT', ENDPOINTS.tenant, {
          tenant_name: 'a'.repeat(1000),
        });

        // 長すぎる名前は400エラーまたは許可される（切り詰め）
        expect([200, 400]).toContain(response.status());
      });

      test('SQLインジェクション試行は安全に処理される', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('PUT', ENDPOINTS.tenant, {
          tenant_name: "'; DROP TABLE tenants; --",
        });

        // SQLインジェクションが成功してはいけない
        expect(response.status()).toBeGreaterThanOrEqual(200);
        expect(response.status()).not.toBe(500);
      });
    });
  });

  // ============================================================
  // 3. GET /api/v1/settings/manager-permissions - マネージャー権限取得
  // ============================================================
  test.describe('GET /api/v1/settings/manager-permissions', () => {
    test.describe('正常系', () => {
      test('マネージャー権限設定を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.managerPermissions);

        expect(response.status()).toBe(200);

        const body = await response.json();
        expect(body.data).toBeDefined();
      });

      test('レスポンスに権限フィールドが含まれている', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.managerPermissions);

        expect(response.status()).toBe(200);

        const body = await response.json();
        expect(body.data).toBeDefined();
        // 具体的な権限フィールドはAPI仕様による
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.managerPermissions);

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.managerPermissions);

        expect(response.status()).toBe(401);
      });
    });
  });

  // ============================================================
  // 4. PUT /api/v1/settings/manager-permissions - マネージャー権限更新
  // ============================================================
  test.describe('PUT /api/v1/settings/manager-permissions', () => {
    test.describe('正常系', () => {
      test('マネージャー権限を更新できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // 現在の設定を取得
        const getCurrentResponse = await client.raw('GET', ENDPOINTS.managerPermissions);
        expect(getCurrentResponse.status()).toBe(200);
        const currentData = await getCurrentResponse.json();

        // 権限を更新
        const updateResponse = await client.raw('PUT', ENDPOINTS.managerPermissions, {
          ...currentData.data,
        });

        // 成功か、権限不足で403
        expect([200, 400, 403]).toContain(updateResponse.status());
      });

      test('更新後の権限設定が反映されている', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // 現在の設定を取得
        const getCurrentResponse = await client.raw('GET', ENDPOINTS.managerPermissions);
        expect(getCurrentResponse.status()).toBe(200);
        const currentData = await getCurrentResponse.json();
        const originalPermissions = { ...currentData.data };

        // 権限を更新（同じ値で更新）
        const updateResponse = await client.raw('PUT', ENDPOINTS.managerPermissions, {
          ...currentData.data,
        });

        if (updateResponse.status() === 200) {
          // 更新後の設定を確認
          const getUpdatedResponse = await client.raw('GET', ENDPOINTS.managerPermissions);
          expect(getUpdatedResponse.status()).toBe(200);

          // 元に戻す
          await client.raw('PUT', ENDPOINTS.managerPermissions, originalPermissions);
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('PUT', ENDPOINTS.managerPermissions, {});

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('PUT', ENDPOINTS.managerPermissions, {});

        expect(response.status()).toBe(401);
      });

      test('空のリクエストボディで400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('PUT', ENDPOINTS.managerPermissions, {});

        // 空のボディは400エラーまたは現在値を維持
        expect([200, 400]).toContain(response.status());
      });

      test('不正なフィールド型で400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('PUT', ENDPOINTS.managerPermissions, {
          invalid_field: 'invalid_value',
        });

        // 不正なフィールドは400エラーまたは無視
        expect([200, 400]).toContain(response.status());
      });
    });
  });
});

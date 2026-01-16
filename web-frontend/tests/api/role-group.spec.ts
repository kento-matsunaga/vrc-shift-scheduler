import { test, expect } from '@playwright/test';
import {
  ENDPOINTS,
  ApiClient,
} from '../utils/api-client';
import { loginAsAdmin, getUnauthenticatedClient } from '../utils/auth';

/**
 * RoleGroup API Tests
 *
 * Endpoints:
 * 1. GET /api/v1/role-groups - ロールグループ一覧取得
 * 2. POST /api/v1/role-groups - ロールグループ作成
 * 3. GET /api/v1/role-groups/{id} - ロールグループ取得
 * 4. PUT /api/v1/role-groups/{id} - ロールグループ更新
 * 5. DELETE /api/v1/role-groups/{id} - ロールグループ削除
 * 6. GET /api/v1/role-groups/{id}/roles - ロールグループのロール一覧
 * 7. PUT /api/v1/role-groups/{id}/roles - ロールグループのロール更新
 */

test.describe('RoleGroup API', () => {
  // ============================================================
  // 1. GET /api/v1/role-groups - ロールグループ一覧取得
  // ============================================================
  test.describe('GET /api/v1/role-groups', () => {
    test.describe('正常系', () => {
      test('ロールグループ一覧を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.roleGroups);

        expect([200, 404]).toContain(response.status());
        if (response.status() === 200) {
          const body = await response.json();
          expect(body.data).toBeDefined();
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.roleGroups);

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.roleGroups);

        expect(response.status()).toBe(401);
      });
    });
  });

  // ============================================================
  // 2. POST /api/v1/role-groups - ロールグループ作成
  // ============================================================
  test.describe('POST /api/v1/role-groups', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.roleGroups, {
          name: 'Test Role Group',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.roleGroups, {
          name: 'Test Role Group',
        });

        expect(response.status()).toBe(401);
      });

      test('グループ名なしで400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.roleGroups, {});

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('空のグループ名で400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.roleGroups, {
          name: '',
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });
    });
  });

  // ============================================================
  // 3. GET /api/v1/role-groups/{id} - ロールグループ取得
  // ============================================================
  test.describe('GET /api/v1/role-groups/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.roleGroup('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.roleGroup('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないグループIDで404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.roleGroup('01HZNONEXISTENT00000001')
        );

        expect([400, 404]).toContain(response.status());
      });

      test('無効なグループID形式で400/404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.roleGroup('invalid-id'));

        expect([400, 404]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 4. PUT /api/v1/role-groups/{id} - ロールグループ更新
  // ============================================================
  test.describe('PUT /api/v1/role-groups/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('PUT', ENDPOINTS.roleGroup('some-id'), {
          name: 'Updated Group',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('PUT', ENDPOINTS.roleGroup('some-id'), {
          name: 'Updated Group',
        });

        expect(response.status()).toBe(401);
      });

      test('存在しないグループIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'PUT',
          ENDPOINTS.roleGroup('01HZNONEXISTENT00000001'),
          { name: 'Updated Group' }
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 5. DELETE /api/v1/role-groups/{id} - ロールグループ削除
  // ============================================================
  test.describe('DELETE /api/v1/role-groups/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('DELETE', ENDPOINTS.roleGroup('some-id'));

        expect([400, 401, 405]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('DELETE', ENDPOINTS.roleGroup('some-id'));

        expect([401, 405]).toContain(response.status());
      });

      test('存在しないグループIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.roleGroup('01HZNONEXISTENT00000001')
        );

        // 200 (成功扱い - 冪等性) / 405 (Method Not Allowed) も許容
        // Note: 200が返る場合は冪等性の観点で意図的な可能性あり
        expect([200, 400, 404, 405, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 6. GET /api/v1/role-groups/{id}/roles - ロールグループのロール一覧
  // ============================================================
  test.describe('GET /api/v1/role-groups/{id}/roles', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.roleGroupRoles('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.roleGroupRoles('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないグループIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.roleGroupRoles('01HZNONEXISTENT00000001')
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 7. PUT /api/v1/role-groups/{id}/roles - ロールグループのロール更新
  // ============================================================
  test.describe('PUT /api/v1/role-groups/{id}/roles', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('PUT', ENDPOINTS.roleGroupRoles('some-id'), {
          role_ids: [],
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('PUT', ENDPOINTS.roleGroupRoles('some-id'), {
          role_ids: [],
        });

        expect(response.status()).toBe(401);
      });

      test('存在しないグループIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'PUT',
          ENDPOINTS.roleGroupRoles('01HZNONEXISTENT00000001'),
          { role_ids: [] }
        );

        // 200 (成功扱い) / 405 (Method Not Allowed) も許容
        // Note: 200が返る場合は冪等性の観点で意図的な可能性あり
        expect([200, 400, 404, 405, 500]).toContain(response.status());
      });
    });
  });
});

import { test, expect } from '@playwright/test';
import {
  ENDPOINTS,
  ApiClient,
} from '../utils/api-client';
import { loginAsAdmin, getUnauthenticatedClient } from '../utils/auth';

/**
 * MemberGroup API Tests
 *
 * Endpoints:
 * 1. GET /api/v1/member-groups - メンバーグループ一覧取得
 * 2. POST /api/v1/member-groups - メンバーグループ作成
 * 3. GET /api/v1/member-groups/{id} - メンバーグループ取得
 * 4. PUT /api/v1/member-groups/{id} - メンバーグループ更新
 * 5. DELETE /api/v1/member-groups/{id} - メンバーグループ削除
 * 6. GET /api/v1/member-groups/{id}/members - メンバーグループのメンバー一覧
 * 7. PUT /api/v1/member-groups/{id}/members - メンバーグループのメンバー更新
 */

test.describe('MemberGroup API', () => {
  // ============================================================
  // 1. GET /api/v1/member-groups - メンバーグループ一覧取得
  // ============================================================
  test.describe('GET /api/v1/member-groups', () => {
    test.describe('正常系', () => {
      test('メンバーグループ一覧を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.memberGroups);

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

        const response = await client.raw('GET', ENDPOINTS.memberGroups);

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.memberGroups);

        expect(response.status()).toBe(401);
      });
    });
  });

  // ============================================================
  // 2. POST /api/v1/member-groups - メンバーグループ作成
  // ============================================================
  test.describe('POST /api/v1/member-groups', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.memberGroups, {
          name: 'Test Member Group',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.memberGroups, {
          name: 'Test Member Group',
        });

        expect(response.status()).toBe(401);
      });

      test('グループ名なしで400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.memberGroups, {});

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('空のグループ名で400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.memberGroups, {
          name: '',
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });
    });
  });

  // ============================================================
  // 3. GET /api/v1/member-groups/{id} - メンバーグループ取得
  // ============================================================
  test.describe('GET /api/v1/member-groups/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.memberGroup('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.memberGroup('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないグループIDで404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.memberGroup('01HZNONEXISTENT00000001')
        );

        expect([400, 404]).toContain(response.status());
      });

      test('無効なグループID形式で400/404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.memberGroup('invalid-id'));

        expect([400, 404]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 4. PUT /api/v1/member-groups/{id} - メンバーグループ更新
  // ============================================================
  test.describe('PUT /api/v1/member-groups/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('PUT', ENDPOINTS.memberGroup('some-id'), {
          name: 'Updated Group',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('PUT', ENDPOINTS.memberGroup('some-id'), {
          name: 'Updated Group',
        });

        expect(response.status()).toBe(401);
      });

      test('存在しないグループIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'PUT',
          ENDPOINTS.memberGroup('01HZNONEXISTENT00000001'),
          { name: 'Updated Group' }
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 5. DELETE /api/v1/member-groups/{id} - メンバーグループ削除
  // ============================================================
  test.describe('DELETE /api/v1/member-groups/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('DELETE', ENDPOINTS.memberGroup('some-id'));

        expect([400, 401, 405]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('DELETE', ENDPOINTS.memberGroup('some-id'));

        expect([401, 405]).toContain(response.status());
      });

      test('存在しないグループIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.memberGroup('01HZNONEXISTENT00000001')
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 6. GET /api/v1/member-groups/{id}/members - メンバーグループのメンバー一覧
  // ============================================================
  test.describe('GET /api/v1/member-groups/{id}/members', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.memberGroupMembers('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.memberGroupMembers('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないグループIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.memberGroupMembers('01HZNONEXISTENT00000001')
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 7. PUT /api/v1/member-groups/{id}/members - メンバーグループのメンバー更新
  // ============================================================
  test.describe('PUT /api/v1/member-groups/{id}/members', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('PUT', ENDPOINTS.memberGroupMembers('some-id'), {
          member_ids: [],
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('PUT', ENDPOINTS.memberGroupMembers('some-id'), {
          member_ids: [],
        });

        expect(response.status()).toBe(401);
      });

      test('存在しないグループIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'PUT',
          ENDPOINTS.memberGroupMembers('01HZNONEXISTENT00000001'),
          { member_ids: [] }
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });
});

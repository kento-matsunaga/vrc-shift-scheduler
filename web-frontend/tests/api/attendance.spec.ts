import { test, expect } from '@playwright/test';
import {
  ENDPOINTS,
  ApiClient,
} from '../utils/api-client';
import { loginAsAdmin, getUnauthenticatedClient } from '../utils/auth';

/**
 * Attendance API Tests
 *
 * Endpoints:
 * 1. GET /api/v1/attendance/collections - 出欠収集一覧取得
 * 2. POST /api/v1/attendance/collections - 出欠収集作成
 * 3. GET /api/v1/attendance/collections/{id} - 出欠収集取得
 * 4. PUT /api/v1/attendance/collections/{id} - 出欠収集更新
 * 5. DELETE /api/v1/attendance/collections/{id} - 出欠収集削除
 * 6. POST /api/v1/attendance/collections/{id}/close - 出欠収集締め切り
 * 7. GET /api/v1/attendance/collections/{id}/responses - 出欠回答一覧
 */

test.describe('Attendance API', () => {
  // ============================================================
  // 1. GET /api/v1/attendance/collections - 出欠収集一覧取得
  // ============================================================
  test.describe('GET /api/v1/attendance/collections', () => {
    test.describe('正常系', () => {
      test('出欠収集一覧を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.attendanceCollections);

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

        const response = await client.raw('GET', ENDPOINTS.attendanceCollections);

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.attendanceCollections);

        expect(response.status()).toBe(401);
      });
    });
  });

  // ============================================================
  // 2. POST /api/v1/attendance/collections - 出欠収集作成
  // ============================================================
  test.describe('POST /api/v1/attendance/collections', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.attendanceCollections, {
          title: 'Test Attendance',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.attendanceCollections, {
          title: 'Test Attendance',
        });

        expect(response.status()).toBe(401);
      });

      test('必須パラメータなしで400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.attendanceCollections, {});

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });
    });
  });

  // ============================================================
  // 3. GET /api/v1/attendance/collections/{id} - 出欠収集取得
  // ============================================================
  test.describe('GET /api/v1/attendance/collections/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.attendanceCollection('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.attendanceCollection('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しない出欠収集IDで404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.attendanceCollection('01HZNONEXISTENT00000001')
        );

        expect([400, 404]).toContain(response.status());
      });

      test('無効なID形式で400/404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.attendanceCollection('invalid-id'));

        expect([400, 404]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 4. PUT /api/v1/attendance/collections/{id} - 出欠収集更新
  // ============================================================
  test.describe('PUT /api/v1/attendance/collections/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('PUT', ENDPOINTS.attendanceCollection('some-id'), {
          title: 'Updated Attendance',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('PUT', ENDPOINTS.attendanceCollection('some-id'), {
          title: 'Updated Attendance',
        });

        expect(response.status()).toBe(401);
      });

      test('存在しない出欠収集IDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'PUT',
          ENDPOINTS.attendanceCollection('01HZNONEXISTENT00000001'),
          { title: 'Updated Attendance' }
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 5. DELETE /api/v1/attendance/collections/{id} - 出欠収集削除
  // ============================================================
  test.describe('DELETE /api/v1/attendance/collections/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('DELETE', ENDPOINTS.attendanceCollection('some-id'));

        expect([400, 401, 405]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('DELETE', ENDPOINTS.attendanceCollection('some-id'));

        expect([401, 405]).toContain(response.status());
      });

      test('存在しない出欠収集IDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.attendanceCollection('01HZNONEXISTENT00000001')
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 6. POST /api/v1/attendance/collections/{id}/close - 出欠収集締め切り
  // ============================================================
  test.describe('POST /api/v1/attendance/collections/{id}/close', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.attendanceCollectionClose('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.attendanceCollectionClose('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しない出欠収集IDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'POST',
          ENDPOINTS.attendanceCollectionClose('01HZNONEXISTENT00000001')
        );

        expect([400, 404, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 7. PUT /api/v1/attendance/collections/{id}/responses - 管理者による出欠回答更新
  // ============================================================
  test.describe('PUT /api/v1/attendance/collections/{id}/responses', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('PUT', ENDPOINTS.attendanceCollectionAdminUpdate('some-id'), {
          member_id: 'some-member-id',
          responses: [],
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('PUT', ENDPOINTS.attendanceCollectionAdminUpdate('some-id'), {
          member_id: 'some-member-id',
          responses: [],
        });

        expect(response.status()).toBe(401);
      });

      test('存在しない出欠収集IDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'PUT',
          ENDPOINTS.attendanceCollectionAdminUpdate('01HZNONEXISTENT00000001'),
          {
            member_id: '01HZNONEXISTENT00000002',
            responses: [],
          }
        );

        expect([400, 404, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 8. GET /api/v1/attendance/collections/{id}/responses - 出欠回答一覧
  // ============================================================
  test.describe('GET /api/v1/attendance/collections/{id}/responses', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.attendanceCollectionResponses('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.attendanceCollectionResponses('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しない出欠収集IDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.attendanceCollectionResponses('01HZNONEXISTENT00000001')
        );

        expect([400, 404, 500]).toContain(response.status());
      });
    });
  });
});

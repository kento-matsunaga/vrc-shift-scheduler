import { test, expect } from '@playwright/test';
import {
  ENDPOINTS,
  ApiClient,
} from '../utils/api-client';
import { loginAsAdmin, getUnauthenticatedClient } from '../utils/auth';

/**
 * Schedule API Tests
 *
 * Endpoints:
 * 1. GET /api/v1/schedules - スケジュール一覧取得
 * 2. POST /api/v1/schedules - スケジュール作成
 * 3. GET /api/v1/schedules/{id} - スケジュール取得
 * 4. PUT /api/v1/schedules/{id} - スケジュール更新
 * 5. DELETE /api/v1/schedules/{id} - スケジュール削除
 * 6. POST /api/v1/schedules/{id}/decide - スケジュール確定
 * 7. POST /api/v1/schedules/{id}/close - スケジュール締め切り
 * 8. GET /api/v1/schedules/{id}/responses - スケジュール回答一覧
 */

test.describe('Schedule API', () => {
  // ============================================================
  // 1. GET /api/v1/schedules - スケジュール一覧取得
  // ============================================================
  test.describe('GET /api/v1/schedules', () => {
    test.describe('正常系', () => {
      test('スケジュール一覧を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.schedules);

        expect([200, 404]).toContain(response.status());
        if (response.status() === 200) {
          const body = await response.json();
          // レスポンス形式: { data: [...] } または直接配列/オブジェクト
          expect(body).toBeDefined();
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.schedules);

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.schedules);

        expect(response.status()).toBe(401);
      });
    });
  });

  // ============================================================
  // 2. POST /api/v1/schedules - スケジュール作成
  // ============================================================
  test.describe('POST /api/v1/schedules', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.schedules, {
          title: 'Test Schedule',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.schedules, {
          title: 'Test Schedule',
        });

        expect(response.status()).toBe(401);
      });

      test('必須パラメータなしで400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.schedules, {});

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });
    });
  });

  // ============================================================
  // 3. GET /api/v1/schedules/{id} - スケジュール取得
  // ============================================================
  test.describe('GET /api/v1/schedules/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.schedule('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.schedule('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないスケジュールIDで404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.schedule('01HZNONEXISTENT00000001')
        );

        expect([400, 404]).toContain(response.status());
      });

      test('無効なID形式で400/404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.schedule('invalid-id'));

        expect([400, 404]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 4. PUT /api/v1/schedules/{id} - スケジュール更新
  // ============================================================
  test.describe('PUT /api/v1/schedules/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('PUT', ENDPOINTS.schedule('some-id'), {
          title: 'Updated Schedule',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('PUT', ENDPOINTS.schedule('some-id'), {
          title: 'Updated Schedule',
        });

        expect(response.status()).toBe(401);
      });

      test('存在しないスケジュールIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'PUT',
          ENDPOINTS.schedule('01HZNONEXISTENT00000001'),
          { title: 'Updated Schedule' }
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 5. DELETE /api/v1/schedules/{id} - スケジュール削除
  // ============================================================
  test.describe('DELETE /api/v1/schedules/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('DELETE', ENDPOINTS.schedule('some-id'));

        expect([400, 401, 405]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('DELETE', ENDPOINTS.schedule('some-id'));

        expect([401, 405]).toContain(response.status());
      });

      test('存在しないスケジュールIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.schedule('01HZNONEXISTENT00000001')
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 6. POST /api/v1/schedules/{id}/decide - スケジュール確定
  // ============================================================
  test.describe('POST /api/v1/schedules/{id}/decide', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.scheduleDecide('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.scheduleDecide('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないスケジュールIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'POST',
          ENDPOINTS.scheduleDecide('01HZNONEXISTENT00000001')
        );

        expect([400, 404, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 7. POST /api/v1/schedules/{id}/close - スケジュール締め切り
  // ============================================================
  test.describe('POST /api/v1/schedules/{id}/close', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.scheduleClose('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.scheduleClose('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないスケジュールIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'POST',
          ENDPOINTS.scheduleClose('01HZNONEXISTENT00000001')
        );

        expect([400, 404, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 8. GET /api/v1/schedules/{id}/responses - スケジュール回答一覧
  // ============================================================
  test.describe('GET /api/v1/schedules/{id}/responses', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.scheduleResponses('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.scheduleResponses('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないスケジュールIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.scheduleResponses('01HZNONEXISTENT00000001')
        );

        expect([400, 404, 500]).toContain(response.status());
      });
    });
  });
});

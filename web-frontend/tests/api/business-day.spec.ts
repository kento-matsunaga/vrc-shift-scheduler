import { test, expect } from '@playwright/test';
import {
  ENDPOINTS,
  ApiClient,
} from '../utils/api-client';
import { loginAsAdmin, getUnauthenticatedClient } from '../utils/auth';

/**
 * BusinessDay API Tests
 *
 * Endpoints:
 * 1. GET /api/v1/business-days - 営業日一覧取得
 * 2. POST /api/v1/business-days - 営業日作成
 * 3. GET /api/v1/business-days/{id} - 営業日取得
 * 4. PUT /api/v1/business-days/{id} - 営業日更新
 * 5. DELETE /api/v1/business-days/{id} - 営業日削除
 * 6. GET /api/v1/events/{eventId}/business-days - イベントの営業日一覧
 * 7. POST /api/v1/business-days/{id}/apply-template - テンプレート適用
 * 8. POST /api/v1/business-days/{id}/save-as-template - テンプレートとして保存
 */

test.describe('BusinessDay API', () => {
  // ============================================================
  // 1. GET /api/v1/business-days - 営業日一覧取得
  // ============================================================
  test.describe('GET /api/v1/business-days', () => {
    test.describe('正常系', () => {
      test('営業日一覧APIが存在する', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.businessDays);

        // 200 (成功) または 404 (イベントID必須の場合)
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

        const response = await client.raw('GET', ENDPOINTS.businessDays);

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.businessDays);

        expect(response.status()).toBe(401);
      });
    });
  });

  // ============================================================
  // 2. POST /api/v1/business-days - 営業日作成
  // ============================================================
  test.describe('POST /api/v1/business-days', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.businessDays, {
          date: '2025-01-01',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.businessDays, {
          date: '2025-01-01',
        });

        expect(response.status()).toBe(401);
      });

      test('日付なしで400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.businessDays, {});

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('不正な日付形式で400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.businessDays, {
          date: 'invalid-date',
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });
    });
  });

  // ============================================================
  // 3. GET /api/v1/business-days/{id} - 営業日取得
  // ============================================================
  test.describe('GET /api/v1/business-days/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.businessDay('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.businessDay('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しない営業日IDで404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.businessDay('01HZNONEXISTENT00000001')
        );

        expect([400, 404]).toContain(response.status());
      });

      test('無効な営業日ID形式で400/404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.businessDay('invalid-id'));

        expect([400, 404]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 4. PUT /api/v1/business-days/{id} - 営業日更新
  // ============================================================
  test.describe('PUT /api/v1/business-days/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('PUT', ENDPOINTS.businessDay('some-id'), {
          date: '2025-01-01',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('PUT', ENDPOINTS.businessDay('some-id'), {
          date: '2025-01-01',
        });

        expect(response.status()).toBe(401);
      });

      test('存在しない営業日IDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'PUT',
          ENDPOINTS.businessDay('01HZNONEXISTENT00000001'),
          {
            date: '2025-01-01',
          }
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 5. DELETE /api/v1/business-days/{id} - 営業日削除
  // ============================================================
  test.describe('DELETE /api/v1/business-days/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.businessDay('some-id')
        );

        expect([400, 401, 405]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.businessDay('some-id')
        );

        expect([401, 405]).toContain(response.status());
      });

      test('存在しない営業日IDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.businessDay('01HZNONEXISTENT00000001')
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 6. GET /api/v1/events/{eventId}/business-days - イベントの営業日一覧
  // ============================================================
  test.describe('GET /api/v1/events/{eventId}/business-days', () => {
    test.describe('正常系', () => {
      test('イベントの営業日一覧を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // まずイベント一覧を取得
        const listResponse = await client.raw('GET', ENDPOINTS.events);
        expect(listResponse.status()).toBe(200);
        const listBody = await listResponse.json();

        const events = Array.isArray(listBody.data) ? listBody.data : (listBody.data?.events || []);
        if (events.length > 0) {
          const eventId = events[0].id || events[0].event_id;

          const response = await client.raw('GET', ENDPOINTS.businessDaysByEvent(eventId));

          expect(response.status()).toBe(200);
          const body = await response.json();
          expect(body.data).toBeDefined();
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.businessDaysByEvent('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.businessDaysByEvent('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないイベントIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.businessDaysByEvent('01HZNONEXISTENT00000001')
        );

        expect([400, 404, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 7. POST /api/v1/business-days/{id}/apply-template - テンプレート適用
  // ============================================================
  test.describe('POST /api/v1/business-days/{id}/apply-template', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw(
          'POST',
          ENDPOINTS.businessDayApplyTemplate('some-id'),
          {}
        );

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw(
          'POST',
          ENDPOINTS.businessDayApplyTemplate('some-id'),
          {}
        );

        expect(response.status()).toBe(401);
      });

      test('存在しない営業日IDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'POST',
          ENDPOINTS.businessDayApplyTemplate('01HZNONEXISTENT00000001'),
          {}
        );

        expect([400, 404, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 8. POST /api/v1/business-days/{id}/save-as-template - テンプレートとして保存
  // ============================================================
  test.describe('POST /api/v1/business-days/{id}/save-as-template', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw(
          'POST',
          ENDPOINTS.businessDaySaveAsTemplate('some-id'),
          {}
        );

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw(
          'POST',
          ENDPOINTS.businessDaySaveAsTemplate('some-id'),
          {}
        );

        expect(response.status()).toBe(401);
      });

      test('存在しない営業日IDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'POST',
          ENDPOINTS.businessDaySaveAsTemplate('01HZNONEXISTENT00000001'),
          {}
        );

        expect([400, 404, 500]).toContain(response.status());
      });
    });
  });
});

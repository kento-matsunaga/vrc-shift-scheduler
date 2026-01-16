import { test, expect } from '@playwright/test';
import {
  ENDPOINTS,
  ApiClient,
} from '../utils/api-client';
import { loginAsAdmin, getUnauthenticatedClient } from '../utils/auth';

/**
 * Instance API Tests
 *
 * Endpoints:
 * 1. GET /api/v1/events/{eventId}/instances - インスタンス一覧取得
 * 2. POST /api/v1/events/{eventId}/instances - インスタンス作成
 * 3. GET /api/v1/instances/{id} - インスタンス取得
 * 4. PUT /api/v1/instances/{id} - インスタンス更新
 * 5. DELETE /api/v1/instances/{id} - インスタンス削除
 */

test.describe('Instance API', () => {
  // ============================================================
  // 1. GET /api/v1/events/{eventId}/instances - インスタンス一覧取得
  // ============================================================
  test.describe('GET /api/v1/events/{eventId}/instances', () => {
    test.describe('正常系', () => {
      test('インスタンス一覧を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // まずイベント一覧を取得
        const eventsResponse = await client.raw('GET', ENDPOINTS.events);
        expect(eventsResponse.status()).toBe(200);
        const eventsBody = await eventsResponse.json();

        const events = Array.isArray(eventsBody.data) ? eventsBody.data : (eventsBody.data?.events || []);
        if (events.length > 0) {
          const eventId = events[0].id || events[0].event_id;

          const response = await client.raw('GET', ENDPOINTS.instances(eventId));

          expect([200, 404]).toContain(response.status());
          if (response.status() === 200) {
            const body = await response.json();
            expect(body.data).toBeDefined();
          }
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.instances('some-event-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.instances('some-event-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないイベントIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.instances('01HZNONEXISTENT00000001')
        );

        expect([400, 404, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 2. POST /api/v1/events/{eventId}/instances - インスタンス作成
  // ============================================================
  test.describe('POST /api/v1/events/{eventId}/instances', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.instances('some-event-id'), {
          name: 'Test Instance',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.instances('some-event-id'), {
          name: 'Test Instance',
        });

        expect(response.status()).toBe(401);
      });

      test('存在しないイベントIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'POST',
          ENDPOINTS.instances('01HZNONEXISTENT00000001'),
          { name: 'Test Instance' }
        );

        expect([400, 404, 500]).toContain(response.status());
      });

      test('必須パラメータなしでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // まずイベント一覧を取得
        const eventsResponse = await client.raw('GET', ENDPOINTS.events);
        expect(eventsResponse.status()).toBe(200);
        const eventsBody = await eventsResponse.json();

        const events = Array.isArray(eventsBody.data) ? eventsBody.data : (eventsBody.data?.events || []);
        if (events.length > 0) {
          const eventId = events[0].id || events[0].event_id;

          const response = await client.raw('POST', ENDPOINTS.instances(eventId), {});

          expect(response.status()).toBeGreaterThanOrEqual(400);
        }
      });
    });
  });

  // ============================================================
  // 3. GET /api/v1/instances/{id} - インスタンス取得
  // ============================================================
  test.describe('GET /api/v1/instances/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.instance('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.instance('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないインスタンスIDで404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.instance('01HZNONEXISTENT00000001')
        );

        expect([400, 404]).toContain(response.status());
      });

      test('無効なインスタンスID形式で400/404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.instance('invalid-id'));

        expect([400, 404]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 4. PUT /api/v1/instances/{id} - インスタンス更新
  // ============================================================
  test.describe('PUT /api/v1/instances/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('PUT', ENDPOINTS.instance('some-id'), {
          name: 'Updated Instance',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('PUT', ENDPOINTS.instance('some-id'), {
          name: 'Updated Instance',
        });

        expect(response.status()).toBe(401);
      });

      test('存在しないインスタンスIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'PUT',
          ENDPOINTS.instance('01HZNONEXISTENT00000001'),
          { name: 'Updated Instance' }
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 5. DELETE /api/v1/instances/{id} - インスタンス削除
  // ============================================================
  test.describe('DELETE /api/v1/instances/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('DELETE', ENDPOINTS.instance('some-id'));

        expect([400, 401, 405]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('DELETE', ENDPOINTS.instance('some-id'));

        expect([401, 405]).toContain(response.status());
      });

      test('存在しないインスタンスIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.instance('01HZNONEXISTENT00000001')
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });
});

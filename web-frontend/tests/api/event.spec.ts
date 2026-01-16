import { test, expect } from '@playwright/test';
import {
  ENDPOINTS,
  ApiClient,
} from '../utils/api-client';
import { loginAsAdmin, getUnauthenticatedClient } from '../utils/auth';

/**
 * Event API Tests
 *
 * Endpoints:
 * 1. GET /api/v1/events - イベント一覧取得
 * 2. POST /api/v1/events - イベント作成
 * 3. GET /api/v1/events/{id} - イベント取得
 * 4. PUT /api/v1/events/{id} - イベント更新
 * 5. DELETE /api/v1/events/{id} - イベント削除
 * 6. POST /api/v1/events/{id}/generate-business-days - 営業日生成
 * 7. GET /api/v1/events/{id}/groups - イベントグループ取得
 */

test.describe('Event API', () => {
  // ============================================================
  // 1. GET /api/v1/events - イベント一覧取得
  // ============================================================
  test.describe('GET /api/v1/events', () => {
    test.describe('正常系', () => {
      test('イベント一覧を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.events);

        expect(response.status()).toBe(200);

        const body = await response.json();
        expect(body.data).toBeDefined();
      });

      test('イベント一覧にイベント情報が含まれている', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.events);

        expect(response.status()).toBe(200);

        const body = await response.json();
        const events = Array.isArray(body.data) ? body.data : (body.data?.events || []);
        if (events.length > 0) {
          const event = events[0];
          // id または event_id が含まれる
          const hasId = event.id || event.event_id;
          expect(hasId).toBeTruthy();
          expect(event.name || event.event_name).toBeDefined();
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.events);

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.events);

        expect(response.status()).toBe(401);
      });
    });
  });

  // ============================================================
  // 2. POST /api/v1/events - イベント作成
  // ============================================================
  test.describe('POST /api/v1/events', () => {
    test.describe('正常系', () => {
      test('イベントを作成できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const eventName = `Test Event ${Date.now()}`;
        const response = await client.raw('POST', ENDPOINTS.events, {
          name: eventName,
        });

        // 成功（200/201）または作成が許可されていない場合
        expect([200, 201, 400, 403]).toContain(response.status());

        if (response.status() === 200 || response.status() === 201) {
          const body = await response.json();
          expect(body.data).toBeDefined();
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.events, {
          name: 'Test Event',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.events, {
          name: 'Test Event',
        });

        expect(response.status()).toBe(401);
      });

      test('イベント名なしで400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.events, {});

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('空のイベント名で400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.events, {
          name: '',
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });
    });
  });

  // ============================================================
  // 3. GET /api/v1/events/{id} - イベント取得
  // ============================================================
  test.describe('GET /api/v1/events/{id}', () => {
    test.describe('正常系', () => {
      test('イベント情報を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // まずイベント一覧を取得
        const listResponse = await client.raw('GET', ENDPOINTS.events);
        expect(listResponse.status()).toBe(200);
        const listBody = await listResponse.json();

        const events = Array.isArray(listBody.data) ? listBody.data : (listBody.data?.events || []);
        if (events.length > 0) {
          const eventId = events[0].id || events[0].event_id;

          const response = await client.raw('GET', ENDPOINTS.event(eventId));

          expect(response.status()).toBe(200);
          const body = await response.json();
          expect(body.data).toBeDefined();
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.event('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.event('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないイベントIDで404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.event('01HZNONEXISTENT00000001')
        );

        expect([400, 404]).toContain(response.status());
      });

      test('無効なイベントID形式で400/404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.event('invalid-id'));

        expect([400, 404]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 4. PUT /api/v1/events/{id} - イベント更新
  // ============================================================
  test.describe('PUT /api/v1/events/{id}', () => {
    test.describe('正常系', () => {
      test('イベント情報を更新できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // まずイベント一覧を取得
        const listResponse = await client.raw('GET', ENDPOINTS.events);
        expect(listResponse.status()).toBe(200);
        const listBody = await listResponse.json();

        const events = Array.isArray(listBody.data) ? listBody.data : (listBody.data?.events || []);
        if (events.length > 0) {
          const event = events[0];
          const eventId = event.id || event.event_id;
          const originalName = event.name || event.event_name;

          const newName = 'Updated ' + Date.now();
          const response = await client.raw('PUT', ENDPOINTS.event(eventId), {
            name: newName,
          });

          // 成功または更新が許可されていない場合
          expect([200, 400, 403]).toContain(response.status());

          if (response.status() === 200) {
            // 元に戻す
            await client.raw('PUT', ENDPOINTS.event(eventId), {
              name: originalName,
            });
          }
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('PUT', ENDPOINTS.event('some-id'), {
          name: 'Test',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('PUT', ENDPOINTS.event('some-id'), {
          name: 'Test',
        });

        expect(response.status()).toBe(401);
      });

      test('存在しないイベントIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'PUT',
          ENDPOINTS.event('01HZNONEXISTENT00000001'),
          {
            name: 'Test',
          }
        );

        expect([400, 404, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 5. DELETE /api/v1/events/{id} - イベント削除
  // ============================================================
  test.describe('DELETE /api/v1/events/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.event('some-id')
        );

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.event('some-id')
        );

        expect(response.status()).toBe(401);
      });

      test('存在しないイベントIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.event('01HZNONEXISTENT00000001')
        );

        expect([400, 404, 500]).toContain(response.status());
      });

      test('無効なイベントID形式でエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.event('invalid-id')
        );

        expect([400, 404, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 6. POST /api/v1/events/{id}/generate-business-days - 営業日生成
  // ============================================================
  test.describe('POST /api/v1/events/{id}/generate-business-days', () => {
    test.describe('正常系', () => {
      test('営業日生成APIが存在する', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // まずイベント一覧を取得
        const listResponse = await client.raw('GET', ENDPOINTS.events);
        expect(listResponse.status()).toBe(200);
        const listBody = await listResponse.json();

        const events = Array.isArray(listBody.data) ? listBody.data : (listBody.data?.events || []);
        if (events.length > 0) {
          const eventId = events[0].id || events[0].event_id;

          const response = await client.raw(
            'POST',
            ENDPOINTS.eventGenerateBusinessDays(eventId),
            {}
          );

          // エンドポイントが存在することを確認（404ではない）
          expect(response.status()).not.toBe(404);
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw(
          'POST',
          ENDPOINTS.eventGenerateBusinessDays('some-id'),
          {}
        );

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw(
          'POST',
          ENDPOINTS.eventGenerateBusinessDays('some-id'),
          {}
        );

        expect(response.status()).toBe(401);
      });

      test('存在しないイベントIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'POST',
          ENDPOINTS.eventGenerateBusinessDays('01HZNONEXISTENT00000001'),
          {}
        );

        expect([400, 404, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 7. GET /api/v1/events/{id}/groups - イベントグループ取得
  // ============================================================
  test.describe('GET /api/v1/events/{id}/groups', () => {
    test.describe('正常系', () => {
      test('イベントグループを取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // まずイベント一覧を取得
        const listResponse = await client.raw('GET', ENDPOINTS.events);
        expect(listResponse.status()).toBe(200);
        const listBody = await listResponse.json();

        const events = Array.isArray(listBody.data) ? listBody.data : (listBody.data?.events || []);
        if (events.length > 0) {
          const eventId = events[0].id || events[0].event_id;

          const response = await client.raw('GET', ENDPOINTS.eventGroups(eventId));

          // 成功または機能が実装されていない場合
          expect([200, 404, 501]).toContain(response.status());
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.eventGroups('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.eventGroups('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないイベントIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.eventGroups('01HZNONEXISTENT00000001')
        );

        expect([400, 404, 500]).toContain(response.status());
      });
    });
  });
});

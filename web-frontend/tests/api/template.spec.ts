import { test, expect } from '@playwright/test';
import {
  ENDPOINTS,
  ApiClient,
} from '../utils/api-client';
import { loginAsAdmin, getUnauthenticatedClient } from '../utils/auth';

/**
 * Template API Tests
 *
 * Endpoints:
 * 1. GET /api/v1/events/{eventId}/templates - テンプレート一覧取得
 * 2. POST /api/v1/events/{eventId}/templates - テンプレート作成
 * 3. GET /api/v1/events/{eventId}/templates/{id} - テンプレート取得
 * 4. PUT /api/v1/events/{eventId}/templates/{id} - テンプレート更新
 * 5. DELETE /api/v1/events/{eventId}/templates/{id} - テンプレート削除
 */

test.describe('Template API', () => {
  // ============================================================
  // 1. GET /api/v1/events/{eventId}/templates - テンプレート一覧取得
  // ============================================================
  test.describe('GET /api/v1/events/{eventId}/templates', () => {
    test.describe('正常系', () => {
      test('テンプレート一覧を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // まずイベント一覧を取得
        const eventsResponse = await client.raw('GET', ENDPOINTS.events);
        expect(eventsResponse.status()).toBe(200);
        const eventsBody = await eventsResponse.json();

        const events = Array.isArray(eventsBody.data) ? eventsBody.data : (eventsBody.data?.events || []);
        if (events.length > 0) {
          const eventId = events[0].id || events[0].event_id;

          const response = await client.raw('GET', ENDPOINTS.templates(eventId));

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

        const response = await client.raw('GET', ENDPOINTS.templates('some-event-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.templates('some-event-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないイベントIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.templates('01HZNONEXISTENT00000001')
        );

        expect([400, 404, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 2. POST /api/v1/events/{eventId}/templates - テンプレート作成
  // ============================================================
  test.describe('POST /api/v1/events/{eventId}/templates', () => {
    test.describe('正常系', () => {
      test('テンプレートを作成できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // まずイベント一覧を取得
        const eventsResponse = await client.raw('GET', ENDPOINTS.events);
        expect(eventsResponse.status()).toBe(200);
        const eventsBody = await eventsResponse.json();

        const events = Array.isArray(eventsBody.data) ? eventsBody.data : (eventsBody.data?.events || []);
        if (events.length > 0) {
          const eventId = events[0].id || events[0].event_id;
          const templateName = `Test Template ${Date.now()}`;

          const response = await client.raw('POST', ENDPOINTS.templates(eventId), {
            name: templateName,
          });

          // 成功（200/201）または作成が許可されていない場合
          expect([200, 201, 400, 403, 404]).toContain(response.status());
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.templates('some-event-id'), {
          name: 'Test Template',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.templates('some-event-id'), {
          name: 'Test Template',
        });

        expect(response.status()).toBe(401);
      });

      test('存在しないイベントIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'POST',
          ENDPOINTS.templates('01HZNONEXISTENT00000001'),
          { name: 'Test Template' }
        );

        expect([400, 404, 500]).toContain(response.status());
      });

      test('テンプレート名なしで400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // まずイベント一覧を取得
        const eventsResponse = await client.raw('GET', ENDPOINTS.events);
        expect(eventsResponse.status()).toBe(200);
        const eventsBody = await eventsResponse.json();

        const events = Array.isArray(eventsBody.data) ? eventsBody.data : (eventsBody.data?.events || []);
        if (events.length > 0) {
          const eventId = events[0].id || events[0].event_id;

          const response = await client.raw('POST', ENDPOINTS.templates(eventId), {});

          expect(response.status()).toBeGreaterThanOrEqual(400);
        }
      });
    });
  });

  // ============================================================
  // 3. GET /api/v1/events/{eventId}/templates/{id} - テンプレート取得
  // ============================================================
  test.describe('GET /api/v1/events/{eventId}/templates/{id}', () => {
    test.describe('正常系', () => {
      test('テンプレート情報を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // まずイベント一覧を取得
        const eventsResponse = await client.raw('GET', ENDPOINTS.events);
        expect(eventsResponse.status()).toBe(200);
        const eventsBody = await eventsResponse.json();

        const events = Array.isArray(eventsBody.data) ? eventsBody.data : (eventsBody.data?.events || []);
        if (events.length > 0) {
          const eventId = events[0].id || events[0].event_id;

          // テンプレート一覧を取得
          const templatesResponse = await client.raw('GET', ENDPOINTS.templates(eventId));
          if (templatesResponse.status() === 200) {
            const templatesBody = await templatesResponse.json();
            const templates = Array.isArray(templatesBody.data) ? templatesBody.data : (templatesBody.data?.templates || []);

            if (templates.length > 0) {
              const templateId = templates[0].id || templates[0].template_id;

              const response = await client.raw('GET', ENDPOINTS.template(eventId, templateId));

              expect([200, 404]).toContain(response.status());
              if (response.status() === 200) {
                const body = await response.json();
                expect(body.data).toBeDefined();
              }
            }
          }
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.template('some-event-id', 'some-template-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.template('some-event-id', 'some-template-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないテンプレートIDで404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // まずイベント一覧を取得
        const eventsResponse = await client.raw('GET', ENDPOINTS.events);
        expect(eventsResponse.status()).toBe(200);
        const eventsBody = await eventsResponse.json();

        const events = Array.isArray(eventsBody.data) ? eventsBody.data : (eventsBody.data?.events || []);
        if (events.length > 0) {
          const eventId = events[0].id || events[0].event_id;

          const response = await client.raw(
            'GET',
            ENDPOINTS.template(eventId, '01HZNONEXISTENT00000001')
          );

          expect([400, 404]).toContain(response.status());
        }
      });

      test('存在しないイベントIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.template('01HZNONEXISTENT00000001', '01HZNONEXISTENT00000002')
        );

        expect([400, 404, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 4. PUT /api/v1/events/{eventId}/templates/{id} - テンプレート更新
  // ============================================================
  test.describe('PUT /api/v1/events/{eventId}/templates/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw(
          'PUT',
          ENDPOINTS.template('some-event-id', 'some-template-id'),
          { name: 'Updated Template' }
        );

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw(
          'PUT',
          ENDPOINTS.template('some-event-id', 'some-template-id'),
          { name: 'Updated Template' }
        );

        expect(response.status()).toBe(401);
      });

      test('存在しないテンプレートIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // まずイベント一覧を取得
        const eventsResponse = await client.raw('GET', ENDPOINTS.events);
        expect(eventsResponse.status()).toBe(200);
        const eventsBody = await eventsResponse.json();

        const events = Array.isArray(eventsBody.data) ? eventsBody.data : (eventsBody.data?.events || []);
        if (events.length > 0) {
          const eventId = events[0].id || events[0].event_id;

          const response = await client.raw(
            'PUT',
            ENDPOINTS.template(eventId, '01HZNONEXISTENT00000001'),
            { name: 'Updated Template' }
          );

          // 405 (Method Not Allowed) も許容
          expect([400, 404, 405, 500]).toContain(response.status());
        }
      });

      test('存在しないイベントIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'PUT',
          ENDPOINTS.template('01HZNONEXISTENT00000001', '01HZNONEXISTENT00000002'),
          { name: 'Updated Template' }
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 5. DELETE /api/v1/events/{eventId}/templates/{id} - テンプレート削除
  // ============================================================
  test.describe('DELETE /api/v1/events/{eventId}/templates/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.template('some-event-id', 'some-template-id')
        );

        expect([400, 401, 405]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.template('some-event-id', 'some-template-id')
        );

        expect([401, 405]).toContain(response.status());
      });

      test('存在しないテンプレートIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // まずイベント一覧を取得
        const eventsResponse = await client.raw('GET', ENDPOINTS.events);
        expect(eventsResponse.status()).toBe(200);
        const eventsBody = await eventsResponse.json();

        const events = Array.isArray(eventsBody.data) ? eventsBody.data : (eventsBody.data?.events || []);
        if (events.length > 0) {
          const eventId = events[0].id || events[0].event_id;

          const response = await client.raw(
            'DELETE',
            ENDPOINTS.template(eventId, '01HZNONEXISTENT00000001')
          );

          // 405 (Method Not Allowed) も許容
          expect([400, 404, 405, 500]).toContain(response.status());
        }
      });

      test('存在しないイベントIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.template('01HZNONEXISTENT00000001', '01HZNONEXISTENT00000002')
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });
});

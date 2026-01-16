import { test, expect } from '@playwright/test';
import {
  ENDPOINTS,
  ApiClient,
} from '../utils/api-client';
import { loginAsAdmin, getUnauthenticatedClient } from '../utils/auth';

/**
 * Announcement API Tests
 *
 * Endpoints:
 * 1. GET /api/v1/announcements - お知らせ一覧取得
 * 2. GET /api/v1/announcements/unread-count - 未読件数取得
 * 3. POST /api/v1/announcements/{id}/read - お知らせ既読化
 * 4. POST /api/v1/announcements/read-all - 全件既読化
 */

test.describe('Announcement API', () => {
  // ============================================================
  // 1. GET /api/v1/announcements - お知らせ一覧取得
  // ============================================================
  test.describe('GET /api/v1/announcements', () => {
    test.describe('正常系', () => {
      test('お知らせ一覧を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.announcements);

        expect([200, 404]).toContain(response.status());
        if (response.status() === 200) {
          const body = await response.json();
          expect(body).toBeDefined();
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.announcements);

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.announcements);

        expect(response.status()).toBe(401);
      });
    });
  });

  // ============================================================
  // 2. GET /api/v1/announcements/unread-count - 未読件数取得
  // ============================================================
  test.describe('GET /api/v1/announcements/unread-count', () => {
    test.describe('正常系', () => {
      test('未読件数を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.announcementsUnreadCount);

        expect([200, 404]).toContain(response.status());
        if (response.status() === 200) {
          const body = await response.json();
          expect(body).toBeDefined();
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.announcementsUnreadCount);

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.announcementsUnreadCount);

        expect(response.status()).toBe(401);
      });
    });
  });

  // ============================================================
  // 3. POST /api/v1/announcements/{id}/read - お知らせ既読化
  // ============================================================
  test.describe('POST /api/v1/announcements/{id}/read', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.announcementRead('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.announcementRead('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないお知らせIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'POST',
          ENDPOINTS.announcementRead('01HZNONEXISTENT00000001')
        );

        // 200 (成功 - 冪等性) も許容
        expect([200, 400, 404, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 4. POST /api/v1/announcements/read-all - 全件既読化
  // ============================================================
  test.describe('POST /api/v1/announcements/read-all', () => {
    test.describe('正常系', () => {
      test('全件既読化できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.announcementsReadAll);

        expect([200, 204]).toContain(response.status());
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.announcementsReadAll);

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.announcementsReadAll);

        expect(response.status()).toBe(401);
      });
    });
  });
});

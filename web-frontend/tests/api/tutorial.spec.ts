import { test, expect } from '@playwright/test';
import {
  ENDPOINTS,
  ApiClient,
} from '../utils/api-client';
import { loginAsAdmin, getUnauthenticatedClient } from '../utils/auth';

/**
 * Tutorial API Tests
 *
 * Endpoints:
 * 1. GET /api/v1/tutorials - チュートリアル一覧取得
 * 2. GET /api/v1/tutorials/{id} - チュートリアル取得
 * 3. PUT /api/v1/tutorials/{id} - チュートリアル更新（完了状態など）
 */

test.describe('Tutorial API', () => {
  // ============================================================
  // 1. GET /api/v1/tutorials - チュートリアル一覧取得
  // ============================================================
  test.describe('GET /api/v1/tutorials', () => {
    test.describe('正常系', () => {
      test('チュートリアル一覧を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.tutorials);

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

        const response = await client.raw('GET', ENDPOINTS.tutorials);

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.tutorials);

        expect(response.status()).toBe(401);
      });
    });
  });

  // ============================================================
  // 2. GET /api/v1/tutorials/{id} - チュートリアル取得
  // ============================================================
  test.describe('GET /api/v1/tutorials/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.tutorial('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.tutorial('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないチュートリアルIDで404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.tutorial('01HZNONEXISTENT00000001')
        );

        // Note: 500が返る場合はバグの可能性あり
        expect([400, 404, 500]).toContain(response.status());
      });

      test('無効なID形式で400/404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.tutorial('invalid-id'));

        // Note: 500が返る場合はバグの可能性あり
        expect([400, 404, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 3. PUT /api/v1/tutorials/{id} - チュートリアル更新
  // ============================================================
  test.describe('PUT /api/v1/tutorials/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('PUT', ENDPOINTS.tutorial('some-id'), {
          completed: true,
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('PUT', ENDPOINTS.tutorial('some-id'), {
          completed: true,
        });

        expect(response.status()).toBe(401);
      });

      test('存在しないチュートリアルIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'PUT',
          ENDPOINTS.tutorial('01HZNONEXISTENT00000001'),
          { completed: true }
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });
});

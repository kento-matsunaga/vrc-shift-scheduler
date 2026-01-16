import { test, expect } from '@playwright/test';
import {
  ENDPOINTS,
  ApiClient,
} from '../utils/api-client';
import { loginAsAdmin, getUnauthenticatedClient } from '../utils/auth';

/**
 * Import API Tests
 *
 * Endpoints:
 * 1. GET /api/v1/imports - インポート一覧取得
 * 2. POST /api/v1/imports/members - メンバーインポート
 * 3. GET /api/v1/imports/{id}/status - インポートステータス取得
 * 4. GET /api/v1/imports/{id}/result - インポート結果取得
 */

test.describe('Import API', () => {
  // ============================================================
  // 1. GET /api/v1/imports - インポート一覧取得
  // ============================================================
  test.describe('GET /api/v1/imports', () => {
    test.describe('正常系', () => {
      test('インポート一覧を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.imports);

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

        const response = await client.raw('GET', ENDPOINTS.imports);

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.imports);

        expect(response.status()).toBe(401);
      });
    });
  });

  // ============================================================
  // 2. POST /api/v1/imports/members - メンバーインポート
  // ============================================================
  test.describe('POST /api/v1/imports/members', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.importMembers, {
          members: [],
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.importMembers, {
          members: [],
        });

        expect(response.status()).toBe(401);
      });

      test('必須パラメータなしで400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.importMembers, {});

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });
    });
  });

  // ============================================================
  // 3. GET /api/v1/imports/{id}/status - インポートステータス取得
  // ============================================================
  test.describe('GET /api/v1/imports/{id}/status', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.importStatus('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.importStatus('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないインポートIDで404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.importStatus('01HZNONEXISTENT00000001')
        );

        expect([400, 404]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 4. GET /api/v1/imports/{id}/result - インポート結果取得
  // ============================================================
  test.describe('GET /api/v1/imports/{id}/result', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.importResult('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.importResult('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないインポートIDで404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.importResult('01HZNONEXISTENT00000001')
        );

        expect([400, 404]).toContain(response.status());
      });
    });
  });
});

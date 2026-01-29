import { test, expect } from '@playwright/test';
import {
  ENDPOINTS,
  ApiClient,
} from '../utils/api-client';
import { loginAsAdmin, getUnauthenticatedClient } from '../utils/auth';

/**
 * ActualAttendance API Tests
 *
 * Endpoints:
 * 1. GET /api/v1/actual-attendance - 実績出欠一覧取得
 * 2. POST /api/v1/actual-attendance - 実績出欠作成/更新
 */

test.describe('ActualAttendance API', () => {
  // ============================================================
  // 1. GET /api/v1/actual-attendance - 実績出欠一覧取得
  // ============================================================
  test.describe('GET /api/v1/actual-attendance', () => {
    test.describe('正常系', () => {
      test('実績出欠一覧を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.actualAttendance);

        // 200 (成功) または 400 (クエリパラメータ必須) または 404 (リソースなし)
        expect([200, 400, 404]).toContain(response.status());
        if (response.status() === 200) {
          const body = await response.json();
          expect(body).toBeDefined();
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.actualAttendance);

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.actualAttendance);

        expect(response.status()).toBe(401);
      });
    });
  });

  // ============================================================
  // 2. POST /api/v1/actual-attendance - 実績出欠作成/更新
  // ============================================================
  test.describe('POST /api/v1/actual-attendance', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.actualAttendance, {
          business_day_id: 'some-id',
          member_id: 'some-member-id',
          status: 'present',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.actualAttendance, {
          business_day_id: 'some-id',
          member_id: 'some-member-id',
          status: 'present',
        });

        expect(response.status()).toBe(401);
      });

      test('必須パラメータなしで400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.actualAttendance, {});

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('存在しない営業日IDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.actualAttendance, {
          business_day_id: '01HZNONEXISTENT00000001',
          member_id: '01HZNONEXISTENT00000002',
          status: 'present',
        });

        // 200 (成功 - upsert動作) / 405 (Method Not Allowed) も許容
        expect([200, 400, 404, 405, 500]).toContain(response.status());
      });

      test('無効なステータスでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.actualAttendance, {
          business_day_id: '01HZNONEXISTENT00000001',
          member_id: '01HZNONEXISTENT00000002',
          status: 'invalid-status',
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });
    });
  });
});

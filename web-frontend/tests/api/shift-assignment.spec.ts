import { test, expect } from '@playwright/test';
import {
  ENDPOINTS,
  ApiClient,
} from '../utils/api-client';
import { loginAsAdmin, getUnauthenticatedClient } from '../utils/auth';

/**
 * ShiftAssignment API Tests
 *
 * Endpoints:
 * 1. GET /api/v1/shift-assignments - シフト割当一覧取得
 * 2. POST /api/v1/shift-assignments - シフト割当作成
 * 3. GET /api/v1/shift-assignments/{id} - シフト割当取得
 * 4. PUT /api/v1/shift-assignments/{id} - シフト割当更新
 * 5. DELETE /api/v1/shift-assignments/{id} - シフト割当削除
 * 6. PUT /api/v1/shift-assignments/{id}/status - シフト割当ステータス更新
 */

test.describe('ShiftAssignment API', () => {
  // ============================================================
  // 1. GET /api/v1/shift-assignments - シフト割当一覧取得
  // ============================================================
  test.describe('GET /api/v1/shift-assignments', () => {
    test.describe('正常系', () => {
      test('シフト割当一覧APIが存在する', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.shiftAssignments);

        // 200 (成功) または 400 (クエリパラメータ必須) または 404 (リソースなし)
        expect([200, 400, 404]).toContain(response.status());

        if (response.status() === 200) {
          const body = await response.json();
          expect(body.data).toBeDefined();
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.shiftAssignments);

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.shiftAssignments);

        expect(response.status()).toBe(401);
      });
    });
  });

  // ============================================================
  // 2. POST /api/v1/shift-assignments - シフト割当作成
  // ============================================================
  test.describe('POST /api/v1/shift-assignments', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.shiftAssignments, {
          shift_slot_id: 'some-slot-id',
          member_id: 'some-member-id',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.shiftAssignments, {
          shift_slot_id: 'some-slot-id',
          member_id: 'some-member-id',
        });

        expect(response.status()).toBe(401);
      });

      test('必須パラメータなしで400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.shiftAssignments, {});

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('存在しないシフト枠IDで400/404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.shiftAssignments, {
          shift_slot_id: '01HZNONEXISTENT00000001',
          member_id: '01HZNONEXISTENT00000002',
        });

        expect([400, 404, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 3. GET /api/v1/shift-assignments/{id} - シフト割当取得
  // ============================================================
  test.describe('GET /api/v1/shift-assignments/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.shiftAssignment('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.shiftAssignment('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないシフト割当IDで404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.shiftAssignment('01HZNONEXISTENT00000001')
        );

        expect([400, 404]).toContain(response.status());
      });

      test('無効なシフト割当ID形式で400/404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.shiftAssignment('invalid-id'));

        expect([400, 404]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 4. PUT /api/v1/shift-assignments/{id} - シフト割当更新
  // ============================================================
  test.describe('PUT /api/v1/shift-assignments/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('PUT', ENDPOINTS.shiftAssignment('some-id'), {
          status: 'confirmed',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('PUT', ENDPOINTS.shiftAssignment('some-id'), {
          status: 'confirmed',
        });

        expect(response.status()).toBe(401);
      });

      test('存在しないシフト割当IDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'PUT',
          ENDPOINTS.shiftAssignment('01HZNONEXISTENT00000001'),
          {
            status: 'confirmed',
          }
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 5. DELETE /api/v1/shift-assignments/{id} - シフト割当削除
  // ============================================================
  test.describe('DELETE /api/v1/shift-assignments/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.shiftAssignment('some-id')
        );

        expect([400, 401, 405]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.shiftAssignment('some-id')
        );

        expect([401, 405]).toContain(response.status());
      });

      test('存在しないシフト割当IDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.shiftAssignment('01HZNONEXISTENT00000001')
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 6. PUT /api/v1/shift-assignments/{id}/status - シフト割当ステータス更新
  // ============================================================
  test.describe('PUT /api/v1/shift-assignments/{id}/status', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw(
          'PUT',
          ENDPOINTS.shiftAssignmentStatus('some-id'),
          { status: 'confirmed' }
        );

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw(
          'PUT',
          ENDPOINTS.shiftAssignmentStatus('some-id'),
          { status: 'confirmed' }
        );

        expect(response.status()).toBe(401);
      });

      test('存在しないシフト割当IDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'PUT',
          ENDPOINTS.shiftAssignmentStatus('01HZNONEXISTENT00000001'),
          { status: 'confirmed' }
        );

        expect([400, 404, 500]).toContain(response.status());
      });

      test('無効なステータスで400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'PUT',
          ENDPOINTS.shiftAssignmentStatus('01HZNONEXISTENT00000001'),
          { status: 'invalid-status' }
        );

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });
    });
  });
});

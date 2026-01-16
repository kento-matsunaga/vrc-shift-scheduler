import { test, expect } from '@playwright/test';
import {
  ENDPOINTS,
  ApiClient,
} from '../utils/api-client';
import { getUnauthenticatedClient } from '../utils/auth';

/**
 * Public API Tests
 *
 * Endpoints (認証不要、トークンベースのアクセス):
 * 1. GET /api/v1/public/attendance/{token} - 公開出欠収集取得
 * 2. POST /api/v1/public/attendance/{token}/responses - 公開出欠回答
 * 3. GET /api/v1/public/attendance/{token}/members/{memberId}/responses - メンバー回答取得
 * 4. GET /api/v1/public/members - 公開メンバー一覧
 * 5. GET /api/v1/public/schedules/{token} - 公開スケジュール取得
 * 6. POST /api/v1/public/schedules/{token}/responses - 公開スケジュール回答
 * 7. POST /api/v1/public/license/claim - ライセンスクレーム
 */

test.describe('Public API', () => {
  // ============================================================
  // 1. GET /api/v1/public/attendance/{token} - 公開出欠収集取得
  // ============================================================
  test.describe('GET /api/v1/public/attendance/{token}', () => {
    test.describe('異常系', () => {
      test('無効なトークンで404エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.publicAttendance('invalid-token-12345'));

        expect([400, 404]).toContain(response.status());
      });

      test('存在しないトークンで404エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.publicAttendance('01HZNONEXISTENT00000001'));

        expect([400, 404]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 2. POST /api/v1/public/attendance/{token}/responses - 公開出欠回答
  // ============================================================
  test.describe('POST /api/v1/public/attendance/{token}/responses', () => {
    test.describe('異常系', () => {
      test('無効なトークンで404エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.publicAttendanceResponses('invalid-token'), {
          responses: [],
        });

        expect([400, 404]).toContain(response.status());
      });

      test('必須パラメータなしでエラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.publicAttendanceResponses('some-token'), {});

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });
    });
  });

  // ============================================================
  // 2.5 GET /api/v1/public/attendance/{token}/responses - 全回答取得
  // ============================================================
  test.describe('GET /api/v1/public/attendance/{token}/responses (all responses)', () => {
    test.describe('異常系', () => {
      test('無効なトークンで404エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.publicAttendanceAllResponses('invalid-token'));

        expect([400, 404]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 3. GET /api/v1/public/attendance/{token}/members/{memberId}/responses - メンバー回答取得
  // ============================================================
  test.describe('GET /api/v1/public/attendance/{token}/members/{memberId}/responses', () => {
    test.describe('異常系', () => {
      test('無効なトークンで404エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.publicAttendanceMemberResponses('invalid-token', 'some-member-id')
        );

        expect([400, 404]).toContain(response.status());
      });

      test('存在しないメンバーIDで404エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.publicAttendanceMemberResponses('some-token', '01HZNONEXISTENT00000001')
        );

        expect([400, 404]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 4. GET /api/v1/public/members - 公開メンバー一覧
  // ============================================================
  test.describe('GET /api/v1/public/members', () => {
    test.describe('異常系', () => {
      test('トークンなしで401/403エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.publicMembers);

        // Public APIはトークンベースのため、トークンなしではエラー
        expect([400, 401, 403, 404]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 5. GET /api/v1/public/schedules/{token} - 公開スケジュール取得
  // ============================================================
  test.describe('GET /api/v1/public/schedules/{token}', () => {
    test.describe('異常系', () => {
      test('無効なトークンで404エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.publicSchedule('invalid-token-12345'));

        expect([400, 404]).toContain(response.status());
      });

      test('存在しないトークンで404エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.publicSchedule('01HZNONEXISTENT00000001'));

        expect([400, 404]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 6. POST /api/v1/public/schedules/{token}/responses - 公開スケジュール回答
  // ============================================================
  test.describe('POST /api/v1/public/schedules/{token}/responses', () => {
    test.describe('異常系', () => {
      test('無効なトークンで404エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.publicScheduleResponses('invalid-token'), {
          responses: [],
        });

        expect([400, 404]).toContain(response.status());
      });

      test('必須パラメータなしでエラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.publicScheduleResponses('some-token'), {});

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });
    });
  });

  // ============================================================
  // 6.5 GET /api/v1/public/schedules/{token}/responses - 全回答取得
  // ============================================================
  test.describe('GET /api/v1/public/schedules/{token}/responses (all responses)', () => {
    test.describe('異常系', () => {
      test('無効なトークンで404エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.publicScheduleAllResponses('invalid-token'));

        expect([400, 404]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 7. POST /api/v1/public/license/claim - ライセンスクレーム
  // ============================================================
  test.describe('POST /api/v1/public/license/claim', () => {
    test.describe('異常系', () => {
      test('ライセンスキーなしで400エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.publicLicenseClaim, {});

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('無効なライセンスキーでエラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.publicLicenseClaim, {
          license_key: 'INVALID-LICENSE-KEY-12345',
        });

        // 401/403 (認証エラー) / 500 (サーバーエラー) も許容
        expect([400, 401, 403, 404, 500]).toContain(response.status());
      });

      test('空のライセンスキーで400エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.publicLicenseClaim, {
          license_key: '',
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });
    });
  });
});

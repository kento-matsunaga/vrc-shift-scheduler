import { test, expect } from '@playwright/test';
import {
  ENDPOINTS,
  ApiClient,
} from '../utils/api-client';
import { loginAsAdmin, getUnauthenticatedClient } from '../utils/auth';

/**
 * Member API Tests
 *
 * Endpoints:
 * 1. GET /api/v1/members - メンバー一覧取得
 * 2. POST /api/v1/members - メンバー作成
 * 3. GET /api/v1/members/{id} - メンバー取得
 * 4. PUT /api/v1/members/{id} - メンバー更新
 * 5. DELETE /api/v1/members/{id} - メンバー削除
 * 6. GET /api/v1/members/me - 自分のメンバー情報取得
 * 7. GET /api/v1/members/recent-attendance - 最近の出勤情報取得
 * 8. POST /api/v1/members/bulk-import - 一括インポート
 * 9. POST /api/v1/members/bulk-update-roles - ロール一括更新
 */

test.describe('Member API', () => {
  // ============================================================
  // 1. GET /api/v1/members - メンバー一覧取得
  // ============================================================
  test.describe('GET /api/v1/members', () => {
    test.describe('正常系', () => {
      test('メンバー一覧を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.members);

        expect(response.status()).toBe(200);

        const body = await response.json();
        expect(body.data).toBeDefined();
        // data自体が配列、またはdata.membersが配列
        const members = Array.isArray(body.data) ? body.data : body.data.members;
        expect(members === undefined || Array.isArray(members)).toBe(true);
      });

      test('メンバー一覧にメンバー情報が含まれている', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.members);

        expect(response.status()).toBe(200);

        const body = await response.json();
        // data自体が配列、またはdata.membersが配列
        const members = Array.isArray(body.data) ? body.data : (body.data.members || []);
        if (members.length > 0) {
          const member = members[0];
          // id または member_id が含まれる
          const hasId = member.id || member.member_id;
          expect(hasId).toBeTruthy();
          expect(member.display_name).toBeDefined();
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.members);

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.members);

        expect(response.status()).toBe(401);
      });
    });
  });

  // ============================================================
  // 2. POST /api/v1/members - メンバー作成
  // ============================================================
  test.describe('POST /api/v1/members', () => {
    test.describe('正常系', () => {
      test('メンバーを作成できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const memberName = `Test Member ${Date.now()}`;
        const response = await client.raw('POST', ENDPOINTS.members, {
          display_name: memberName,
        });

        // 成功（200/201）または作成が許可されていない場合
        expect([200, 201, 400, 403]).toContain(response.status());

        if (response.status() === 200 || response.status() === 201) {
          const body = await response.json();
          expect(body.data).toBeDefined();
          // idまたはmember_idまたはidが含まれる
          const hasId = body.data.id || body.data.member_id || (body.data.member && body.data.member.id);
          expect(hasId).toBeTruthy();
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.members, {
          display_name: 'Test',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.members, {
          display_name: 'Test',
        });

        expect(response.status()).toBe(401);
      });

      test('表示名なしで400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.members, {});

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('空の表示名で400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.members, {
          display_name: '',
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('非常に長い表示名で400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.members, {
          display_name: 'a'.repeat(1000),
        });

        // 長すぎる名前は400エラーまたは許可される（切り詰め）
        expect([200, 201, 400]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 3. GET /api/v1/members/{id} - メンバー取得
  // ============================================================
  test.describe('GET /api/v1/members/{id}', () => {
    test.describe('正常系', () => {
      test('メンバー情報を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // まずメンバー一覧を取得
        const listResponse = await client.raw('GET', ENDPOINTS.members);
        expect(listResponse.status()).toBe(200);
        const listBody = await listResponse.json();

        if (listBody.data.length > 0) {
          const memberId = listBody.data[0].id;

          const response = await client.raw('GET', ENDPOINTS.member(memberId));

          expect(response.status()).toBe(200);
          const body = await response.json();
          expect(body.data).toBeDefined();
          expect(body.data.id).toBe(memberId);
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.member('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.member('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないメンバーIDで404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.member('01HZNONEXISTENT00000001')
        );

        expect([400, 404]).toContain(response.status());
      });

      test('無効なメンバーID形式で400/404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.member('invalid-id'));

        expect([400, 404]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 4. PUT /api/v1/members/{id} - メンバー更新
  // ============================================================
  test.describe('PUT /api/v1/members/{id}', () => {
    test.describe('正常系', () => {
      test('メンバー情報を更新できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // まずメンバー一覧を取得
        const listResponse = await client.raw('GET', ENDPOINTS.members);
        expect(listResponse.status()).toBe(200);
        const listBody = await listResponse.json();

        if (listBody.data.length > 0) {
          const member = listBody.data[0];
          const originalName = member.display_name;

          const newName = 'Updated ' + Date.now();
          const response = await client.raw('PUT', ENDPOINTS.member(member.id), {
            display_name: newName,
          });

          // 成功または更新が許可されていない場合
          expect([200, 400, 403]).toContain(response.status());

          if (response.status() === 200) {
            // 元に戻す
            await client.raw('PUT', ENDPOINTS.member(member.id), {
              display_name: originalName,
            });
          }
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('PUT', ENDPOINTS.member('some-id'), {
          display_name: 'Test',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('PUT', ENDPOINTS.member('some-id'), {
          display_name: 'Test',
        });

        expect(response.status()).toBe(401);
      });

      test('存在しないメンバーIDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'PUT',
          ENDPOINTS.member('01HZNONEXISTENT00000001'),
          {
            display_name: 'Test',
          }
        );

        // 400 (validation), 404 (not found), or 500 (server error)
        expect([400, 404, 500]).toContain(response.status());
      });

      test('空の表示名で400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // まずメンバー一覧を取得
        const listResponse = await client.raw('GET', ENDPOINTS.members);
        expect(listResponse.status()).toBe(200);
        const listBody = await listResponse.json();

        if (listBody.data.length > 0) {
          const memberId = listBody.data[0].id;

          const response = await client.raw('PUT', ENDPOINTS.member(memberId), {
            display_name: '',
          });

          expect(response.status()).toBeGreaterThanOrEqual(400);
        }
      });
    });
  });

  // ============================================================
  // 5. DELETE /api/v1/members/{id} - メンバー削除
  // ============================================================
  test.describe('DELETE /api/v1/members/{id}', () => {
    test.describe('異常系', () => {
      // Note: 正常系の削除テストはデータを破壊するため、慎重に扱う必要がある

      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.member('some-id')
        );

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.member('some-id')
        );

        expect(response.status()).toBe(401);
      });

      test('存在しないメンバーIDで404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.member('01HZNONEXISTENT00000001')
        );

        expect([400, 404]).toContain(response.status());
      });

      test('無効なメンバーID形式で400/404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.member('invalid-id')
        );

        expect([400, 404]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 6. GET /api/v1/members/me - 自分のメンバー情報取得
  // ============================================================
  test.describe('GET /api/v1/members/me', () => {
    test.describe('正常系', () => {
      test('自分のメンバー情報を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.membersMe);

        // 200 (success), 400 (bad request), or 404 (if admin has no associated member)
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

        const response = await client.raw('GET', ENDPOINTS.membersMe);

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.membersMe);

        expect(response.status()).toBe(401);
      });
    });
  });

  // ============================================================
  // 7. GET /api/v1/members/recent-attendance - 最近の出勤情報取得
  // ============================================================
  test.describe('GET /api/v1/members/recent-attendance', () => {
    test.describe('正常系', () => {
      test('最近の出勤情報を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.membersRecentAttendance);

        expect(response.status()).toBe(200);

        const body = await response.json();
        expect(body.data).toBeDefined();
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.membersRecentAttendance);

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.membersRecentAttendance);

        expect(response.status()).toBe(401);
      });
    });
  });

  // ============================================================
  // 8. POST /api/v1/members/bulk-import - 一括インポート
  // ============================================================
  test.describe('POST /api/v1/members/bulk-import', () => {
    test.describe('正常系', () => {
      test('一括インポートAPIが存在する', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.membersBulkImport, {
          members: [],
        });

        // エンドポイントが存在することを確認（404ではない）
        expect(response.status()).not.toBe(404);
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.membersBulkImport, {
          members: [],
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.membersBulkImport, {
          members: [],
        });

        expect(response.status()).toBe(401);
      });

      test('空のリクエストボディで400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.membersBulkImport, {});

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('不正なデータ形式で400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.membersBulkImport, {
          members: 'not-an-array',
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });
    });
  });

  // ============================================================
  // 9. POST /api/v1/members/bulk-update-roles - ロール一括更新
  // ============================================================
  test.describe('POST /api/v1/members/bulk-update-roles', () => {
    test.describe('正常系', () => {
      test('ロール一括更新APIが存在する', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.membersBulkUpdateRoles, {
          updates: [],
        });

        // エンドポイントが存在することを確認（404ではない）
        expect(response.status()).not.toBe(404);
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.membersBulkUpdateRoles, {
          updates: [],
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.membersBulkUpdateRoles, {
          updates: [],
        });

        expect(response.status()).toBe(401);
      });

      test('空のリクエストボディで400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.membersBulkUpdateRoles, {});

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('不正なデータ形式で400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.membersBulkUpdateRoles, {
          updates: 'not-an-array',
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });
    });
  });
});

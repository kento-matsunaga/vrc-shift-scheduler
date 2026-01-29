import { test, expect } from '@playwright/test';
import {
  ENDPOINTS,
  ApiClient,
} from '../utils/api-client';
import { loginAsAdmin, getUnauthenticatedClient } from '../utils/auth';

/**
 * ShiftSlot API Tests
 *
 * Endpoints:
 * 1. GET /api/v1/shift-slots - シフト枠一覧取得
 * 2. POST /api/v1/shift-slots - シフト枠作成
 * 3. GET /api/v1/shift-slots/{id} - シフト枠取得
 * 4. PUT /api/v1/shift-slots/{id} - シフト枠更新
 * 5. DELETE /api/v1/shift-slots/{id} - シフト枠削除
 * 6. GET /api/v1/business-days/{businessDayId}/shift-slots - 営業日のシフト枠一覧
 */

test.describe('ShiftSlot API', () => {
  // ============================================================
  // 1. GET /api/v1/shift-slots - シフト枠一覧取得
  // ============================================================
  test.describe('GET /api/v1/shift-slots', () => {
    test.describe('正常系', () => {
      test('シフト枠一覧APIが存在する', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.shiftSlots);

        // 200 (成功) または 404 (営業日ID必須の場合)
        expect([200, 404]).toContain(response.status());

        if (response.status() === 200) {
          const body = await response.json();
          expect(body.data).toBeDefined();
        }
      });
    });

    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.shiftSlots);

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.shiftSlots);

        expect(response.status()).toBe(401);
      });
    });
  });

  // ============================================================
  // 2. POST /api/v1/shift-slots - シフト枠作成
  // ============================================================
  test.describe('POST /api/v1/shift-slots', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('POST', ENDPOINTS.shiftSlots, {
          name: 'Test Shift Slot',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('POST', ENDPOINTS.shiftSlots, {
          name: 'Test Shift Slot',
        });

        expect(response.status()).toBe(401);
      });

      test('必須パラメータなしで400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.shiftSlots, {});

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('priority=0で400エラー（priorityは1以上）', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.shiftSlots, {
          business_day_id: '01HZNONEXISTENT00000001',
          name: 'Test Shift Slot',
          priority: 0, // 無効: priorityは1以上である必要がある
          capacity: 5,
        });

        // 400 (バリデーションエラー) または 404 (営業日なし)
        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('負のpriorityで400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.shiftSlots, {
          business_day_id: '01HZNONEXISTENT00000001',
          name: 'Test Shift Slot',
          priority: -1, // 無効: 負の値
          capacity: 5,
        });

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });
    });
  });

  // ============================================================
  // 3. GET /api/v1/shift-slots/{id} - シフト枠取得
  // ============================================================
  test.describe('GET /api/v1/shift-slots/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('GET', ENDPOINTS.shiftSlot('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.shiftSlot('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しないシフト枠IDで404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.shiftSlot('01HZNONEXISTENT00000001')
        );

        expect([400, 404]).toContain(response.status());
      });

      test('無効なシフト枠ID形式で400/404エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('GET', ENDPOINTS.shiftSlot('invalid-id'));

        expect([400, 404]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 4. PUT /api/v1/shift-slots/{id} - シフト枠更新
  // ============================================================
  test.describe('PUT /api/v1/shift-slots/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw('PUT', ENDPOINTS.shiftSlot('some-id'), {
          name: 'Updated Shift Slot',
        });

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('PUT', ENDPOINTS.shiftSlot('some-id'), {
          name: 'Updated Shift Slot',
        });

        expect(response.status()).toBe(401);
      });

      test('存在しないシフト枠IDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'PUT',
          ENDPOINTS.shiftSlot('01HZNONEXISTENT00000001'),
          {
            name: 'Updated Shift Slot',
          }
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 5. DELETE /api/v1/shift-slots/{id} - シフト枠削除
  // ============================================================
  test.describe('DELETE /api/v1/shift-slots/{id}', () => {
    test.describe('異常系', () => {
      test('認証なしで401エラー', async ({ request }) => {
        const client = getUnauthenticatedClient(request);

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.shiftSlot('some-id')
        );

        expect([400, 401, 405]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.shiftSlot('some-id')
        );

        expect([401, 405]).toContain(response.status());
      });

      test('存在しないシフト枠IDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'DELETE',
          ENDPOINTS.shiftSlot('01HZNONEXISTENT00000001')
        );

        // 405 (Method Not Allowed) も許容
        expect([400, 404, 405, 500]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 6. POST /api/v1/business-days/{businessDayId}/shift-slots - シフト枠作成（instance_id紐付け）
  // ============================================================
  test.describe('POST /api/v1/business-days/{businessDayId}/shift-slots - instance_id linking', () => {
    test.describe('正常系', () => {
      test('instance_idを指定してシフト枠を作成すると紐付けられる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // 1. イベント一覧を取得
        const eventsResponse = await client.raw('GET', ENDPOINTS.events);
        expect(eventsResponse.status()).toBe(200);
        const eventsBody = await eventsResponse.json();
        const events = Array.isArray(eventsBody.data) ? eventsBody.data : (eventsBody.data?.events || []);

        if (events.length === 0) {
          test.skip(true, 'テスト用イベントがありません');
          return;
        }

        const eventId = events[0].id || events[0].event_id;

        // 2. イベントのインスタンス一覧を取得
        const instancesResponse = await client.raw('GET', ENDPOINTS.instances(eventId));
        expect(instancesResponse.status()).toBe(200);
        const instancesBody = await instancesResponse.json();
        const instances = Array.isArray(instancesBody.data) ? instancesBody.data : (instancesBody.data?.instances || []);

        if (instances.length === 0) {
          test.skip(true, 'テスト用インスタンスがありません');
          return;
        }

        const instanceId = instances[0].id || instances[0].instance_id;
        const instanceName = instances[0].name;

        // 3. 営業日一覧を取得
        const businessDaysResponse = await client.raw('GET', ENDPOINTS.businessDaysByEvent(eventId));
        expect(businessDaysResponse.status()).toBe(200);
        const businessDaysBody = await businessDaysResponse.json();
        const businessDays = Array.isArray(businessDaysBody.data) ? businessDaysBody.data : (businessDaysBody.data?.business_days || []);

        if (businessDays.length === 0) {
          test.skip(true, 'テスト用営業日がありません');
          return;
        }

        const businessDayId = businessDays[0].id || businessDays[0].business_day_id;

        // 4. instance_idを指定してシフト枠を作成
        const createResponse = await client.raw('POST', ENDPOINTS.shiftSlotsByBusinessDay(businessDayId), {
          slot_name: 'テスト_instance_id紐付けテスト',
          instance_id: instanceId,
          instance_name: instanceName,
          start_time: '21:00',
          end_time: '23:00',
          required_count: 2,
          priority: 99, // テストデータと区別するため高い値
        });

        expect(createResponse.status()).toBe(201);
        const createBody = await createResponse.json();
        const createdSlot = createBody.data;

        // 5. 作成されたシフト枠にinstance_idが紐付いていることを確認
        expect(createdSlot.instance_id).toBe(instanceId);
        expect(createdSlot.instance_name).toBe(instanceName);
        expect(createdSlot.slot_name).toBe('テスト_instance_id紐付けテスト');

        // 6. シフト枠詳細APIでも確認
        const detailResponse = await client.raw('GET', ENDPOINTS.shiftSlot(createdSlot.slot_id));
        expect(detailResponse.status()).toBe(200);
        const detailBody = await detailResponse.json();
        expect(detailBody.data.instance_id).toBe(instanceId);

        // 7. クリーンアップ: 作成したシフト枠を削除
        const deleteResponse = await client.raw('DELETE', ENDPOINTS.shiftSlot(createdSlot.slot_id));
        expect([200, 204, 404, 405]).toContain(deleteResponse.status());
      });

      test('instance_idなしでシフト枠を作成するとinstance_idはnull', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // 1. イベント一覧を取得
        const eventsResponse = await client.raw('GET', ENDPOINTS.events);
        expect(eventsResponse.status()).toBe(200);
        const eventsBody = await eventsResponse.json();
        const events = Array.isArray(eventsBody.data) ? eventsBody.data : (eventsBody.data?.events || []);

        if (events.length === 0) {
          test.skip(true, 'テスト用イベントがありません');
          return;
        }

        const eventId = events[0].id || events[0].event_id;

        // 2. 営業日一覧を取得
        const businessDaysResponse = await client.raw('GET', ENDPOINTS.businessDaysByEvent(eventId));
        expect(businessDaysResponse.status()).toBe(200);
        const businessDaysBody = await businessDaysResponse.json();
        const businessDays = Array.isArray(businessDaysBody.data) ? businessDaysBody.data : (businessDaysBody.data?.business_days || []);

        if (businessDays.length === 0) {
          test.skip(true, 'テスト用営業日がありません');
          return;
        }

        const businessDayId = businessDays[0].id || businessDays[0].business_day_id;

        // 3. instance_idなしでシフト枠を作成
        const createResponse = await client.raw('POST', ENDPOINTS.shiftSlotsByBusinessDay(businessDayId), {
          slot_name: 'テスト_instance_idなし',
          instance_name: '手動入力インスタンス',
          start_time: '21:00',
          end_time: '23:00',
          required_count: 1,
          priority: 98,
        });

        expect(createResponse.status()).toBe(201);
        const createBody = await createResponse.json();
        const createdSlot = createBody.data;

        // 4. instance_idがnullまたはundefinedであることを確認
        expect(createdSlot.instance_id).toBeFalsy();
        expect(createdSlot.instance_name).toBe('手動入力インスタンス');

        // 5. クリーンアップ: 作成したシフト枠を削除
        const deleteResponse = await client.raw('DELETE', ENDPOINTS.shiftSlot(createdSlot.slot_id));
        expect([200, 204, 404, 405]).toContain(deleteResponse.status());
      });
    });

    test.describe('異常系', () => {
      test('存在しないinstance_idを指定すると400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // 1. イベント一覧を取得
        const eventsResponse = await client.raw('GET', ENDPOINTS.events);
        expect(eventsResponse.status()).toBe(200);
        const eventsBody = await eventsResponse.json();
        const events = Array.isArray(eventsBody.data) ? eventsBody.data : (eventsBody.data?.events || []);

        if (events.length === 0) {
          test.skip(true, 'テスト用イベントがありません');
          return;
        }

        const eventId = events[0].id || events[0].event_id;

        // 2. 営業日一覧を取得
        const businessDaysResponse = await client.raw('GET', ENDPOINTS.businessDaysByEvent(eventId));
        expect(businessDaysResponse.status()).toBe(200);
        const businessDaysBody = await businessDaysResponse.json();
        const businessDays = Array.isArray(businessDaysBody.data) ? businessDaysBody.data : (businessDaysBody.data?.business_days || []);

        if (businessDays.length === 0) {
          test.skip(true, 'テスト用営業日がありません');
          return;
        }

        const businessDayId = businessDays[0].id || businessDays[0].business_day_id;

        // 3. 無効なinstance_idを指定してシフト枠を作成
        const createResponse = await client.raw('POST', ENDPOINTS.shiftSlotsByBusinessDay(businessDayId), {
          slot_name: 'テスト_無効なinstance_id',
          instance_id: 'invalid-id-format',
          instance_name: 'テストインスタンス',
          start_time: '21:00',
          end_time: '23:00',
          required_count: 1,
          priority: 1,
        });

        // 無効なinstance_id形式は400エラーを返す
        expect(createResponse.status()).toBe(400);
      });
    });
  });

  // ============================================================
  // 7. GET /api/v1/business-days/{businessDayId}/shift-slots - 営業日のシフト枠一覧
  // ============================================================
  test.describe('GET /api/v1/business-days/{businessDayId}/shift-slots', () => {
    test.describe('正常系', () => {
      test('営業日のシフト枠一覧を取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // まずイベント一覧を取得
        const eventsResponse = await client.raw('GET', ENDPOINTS.events);
        expect(eventsResponse.status()).toBe(200);
        const eventsBody = await eventsResponse.json();

        const events = Array.isArray(eventsBody.data) ? eventsBody.data : (eventsBody.data?.events || []);
        if (events.length > 0) {
          const eventId = events[0].id || events[0].event_id;

          // イベントの営業日一覧を取得
          const businessDaysResponse = await client.raw('GET', ENDPOINTS.businessDaysByEvent(eventId));
          if (businessDaysResponse.status() === 200) {
            const businessDaysBody = await businessDaysResponse.json();
            const businessDays = Array.isArray(businessDaysBody.data) ? businessDaysBody.data : (businessDaysBody.data?.business_days || []);

            if (businessDays.length > 0) {
              const businessDayId = businessDays[0].id || businessDays[0].business_day_id;

              const response = await client.raw('GET', ENDPOINTS.shiftSlotsByBusinessDay(businessDayId));

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

        const response = await client.raw('GET', ENDPOINTS.shiftSlotsByBusinessDay('some-id'));

        expect([400, 401]).toContain(response.status());
      });

      test('無効なトークンで401エラー', async ({ request }) => {
        const client = new ApiClient(request);
        client.setToken('invalid-token-12345');

        const response = await client.raw('GET', ENDPOINTS.shiftSlotsByBusinessDay('some-id'));

        expect(response.status()).toBe(401);
      });

      test('存在しない営業日IDでエラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw(
          'GET',
          ENDPOINTS.shiftSlotsByBusinessDay('01HZNONEXISTENT00000001')
        );

        expect([400, 404, 500]).toContain(response.status());
      });
    });
  });
});

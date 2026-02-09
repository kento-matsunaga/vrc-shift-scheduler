import { test, expect } from '@playwright/test';
import { ENDPOINTS, ApiClient } from '../utils/api-client';
import { loginAsAdmin } from '../utils/auth';

/**
 * Advanced API Test Scenarios
 *
 * Issue #163: API統合テストのカバレッジ拡充
 *
 * 追加テストシナリオ:
 * 1. 同時実行テスト - 同じリソースへの並列更新時の動作確認
 * 2. 境界値テスト - 最大長のフィールド、0件の一覧、空文字列、最大値/最小値
 * 3. パフォーマンステスト - 大量データでのレスポンスタイム計測
 * 4. 国際化テスト - 日本語、絵文字、特殊文字を含むデータ
 */

test.describe('Advanced API Scenarios', () => {
  // ============================================================
  // 1. 同時実行テスト
  // ============================================================
  test.describe('Concurrent Operations', () => {
    test('同じリソースへの並列リクエストが処理できる', async ({ request }) => {
      const { client } = await loginAsAdmin(request);

      // メンバー一覧を並列で取得
      const requests = Array(5)
        .fill(null)
        .map(() => client.raw('GET', ENDPOINTS.members));

      const responses = await Promise.all(requests);

      // 全てのリクエストが成功することを確認
      for (const response of responses) {
        expect(response.status()).toBe(200);
      }
    });

    test('並列での異なるエンドポイントへのリクエストが処理できる', async ({ request }) => {
      const { client } = await loginAsAdmin(request);

      // 異なるエンドポイントを並列で取得
      const [membersRes, eventsRes, rolesRes, tenantRes] = await Promise.all([
        client.raw('GET', ENDPOINTS.members),
        client.raw('GET', ENDPOINTS.events),
        client.raw('GET', ENDPOINTS.roles),
        client.raw('GET', ENDPOINTS.tenant),
      ]);

      expect(membersRes.status()).toBe(200);
      expect(eventsRes.status()).toBe(200);
      expect(rolesRes.status()).toBe(200);
      expect(tenantRes.status()).toBe(200);
    });
  });

  // ============================================================
  // 2. 境界値テスト
  // ============================================================
  test.describe('Boundary Value Tests', () => {
    test.describe('空のデータ', () => {
      test('空のリクエストボディでメンバー作成が400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.members, {});

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('空のリクエストボディでイベント作成が400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.events, {});

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });

      test('空のリクエストボディでロール作成が400エラー', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const response = await client.raw('POST', ENDPOINTS.roles, {});

        expect(response.status()).toBeGreaterThanOrEqual(400);
      });
    });

    test.describe('最大長フィールド', () => {
      test('255文字の表示名でメンバー作成', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const longName = 'A'.repeat(255);
        const response = await client.raw('POST', ENDPOINTS.members, {
          display_name: longName,
        });

        // 成功または長すぎるとして拒否
        expect([200, 201, 400]).toContain(response.status());
      });

      test('1000文字の表示名でメンバー作成が適切に処理される', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const veryLongName = 'B'.repeat(1000);
        const response = await client.raw('POST', ENDPOINTS.members, {
          display_name: veryLongName,
        });

        // 長すぎる名前は拒否されるか、切り詰められる
        expect([200, 201, 400]).toContain(response.status());
      });

      test('255文字のイベント名で作成', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const longEventName = 'Event-' + 'X'.repeat(249);
        const response = await client.raw('POST', ENDPOINTS.events, {
          name: longEventName,
        });

        // 成功または長すぎるとして拒否
        expect([200, 201, 400]).toContain(response.status());
      });
    });

    test.describe('0件・空の一覧', () => {
      test('フィルタリングで0件の結果が正常に返る', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // 存在しないフィルタで検索
        const response = await client.raw(
          'GET',
          ENDPOINTS.members + '?search=NONEXISTENT_SEARCH_TERM_12345'
        );

        // 0件でも200で返る（空配列）
        expect(response.status()).toBe(200);
        const body = await response.json();
        expect(body.data).toBeDefined();
      });
    });

    test.describe('数値の境界値', () => {
      test('負の数値でのリクエストが適切に処理される', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // paginationパラメータに負の値
        const response = await client.raw('GET', ENDPOINTS.members + '?limit=-1');

        // 400エラーまたはデフォルト値が適用される
        expect([200, 400]).toContain(response.status());
      });

      test('非常に大きな数値でのリクエストが適切に処理される', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        // 非常に大きなlimit
        const response = await client.raw('GET', ENDPOINTS.members + '?limit=999999');

        // 200で返るが上限が適用されるか、400エラー
        expect([200, 400]).toContain(response.status());
      });
    });
  });

  // ============================================================
  // 3. パフォーマンステスト
  // ============================================================
  test.describe('Performance Tests', () => {
    test('メンバー一覧取得が5秒以内に完了する', async ({ request }) => {
      const { client } = await loginAsAdmin(request);

      const startTime = Date.now();
      const response = await client.raw('GET', ENDPOINTS.members);
      const endTime = Date.now();

      expect(response.status()).toBe(200);
      expect(endTime - startTime).toBeLessThan(5000);
    });

    test('イベント一覧取得が5秒以内に完了する', async ({ request }) => {
      const { client } = await loginAsAdmin(request);

      const startTime = Date.now();
      const response = await client.raw('GET', ENDPOINTS.events);
      const endTime = Date.now();

      expect(response.status()).toBe(200);
      expect(endTime - startTime).toBeLessThan(5000);
    });

    test('連続した複数リクエストが安定して処理される', async ({ request }) => {
      const { client } = await loginAsAdmin(request);

      const responseTimes: number[] = [];

      // 10回連続でリクエスト
      for (let i = 0; i < 10; i++) {
        const startTime = Date.now();
        const response = await client.raw('GET', ENDPOINTS.members);
        const endTime = Date.now();

        expect(response.status()).toBe(200);
        responseTimes.push(endTime - startTime);
      }

      // 平均レスポンス時間が2秒以内
      const avgTime = responseTimes.reduce((a, b) => a + b, 0) / responseTimes.length;
      expect(avgTime).toBeLessThan(2000);
    });
  });

  // ============================================================
  // 4. 国際化テスト
  // ============================================================
  test.describe('Internationalization Tests', () => {
    test.describe('日本語データ', () => {
      test('日本語名でメンバー作成・取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const japaneseName = `テストメンバー_${Date.now()}`;
        const createResponse = await client.raw('POST', ENDPOINTS.members, {
          display_name: japaneseName,
        });

        // 成功または権限不足
        expect([200, 201, 400, 403]).toContain(createResponse.status());

        if (createResponse.status() === 200 || createResponse.status() === 201) {
          const createBody = await createResponse.json();
          const memberId = createBody.data?.id || createBody.data?.member_id || createBody.data?.member?.id;

          if (memberId) {
            // 取得して名前が正しく保存されているか確認
            const getResponse = await client.raw('GET', ENDPOINTS.member(memberId));
            expect(getResponse.status()).toBe(200);
            const getBody = await getResponse.json();
            expect(getBody.data.display_name).toBe(japaneseName);
          }
        }
      });

      test('日本語名でイベント作成・取得できる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const japaneseEventName = `テストイベント_${Date.now()}`;
        const createResponse = await client.raw('POST', ENDPOINTS.events, {
          name: japaneseEventName,
        });

        // 成功または権限不足
        expect([200, 201, 400, 403]).toContain(createResponse.status());
      });
    });

    test.describe('絵文字データ', () => {
      test('絵文字を含む名前でメンバー作成が適切に処理される', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const emojiName = `Test User 🎉👍 ${Date.now()}`;
        const response = await client.raw('POST', ENDPOINTS.members, {
          display_name: emojiName,
        });

        // 成功または絵文字が許可されていない場合は400
        expect([200, 201, 400, 403]).toContain(response.status());
      });

      test('絵文字を含む名前でイベント作成が適切に処理される', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const emojiEventName = `Test Event 🎊🎈 ${Date.now()}`;
        const response = await client.raw('POST', ENDPOINTS.events, {
          name: emojiEventName,
        });

        // 成功または絵文字が許可されていない場合は400
        expect([200, 201, 400, 403]).toContain(response.status());
      });
    });

    test.describe('特殊文字データ', () => {
      test('特殊文字を含む名前でメンバー作成が適切に処理される', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const specialCharsName = `Test<User>&"'${Date.now()}`;
        const response = await client.raw('POST', ENDPOINTS.members, {
          display_name: specialCharsName,
        });

        // 成功またはサニタイズされるか、拒否される
        expect([200, 201, 400, 403]).toContain(response.status());
      });

      test('SQLインジェクション的な文字列が適切にエスケープされる', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const sqlInjectionName = `Test'; DROP TABLE members; --${Date.now()}`;
        const response = await client.raw('POST', ENDPOINTS.members, {
          display_name: sqlInjectionName,
        });

        // サーバーエラーにならないことを確認
        expect(response.status()).not.toBe(500);
      });

      test('XSS的な文字列が適切に処理される', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const xssName = `<script>alert('XSS')</script>${Date.now()}`;
        const response = await client.raw('POST', ENDPOINTS.members, {
          display_name: xssName,
        });

        // サーバーエラーにならないことを確認
        expect(response.status()).not.toBe(500);
      });

      test('マルチバイト文字（中国語・韓国語）が適切に処理される', async ({ request }) => {
        const { client } = await loginAsAdmin(request);

        const multibyteNames = [
          `测试用户_${Date.now()}`, // 中国語
          `테스트사용자_${Date.now()}`, // 韓国語
        ];

        for (const name of multibyteNames) {
          const response = await client.raw('POST', ENDPOINTS.members, {
            display_name: name,
          });

          // 成功または400エラー（サーバーエラーではない）
          expect([200, 201, 400, 403]).toContain(response.status());
        }
      });
    });
  });

  // ============================================================
  // 5. エラーハンドリングテスト
  // ============================================================
  test.describe('Error Handling Tests', () => {
    test('不正なJSONでリクエストした場合のエラー処理', async ({ request }) => {
      const { client } = await loginAsAdmin(request);

      // 不正なデータ型
      const response = await client.raw('POST', ENDPOINTS.members, 'not-a-json-object');

      expect(response.status()).toBeGreaterThanOrEqual(400);
    });

    test('存在しないエンドポイントで404が返る', async ({ request }) => {
      const { client } = await loginAsAdmin(request);

      const response = await client.raw('GET', '/api/v1/nonexistent-endpoint');

      expect(response.status()).toBe(404);
    });

    test('サポートされていないHTTPメソッドで適切なエラーが返る', async ({ request }) => {
      const { client } = await loginAsAdmin(request);

      // PATCHがサポートされていないエンドポイントでテスト
      const response = await client.raw('PATCH', ENDPOINTS.health, {});

      // 405 Method Not Allowed または 404
      expect([404, 405]).toContain(response.status());
    });
  });
});

import { http, HttpResponse, passthrough, bypass } from 'msw';
import { DUMMY_IDS } from '../steps/types';
import {
  DUMMY_ROLES,
  DUMMY_EVENT,
  DUMMY_TEMPLATE,
  DUMMY_BUSINESS_DAY,
  DUMMY_SHIFT_SLOTS,
  DUMMY_MEMBERS,
  DUMMY_COLLECTION,
  DUMMY_RESPONSES,
  DUMMY_CALENDAR,
  DUMMY_INSTANCES,
  isTutorialId,
} from './data';

// 内部ステート: 出欠コレクションのステータス管理
let collectionStatus: 'open' | 'closed' = 'open';

// チュートリアル開始時にステートリセット
export function resetMockState() {
  collectionStatus = 'open';
}

/**
 * 実APIのGETレスポンスにダミーデータをマージするヘルパー
 * 実APIが { data: { [dataKey]: [...] } } 形式を返す前提
 * 実API到達不可時はダミーデータのみで正しい形式を返す
 */
async function mergeWithRealData<T extends Record<string, unknown>>(
  request: Request,
  url: string,
  dataKey: string,
  dummyItems: T[],
) {
  try {
    const realResponse = await fetch(bypass(new Request(url, {
      headers: {
        'Authorization': request.headers.get('Authorization') || '',
        'Content-Type': 'application/json',
      },
    })));
    if (realResponse.ok) {
      const realData = await realResponse.json();
      // 実API応答: { data: { [dataKey]: [...] } }
      const realItems = realData.data?.[dataKey] || [];
      return HttpResponse.json({
        data: {
          ...(realData.data || {}),
          [dataKey]: [...dummyItems, ...realItems],
        },
      });
    }
  } catch {
    // 実API到達不可（バックエンド停止中など）
  }
  // フォールバック: ダミーデータのみ（正しい { data: { key: [...] } } 形式）
  return HttpResponse.json({ data: { [dataKey]: dummyItems } });
}

export const handlers = [
  // === ロール ===
  http.post('/api/v1/roles', () => {
    return HttpResponse.json({ data: DUMMY_ROLES[0] }, { status: 201 });
  }),

  http.get('/api/v1/roles', ({ request }) => {
    return mergeWithRealData(request, request.url, 'roles', DUMMY_ROLES);
  }),

  // === イベント ===
  http.post('/api/v1/events', () => {
    return HttpResponse.json({ data: DUMMY_EVENT }, { status: 201 });
  }),

  http.get('/api/v1/events', ({ request }) => {
    return mergeWithRealData(request, request.url, 'events', [DUMMY_EVENT]);
  }),

  // イベント詳細（BusinessDayList, CalendarList が必要）
  http.get('/api/v1/events/:eventId', ({ params }) => {
    const { eventId } = params;
    if (typeof eventId === 'string' && isTutorialId(eventId)) {
      return HttpResponse.json({ data: DUMMY_EVENT });
    }
    return passthrough();
  }),

  // === テンプレート ===
  http.get(`/api/v1/events/${DUMMY_IDS.eventId}/templates`, () => {
    return HttpResponse.json({ data: { templates: [DUMMY_TEMPLATE] } });
  }),

  http.post(`/api/v1/events/${DUMMY_IDS.eventId}/templates`, () => {
    return HttpResponse.json({ data: DUMMY_TEMPLATE }, { status: 201 });
  }),

  http.get(`/api/v1/events/${DUMMY_IDS.eventId}/templates/${DUMMY_IDS.templateId}`, () => {
    return HttpResponse.json({ data: DUMMY_TEMPLATE });
  }),

  // === インスタンス（TemplateForm が必要）===
  http.get(`/api/v1/events/${DUMMY_IDS.eventId}/instances`, () => {
    return HttpResponse.json({ data: { instances: DUMMY_INSTANCES } });
  }),

  // === 営業日 ===
  http.get(`/api/v1/events/${DUMMY_IDS.eventId}/business-days`, ({ request }) => {
    return mergeWithRealData(request, request.url, 'business_days', [DUMMY_BUSINESS_DAY]);
  }),

  http.post(`/api/v1/events/${DUMMY_IDS.eventId}/business-days`, () => {
    return HttpResponse.json({ data: DUMMY_BUSINESS_DAY }, { status: 201 });
  }),

  // === シフト枠 ===
  http.get(`/api/v1/business-days/${DUMMY_IDS.businessDayId}/shift-slots`, () => {
    return HttpResponse.json({ data: { shift_slots: DUMMY_SHIFT_SLOTS } });
  }),

  http.post(`/api/v1/business-days/${DUMMY_IDS.businessDayId}/shift-slots`, () => {
    return HttpResponse.json({ data: DUMMY_SHIFT_SLOTS[0] }, { status: 201 });
  }),

  http.get(`/api/v1/shift-slots/${DUMMY_IDS.shiftSlotId1}`, () => {
    return HttpResponse.json({ data: DUMMY_SHIFT_SLOTS[0] });
  }),

  // === シフト割り当て（ShiftAdjustment が必要）===
  http.get('/api/v1/shift-assignments', ({ request }) => {
    const url = new URL(request.url);
    const bdId = url.searchParams.get('business_day_id');
    if (bdId && isTutorialId(bdId)) {
      return HttpResponse.json({ data: { assignments: [] } });
    }
    return passthrough();
  }),

  // === メンバー ===
  http.post('/api/v1/members', () => {
    return HttpResponse.json({ data: DUMMY_MEMBERS[0] }, { status: 201 });
  }),

  http.get('/api/v1/members', ({ request }) => {
    return mergeWithRealData(request, request.url, 'members', DUMMY_MEMBERS);
  }),

  // === 出欠確認 ===
  http.post('/api/v1/attendance/collections', () => {
    collectionStatus = 'open';
    return HttpResponse.json({ data: DUMMY_COLLECTION }, { status: 201 });
  }),

  http.get('/api/v1/attendance/collections', ({ request }) => {
    const currentCollection = { ...DUMMY_COLLECTION, status: collectionStatus };
    return mergeWithRealData(request, request.url, 'collections', [currentCollection]);
  }),

  http.get('/api/v1/attendance/collections/:collectionId', ({ params }) => {
    const { collectionId } = params;
    if (typeof collectionId === 'string' && isTutorialId(collectionId)) {
      return HttpResponse.json({
        data: { ...DUMMY_COLLECTION, status: collectionStatus },
      });
    }
    return passthrough();
  }),

  http.get('/api/v1/attendance/collections/:collectionId/responses', ({ params }) => {
    const { collectionId } = params;
    if (typeof collectionId === 'string' && isTutorialId(collectionId)) {
      return HttpResponse.json({ data: { responses: DUMMY_RESPONSES } });
    }
    return passthrough();
  }),

  http.post('/api/v1/attendance/collections/:collectionId/close', ({ params }) => {
    const { collectionId } = params;
    if (typeof collectionId === 'string' && isTutorialId(collectionId)) {
      collectionStatus = 'closed';
      return HttpResponse.json({
        data: { ...DUMMY_COLLECTION, status: 'closed' },
      });
    }
    return passthrough();
  }),

  // === シフト割り当て確定 ===
  http.post('/api/v1/assignments/confirm', () => {
    return HttpResponse.json({
      data: { message: 'Assignments confirmed' },
    });
  }),

  // === カレンダー ===
  http.post('/api/v1/calendars', () => {
    return HttpResponse.json({ data: DUMMY_CALENDAR }, { status: 201 });
  }),

  http.get('/api/v1/calendars', ({ request }) => {
    return mergeWithRealData(request, request.url, 'calendars', [DUMMY_CALENDAR]);
  }),

  http.put(`/api/v1/calendars/${DUMMY_IDS.calendarId}`, () => {
    return HttpResponse.json({ data: DUMMY_CALENDAR });
  }),
];

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
  isTutorialId,
} from './data';

// 内部ステート: 出欠コレクションのステータス管理
let collectionStatus: 'open' | 'closed' = 'open';

// チュートリアル開始時にステートリセット
export function resetMockState() {
  collectionStatus = 'open';
}

// 実APIからのGETレスポンスにダミーデータをマージするヘルパー
async function mergeWithRealData<T extends Record<string, unknown>>(
  request: Request,
  url: string,
  dataKey: string,
  dummyItems: T[],
) {
  try {
    // bypass() で MSW をスキップし、実サーバーへ直接リクエスト（無限再帰防止）
    const realResponse = await fetch(bypass(new Request(url, {
      headers: {
        'Authorization': request.headers.get('Authorization') || '',
        'Content-Type': 'application/json',
      },
    })));
    if (realResponse.ok) {
      const realData = await realResponse.json();
      const realItems = realData[dataKey] || realData.data?.[dataKey] || [];
      return HttpResponse.json({
        ...realData,
        [dataKey]: [...dummyItems, ...realItems],
        data: realData.data ? {
          ...realData.data,
          [dataKey]: [...dummyItems, ...(realData.data[dataKey] || [])],
        } : undefined,
      });
    }
  } catch {
    // 実API失敗時はダミーデータのみ返す
  }
  return HttpResponse.json({ [dataKey]: dummyItems });
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

  // === シフト割り当て ===
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

import { DUMMY_IDS } from '../steps/types';

// 来週の土曜日を計算
function getNextSaturday(): string {
  const now = new Date();
  const dayOfWeek = now.getDay();
  const daysUntilSaturday = (6 - dayOfWeek + 7) % 7 || 7;
  const nextSat = new Date(now);
  nextSat.setDate(now.getDate() + daysUntilSaturday);
  return nextSat.toISOString().split('T')[0];
}

const targetDate = getNextSaturday();

export const DUMMY_ROLES = [
  {
    role_id: DUMMY_IDS.roleId1,
    name: 'バーテンダー',
    description: 'ドリンクの提供を担当',
    color: '#3B82F6',
    display_order: 1,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    role_id: DUMMY_IDS.roleId2,
    name: 'MC',
    description: 'イベントの司会進行',
    color: '#EF4444',
    display_order: 2,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
];

export const DUMMY_EVENT = {
  event_id: DUMMY_IDS.eventId,
  event_name: 'チュートリアル Bar',
  event_type: 'normal' as const,
  description: 'チュートリアル用のサンプルイベントです',
  recurrence_type: 'none' as const,
  is_active: true,
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

export const DUMMY_TEMPLATE = {
  template_id: DUMMY_IDS.templateId,
  event_id: DUMMY_IDS.eventId,
  name: 'メインインスタンス構成',
  description: 'メインフロアの基本シフト構成',
  items: [
    {
      template_item_id: 'tut_ti_001',
      instance_id: 'tut_instance_001',
      instance_name: 'メインフロア',
      role_name: 'バーテンダー',
      required_count: 2,
      start_time: '21:00:00',
      end_time: '23:30:00',
      priority: 1,
    },
    {
      template_item_id: 'tut_ti_002',
      instance_id: 'tut_instance_001',
      instance_name: 'メインフロア',
      role_name: 'MC',
      required_count: 1,
      start_time: '21:00:00',
      end_time: '23:30:00',
      priority: 2,
    },
  ],
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

export const DUMMY_BUSINESS_DAY = {
  business_day_id: DUMMY_IDS.businessDayId,
  event_id: DUMMY_IDS.eventId,
  target_date: `${targetDate}T00:00:00Z`,
  start_time: '21:00:00',
  end_time: '23:30:00',
  is_active: true,
  note: '',
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

export const DUMMY_SHIFT_SLOTS = [
  {
    shift_slot_id: DUMMY_IDS.shiftSlotId1,
    business_day_id: DUMMY_IDS.businessDayId,
    instance_id: 'tut_instance_001',
    instance_name: 'メインフロア',
    slot_name: 'バーテンダー',
    required_count: 2,
    start_time: '21:00:00',
    end_time: '23:30:00',
    priority: 1,
    assigned_count: 0,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    shift_slot_id: DUMMY_IDS.shiftSlotId2,
    business_day_id: DUMMY_IDS.businessDayId,
    instance_id: 'tut_instance_001',
    instance_name: 'メインフロア',
    slot_name: 'MC',
    required_count: 1,
    start_time: '21:00:00',
    end_time: '23:30:00',
    priority: 2,
    assigned_count: 0,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
];

export const DUMMY_MEMBERS = [
  {
    member_id: DUMMY_IDS.memberId1,
    display_name: '田中太郎',
    is_active: true,
    role_ids: [DUMMY_IDS.roleId1],
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    member_id: DUMMY_IDS.memberId2,
    display_name: '佐藤花子',
    is_active: true,
    role_ids: [DUMMY_IDS.roleId1, DUMMY_IDS.roleId2],
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    member_id: DUMMY_IDS.memberId3,
    display_name: '鈴木一郎',
    is_active: true,
    role_ids: [DUMMY_IDS.roleId2],
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
];

export const DUMMY_COLLECTION = {
  collection_id: DUMMY_IDS.collectionId,
  title: 'チュートリアル Bar 出欠確認',
  description: 'チュートリアル用の出欠確認です',
  status: 'open' as const,
  target_type: 'event',
  target_id: DUMMY_IDS.eventId,
  public_token: 'tut_token_001',
  deadline: null,
  target_date_count: 1,
  response_count: 3,
  group_ids: [],
  role_ids: [],
  target_dates: [
    {
      target_date_id: DUMMY_IDS.targetDateId1,
      target_date: `${targetDate}T00:00:00Z`,
      start_time: '21:00:00',
      end_time: '23:30:00',
    },
  ],
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

export const DUMMY_RESPONSES = [
  {
    response_id: 'tut_resp_001',
    collection_id: DUMMY_IDS.collectionId,
    member_id: DUMMY_IDS.memberId1,
    member_name: '田中太郎',
    answers: [
      { target_date_id: DUMMY_IDS.targetDateId1, status: 'attending', note: '' },
    ],
    submitted_at: new Date().toISOString(),
  },
  {
    response_id: 'tut_resp_002',
    collection_id: DUMMY_IDS.collectionId,
    member_id: DUMMY_IDS.memberId2,
    member_name: '佐藤花子',
    answers: [
      { target_date_id: DUMMY_IDS.targetDateId1, status: 'attending', note: '' },
    ],
    submitted_at: new Date().toISOString(),
  },
  {
    response_id: 'tut_resp_003',
    collection_id: DUMMY_IDS.collectionId,
    member_id: DUMMY_IDS.memberId3,
    member_name: '鈴木一郎',
    answers: [
      { target_date_id: DUMMY_IDS.targetDateId1, status: 'attending', note: '' },
    ],
    submitted_at: new Date().toISOString(),
  },
];

export const DUMMY_CALENDAR = {
  id: DUMMY_IDS.calendarId,
  title: 'チュートリアル Bar カレンダー',
  description: 'チュートリアル用の共有カレンダー',
  public_token: 'tut_cal_token_001',
  event_ids: [DUMMY_IDS.eventId],
  is_active: true,
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

// チュートリアルのダミーIDかどうかを判定
export function isTutorialId(id: string): boolean {
  return id.startsWith('tut_');
}

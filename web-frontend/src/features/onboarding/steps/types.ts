export type OnboardingPhase =
  | 'idle'
  | 'sidebar'
  | 'role'
  | 'event'
  | 'template'
  | 'businessDay'
  | 'shiftSlot'
  | 'member'
  | 'attendance'
  | 'attendanceResponse'
  | 'attendanceDetail'
  | 'shiftAdjustment'
  | 'calendar'
  | 'summary'
  | 'complete';

export const PHASE_ORDER: OnboardingPhase[] = [
  'sidebar',
  'role',
  'event',
  'template',
  'businessDay',
  'shiftSlot',
  'member',
  'attendance',
  'attendanceResponse',
  'attendanceDetail',
  'shiftAdjustment',
  'calendar',
  'summary',
  'complete',
];

export interface DummyIds {
  roleId1: string;
  roleId2: string;
  eventId: string;
  templateId: string;
  businessDayId: string;
  shiftSlotId1: string;
  shiftSlotId2: string;
  memberId1: string;
  memberId2: string;
  memberId3: string;
  collectionId: string;
  targetDateId1: string;
  calendarId: string;
}

export const DUMMY_IDS: DummyIds = {
  roleId1: 'tut_role_001',
  roleId2: 'tut_role_002',
  eventId: 'tut_event_001',
  templateId: 'tut_template_001',
  businessDayId: 'tut_bd_001',
  shiftSlotId1: 'tut_slot_001',
  shiftSlotId2: 'tut_slot_002',
  memberId1: 'tut_member_001',
  memberId2: 'tut_member_002',
  memberId3: 'tut_member_003',
  collectionId: 'tut_collection_001',
  targetDateId1: 'tut_td_001',
  calendarId: 'tut_calendar_001',
};

export interface OnboardingState {
  isActive: boolean;
  currentPhase: OnboardingPhase;
  dummyIds: DummyIds;
  mswReady: boolean;
}

export type OnboardingAction =
  | { type: 'START' }
  | { type: 'STOP' }
  | { type: 'SET_PHASE'; phase: OnboardingPhase }
  | { type: 'NEXT_PHASE' }
  | { type: 'SET_MSW_READY'; ready: boolean }
  | { type: 'RESTORE'; state: OnboardingState };

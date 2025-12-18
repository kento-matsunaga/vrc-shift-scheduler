// API レスポンスの共通型
export interface ApiResponse<T> {
  data: T;
}

export interface ApiError {
  error: {
    code: string;
    message: string;
    details?: Record<string, unknown>;
  };
}

// Event 関連
export interface Event {
  event_id: string;
  tenant_id: string;
  event_name: string;
  event_type: 'normal' | 'special';
  description: string;
  is_active: boolean;
  recurrence_type: 'none' | 'weekly' | 'biweekly';
  recurrence_start_date?: string; // YYYY-MM-DD
  recurrence_day_of_week?: number; // 0-6: 日曜=0, 土曜=6
  default_start_time?: string; // HH:MM:SS
  default_end_time?: string; // HH:MM:SS
  created_at: string;
  updated_at: string;
}

export interface GenerateBusinessDaysResponse {
  generated_count: number;
  message: string;
  event: Event;
}

export interface EventListResponse {
  events: Event[];
  count: number;
}

// EventBusinessDay 関連
export interface BusinessDay {
  business_day_id: string;
  tenant_id: string;
  event_id: string;
  target_date: string; // YYYY-MM-DD
  start_time: string; // HH:MM:SS
  end_time: string; // HH:MM:SS
  occurrence_type: 'recurring' | 'special';
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface BusinessDayListResponse {
  business_days: BusinessDay[];
  count: number;
}

// ShiftSlot 関連
export interface ShiftSlot {
  slot_id: string;
  tenant_id: string;
  business_day_id: string;
  position_id: string;
  slot_name: string;
  instance_name: string;
  start_time: string; // HH:MM:SS
  end_time: string; // HH:MM:SS
  required_count: number;
  assigned_count?: number; // API から取得時に含まれる
  priority: number;
  is_overnight: boolean;
  created_at: string;
  updated_at: string;
}

export interface ShiftSlotListResponse {
  shift_slots: ShiftSlot[];
  count: number;
}

// ShiftAssignment 関連
export interface ShiftAssignment {
  assignment_id: string;
  tenant_id: string;
  slot_id: string;
  member_id: string;
  member_display_name?: string; // JOIN で取得する場合
  slot_name?: string; // JOIN で取得する場合
  target_date?: string; // JOIN で取得する場合
  start_time?: string; // JOIN で取得する場合
  end_time?: string; // JOIN で取得する場合
  assignment_status: 'confirmed' | 'cancelled';
  assignment_method: 'auto' | 'manual';
  is_outside_preference: boolean;
  assigned_at: string;
  cancelled_at?: string;
  created_at: string;
  updated_at: string;
}

export interface ShiftAssignmentListResponse {
  assignments: ShiftAssignment[];
  count: number;
}

// Member 関連
export interface Member {
  member_id: string;
  tenant_id: string;
  display_name: string;
  discord_user_id?: string;
  email?: string;
  is_active: boolean;
  role_ids?: string[]; // Assigned role IDs
  created_at: string;
  updated_at: string;
}

export interface MemberListResponse {
  members: Member[];
  count: number;
}

// Recent Attendance 関連
export interface TargetDateInfo {
  target_date_id: string;
  target_date: string; // ISO 8601
  display_order: number;
}

export interface MemberAttendanceStatus {
  member_id: string;
  member_name: string;
  attendance_map: Record<string, string>; // target_date_id -> "attending" | "absent" | ""
}

export interface RecentAttendanceResponse {
  target_dates: TargetDateInfo[];
  member_attendances: MemberAttendanceStatus[];
}

// Position 関連
export interface Position {
  position_id: string;
  tenant_id: string;
  position_name: string;
  description: string;
  display_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface PositionListResponse {
  positions: Position[];
  count: number;
}

// ShiftSlotTemplate 関連
export interface TemplateItem {
  item_id?: string;
  position_id: string;
  slot_name: string;
  instance_name: string;
  start_time: string; // HH:MM:SS
  end_time: string; // HH:MM:SS
  required_count: number;
  priority: number;
}

export interface Template {
  template_id: string;
  tenant_id: string;
  event_id: string;
  template_name: string;
  description: string;
  items: TemplateItem[];
  created_at: string;
  updated_at: string;
}

export interface TemplateListResponse {
  templates: Template[];
  count: number;
}

export interface CreateTemplateRequest {
  template_name: string;
  description: string;
  items: TemplateItem[];
}

export interface UpdateTemplateRequest {
  template_name: string;
  description: string;
  items: TemplateItem[];
}

export interface SaveAsTemplateRequest {
  template_name: string;
  description: string;
}


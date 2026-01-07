/**
 * 公開API（認証不要）
 * 出欠確認・日程調整の公開回答ページ用
 */

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';

/**
 * 公開APIエラー
 */
export class PublicApiError extends Error {
  public statusCode: number;

  constructor(message: string, statusCode: number) {
    super(message);
    this.name = 'PublicApiError';
    this.statusCode = statusCode;
  }

  isNotFound(): boolean {
    return this.statusCode === 404;
  }

  isBadRequest(): boolean {
    return this.statusCode === 400;
  }

  isForbidden(): boolean {
    return this.statusCode === 403;
  }
}

/**
 * 公開API用の汎用リクエストヘルパー
 */
async function publicRequest<T>(
  method: string,
  path: string,
  body?: unknown
): Promise<T> {
  try {
    const res = await fetch(`${API_BASE_URL}${path}`, {
      method,
      headers: {
        'Content-Type': 'application/json',
      },
      body: body ? JSON.stringify(body) : undefined,
    });

    if (!res.ok) {
      const errorText = await res.text().catch(() => 'Unknown error');
      throw new PublicApiError(errorText, res.status);
    }

    if (res.status === 204) {
      return null as T;
    }

    return await res.json();
  } catch (error) {
    if (error instanceof PublicApiError) {
      throw error;
    }
    throw new PublicApiError(
      error instanceof Error ? error.message : 'Network error',
      0
    );
  }
}

// ==========================================
// 出欠確認 公開API
// ==========================================

export interface TargetDate {
  target_date_id: string;
  target_date: string; // ISO 8601 format
  start_time?: string; // HH:MM format (optional)
  end_time?: string;   // HH:MM format (optional)
  display_order: number;
}

export interface AttendanceCollection {
  collection_id: string;
  tenant_id: string;
  title: string;
  description: string;
  target_type: string;
  target_id: string;
  target_dates?: TargetDate[]; // Target dates with IDs
  public_token: string;
  status: 'open' | 'closed';
  deadline?: string;
  group_ids?: string[]; // Target group IDs
  role_ids?: string[]; // Target role IDs
  created_at: string;
  updated_at: string;
}

export interface Member {
  member_id: string;
  tenant_id: string;
  display_name: string;
  discord_user_id?: string;
  email?: string;
  is_active: boolean;
}

export interface AttendanceSubmitRequest {
  member_id: string;
  target_date_id: string;
  response: 'attending' | 'absent' | 'undecided';
  note?: string;
  available_from?: string; // 参加可能開始時間 (HH:MM)
  available_to?: string;   // 参加可能終了時間 (HH:MM)
}

export interface AttendanceSubmitResponse {
  response_id: string;
  collection_id: string;
  member_id: string;
  response: string;
  note: string;
  responded_at: string;
}

/**
 * 出欠確認情報を取得（公開）
 */
export async function getAttendanceByToken(token: string): Promise<AttendanceCollection> {
  const response = await publicRequest<{ data: AttendanceCollection }>('GET', `/api/v1/public/attendance/${token}`);
  return response.data;
}

/**
 * メンバー一覧を取得（出欠確認用）
 * NOTE: MVPでは簡易実装として公開APIでメンバー一覧を取得可能
 * @param tenantId テナントID
 * @param groupIds 対象グループID - 指定するとそのグループに属するメンバーのみ返す
 * @param roleIds 対象ロールID - 指定するとそのロールを持つメンバーのみ返す
 * 両方指定した場合は AND 条件（両方に該当するメンバーのみ）
 */
export async function getMembers(
  tenantId: string,
  groupIds?: string[],
  roleIds?: string[]
): Promise<{ data: { members: Member[] } }> {
  let url = `/api/v1/public/members?tenant_id=${tenantId}`;
  if (groupIds && groupIds.length > 0) {
    url += `&group_ids=${groupIds.join(',')}`;
  }
  if (roleIds && roleIds.length > 0) {
    url += `&role_ids=${roleIds.join(',')}`;
  }
  return publicRequest<{ data: { members: Member[] } }>('GET', url);
}

/**
 * 出欠回答を送信（公開）
 */
export async function submitAttendanceResponse(
  token: string,
  data: AttendanceSubmitRequest
): Promise<AttendanceSubmitResponse> {
  const response = await publicRequest<{ data: AttendanceSubmitResponse }>(
    'POST',
    `/api/v1/public/attendance/${token}/responses`,
    data
  );
  return response.data;
}

/**
 * メンバーの既存回答を取得（公開）
 */
export interface MemberAttendanceResponse {
  target_date_id: string;
  response: 'attending' | 'absent' | 'undecided';
  note: string;
  available_from?: string;
  available_to?: string;
}

export interface MemberResponsesResult {
  member_id: string;
  responses: MemberAttendanceResponse[];
}

export async function getMemberAttendanceResponses(
  token: string,
  memberId: string
): Promise<MemberResponsesResult> {
  const response = await publicRequest<{ data: MemberResponsesResult }>(
    'GET',
    `/api/v1/public/attendance/${token}/members/${memberId}/responses`
  );
  return response.data;
}

// ==========================================
// 日程調整 公開API
// ==========================================

export interface DateSchedule {
  schedule_id: string;
  tenant_id: string;
  title: string;
  description: string;
  event_id?: string;
  public_token: string;
  status: 'open' | 'closed' | 'decided';
  deadline?: string;
  decided_candidate_id?: string;
  candidates: ScheduleCandidate[];
  group_ids?: string[]; // Target group IDs
  created_at: string;
  updated_at: string;
}

export interface ScheduleCandidate {
  candidate_id: string;
  date: string;
  start_time?: string;
  end_time?: string;
}

export interface ScheduleResponseInput {
  candidate_id: string;
  availability: 'available' | 'unavailable' | 'maybe';
  note?: string;
}

export interface ScheduleSubmitRequest {
  member_id: string;
  responses: ScheduleResponseInput[];
}

export interface ScheduleSubmitResponse {
  schedule_id: string;
  member_id: string;
  responded_at: string;
}

/**
 * 日程調整情報を取得（公開）
 */
export async function getScheduleByToken(token: string): Promise<DateSchedule> {
  const response = await publicRequest<{ data: DateSchedule }>('GET', `/api/v1/public/schedules/${token}`);
  return response.data;
}

/**
 * 日程調整回答を送信（公開）
 */
export async function submitScheduleResponse(
  token: string,
  data: ScheduleSubmitRequest
): Promise<ScheduleSubmitResponse> {
  const response = await publicRequest<{ data: ScheduleSubmitResponse }>(
    'POST',
    `/api/v1/public/schedules/${token}/responses`,
    data
  );
  return response.data;
}

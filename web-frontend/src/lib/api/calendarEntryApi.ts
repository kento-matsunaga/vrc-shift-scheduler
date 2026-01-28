import { apiClient } from '../apiClient';
import type { ApiResponse } from '../../types/api';

export interface CalendarEntry {
  entry_id: string;
  calendar_id: string;
  title: string;
  date: string;
  start_time?: string;
  end_time?: string;
  note: string;
  created_at: string;
  updated_at: string;
}

export interface CreateCalendarEntryInput {
  title: string;
  date: string;
  start_time?: string;
  end_time?: string;
  note?: string;
}

export interface CalendarEntryListResponse {
  entries: CalendarEntry[];
}

export async function createCalendarEntry(
  calendarId: string,
  input: CreateCalendarEntryInput
): Promise<CalendarEntry> {
  const res = await apiClient.post<ApiResponse<CalendarEntry>>(
    `/api/v1/calendars/${calendarId}/entries`,
    input
  );
  return res.data;
}

export async function listCalendarEntries(calendarId: string): Promise<CalendarEntry[]> {
  const res = await apiClient.get<ApiResponse<CalendarEntryListResponse>>(
    `/api/v1/calendars/${calendarId}/entries`
  );
  return res.data.entries || [];
}

export async function updateCalendarEntry(
  calendarId: string,
  entryId: string,
  input: CreateCalendarEntryInput
): Promise<CalendarEntry> {
  const res = await apiClient.put<ApiResponse<CalendarEntry>>(
    `/api/v1/calendars/${calendarId}/entries/${entryId}`,
    input
  );
  return res.data;
}

export async function deleteCalendarEntry(calendarId: string, entryId: string): Promise<void> {
  await apiClient.delete(`/api/v1/calendars/${calendarId}/entries/${entryId}`);
}

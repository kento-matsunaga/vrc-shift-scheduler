import { apiClient } from '../apiClient';
import type {
  ApiResponse,
  Template,
  TemplateListResponse,
  CreateTemplateRequest,
  UpdateTemplateRequest,
  SaveAsTemplateRequest
} from '../../types/api';

/**
 * テンプレート一覧を取得
 */
export async function listTemplates(eventId: string): Promise<Template[]> {
  const res = await apiClient.get<ApiResponse<TemplateListResponse>>(
    `/api/v1/events/${eventId}/templates`
  );
  return res.data.templates;
}

/**
 * テンプレート詳細を取得
 */
export async function getTemplate(
  eventId: string,
  templateId: string
): Promise<Template> {
  const res = await apiClient.get<ApiResponse<Template>>(
    `/api/v1/events/${eventId}/templates/${templateId}`
  );
  return res.data;
}

/**
 * テンプレートを作成
 */
export async function createTemplate(
  eventId: string,
  data: CreateTemplateRequest
): Promise<Template> {
  const res = await apiClient.post<ApiResponse<Template>>(
    `/api/v1/events/${eventId}/templates`,
    data
  );
  return res.data;
}

/**
 * テンプレートを更新
 */
export async function updateTemplate(
  eventId: string,
  templateId: string,
  data: UpdateTemplateRequest
): Promise<Template> {
  const res = await apiClient.put<ApiResponse<Template>>(
    `/api/v1/events/${eventId}/templates/${templateId}`,
    data
  );
  return res.data;
}

/**
 * テンプレートを削除
 */
export async function deleteTemplate(
  eventId: string,
  templateId: string
): Promise<void> {
  await apiClient.delete(`/api/v1/events/${eventId}/templates/${templateId}`);
}

/**
 * 営業日からテンプレートを作成
 */
export async function saveBusinessDayAsTemplate(
  businessDayId: string,
  data: SaveAsTemplateRequest
): Promise<Template> {
  const res = await apiClient.post<ApiResponse<Template>>(
    `/api/v1/business-days/${businessDayId}/save-as-template`,
    data
  );
  return res.data;
}

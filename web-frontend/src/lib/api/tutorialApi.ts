import { apiClient } from '../apiClient';
import type { ApiResponse } from '../../types/api';

export interface Tutorial {
  id: string;
  category: string;
  title: string;
  body: string;
  display_order: number;
  is_published: boolean;
  created_at: string;
}

export interface TutorialListResponse {
  tutorials: Tutorial[];
}

/**
 * チュートリアル一覧取得
 */
export async function getTutorials(): Promise<Tutorial[]> {
  const res = await apiClient.get<ApiResponse<TutorialListResponse>>('/api/v1/tutorials');
  return res.data.tutorials;
}

/**
 * チュートリアル詳細取得
 */
export async function getTutorial(id: string): Promise<Tutorial> {
  const res = await apiClient.get<ApiResponse<{ tutorial: Tutorial }>>(`/api/v1/tutorials/${id}`);
  return res.data.tutorial;
}

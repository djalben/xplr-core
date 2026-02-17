import apiClient from './axios';

export interface GradeInfo {
  grade: string;
  total_spent: string;
  fee_percent: string;
  next_grade?: string;
  next_spend?: string;
}

// Получить информацию о Grade пользователя
export const getUserGrade = async (): Promise<GradeInfo> => {
  const response = await apiClient.get<GradeInfo>('/user/grade');
  return response.data;
};

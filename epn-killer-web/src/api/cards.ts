import apiClient from './axios';

export interface Card {
  id: number;
  user_id: number;
  team_id?: number;
  bin: string;
  last_4_digits: string;
  card_status: string;
  nickname?: string;
  daily_spend_limit: string;
  failed_auth_count: number;
  card_type?: string;
  auto_replenish_enabled?: boolean;
  auto_replenish_threshold?: string;
  auto_replenish_amount?: string;
  card_balance?: string;
  created_at: string;
}

export interface MassIssueRequest {
  count: number;
  daily_limit: number;
  nickname: string;
  merchant_name: string;
  card_type?: string;
  team_id?: number;
}

export interface CardIssueResult {
  success: boolean;
  status: string;
  card_last_4: string;
  nickname: string;
  message: string;
  card?: Card;
}

export interface MassIssueResponse {
  successful_count: number;
  failed_count: number;
  results: CardIssueResult[];
}

// Выпустить виртуальные карты
export const issueCards = async (data: MassIssueRequest): Promise<MassIssueResponse> => {
  const response = await apiClient.post<MassIssueResponse>('/user/cards/issue', data);
  return response.data;
};

// Получить список карт пользователя
export const getUserCards = async (): Promise<Card[]> => {
  const response = await apiClient.get<Card[]>('/user/cards');
  return response.data;
};

// Установить автопополнение карты
export const setCardAutoReplenishment = async (
  cardId: number,
  data: { enabled: boolean; threshold: number; amount: number }
) => {
  const response = await apiClient.post(`/user/cards/${cardId}/auto-replenishment`, data);
  return response.data;
};

// Отключить автопополнение карты
export const unsetCardAutoReplenishment = async (cardId: number) => {
  const response = await apiClient.delete(`/user/cards/${cardId}/auto-replenishment`);
  return response.data;
};

import apiClient from './axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || '';

export interface Card {
  id: number;
  user_id: number;
  team_id?: number;
  provider_card_id: string;
  external_id?: string;
  bin: string;
  last_4_digits: string;
  card_status: string;
  status?: string;
  nickname?: string;
  service_slug?: string;
  daily_spend_limit: string;
  spend_limit?: string;
  failed_auth_count: number;
  card_type?: string;
  category?: string;
  auto_replenish_enabled?: boolean;
  auto_replenish_threshold?: string;
  auto_replenish_amount?: string;
  card_balance?: string;
  spending_limit?: string;
  spent_from_wallet?: string;
  expiry_date?: string;
  created_at: string;
}

export interface MassIssueRequest {
  count: number;
  daily_limit: number;
  nickname: string;
  merchant_name: string;
  card_type?: string;
  service_slug?: string;
  category?: string;
  team_id?: number;
  price_rub?: number;
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
  const response = await apiClient.post<MassIssueResponse>(`${API_BASE_URL}/user/cards/issue`, data);
  return response.data;
};

// Выпустить персональную карту (subscriptions/travel/premium)
export const issuePersonalCard = async (
  cardType: 'subscriptions' | 'travel' | 'premium',
  priceRub: number
): Promise<MassIssueResponse> => {
  const categoryMap: Record<string, string> = {
    subscriptions: 'services',
    travel: 'travel',
    premium: 'travel',
  };
  const nicknameMap: Record<string, string> = {
    subscriptions: 'Карта для подписок',
    travel: 'Карта для путешествий',
    premium: 'Премиальная карта',
  };

  const data: MassIssueRequest = {
    count: 1,
    daily_limit: cardType === 'premium' ? 50000 : 10000,
    nickname: nicknameMap[cardType],
    merchant_name: 'XPLR Personal',
    card_type: 'MasterCard',
    service_slug: cardType,
    category: categoryMap[cardType],
    price_rub: priceRub,
  };

  return issueCards(data);
};

// Получить список карт пользователя
export const getUserCards = async (): Promise<Card[]> => {
  const response = await apiClient.get<Card[]>(`${API_BASE_URL}/user/cards`);
  return response.data;
};

// Получить детали карты (PAN, CVV, expiry)
export const getCardDetails = async (cardId: number): Promise<{
  card_id: number;
  full_number: string;
  cvv: string;
  expiry: string;
  card_type: string;
  bin: string;
  last_4: string;
}> => {
  const response = await apiClient.get(`${API_BASE_URL}/user/cards/${cardId}/details`);
  return response.data;
};

// Установить автопополнение карты
export const setCardAutoReplenishment = async (
  cardId: number,
  data: { enabled: boolean; threshold: number; amount: number }
) => {
  const response = await apiClient.post(`${API_BASE_URL}/user/cards/${cardId}/auto-replenishment`, data);
  return response.data;
};

// Отключить автопополнение карты
export const unsetCardAutoReplenishment = async (cardId: number) => {
  const response = await apiClient.delete(`${API_BASE_URL}/user/cards/${cardId}/auto-replenishment`);
  return response.data;
};

// Изменить статус карты
export const updateCardStatus = async (cardId: number, status: string): Promise<void> => {
  await apiClient.patch(`${API_BASE_URL}/user/cards/${cardId}/status`, { status });
};

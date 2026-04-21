import apiClient from './axios';

export interface Card {
  id: string;
  user_id: string;
  provider_card_id: string;
  bin: string;
  last_4_digits: string;
  card_status: string;
  status?: string;
  nickname?: string;
  daily_spend_limit: string;
  failed_auth_count: number;
  card_type?: string; // subscriptions | travel | premium (sandbox BFF)
  currency?: string;
  balance?: string;
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
  price_usd?: number;
  currency?: string;
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

// Выпустить персональную карту (subscriptions/travel/premium)
export const issuePersonalCard = async (
  cardType: 'subscriptions' | 'travel' | 'premium',
  priceUsd: number,
  currency: 'USD' | 'EUR'
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
    price_usd: priceUsd,
    currency,
  };

  return issueCards(data);
};

// Получить список карт пользователя
export const getUserCards = async (): Promise<Card[]> => {
  const response = await apiClient.get<Card[]>('/user/cards');
  return response.data;
};

// Получить детали карты (PAN, CVV, expiry)
export const getCardDetails = async (cardId: string): Promise<{
  card_id: string;
  full_number: string;
  cvv: string;
  expiry: string;
  card_type: string;
  bin: string;
  last_4: string;
}> => {
  const response = await apiClient.get(`/user/cards/${cardId}/details`);
  return response.data;
};

// Установить автопополнение карты
export const setCardAutoReplenishment = async (
  cardId: string,
  data: { enabled: boolean; threshold: number; amount: number }
) => {
  const response = await apiClient.post(`/user/cards/${cardId}/auto-replenishment`, data);
  return response.data;
};

// Отключить автопополнение карты
export const unsetCardAutoReplenishment = async (cardId: string) => {
  const response = await apiClient.delete(`/user/cards/${cardId}/auto-replenishment`);
  return response.data;
};

// Изменить статус карты
export const updateCardStatus = async (cardId: string, status: string): Promise<void> => {
  await apiClient.patch(`/user/cards/${cardId}/status`, { status });
};

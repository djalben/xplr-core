import apiClient from './axios';

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
  is_auto_pay_enabled?: boolean;
  created_at: string;
}

export interface CardSubscription {
  id: number;
  merchant_name: string;
  last_amount: string;
  last_currency: string;
  charge_count: number;
  is_allowed: boolean;
  first_seen_at: string;
  last_seen_at: string;
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
  priceUsd: number
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
    currency: 'USD',
  };

  return issueCards(data);
};

// Получить список карт пользователя
export const getUserCards = async (): Promise<Card[]> => {
  const response = await apiClient.get<Card[]>('/user/cards');
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
  const response = await apiClient.get(`/user/cards/${cardId}/details`);
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

// Изменить статус карты
export const updateCardStatus = async (cardId: number, status: string): Promise<void> => {
  await apiClient.patch(`/user/cards/${cardId}/status`, { status });
};

// Toggle auto-pay (recurring/subscription filter)
export const toggleAutoPay = async (cardId: number, enabled: boolean): Promise<void> => {
  await apiClient.patch(`/user/cards/${cardId}/auto-pay`, { enabled });
};

// Get card subscriptions (tracked merchants)
export const getCardSubscriptions = async (cardId: number): Promise<{
  subscriptions: CardSubscription[];
  is_auto_pay_enabled: boolean;
}> => {
  const response = await apiClient.get(`/user/cards/${cardId}/subscriptions`);
  return response.data;
};

// Toggle individual subscription allow/block
export const toggleSubscription = async (cardId: number, subId: number, isAllowed: boolean): Promise<void> => {
  await apiClient.patch(`/user/cards/${cardId}/subscriptions/${subId}`, { is_allowed: isAllowed });
};

// Freeze/unfreeze all subscriptions for a card
export const freezeAllSubscriptions = async (cardId: number, freeze: boolean): Promise<{ is_allowed: boolean }> => {
  const response = await apiClient.post(`/user/cards/${cardId}/freeze-all-subscriptions`, { freeze });
  return response.data;
};

// Check Telegram link status
export const getTelegramStatus = async (): Promise<{ telegram_linked: boolean }> => {
  const response = await apiClient.get('/user/telegram-status');
  return response.data;
};

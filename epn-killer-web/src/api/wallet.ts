import apiClient from './axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || '';

export interface InternalBalance {
  id: number;
  user_id: number;
  master_balance: number;
  updated_at: string;
}

// Получить текущий баланс Кошелька
export const getWallet = async (): Promise<InternalBalance> => {
  const res = await apiClient.get<InternalBalance>(`${API_BASE_URL}/user/wallet`);
  return res.data;
};

// Пополнить Кошелёк из баланса пользователя
export const topUpWallet = async (amount: number): Promise<InternalBalance> => {
  const res = await apiClient.post<InternalBalance>(`${API_BASE_URL}/user/wallet/topup`, { amount });
  return res.data;
};

// Установить лимит списания карты из Кошелька
export const setCardSpendingLimit = async (cardId: number, spendingLimit: number): Promise<void> => {
  await apiClient.patch(`${API_BASE_URL}/user/cards/${cardId}/spending-limit`, {
    spending_limit: spendingLimit,
  });
};

// Перевести средства из Кошелька на карту (внутренний перевод)
export const transferWalletToCard = async (
  cardId: string,
  amount: number,
  currency: string
): Promise<InternalBalance> => {
  const res = await apiClient.post<InternalBalance>(
    `${API_BASE_URL}/user/wallet/transfer-to-card`,
    { card_id: cardId, amount, currency }
  );
  return res.data;
};

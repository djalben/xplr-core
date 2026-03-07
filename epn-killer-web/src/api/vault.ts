import apiClient from './axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || '';

export interface InternalBalance {
  id: number;
  user_id: number;
  master_balance: number;
  updated_at: string;
}

// Получить текущий баланс Сейфа
export const getVault = async (): Promise<InternalBalance> => {
  const res = await apiClient.get<InternalBalance>(`${API_BASE_URL}/user/vault`);
  return res.data;
};

// Пополнить Сейф из баланса пользователя
export const topUpVault = async (amount: number): Promise<InternalBalance> => {
  const res = await apiClient.post<InternalBalance>(`${API_BASE_URL}/user/vault/topup`, { amount });
  return res.data;
};

// Установить лимит списания карты из Сейфа
export const setCardSpendingLimit = async (cardId: number, spendingLimit: number): Promise<void> => {
  await apiClient.patch(`${API_BASE_URL}/user/cards/${cardId}/spending-limit`, {
    spending_limit: spendingLimit,
  });
};

import apiClient from './axios';

export interface InternalBalance {
  id: number;
  user_id: number;
  master_balance: number;
  auto_topup_enabled: boolean;
  updated_at: string;
}

// Получить текущий баланс Кошелька
export const getWallet = async (): Promise<InternalBalance> => {
  const res = await apiClient.get<InternalBalance>('/user/wallet');
  return res.data;
};

// Пополнить Кошелёк из баланса пользователя
export const topUpWallet = async (amount: number): Promise<InternalBalance> => {
  const res = await apiClient.post<InternalBalance>('/user/wallet/topup', { amount });
  return res.data;
};

// Установить лимит списания карты из Кошелька
export const setCardSpendingLimit = async (cardId: number, spendingLimit: number): Promise<void> => {
  await apiClient.patch(`/user/cards/${cardId}/spending-limit`, {
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
    '/user/wallet/transfer-to-card',
    { card_id: cardId, amount, currency }
  );
  return res.data;
};

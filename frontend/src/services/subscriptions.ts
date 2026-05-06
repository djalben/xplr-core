import apiClient from './axios';

export interface CardSubscription {
  id: string;
  userId: string;
  cardId: string;
  merchantName: string;
  merchantKey: string;
  lastAmount: string;
  lastCurrency: string;
  chargeCount: number;
  firstSeenAt: string;
  lastSeenAt: string;
  isBlocked: boolean;
  blockedAt?: string | null;
}

export async function listSubscriptions(): Promise<CardSubscription[]> {
  const res = await apiClient.get('/subscriptions');
  return res.data?.subscriptions ?? [];
}

export async function setSubscriptionBlocked(id: string, isBlocked: boolean): Promise<void> {
  await apiClient.patch(`/subscriptions/${id}`, { is_blocked: isBlocked });
}

export async function setBlockedByCard(cardId: string, isBlocked: boolean): Promise<void> {
  await apiClient.post(`/subscriptions/block-by-card/${cardId}`, { is_blocked: isBlocked });
}


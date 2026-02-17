import apiClient from './axios';

export interface ReferralStats {
  total_referrals: number;
  active_referrals: number;
  total_commission: string;
  referral_code: string;
}

// Получить статистику реферальной программы
export const getReferralStats = async (): Promise<ReferralStats> => {
  const response = await apiClient.get<ReferralStats>('/user/referrals');
  return response.data;
};

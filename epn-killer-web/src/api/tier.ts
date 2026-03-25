import axios from 'axios';

const baseURL = '/api/v1';

export interface TierInfo {
  tier: 'standard' | 'gold';
  tier_expires_at: string | null;
  card_limit: number;
  current_cards: number;
  gold_price: string;
  gold_duration: string;
  can_issue_more: boolean;
}

export const getTierInfo = async (): Promise<TierInfo> => {
  const response = await axios.get(`${baseURL}/user/tier-info`);
  return response.data;
};

export const upgradeTier = async (): Promise<{ status: string; tier: string; expires_at: string; paid: number }> => {
  const response = await axios.post(`${baseURL}/user/upgrade-tier`);
  return response.data;
};

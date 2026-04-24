import apiClient from './axios';

export interface SystemSetting {
  key: string;
  value: string;
  description: string;
}

export interface SBPStatus {
  enabled: boolean;
}

export const getSBPStatus = async (): Promise<SBPStatus> => {
  const response = await apiClient.get('/sbp-status');
  return { enabled: Boolean(response.data.sbp_enabled ?? response.data.enabled) };
};

export const getSystemSettings = async (): Promise<SystemSetting[]> => {
  const response = await apiClient.get('/admin/system-settings');
  return response.data;
};

export const updateSystemSetting = async (key: string, value: string): Promise<void> => {
  await apiClient.patch(`/admin/system-settings/${key}`, { value });
};

import axios from 'axios';
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
  const response = await axios.get('/api/v1/sbp-status');
  return response.data;
};

export const getSystemSettings = async (): Promise<SystemSetting[]> => {
  const response = await apiClient.get('/admin/system-settings');
  return response.data;
};

export const updateSystemSetting = async (key: string, value: string): Promise<void> => {
  await apiClient.patch(`/admin/system-settings/${key}`, { value });
};

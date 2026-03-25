import axios from 'axios';

const baseURL = '/api/v1';

export interface SystemSetting {
  key: string;
  value: string;
  description: string;
}

export interface SBPStatus {
  enabled: boolean;
}

export const getSBPStatus = async (): Promise<SBPStatus> => {
  const response = await axios.get(`${baseURL}/sbp-status`);
  return response.data;
};

export const getSystemSettings = async (): Promise<SystemSetting[]> => {
  const response = await axios.get(`${baseURL}/admin/system-settings`);
  return response.data;
};

export const updateSystemSetting = async (key: string, value: string): Promise<void> => {
  await axios.patch(`${baseURL}/admin/system-settings/${key}`, { value });
};

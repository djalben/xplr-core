import apiClient from './axios';

export interface User {
  id: number;
  email: string;
  balance: string;
  status: string;
  api_key?: string;
}

export interface DepositRequest {
  amount: number;
}

export interface DepositResponse {
  new_balance: string;
}

// Получить данные текущего пользователя
export const getCurrentUser = async (): Promise<User> => {
  const response = await apiClient.get<User>('/user/me');
  return response.data;
};

// Пополнить баланс
export const deposit = async (data: DepositRequest): Promise<DepositResponse> => {
  const response = await apiClient.post<DepositResponse>('/user/deposit', data);
  return response.data;
};

// Получить отчет по транзакциям
export const getTransactionReport = async () => {
  const response = await apiClient.get('/user/report');
  return response.data;
};

// Создать новый API ключ
export const createAPIKey = async () => {
  const response = await apiClient.post('/user/api-key');
  return response.data;
};

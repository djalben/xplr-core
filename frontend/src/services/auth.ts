import apiClient from './axios';

export interface RegisterRequest {
  email: string;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthResponse {
  token?: string;
  mfaRequired?: boolean;
  mfaToken?: string;
  user?: {
    id: number | string;
    email: string;
    balance?: string;
    status?: string;
    is_admin?: boolean;
    role?: string;
    is_verified?: boolean;
  };
  message?: string;
  email?: string;
  email_verified?: boolean;
}

// Регистрация нового пользователя
export const register = async (data: RegisterRequest): Promise<AuthResponse> => {
  const response = await apiClient.post<AuthResponse>('/auth/register', data);

  if (response.data.token) {
    localStorage.setItem('token', response.data.token);
  }

  return response.data;
};

// Вход в систему
export const login = async (data: LoginRequest): Promise<AuthResponse> => {
  const response = await apiClient.post<AuthResponse>('/auth/login', data);

  if (response.data.token) {
    localStorage.setItem('token', response.data.token);
  }

  return response.data;
};

export interface LoginMFARequest {
  mfaToken: string;
  totpCode: string;
  rememberDevice?: boolean;
}

export const loginMFA = async (data: LoginMFARequest): Promise<AuthResponse> => {
  const response = await apiClient.post<AuthResponse>('/auth/login/mfa', data);

  if (response.data.token) {
    localStorage.setItem('token', response.data.token);
  }

  return response.data;
};

// Выход из системы
export const logout = (): void => {
  localStorage.removeItem('token');
};

// Запрос сброса пароля
export const requestPasswordReset = async (email: string): Promise<void> => {
  await apiClient.post('/auth/reset-password-request', { email });
};

// Установка нового пароля по токену
export const resetPassword = async (token: string, password: string): Promise<void> => {
  await apiClient.post('/auth/reset-password', { token, password });
};

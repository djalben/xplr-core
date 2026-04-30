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
  token: string;
  requires_2fa?: boolean;
  half_auth_token?: string;
  requires_otp?: boolean;
  email?: string;
  user: {
    id: number;
    email: string;
    balance: string;
    status: string;
    is_admin?: boolean;
    role?: string;
    is_verified?: boolean;
  };
}

export interface Verify2FARequest {
  half_auth_token: string;
  code: string;
  remember_device: boolean;
  fingerprint?: string;
}

// Регистрация нового пользователя (returns requires_otp, NO token)
export const register = async (data: RegisterRequest): Promise<AuthResponse> => {
  const response = await apiClient.post<AuthResponse>('/auth/register', data);
  // No token stored — user must verify OTP first
  return response.data;
};

// Verify 6-digit OTP code after registration
export const verifyOTP = async (email: string, code: string): Promise<{ status: string; message: string }> => {
  const response = await apiClient.post('/auth/verify-otp', { email, code });
  return response.data;
};

// Resend OTP code
export const resendOTP = async (email: string): Promise<{ status: string; message: string }> => {
  const response = await apiClient.post('/auth/resend-otp', { email });
  return response.data;
};

// Вход в систему (Stage 1 — may return requires_2fa)
export const login = async (data: LoginRequest): Promise<AuthResponse> => {
  const response = await apiClient.post<AuthResponse>('/auth/login', data);

  // If 2FA required, don't store token yet
  if (response.data.requires_2fa) {
    return response.data;
  }

  if (response.data.token) {
    localStorage.setItem('token', response.data.token);
  }

  return response.data;
};

// Вход Stage 2 — 2FA verification
export const verify2FA = async (data: Verify2FARequest): Promise<AuthResponse> => {
  const response = await apiClient.post<AuthResponse>('/auth/2fa/verify', data);

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

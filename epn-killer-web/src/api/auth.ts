import apiClient from './axios';
import AsyncStorage from '@react-native-async-storage/async-storage';

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
  user: {
    id: number;
    email: string;
    balance: string;
    status: string;
  };
}

// Регистрация нового пользователя
export const register = async (data: RegisterRequest): Promise<AuthResponse> => {
  const response = await apiClient.post<AuthResponse>('/auth/register', data);

  // Сохраняем токен в AsyncStorage
  if (response.data.token) {
    await AsyncStorage.setItem('jwt_token', response.data.token);
  }

  return response.data;
};

// Вход в систему
export const login = async (data: LoginRequest): Promise<AuthResponse> => {
  const response = await apiClient.post<AuthResponse>('/auth/login', data);

  // Сохраняем токен в AsyncStorage
  if (response.data.token) {
    await AsyncStorage.setItem('jwt_token', response.data.token);
  }

  return response.data;
};

// Выход из системы
export const logout = async (): Promise<void> => {
  await AsyncStorage.removeItem('jwt_token');
};

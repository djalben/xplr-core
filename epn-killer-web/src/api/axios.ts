import axios from 'axios';

// Supabase configuration (from .env via Vite)
export const SUPABASE_URL = import.meta.env.VITE_SUPABASE_URL || '';
export const SUPABASE_ANON_KEY = import.meta.env.VITE_SUPABASE_ANON_KEY || '';

// Backend API URL — same Vercel domain in production, localhost for dev
export const API_BASE_URL =
  typeof window !== 'undefined' && window.location?.hostname === 'localhost'
    ? 'http://localhost:8080/api/v1'
    : '/api/v1';

// Создаем экземпляр axios с базовой конфигурацией
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  withCredentials: false,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor - автоматически добавляет JWT токен к каждому запросу
apiClient.interceptors.request.use(
  (config) => {
    try {
      const token = localStorage.getItem('token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    } catch (error) {
      console.error('Error reading token from storage:', error);
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor - обработка ошибок
apiClient.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      // Токен невалиден или истек - очищаем хранилище
      localStorage.removeItem('token');
      // Здесь можно добавить редирект н�� страницу логина
    }
    return Promise.reject(error);
  }
);

export default apiClient;

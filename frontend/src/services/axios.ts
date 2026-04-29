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

// Silent refresh state — prevents infinite retry loops
let isRefreshing = false;
let refreshSubscribers: ((token: string) => void)[] = [];

function onRefreshed(token: string) {
  refreshSubscribers.forEach(cb => cb(token));
  refreshSubscribers = [];
}

function addRefreshSubscriber(cb: (token: string) => void) {
  refreshSubscribers.push(cb);
}

// Response interceptor - обработка ошибок + silent token refresh
apiClient.interceptors.response.use(
  (response) => {
    // Detect HTML responses (Vercel security checkpoint, etc.)
    const data = response.data;
    if (typeof data === 'string' && (data.includes('<!DOCTYPE') || data.includes('<html') || data.includes('<head'))) {
      return Promise.reject({
        response: { status: 503, data: 'Технические работы на стороне сервера. Пожалуйста, обновите страницу через минуту.' },
        message: 'Технические работы на стороне сервера. Пожалуйста, обновите страницу через минуту.',
      });
    }
    return response;
  },
  async (error) => {
    // Also catch HTML in error responses
    const data = error.response?.data;
    if (typeof data === 'string' && (data.includes('<!DOCTYPE') || data.includes('<html') || data.includes('<head'))) {
      error.response.data = 'Технические работы на стороне сервера. Пожалуйста, обновите страницу через минуту.';
    }

    const originalRequest = error.config;

    // Silent refresh on 401 — try to get a new token before giving up
    if (error.response?.status === 401 && !originalRequest._retry && localStorage.getItem('token')) {
      originalRequest._retry = true;

      if (!isRefreshing) {
        isRefreshing = true;
        try {
          const res = await apiClient.post('/auth/refresh-token');
          const newToken = res.data?.token;
          if (newToken) {
            localStorage.setItem('token', newToken);
            apiClient.defaults.headers.common['Authorization'] = `Bearer ${newToken}`;
            onRefreshed(newToken);
            isRefreshing = false;
            // Retry original request with new token
            originalRequest.headers['Authorization'] = `Bearer ${newToken}`;
            return apiClient(originalRequest);
          }
        } catch {
          // Refresh also failed — clear token
          localStorage.removeItem('token');
          isRefreshing = false;
          refreshSubscribers = [];
          return Promise.reject(error);
        }
      }

      // Another request hit 401 while refresh is in flight — queue it
      return new Promise((resolve) => {
        addRefreshSubscriber((token: string) => {
          originalRequest.headers['Authorization'] = `Bearer ${token}`;
          resolve(apiClient(originalRequest));
        });
      });
    }

    if (error.response?.status === 401) {
      localStorage.removeItem('token');
    }
    return Promise.reject(error);
  }
);

export default apiClient;

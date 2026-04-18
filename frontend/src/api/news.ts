import apiClient from './axios';

export interface NewsItem {
  id: number;
  title: string;
  content: string;
  image_url: string;
  created_at: string;
}

export interface NewsResponse {
  items: NewsItem[];
  total: number;
}

export const getNews = async (limit = 10, offset = 0): Promise<NewsResponse> => {
  const response = await apiClient.get('/user/news', { params: { limit, offset } });
  return response.data;
};

export const getNewsNotifications = async (): Promise<{ enabled: boolean }> => {
  const response = await apiClient.get('/user/news-notifications');
  return response.data;
};

export const updateNewsNotifications = async (enabled: boolean): Promise<void> => {
  await apiClient.patch('/user/news-notifications', { enabled });
};

export const getUnreadNewsCount = async (): Promise<{ count: number }> => {
  const response = await apiClient.get('/user/news/unread-count');
  return response.data;
};

export const markNewsAsRead = async (): Promise<void> => {
  await apiClient.post('/user/news/mark-as-read');
};

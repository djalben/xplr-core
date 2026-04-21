import apiClient from './axios';

export interface StoreCategory {
  id: string;
  slug: string;
  name: string;
  description: string;
  icon: string;
  imageUrl?: string;
  sortOrder: number;
  createdAt?: string;
}

export interface StoreProduct {
  id: string;
  categoryId: string;
  categorySlug: string;
  provider: string;
  externalId: string;
  name: string;
  description: string;
  country: string;
  countryCode: string;
  priceUsd: string;
  oldPrice: string;
  dataGb: string;
  validityDays: number;
  imageUrl: string;
  productType: string;
  inStock: boolean;
  meta: string;
  sortOrder: number;
}

export interface StoreOrder {
  id: string;
  userId: string;
  productId?: string;
  productName: string;
  priceUsd: string;
  status: string;
  activationKey: string;
  qrData: string;
  providerRef: string;
  meta: string;
  createdAt: string;
}

export interface CatalogResponse {
  categories: StoreCategory[];
  products: StoreProduct[];
}

export interface PurchaseResult {
  orderId: string;
  productName: string;
  priceUsd: string;
  activationKey: string;
  qrData: string;
  status: string;
  providerRef: string;
}

export const getStoreCatalog = async (params?: {
  category?: string;
  country?: string;
  search?: string;
}): Promise<CatalogResponse> => {
  const response = await apiClient.get('/user/store/catalog', { params });
  return response.data;
};

export const purchaseProduct = async (productId: string): Promise<PurchaseResult> => {
  const response = await apiClient.post('/user/store/purchase', { productId });
  return response.data;
};

export const getStoreOrders = async (limit = 20): Promise<{ orders: StoreOrder[] }> => {
  const response = await apiClient.get('/user/store/orders', { params: { limit } });
  return response.data;
};

// ── eSIM API ──

export interface ESIMDestination {
  countryCode: string;
  countryName: string;
  flagEmoji: string;
  planCount: number;
}

export interface ESIMPlan {
  planId: string;
  provider: string;
  name: string;
  country: string;
  countryCode: string;
  dataGb: string;
  validityDays: number;
  priceUsd: string;
  oldPrice: string;
  description: string;
  inStock: boolean;
}

export interface ESIMOrderResult {
  orderId: string;
  qrData: string;
  lpa: string;
  smdp: string;
  matchingId: string;
  iccid: string;
  providerRef: string;
}

export const getESIMDestinations = async (search?: string): Promise<{ destinations: ESIMDestination[] }> => {
  const response = await apiClient.get('/user/store/esim/destinations', { params: search ? { search } : {} });
  return response.data;
};

export const getESIMPlans = async (countryCode: string): Promise<{ plans: ESIMPlan[] }> => {
  const response = await apiClient.get('/user/store/esim/plans', { params: { country: countryCode } });
  return response.data;
};

// ── VPN Status API ──

export interface VPNKeyStatus {
  ref: string;
  status: string; // "active" | "traffic_exhausted" | "disabled"
  upload: number;
  download: number;
  used: number;
  total: number;
  remaining: number;
  exhausted: boolean;
  expire_ms: number;
  duration_days: number;
  used_percent: number;
}

export const getVPNKeyStatus = async (ref: string): Promise<VPNKeyStatus> => {
  const response = await apiClient.get('/user/store/vpn-status', { params: { ref } });
  return response.data;
};

export const orderESIM = async (plan: ESIMPlan): Promise<ESIMOrderResult> => {
  const response = await apiClient.post('/user/store/esim/order', { planId: plan.planId });
  return response.data;
};

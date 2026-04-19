import apiClient from './axios';

export interface StoreCategory {
  id: number;
  slug: string;
  name: string;
  description: string;
  icon: string;
  sort_order: number;
}

export interface StoreProduct {
  id: number;
  category_id: number;
  category_slug: string;
  provider: string;
  external_id: string;
  name: string;
  description: string;
  country: string;
  country_code: string;
  price_usd: string;
  old_price: string;
  data_gb: string;
  validity_days: number;
  image_url: string;
  product_type: string;
  in_stock: boolean;
  meta: Record<string, unknown>;
  sort_order: number;
}

export interface StoreOrder {
  id: number;
  user_id: number;
  product_id: number;
  product_name: string;
  price_usd: string;
  status: string;
  activation_key: string;
  qr_data: string;
  provider_ref: string;
  created_at: string;
}

export interface CatalogResponse {
  categories: StoreCategory[];
  products: StoreProduct[];
}

export interface PurchaseResult {
  order_id: number;
  product_name: string;
  price_usd: string;
  activation_key: string;
  qr_data: string;
  status: string;
}

export const getStoreCatalog = async (params?: {
  category?: string;
  country?: string;
  search?: string;
}): Promise<CatalogResponse> => {
  const response = await apiClient.get('/user/store/catalog', { params });
  return response.data;
};

export const purchaseProduct = async (productId: number): Promise<PurchaseResult> => {
  const response = await apiClient.post('/user/store/purchase', { product_id: productId });
  return response.data;
};

export const getStoreOrders = async (limit = 20): Promise<{ orders: StoreOrder[] }> => {
  const response = await apiClient.get('/user/store/orders', { params: { limit } });
  return response.data;
};

// ── eSIM API ──

export interface ESIMDestination {
  country_code: string;
  country_name: string;
  flag_emoji: string;
  plan_count: number;
}

export interface ESIMPlan {
  plan_id: string;
  provider: string;
  name: string;
  country: string;
  country_code: string;
  data_gb: string;
  validity_days: number;
  price_usd: number;
  old_price: number;
  description: string;
  in_stock: boolean;
}

export interface ESIMOrderResult {
  order_id: number;
  product_name: string;
  price_usd: string;
  qr_data: string;
  lpa: string;
  smdp: string;
  matching_id: string;
  iccid: string;
  provider_ref: string;
  status: string;
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
  const response = await apiClient.post('/user/store/esim/order', {
    plan_id: plan.plan_id,
    plan_name: plan.name,
    country: plan.country,
    country_code: plan.country_code,
    data_gb: plan.data_gb,
    validity_days: plan.validity_days,
    price_usd: plan.price_usd,
  });
  return response.data;
};

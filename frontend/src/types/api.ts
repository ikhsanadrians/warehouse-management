export interface ApiEnvelope<TData, TMeta = unknown> {
  success: boolean;
  message: string;
  data?: TData;
  meta?: TMeta;
  errors?: Array<{
    field?: string;
    reason: string;
  }>;
}

export interface HealthStatus {
  status: string;
}

export type HealthResponse = ApiEnvelope<HealthStatus>;

export interface Item {
  id: number;
  sku: string;
  name: string;
  category: string;
  unit: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
}

export interface ItemPayload {
  sku: string;
  name: string;
  category: string;
  unit: string;
}

export interface ListMeta {
  page: number;
  limit: number;
  total: number;
}

export type ItemResponse = ApiEnvelope<Item>;
export type ItemListResponse = ApiEnvelope<Item[], ListMeta>;
export type EmptyResponse = ApiEnvelope<never>;

import type { EmptyResponse, ItemListResponse, ItemPayload, ItemResponse } from "@/types/api";

const API_URL = import.meta.env.VITE_API_URL ?? "";

export class ApiError extends Error {
  status: number;
  errors?: Array<{ field?: string; reason: string }>;

  constructor(message: string, status: number, errors?: Array<{ field?: string; reason: string }>) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.errors = errors;
  }
}

async function request<TResponse>(path: string, options: RequestInit = {}) {
  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  });

  const payload = await response.json().catch(() => null);

  if (!response.ok) {
    throw new ApiError(
      payload?.message ?? "Request failed",
      response.status,
      payload?.errors,
    );
  }

  return payload as TResponse;
}

export function listItems(category?: string) {
  const params = new URLSearchParams({ page: "1", limit: "100" });

  if (category) {
    params.set("category", category);
  }

  return request<ItemListResponse>(`/api/v1/items?${params.toString()}`);
}

export function getItem(id: string) {
  return request<ItemResponse>(`/api/v1/items/${id}`);
}

export function createItem(input: ItemPayload) {
  return request<ItemResponse>("/api/v1/items", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateItem(id: string, input: ItemPayload) {
  return request<ItemResponse>(`/api/v1/items/${id}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function deleteItem(id: number) {
  return request<EmptyResponse>(`/api/v1/items/${id}`, {
    method: "DELETE",
  });
}

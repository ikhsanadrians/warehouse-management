import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { createItem, deleteItem, getItem, listItems, updateItem } from "@/services/api";
import type { ItemPayload } from "@/types/api";

export function useItems(category?: string) {
  return useQuery({
    queryKey: ["items", category ?? "all"],
    queryFn: () => listItems(category),
  });
}

export function useItem(id?: string) {
  return useQuery({
    queryKey: ["items", "detail", id],
    queryFn: () => getItem(id!),
    enabled: Boolean(id),
  });
}

export function useCreateItem() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createItem,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["items"] });
    },
  });
}

export function useUpdateItem(id: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: ItemPayload) => updateItem(id, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["items"] });
      void queryClient.invalidateQueries({ queryKey: ["items", "detail", id] });
    },
  });
}

export function useDeleteItem() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteItem,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["items"] });
    },
  });
}

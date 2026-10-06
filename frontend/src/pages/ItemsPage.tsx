import {
  ChevronLeft,
  ChevronRight,
  Edit2,
  Layers3,
  LayoutGrid,
  Loader2,
  PackageSearch,
  PackagePlus,
  Plus,
  Search,
  Table2,
  Trash2,
  X,
} from "lucide-react";
import { useMemo, useState, type ReactNode } from "react";
import { createPortal } from "react-dom";
import { Link, useMatch, useNavigate, useParams } from "react-router-dom";

import { ItemForm } from "@/components/ItemForm";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { useToast } from "@/components/ToastProvider";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";
import { useCreateItem, useDeleteItem, useItem, useItems, useUpdateItem } from "@/hooks/useItems";
import { ApiError } from "@/services/api";
import type { Item, ItemPayload } from "@/types/api";

const pageSize = 10;
type LayoutMode = "table" | "card";

export function ItemsPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const createMatch = useMatch("/items/new");
  const editMatch = useMatch("/items/:id/edit");
  const modalMode: "create" | "edit" | null = createMatch ? "create" : editMatch ? "edit" : null;
  const [category, setCategory] = useState("");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [layoutMode, setLayoutMode] = useState<LayoutMode>("table");
  const [deleteTarget, setDeleteTarget] = useState<Item | null>(null);
  const debouncedSearch = useDebouncedValue(search, 300);
  const allItemsQuery = useItems();
  const itemsQuery = useItems(category || undefined);
  const itemQuery = useItem(modalMode === "edit" ? id : undefined);
  const createItem = useCreateItem();
  const updateItem = useUpdateItem(id ?? "");
  const deleteItem = useDeleteItem();
  const { showToast } = useToast();

  const items = itemsQuery.data?.data ?? [];
  const categorySource = allItemsQuery.data?.data ?? items;
  const categories = useMemo(
    () => Array.from(new Set(categorySource.map((item) => item.category))).sort(),
    [categorySource],
  );
  const categoryOptions = useMemo(
    () => [
      { label: "All categories", value: "" },
      ...categories.map((itemCategory) => ({ label: itemCategory, value: itemCategory })),
    ],
    [categories],
  );

  const filteredItems = useMemo(() => {
    const keyword = debouncedSearch.trim().toLowerCase();

    if (!keyword) {
      return items;
    }

    return items.filter((item) =>
      [item.sku, item.name].some((value) => value.toLowerCase().includes(keyword)),
    );
  }, [debouncedSearch, items]);

  const totalPages = Math.max(1, Math.ceil(filteredItems.length / pageSize));
  const safePage = Math.min(page, totalPages);
  const visibleItems = filteredItems.slice((safePage - 1) * pageSize, safePage * pageSize);
  const start = filteredItems.length === 0 ? 0 : (safePage - 1) * pageSize + 1;
  const end = Math.min(safePage * pageSize, filteredItems.length);

  function handleCategoryChange(value: string) {
    setCategory(value);
    setPage(1);
  }

  function handleSearchChange(value: string) {
    setSearch(value);
    setPage(1);
  }

  async function confirmDelete() {
    if (!deleteTarget) {
      return;
    }

    try {
      await deleteItem.mutateAsync(deleteTarget.id);
      showToast("Item deleted successfully");
      setDeleteTarget(null);
    } catch (error) {
      showToast(error instanceof ApiError ? error.message : "Failed to delete item", "error");
    }
  }

  function closeModal() {
    navigate("/items");
  }

  async function handleFormSubmit(payload: ItemPayload) {
    try {
      if (modalMode === "edit" && id) {
        await updateItem.mutateAsync(payload);
        showToast("Item updated successfully");
      } else {
        await createItem.mutateAsync(payload);
        showToast("Item created successfully");
      }

      closeModal();
    } catch (error) {
      showToast(error instanceof ApiError ? error.message : "Failed to save item", "error");
    }
  }

  return (
    <section className="space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Warehouse</p>
          <h1 className="mt-1 text-2xl font-semibold tracking-normal">Items</h1>
        </div>
        <div className="flex items-center gap-2">
          <Button asChild>
            <Link to="/items/new">
              <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
              Add Item
            </Link>
          </Button>
        </div>
      </div>

      <div className="grid gap-2 sm:grid-cols-2">
        <Metric
          label="Items"
          value={categorySource.length.toString()}
          icon={<PackageSearch className="h-4 w-4" />}
        />
        <Metric label="Categories" value={categories.length.toString()} icon={<Layers3 className="h-4 w-4" />} />
      </div>

      <div className="rounded-md border bg-card p-3 text-card-foreground">
        <div className="grid gap-3 md:grid-cols-[minmax(280px,1fr)_260px_auto]">
          <label className="relative">
            <span className="sr-only">Search SKU or name</span>
            <Search
              className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
              aria-hidden="true"
            />
            <Input
              className="pl-9"
              value={search}
              onChange={(event) => handleSearchChange(event.target.value)}
              placeholder="Search SKU / Name..."
            />
          </label>

          <Select
            value={category}
            options={categoryOptions}
            onValueChange={handleCategoryChange}
            placeholder="All categories"
          />

          <ViewModeToggle value={layoutMode} onChange={setLayoutMode} />
        </div>
      </div>

      <div className="overflow-hidden rounded-md border bg-card text-card-foreground">
        {itemsQuery.isLoading ? (
          <LoadingTable />
        ) : itemsQuery.isError ? (
          <StateMessage
            title="Could not load items"
            description={
              itemsQuery.error instanceof Error
                ? `${itemsQuery.error.message}. Make sure the backend is running.`
                : "Make sure the backend is running."
            }
          />
        ) : filteredItems.length === 0 ? (
          <StateMessage
            title="No items found"
            description="Create an item or adjust the active filters."
          />
        ) : (
          <>
            {layoutMode === "table" ? (
              <ItemsTable
                items={visibleItems}
                deleting={deleteItem.isPending}
                onDelete={setDeleteTarget}
              />
            ) : (
              <ItemsCardGrid
                items={visibleItems}
                deleting={deleteItem.isPending}
                onDelete={setDeleteTarget}
              />
            )}

            <div className="flex flex-col gap-3 border-t px-4 py-3 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
              <span>
                Showing <strong className="font-medium text-foreground">{start}-{end}</strong> of{" "}
                <strong className="font-medium text-foreground">{filteredItems.length}</strong>
              </span>
              <div className="flex items-center justify-end gap-2">
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  disabled={safePage <= 1}
                  aria-label="Previous page"
                  onClick={() => setPage((current) => Math.max(1, current - 1))}
                >
                  <ChevronLeft className="h-4 w-4" aria-hidden="true" />
                </Button>
                <span className="min-w-16 text-center font-medium text-foreground">
                  {safePage} / {totalPages}
                </span>
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  disabled={safePage >= totalPages}
                  aria-label="Next page"
                  onClick={() => setPage((current) => Math.min(totalPages, current + 1))}
                >
                  <ChevronRight className="h-4 w-4" aria-hidden="true" />
                </Button>
              </div>
            </div>
          </>
        )}
      </div>

      {modalMode ? (
        <ItemModal
          mode={modalMode}
          loading={modalMode === "edit" && itemQuery.isLoading}
          error={modalMode === "edit" && itemQuery.isError ? itemQuery.error : null}
          saving={createItem.isPending || updateItem.isPending}
          item={itemQuery.data?.data}
          onClose={closeModal}
          onSubmit={(payload) => void handleFormSubmit(payload)}
        />
      ) : null}

      {deleteTarget ? (
        <DeleteItemDialog
          item={deleteTarget}
          loading={deleteItem.isPending}
          onCancel={() => setDeleteTarget(null)}
          onConfirm={() => void confirmDelete()}
        />
      ) : null}
    </section>
  );
}

function ViewModeToggle({ value, onChange }: { value: LayoutMode; onChange: (value: LayoutMode) => void }) {
  return (
    <div
      className="grid h-10 w-24 grid-cols-2 rounded-md border bg-background p-1"
      role="group"
      aria-label="Layout mode"
    >
      <button
        type="button"
        className={
          value === "table"
            ? "inline-flex items-center justify-center rounded-sm bg-primary text-primary-foreground"
            : "inline-flex items-center justify-center rounded-sm text-muted-foreground hover:bg-secondary"
        }
        aria-label="Table layout"
        aria-pressed={value === "table"}
        onClick={() => onChange("table")}
      >
        <Table2 className="h-4 w-4" aria-hidden="true" />
      </button>
      <button
        type="button"
        className={
          value === "card"
            ? "inline-flex items-center justify-center rounded-sm bg-primary text-primary-foreground"
            : "inline-flex items-center justify-center rounded-sm text-muted-foreground hover:bg-secondary"
        }
        aria-label="Card layout"
        aria-pressed={value === "card"}
        onClick={() => onChange("card")}
      >
        <LayoutGrid className="h-4 w-4" aria-hidden="true" />
      </button>
    </div>
  );
}

function ItemsTable({
  items,
  deleting,
  onDelete,
}: {
  items: Item[];
  deleting: boolean;
  onDelete: (item: Item) => void;
}) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[720px] text-left text-sm">
        <thead className="border-b bg-secondary text-xs uppercase text-muted-foreground">
          <tr>
            <th className="px-5 py-3 font-semibold">SKU</th>
            <th className="px-5 py-3 font-semibold">Name</th>
            <th className="px-5 py-3 font-semibold">Category</th>
            <th className="px-5 py-3 font-semibold">Unit</th>
            <th className="px-5 py-3 text-right font-semibold">Actions</th>
          </tr>
        </thead>
        <tbody>
          {items.map((item) => (
            <tr key={item.id} className="border-b transition-colors last:border-b-0 hover:bg-muted/30">
              <td className="px-5 py-3.5">
                <DataBadge tone="neutral">{item.sku}</DataBadge>
              </td>
              <td className="px-5 py-3.5">
                <div className="font-medium text-foreground">{item.name}</div>
                <div className="mt-0.5 text-xs text-muted-foreground">ID #{item.id}</div>
              </td>
              <td className="px-5 py-3.5">
                <DataBadge tone="primary">{item.category}</DataBadge>
              </td>
              <td className="px-5 py-3.5">
                <DataBadge tone="neutral">{item.unit}</DataBadge>
              </td>
              <td className="px-5 py-3.5">
                <ItemActions item={item} deleting={deleting} onDelete={onDelete} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function ItemsCardGrid({
  items,
  deleting,
  onDelete,
}: {
  items: Item[];
  deleting: boolean;
  onDelete: (item: Item) => void;
}) {
  return (
    <div className="grid gap-3 p-3 sm:grid-cols-2 lg:grid-cols-3">
      {items.map((item) => (
        <article key={item.id} className="rounded-md border bg-background p-4 text-sm">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <DataBadge tone="neutral">{item.sku}</DataBadge>
              <h2 className="mt-3 truncate text-base font-semibold text-foreground">{item.name}</h2>
              <p className="mt-0.5 text-xs text-muted-foreground">ID #{item.id}</p>
            </div>
            <ItemActions item={item} deleting={deleting} onDelete={onDelete} />
          </div>

          <dl className="mt-5 grid grid-cols-2 gap-3">
            <div>
              <dt className="mb-1 text-xs font-medium text-muted-foreground">Category</dt>
              <dd>
                <DataBadge tone="primary">{item.category}</DataBadge>
              </dd>
            </div>
            <div>
              <dt className="mb-1 text-xs font-medium text-muted-foreground">Unit</dt>
              <dd>
                <DataBadge tone="neutral">{item.unit}</DataBadge>
              </dd>
            </div>
          </dl>
        </article>
      ))}
    </div>
  );
}

function ItemActions({
  item,
  deleting,
  onDelete,
}: {
  item: Item;
  deleting: boolean;
  onDelete: (item: Item) => void;
}) {
  return (
    <div className="flex justify-end gap-2">
      <Button asChild variant="default" size="icon" className="h-8 w-8" aria-label={`Edit ${item.sku}`}>
        <Link to={`/items/${item.id}/edit`}>
          <Edit2 className="h-4 w-4" aria-hidden="true" />
        </Link>
      </Button>
      <Button
        type="button"
        variant="destructive"
        size="icon"
        className="h-8 w-8"
        aria-label={`Delete ${item.sku}`}
        disabled={deleting}
        onClick={() => onDelete(item)}
      >
        {deleting ? (
          <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
        ) : (
          <Trash2 className="h-4 w-4" aria-hidden="true" />
        )}
      </Button>
    </div>
  );
}

function DataBadge({ children, tone = "neutral" }: { children: ReactNode; tone?: "neutral" | "primary" }) {
  return (
    <span
      className={
        tone === "primary"
          ? "inline-flex min-w-14 justify-center rounded-md border border-primary/20 bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary"
          : "inline-flex min-w-14 justify-center rounded-md border border-border bg-secondary px-2 py-0.5 text-xs font-medium text-secondary-foreground"
      }
    >
      {children}
    </span>
  );
}

function DeleteItemDialog({
  item,
  loading,
  onCancel,
  onConfirm,
}: {
  item: Item;
  loading: boolean;
  onCancel: () => void;
  onConfirm: () => void;
}) {
  return createPortal(
    <div
      className="fixed inset-0 z-[60] grid place-items-center bg-slate-950/55 px-4 py-6 backdrop-blur-sm"
      role="presentation"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget && !loading) {
          onCancel();
        }
      }}
    >
      <div
        className="w-full max-w-md overflow-hidden rounded-md border bg-background shadow-2xl"
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="delete-dialog-title"
        aria-describedby="delete-dialog-description"
      >
        <div className="border-b bg-card px-5 py-4 text-card-foreground">
          <div className="flex items-start gap-3">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-destructive/10 text-destructive">
              <Trash2 className="h-5 w-5" aria-hidden="true" />
            </div>
            <div>
              <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Confirm delete</p>
              <h2 id="delete-dialog-title" className="mt-1 text-lg font-semibold">
                Delete item?
              </h2>
            </div>
          </div>
        </div>

        <div className="space-y-4 px-5 py-4">
          <div className="rounded-md border bg-muted/40 px-3 py-2">
            <p className="font-medium text-foreground">{item.name}</p>
            <p className="mt-0.5 font-mono text-xs text-muted-foreground">{item.sku}</p>
          </div>
          <p id="delete-dialog-description" className="text-sm text-muted-foreground">
            This item will be removed from the active item list. Stock history and existing records will not be deleted.
          </p>
        </div>

        <div className="flex flex-col-reverse gap-2 border-t bg-muted/30 px-5 py-4 sm:flex-row sm:justify-end">
          <Button type="button" variant="outline" disabled={loading} onClick={onCancel}>
            Cancel
          </Button>
          <Button type="button" variant="destructive" disabled={loading} onClick={onConfirm}>
            {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" /> : null}
            Delete
          </Button>
        </div>
      </div>
    </div>,
    document.body,
  );
}

function ItemModal({
  mode,
  loading,
  error,
  saving,
  item,
  onClose,
  onSubmit,
}: {
  mode: "create" | "edit";
  loading: boolean;
  error: unknown;
  saving: boolean;
  item?: Item;
  onClose: () => void;
  onSubmit: (payload: ItemPayload) => void;
}) {
  return createPortal(
    <div
      className="fixed inset-0 z-50 grid place-items-center bg-slate-950/55 px-4 py-6 backdrop-blur-sm"
      role="presentation"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) {
          onClose();
        }
      }}
    >
      <div
        className="max-h-[calc(100vh-3rem)] w-full max-w-4xl overflow-auto rounded-md border bg-background shadow-2xl"
        role="dialog"
        aria-modal="true"
        aria-labelledby="item-modal-title"
      >
        <div className="flex items-start justify-between gap-4 border-b bg-card px-5 py-4 text-card-foreground">
          <div className="flex items-start gap-3">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary">
              <PackagePlus className="h-5 w-5" aria-hidden="true" />
            </div>
            <div>
              <h2 id="item-modal-title" className="mt-1 text-xl font-semibold">
                {mode === "edit" ? "Edit Item" : "Create Item"}
              </h2>
            </div>
          </div>
          <Button type="button" variant="ghost" size="icon" aria-label="Close item form" onClick={onClose}>
            <X className="h-4 w-4" aria-hidden="true" />
          </Button>
        </div>

        <div className="p-5">
          {loading ? (
            <div className="flex min-h-72 items-center justify-center rounded-md border bg-card">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" aria-hidden="true" />
            </div>
          ) : error ? (
            <div className="rounded-md border bg-card px-4 py-10 text-center">
              <p className="font-semibold">Could not load item</p>
              <p className="mt-1 text-sm text-muted-foreground">
                {error instanceof Error ? error.message : "Please try again."}
              </p>
            </div>
          ) : (
            <ItemForm initialItem={item} loading={saving} onSubmit={onSubmit} />
          )}
        </div>
      </div>
    </div>,
    document.body,
  );
}

function LoadingTable() {
  return (
    <div className="space-y-3 p-5">
      {Array.from({ length: 6 }).map((_, index) => (
        <div key={index} className="h-12 animate-pulse rounded-md bg-secondary" />
      ))}
    </div>
  );
}

function StateMessage({ title, description }: { title: string; description: string }) {
  return (
    <div className="grid min-h-60 place-items-center px-4 py-10 text-center">
      <div className="max-w-sm">
        <div className="mx-auto mb-3 flex h-11 w-11 items-center justify-center rounded-md bg-secondary text-secondary-foreground">
          <PackageSearch className="h-5 w-5" aria-hidden="true" />
        </div>
        <p className="font-semibold">{title}</p>
        <p className="mt-1 text-sm text-muted-foreground">{description}</p>
      </div>
    </div>
  );
}

function Metric({
  label,
  value,
  helper,
  icon,
}: {
  label: string;
  value: string;
  helper?: string;
  icon: ReactNode;
}) {
  return (
    <div className="rounded-md border bg-card px-4 py-3 text-card-foreground">
      <div className="flex items-center justify-between gap-3">
        <p className="text-sm font-medium text-muted-foreground">{label}</p>
        <span className="flex h-7 w-7 items-center justify-center rounded-md bg-secondary text-secondary-foreground">
          {icon}
        </span>
      </div>
      <div className="mt-2 flex items-end gap-2">
        <p className="text-xl font-semibold">{value}</p>
        {helper ? <p className="pb-1 text-xs text-muted-foreground">{helper}</p> : null}
      </div>
    </div>
  );
}

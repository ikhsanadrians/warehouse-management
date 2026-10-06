import { Loader2, Save } from "lucide-react";
import { useEffect, useState, type FormEvent, type ReactNode } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { Item, ItemPayload } from "@/types/api";

interface ItemFormProps {
  initialItem?: Item;
  loading: boolean;
  onSubmit: (payload: ItemPayload) => void;
}

interface FormState {
  sku: string;
  name: string;
  category: string;
  unit: string;
}

const emptyForm: FormState = {
  sku: "",
  name: "",
  category: "",
  unit: "",
};

export function ItemForm({ initialItem, loading, onSubmit }: ItemFormProps) {
  const [form, setForm] = useState<FormState>(emptyForm);
  const [errors, setErrors] = useState<Partial<Record<keyof FormState, string>>>({});

  useEffect(() => {
    if (initialItem) {
      setForm({
        sku: initialItem.sku,
        name: initialItem.name,
        category: initialItem.category,
        unit: initialItem.unit,
      });
    }
  }, [initialItem]);

  function updateField(field: keyof FormState, value: string) {
    setForm((current) => ({ ...current, [field]: value }));
    setErrors((current) => ({ ...current, [field]: undefined }));
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const labels: Record<keyof FormState, string> = {
      sku: "SKU",
      name: "Name",
      category: "Category",
      unit: "Unit",
    };
    const nextErrors: Partial<Record<keyof FormState, string>> = {};

    (Object.keys(form) as Array<keyof FormState>).forEach((field) => {
      if (!form[field].trim()) {
        nextErrors[field] = `${labels[field]} is required`;
      }
    });

    if (Object.keys(nextErrors).length > 0) {
      setErrors(nextErrors);
      return;
    }

    onSubmit({
      sku: form.sku.trim(),
      name: form.name.trim(),
      category: form.category.trim(),
      unit: form.unit.trim(),
    });
  }

  return (
    <form className="max-w-4xl overflow-hidden rounded-md border bg-card text-card-foreground" onSubmit={handleSubmit}>
      <div className="grid gap-5 p-5 sm:grid-cols-2">
        <Field label="SKU" error={errors.sku}>
          <Input
            value={form.sku}
            onChange={(event) => updateField("sku", event.target.value)}
            disabled={loading}
            autoComplete="off"
            placeholder="SKU001"
          />
        </Field>
        <Field label="Unit" error={errors.unit}>
          <Input
            value={form.unit}
            onChange={(event) => updateField("unit", event.target.value)}
            disabled={loading}
            autoComplete="off"
            placeholder="pcs"
          />
        </Field>
        <Field label="Name" error={errors.name}>
          <Input
            value={form.name}
            onChange={(event) => updateField("name", event.target.value)}
            disabled={loading}
            autoComplete="off"
            placeholder="Keyboard"
          />
        </Field>
        <Field label="Category" error={errors.category}>
          <Input
            value={form.category}
            onChange={(event) => updateField("category", event.target.value)}
            disabled={loading}
            autoComplete="off"
            placeholder="Electronics"
          />
        </Field>
      </div>

      <div className="flex items-center justify-end border-t bg-muted/30 px-5 py-4">
        <Button type="submit" disabled={loading}>
          {loading ? (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
          ) : (
            <Save className="mr-2 h-4 w-4" aria-hidden="true" />
          )}
          Save item
        </Button>
      </div>
    </form>
  );
}

function Field({ label, error, children }: { label: string; error?: string; children: ReactNode }) {
  return (
    <label className="grid gap-2 text-sm font-medium">
      <span>{label}</span>
      {children}
      {error ? <span className="text-xs font-medium text-destructive">{error}</span> : null}
    </label>
  );
}

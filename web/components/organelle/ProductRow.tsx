import type { Product } from "@/api/models";
import { Button } from "@/components/atom";

type ProductRowProps = {
  product: Product;
  quantity: number;
  onChange: (delta: number) => void;
};

export function ProductRow({ product, quantity, onChange }: ProductRowProps) {
  return (
    <div className="flex items-center gap-4 rounded-2xl bg-white p-4 shadow-sm">
      <div className="flex-1">
        <p className="font-medium text-zinc-900">{product.name}</p>
        <p className="text-zinc-500">
          ${parseFloat(product.price ?? "0").toFixed(2)}
        </p>
      </div>
      <div className="flex items-center gap-3">
        <Button
          variant="circle"
          onClick={() => onChange(-1)}
          aria-label={`Decrease ${product.name}`}
        >
          −
        </Button>
        <span className="w-6 text-center font-medium text-zinc-900">
          {quantity}
        </span>
        <Button
          variant="circle"
          onClick={() => onChange(1)}
          aria-label={`Increase ${product.name}`}
        >
          +
        </Button>
      </div>
    </div>
  );
}

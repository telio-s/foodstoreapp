"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useListProducts } from "@/api/query/foodStoreAPI";
import { loadCart, saveCart } from "@/lib/cart";
import { Button, TextInput } from "@/components/atom";
import { ProductRow } from "@/components/organelle";

export function ProductListCell() {
  const router = useRouter();
  const { data, isLoading } = useListProducts();

  const [quantities, setQuantities] = useState<Record<string, number>>({});
  const [memberCardNumber, setMemberCardNumber] = useState("");

  const products =
    data?.status === 200 ? (data.data.data?.items ?? []) : [];
  const errorMessage = data && data.status !== 200 ? data.data.msg : null;

  useEffect(() => {
    if (products.length === 0) return;
    const cart = loadCart();
    if (!cart) return;
    // eslint-disable-next-line react-hooks/set-state-in-effect -- one-time sync from sessionStorage once products are available
    setMemberCardNumber(cart.memberCardNumber);
    setQuantities((prev) => {
      const next = { ...prev };
      for (const item of cart.items) {
        next[item.productId] = item.quantity;
      }
      return next;
    });
  }, [products.length]);

  const updateQuantity = (productId: string, delta: number) => {
    setQuantities((prev) => {
      const current = prev[productId] ?? 0;
      const next = Math.max(0, current + delta);
      return { ...prev, [productId]: next };
    });
  };

  const hasSelection = products.some(
    (p) => p.id && (quantities[p.id] ?? 0) > 0,
  );

  const handleCalculate = () => {
    const items = products
      .filter((p) => p.id && (quantities[p.id] ?? 0) > 0)
      .map((p) => ({
        productId: p.id!,
        name: p.name ?? "",
        price: parseFloat(p.price ?? "0"),
        quantity: quantities[p.id!] ?? 0,
      }));

    saveCart({ items, memberCardNumber });
    router.push("/calculator");
  };

  return (
    <div className="flex flex-col flex-1 items-center bg-zinc-50">
      <main className="flex flex-1 w-full max-w-md flex-col gap-4 px-4 py-6">
        {isLoading && (
          <p className="text-center text-zinc-500">Loading products...</p>
        )}
        {errorMessage && (
          <p className="text-center text-red-500">{errorMessage}</p>
        )}

        <div className="flex flex-col gap-4">
          {products.map((product) => {
            const id = product.id ?? "";
            return (
              <ProductRow
                key={id}
                product={product}
                quantity={quantities[id] ?? 0}
                onChange={(delta) => updateQuantity(id, delta)}
              />
            );
          })}
        </div>

        <TextInput
          id="card-number"
          label="Card Number"
          value={memberCardNumber}
          onChange={(e) => setMemberCardNumber(e.target.value)}
          placeholder="Optional"
        />

        <div className="flex-1" />

        <Button onClick={handleCalculate} disabled={!hasSelection}>
          Calculate
        </Button>
      </main>
    </div>
  );
}

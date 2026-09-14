"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { useCreateOrder } from "@/api/query/foodStoreAPI";
import { loadCart, type Cart } from "@/lib/cart";
import { Button } from "@/components/atom";
import { SummaryCard } from "@/components/organelle";

function round2(value: number) {
  return Math.round(value * 100) / 100;
}

export function OrderSummaryCell() {
  const router = useRouter();
  const [cart, setCart] = useState<Cart | null>(null);
  const requestedRef = useRef(false);
  const { mutate, data, isPending } = useCreateOrder();

  useEffect(() => {
    const stored = loadCart();
    if (!stored || stored.items.length === 0) {
      router.replace("/");
      return;
    }
    // eslint-disable-next-line react-hooks/set-state-in-effect -- one-time sync from sessionStorage on mount
    setCart(stored);

    if (requestedRef.current) return;
    requestedRef.current = true;
    mutate({
      data: {
        member_card_number: stored.memberCardNumber || undefined,
        items: stored.items.map((item) => ({
          product_id: item.productId,
          quantity: item.quantity,
        })),
      },
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (!cart) return null;

  const subtotal = round2(
    cart.items.reduce((sum, item) => sum + item.price * item.quantity, 0),
  );

  const errorMessage = data && data.status !== 201 ? data.data.msg : null;
  const result = data?.status === 201 ? data.data : null;

  const discountAmount = result ? parseFloat(result.discount_amount ?? "0") : 0;
  const totalPrice = result ? parseFloat(result.total_price ?? "0") : 0;
  const memberDiscount = cart.memberCardNumber ? round2(subtotal * 0.1) : 0;
  const pairDiscount = round2(discountAmount - memberDiscount);

  return (
    <div className="flex flex-col flex-1 items-center bg-zinc-50">
      <main className="flex flex-1 w-full max-w-md flex-col gap-4 px-4 py-6">
        <h1 className="text-center text-2xl font-bold text-zinc-900">
          Calculate Total
        </h1>

        {isPending && (
          <p className="text-center text-zinc-500">Calculating...</p>
        )}
        {errorMessage && (
          <p className="text-center text-red-500">{errorMessage}</p>
        )}

        <SummaryCard title="Total (before discount)">
          <p className="text-lg font-bold text-zinc-900">
            ฿{subtotal.toFixed(2)}
          </p>
        </SummaryCard>

        <SummaryCard title="Order Items">
          <div className="flex flex-col gap-1">
            {cart.items.map((item) => (
              <div
                key={item.productId}
                className="flex justify-between text-zinc-600"
              >
                <span>
                  {item.name} x{item.quantity}
                </span>
                <span>฿{(item.price * item.quantity).toFixed(2)}</span>
              </div>
            ))}
          </div>
        </SummaryCard>

        {result && (
          <SummaryCard title="Discounts">
            <div className="flex flex-col gap-1 text-zinc-600">
              <div className="flex justify-between">
                <span>Pair discount</span>
                <span>-฿{pairDiscount.toFixed(2)}</span>
              </div>
              {cart.memberCardNumber && (
                <div className="flex justify-between">
                  <span>Discount from Member Card</span>
                  <span>-฿{memberDiscount.toFixed(2)}</span>
                </div>
              )}
            </div>
          </SummaryCard>
        )}

        {result && (
          <SummaryCard title="Final Total">
            <p className="text-lg font-bold text-zinc-900">
              ฿{totalPrice.toFixed(2)}
            </p>
          </SummaryCard>
        )}

        <div className="flex-1" />

        <Button onClick={() => router.push("/")}>Recalculate</Button>
      </main>
    </div>
  );
}

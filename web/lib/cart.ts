export type CartItem = {
  productId: string;
  name: string;
  price: number;
  quantity: number;
};

export type Cart = {
  items: CartItem[];
  memberCardNumber: string;
};

const STORAGE_KEY = "food-store-cart";

export function saveCart(cart: Cart) {
  sessionStorage.setItem(STORAGE_KEY, JSON.stringify(cart));
}

export function loadCart(): Cart | null {
  const raw = sessionStorage.getItem(STORAGE_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as Cart;
  } catch {
    return null;
  }
}

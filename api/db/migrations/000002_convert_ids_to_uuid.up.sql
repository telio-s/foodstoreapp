DROP TABLE order_items;
DROP TABLE orders;
DROP TABLE products;

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    price TEXT NOT NULL
);

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    member_card_number TEXT,
    total_price TEXT NOT NULL,
    discount_amount TEXT NOT NULL DEFAULT '0',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders (id),
    product_id UUID NOT NULL REFERENCES products (id),
    quantity INTEGER NOT NULL,
    unit_price TEXT NOT NULL
);

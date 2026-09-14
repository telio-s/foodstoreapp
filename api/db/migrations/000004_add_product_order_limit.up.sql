ALTER TABLE products
    ADD COLUMN is_limited BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN last_order_at TIMESTAMPTZ;

-- Red sets are rare: only one order per hour is allowed for them.
UPDATE products SET is_limited = true WHERE name = 'Red';

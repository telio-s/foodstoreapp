-- name: ListProducts :many
SELECT id, name, price, is_limited, last_order_at
FROM products
ORDER BY id;

-- name: GetProductByID :one
SELECT id, name, price, is_limited, last_order_at
FROM products
WHERE id = $1;

-- name: CreateProduct :one
INSERT INTO products (name, price)
VALUES ($1, $2)
RETURNING id, name, price, is_limited, last_order_at;

-- name: ClaimLimitedProduct :one
-- Atomically claims a limited product for an order: last_order_at is only
-- advanced to now() if the product is actually limited and its previous
-- last_order_at (if any) is already outside the 1-hour cooldown. Postgres
-- locks the row for the duration of this statement, so a concurrent claim
-- for the same product waits for this transaction to commit or roll back,
-- then re-evaluates the WHERE clause against the now-committed value --
-- exactly one concurrent claimant can ever win within a given hour.
-- Zero rows returned means the product is still in its cooldown window.
UPDATE products
SET last_order_at = now()
WHERE id = $1
  AND is_limited = true
  AND (
    last_order_at IS NULL
    OR last_order_at <= now() - INTERVAL '1 hour'
  )
RETURNING id, name, price, is_limited, last_order_at;

-- name: CreateOrder :one
INSERT INTO orders (member_card_number, total_price, discount_amount, created_at)
VALUES ($1, $2, $3, now())
RETURNING id, member_card_number, total_price, discount_amount, created_at;

-- name: GetOrderByID :one
SELECT id, member_card_number, total_price, discount_amount, created_at
FROM orders
WHERE id = $1;

-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, product_id, quantity, unit_price)
VALUES ($1, $2, $3, $4)
RETURNING id, order_id, product_id, quantity, unit_price;

-- name: ListOrderItemsByOrderID :many
SELECT id, order_id, product_id, quantity, unit_price
FROM order_items
WHERE order_id = $1;

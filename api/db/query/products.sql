-- name: ListProducts :many
SELECT id, name, price
FROM products
ORDER BY id;

-- name: GetProductByID :one
SELECT id, name, price
FROM products
WHERE id = $1;

-- name: CreateProduct :one
INSERT INTO products (name, price)
VALUES ($1, $2)
RETURNING id, name, price;

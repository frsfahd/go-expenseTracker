-- name: AddUser :one
INSERT INTO users (email, password, username)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE email = $1;


-- name: AddExpense :one
INSERT INTO expenses (name, description, category, amount, user_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListExpenses :many
SELECT * FROM expenses
WHERE user_id = $1;


-- name: FilterExpense :many
SELECT * FROM expenses
WHERE user_id = $1 AND created_at >= NOW() - CAST($2 AS INTERVAL);

-- name: FilterExpenseCustom :many
SELECT id, name, description, category, amount, created_at, updated_at, user_id 
FROM expenses
WHERE user_id = $1 AND created_at BETWEEN $2 AND $3;

-- name: UpdateExpense :one
UPDATE expenses
SET name = $1, description = $2, category = $3, amount = $4
WHERE id = $5 AND user_id = $6
RETURNING *;

-- name: DeleteExpense :one
DELETE FROM expenses
WHERE id = $1 AND user_id = $2
RETURNING *;


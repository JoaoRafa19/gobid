-- name: CreateUser :one
INSERT INTO users (user_name, email, password_hash, bio)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: GetUserById :one
SELECT
    users.id,
    users.user_name,
    users.password_hash,
    users.email,
    users.bio,
    users.created_at,
    users.uptaded_at
FROM users WHERE users.id = $1;

-- name: GetUserByEmail :one
SELECT
    users.id,
    users.user_name,
    users.password_hash,
    users.email,
    users.bio,
    users.created_at,
    users.uptaded_at
FROM users WHERE users.email= $1;
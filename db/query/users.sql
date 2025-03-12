-- name: CreateUser :one
INSERT INTO users (
    email,
    hashed_password
) VALUES (
    $1, $2
) RETURNING user_id;

-- name: FindUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: FindUserById :one
SELECT * FROM users
WHERE user_id = $1 LIMIT 1;
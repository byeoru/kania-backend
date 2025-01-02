-- name: CreateUser :exec
INSERT INTO users (
    email,
    hashed_password,
    nickname
) VALUES (
    $1, $2, $3
);

-- name: FindUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: FindUserById :one
SELECT * FROM users
WHERE user_id = $1 LIMIT 1;
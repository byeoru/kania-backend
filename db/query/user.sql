-- name: CreateUser :exec
INSERT INTO users (
    email,
    hashed_password,
    nickname
) VALUES (
    $1, $2, $3
);

-- name: FindUser :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;
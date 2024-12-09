-- name: CreateUser :exec
INSERT INTO users (
    email,
    hashed_password,
    nickname
) VALUES (
    $1, $2, $3
);
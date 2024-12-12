-- name: CreateRealm :one
INSERT INTO realms (
    name,
    owner_id,
    capital_number,
    political_entity
) VALUES (
    $1, $2, $3, $4
) RETURNING id;

-- name: FindAllRealms :many
SELECT * FROM realms
WHERE owner_id = $1;
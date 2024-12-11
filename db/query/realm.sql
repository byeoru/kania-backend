-- name: FindAllRealms :many
SELECT * FROM realms
WHERE owner_id = $1;
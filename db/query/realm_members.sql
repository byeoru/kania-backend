-- name: CreateRealmMember :exec
INSERT INTO realm_members (
    user_id,
    status,
    private_money
) VALUES (
    $1, $2, $3
);

-- name: GetRealmIdByUserId :one
SELECT R.realm_id FROM realm_members AS RM
LEFT JOIN realms as R
ON RM.user_id = R.owner_id
WHERE user_id = $1
LIMIT 1;
-- name: CreateRealmMember :exec
INSERT INTO realm_members (
    realm_id,
    user_id,
    status,
    private_money
) VALUES (
    $1, $2, $3, $4
);

-- name: GetRealmMembersLevies :many
SELECT sqlc.embed(R), sqlc.embed(L) FROM realm_members AS R
INNER JOIN levies AS L ON R.user_id = L.realm_member_id
WHERE R.realm_id = $1;

-- name: GetRealmIdByUserId :one
SELECT realm_id FROM realm_members
WHERE user_id = $1;
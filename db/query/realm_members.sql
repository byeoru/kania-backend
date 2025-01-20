-- name: CreateRealmMember :one
INSERT INTO realm_members (
    user_id,
    status,
    private_money
) VALUES (
    $1, $2, $3
) RETURNING realm_member_id;

-- name: GetRealmIdByRmId :one
SELECT R.realm_id FROM realm_members AS RM
LEFT JOIN realms as R
ON RM.realm_member_id = R.rm_id
WHERE rm_id = $1
LIMIT 1;

-- name: GetMyRmIdOfSector :one
SELECT S.rm_id
FROM realm_members AS RM
INNER JOIN sectors AS S
ON RM.realm_member_id = S.rm_id AND cell_number = $2
WHERE user_id = $1;
-- name: CreateRealmMember :exec
INSERT INTO realm_members (
    rm_id,
    realm_id,
    status,
    private_money
) VALUES (
    $1, $2, $3, $4
);

-- name: FindRealmMember :one
SELECT * FROM realm_members
WHERE rm_id = $1 LIMIT 1;

-- name: FindFullRealmMember :one
SELECT sqlc.embed(RM), sqlc.embed(MA) 
FROM realm_members AS RM
INNER JOIN member_authorities AS MA
ON RM.rm_id = MA.rm_id
WHERE RM.rm_id = $1 LIMIT 1;

-- name: UpdateRealmMember :exec 
UPDATE realm_members
SET realm_id = $2,
status = $3,
private_money = $4
WHERE rm_id = $1;

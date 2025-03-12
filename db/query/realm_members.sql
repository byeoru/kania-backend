-- name: CreateRealmMember :exec
INSERT INTO realm_members (
    rm_id,
    realm_id,
    nickname,
    status,
    private_coffers
) VALUES (
    $1, $2, $3, $4, $5
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
private_coffers = $4
WHERE rm_id = $1;

-- name: TransferPrivateCoffers :one
WITH deducted AS (
    UPDATE realm_members
    SET private_coffers = FLOOR(private_coffers * (1 - sqlc.arg(reduction_rate)::float))::int
    WHERE rm_id = sqlc.arg(source_rm_id)::bigint
    RETURNING FLOOR(private_coffers / sqlc.arg(reduction_rate)::float)::int - private_coffers AS delta, private_coffers AS source_private_coffers
)
UPDATE realm_members
SET private_coffers = private_coffers + deducted.delta
FROM deducted
WHERE rm_id = sqlc.arg(receiver_rm_id)::bigint
RETURNING deducted.delta, deducted.source_private_coffers, private_coffers AS receiver_private_coffers;
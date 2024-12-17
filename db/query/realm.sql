-- name: CreateRealm :one
INSERT INTO realms (
    name,
    owner_id,
    capital_number,
    political_entity,
    color
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING id;

-- name: FindRealmWithJson :one
SELECT * FROM realms AS R
LEFT JOIN realm_sectors_jsonb AS J 
ON R.id = J.realm_id 
WHERE R.owner_id = $1 LIMIT 1;
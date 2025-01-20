-- name: CreateRealm :one
INSERT INTO realms (
    name,
    rm_id,
    owner_nickname,
    political_entity,
    color,
    population_growth_rate,
    state_coffers,
    census_at,
    tax_collection_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: FindRealmWithJson :one
SELECT 
realm_id, 
name,
owner_nickname,
political_entity, 
color, 
population_growth_rate, 
state_coffers, 
census_at, 
tax_collection_at,
capitals,
J.cells_jsonb
FROM realms AS R
LEFT JOIN realm_sectors_jsonb AS J 
ON R.realm_id = J.realm_sectors_jsonb_id 
WHERE R.rm_id = $1 LIMIT 1;

-- name: FindAllRealmsWithJsonExcludeMe :many
SELECT 
realm_id, 
name, 
owner_nickname, 
capitals, 
political_entity, 
color,
J.cells_jsonb
FROM realms AS R
LEFT JOIN realm_sectors_jsonb AS J 
ON R.realm_id = J.realm_sectors_jsonb_id
WHERE R.rm_id != $1;

-- name: UpdateCensusAt :exec
UPDATE realms
SET census_at = $2
WHERE realm_id = $1;

-- name: GetCensusAndPopulationGrowthRate :one
SELECT census_at, population_growth_rate FROM realms
WHERE realm_id = $1;

-- name: CheckCellOwner :one
SELECT EXISTS (
    SELECT 1
    FROM realms
    WHERE realm_id = $1 AND rm_id = $2
);

-- name: GetRealmId :one
SELECT realm_id FROM realms
WHERE rm_id = $1;

-- name: UpdateStateCoffers :one
UPDATE realms
SET state_coffers = state_coffers - sqlc.arg(deduction)
WHERE realm_id = sqlc.arg(realm_id) AND state_coffers >= sqlc.arg(deduction)
RETURNING state_coffers;

-- name: GetRealmIdWithSector :one
SELECT R.realm_id, name, cell_number FROM realms AS R
LEFT JOIN sectors AS S
ON R.realm_id = S.realm_id AND S.cell_number = $2
WHERE R.rm_id = $1
LIMIT 1;

-- name: GetOurRealmLevies :many
SELECT sqlc.embed(R), sqlc.embed(L) FROM realms AS R
INNER JOIN levies AS L ON R.realm_id = L.realm_id
WHERE R.realm_id = $1;

-- name: RemoveRealm :exec
DELETE FROM realms
WHERE realm_id = $1;

-- name: AddCapital :exec
UPDATE realms
SET capitals = array_append(capitals, sqlc.arg(capital)::int)
WHERE realm_id = $1;

-- name: FindRealm :one
SELECT * FROM realms
WHERE realm_id = $1
LIMIT 1;

-- name: UpdateRealmPoliticalEntityAndRemoveCapital :exec 
UPDATE realms
SET political_entity = sqlc.arg(political_entity)::varchar 
AND capitals = array_remove(capitals, sqlc.arg(remove_capital)::int)
WHERE realm_id = sqlc.arg(realm_id)::bigint;

-- name: RemoveCapital :exec
UPDATE realms
SET capitals = array_remove(capitals, sqlc.arg(remove_capital)::int)
WHERE realm_id = sqlc.arg(realm_id)::bigint;

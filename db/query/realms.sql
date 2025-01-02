-- name: CreateRealm :one
INSERT INTO realms (
    name,
    owner_id,
    owner_nickname,
    capital_number,
    political_entity,
    color,
    population_growth_rate,
    state_coffers,
    census_at,
    tax_collection_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
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
capital_number,
J.cells_jsonb
FROM realms AS R
LEFT JOIN realm_sectors_jsonb AS J 
ON R.realm_id = J.realm_sectors_jsonb_id 
WHERE R.owner_id = $1 LIMIT 1;

-- name: FindAllRealmsWithJsonExcludeMe :many
SELECT 
realm_id, 
name, 
owner_nickname, 
capital_number, 
political_entity, 
color,
J.cells_jsonb
FROM realms AS R
LEFT JOIN realm_sectors_jsonb AS J 
ON R.realm_id = J.realm_sectors_jsonb_id
WHERE R.owner_id != $1;

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
    WHERE realm_id = $1 AND owner_id = $2
);

-- name: GetRealmId :one
SELECT realm_id FROM realms
WHERE owner_id = $1;

-- name: UpdateStateCoffers :one
UPDATE realms
SET state_coffers = state_coffers - sqlc.arg(deduction)
WHERE realm_id = sqlc.arg(realm_id) AND state_coffers >= sqlc.arg(deduction)
RETURNING state_coffers;

-- name: GetRealmIdWithSector :one
SELECT R.realm_id, name, cell_number FROM realms AS R
LEFT JOIN sectors AS S
ON R.realm_id = S.realm_id AND S.cell_number = $2
WHERE R.owner_id = $1;
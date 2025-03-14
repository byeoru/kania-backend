-- name: CreateRealm :one
INSERT INTO realms (
    name,
    owner_rm_id,
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

-- name: UpdateCensusAt :exec
UPDATE realms
SET census_at = $2
WHERE realm_id = $1;

-- name: GetCensusAndPopulationGrowthRate :one
SELECT census_at, population_growth_rate FROM realms
WHERE realm_id = $1;

-- name: UpdateStateCoffers :one
UPDATE realms
SET state_coffers = state_coffers - sqlc.arg(deduction)
WHERE realm_id = sqlc.arg(realm_id) AND state_coffers >= sqlc.arg(deduction)
RETURNING state_coffers;

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
SET political_entity = sqlc.arg(political_entity)::varchar,
capitals = array_remove(capitals, sqlc.arg(remove_capital)::int),
population_growth_rate = sqlc.arg(population_growth_rate)::float
WHERE realm_id = sqlc.arg(realm_id)::bigint;

-- name: RemoveCapital :exec
UPDATE realms
SET capitals = array_remove(capitals, sqlc.arg(remove_capital)::int)
WHERE realm_id = sqlc.arg(realm_id)::bigint;

-- name: GetRealmOwnerRmId :one
SELECT owner_rm_id FROM realms
WHERE realm_id = $1 LIMIT 1;

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
WHERE R.realm_id = $1;

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
WHERE R.realm_id != $1;

-- name: TransferStateCoffers :one
WITH deducted AS (
    UPDATE realms
    SET state_coffers = FLOOR(state_coffers * (1 - sqlc.arg(reduction_rate)::float))::int
    WHERE realm_id = sqlc.arg(source_realm_id)::bigint
    RETURNING FLOOR(state_coffers / sqlc.arg(reduction_rate)::float)::int - state_coffers AS delta, state_coffers AS source_state_coffers
)
UPDATE realms
SET state_coffers = state_coffers + deducted.delta
FROM deducted
WHERE realm_id = sqlc.arg(receiver_realm_id)::bigint
RETURNING deducted.delta, deducted.source_state_coffers, state_coffers AS receiver_state_coffers;
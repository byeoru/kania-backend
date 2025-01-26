-- name: CreateLevy :one
INSERT INTO levies (
    name,
    morale,
    stationed,
    encampment,
    swordmen,
    shield_bearers,
    archers,
    lancers,
    supply_troop,
    movement_speed,
    rm_id,
    realm_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: FindLevy :one
SELECT * FROM levies
WHERE levy_id = $1
LIMIT 1;

-- name: UpdateLevy :exec
UPDATE levies
SET encampment = $2,
swordmen = $3,
shield_bearers = $4,
archers = $5,
lancers = $6,
supply_troop = $7,
movement_speed = $8,
stationed = $9
WHERE levy_id = $1;

-- name: UpdateLevyStatus :exec
UPDATE levies
SET stationed = $2
WHERE levy_id = $1;

-- name: RemoveLevy :exec
DELETE FROM levies
WHERE levy_id = $1;

-- name: UpdateLevyEncampment :exec
UPDATE levies
SET encampment = $2
WHERE levy_id = $1;

-- name: FindEncampmentLevies :many
SELECT * FROM levies
WHERE realm_id = $1 AND encampment = $2;

-- name: RemoveStationedLevies :exec
DELETE FROM levies
WHERE realm_id = $1 AND encampment = $2 AND stationed = true;

-- name: GetEncampmentOfMyLevy :one
SELECT encampment FROM levies
WHERE levy_id = $1 AND rm_id = $2 LIMIT 1;

-- name: FindLevyInfoWithAuthority :one
SELECT L.encampment, L.movement_speed, RM.realm_id, MA.*
FROM levies AS L
INNER JOIN realm_members AS RM
ON L.rm_id = RM.rm_id
INNER JOIN member_authorities AS MA
ON RM.rm_id = MA.rm_id
WHERE L.levy_id = $1
LIMIT 1;

-- name: FindOurRealmLevies :many
SELECT * FROM levies
WHERE realm_id = $1;
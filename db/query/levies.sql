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
    realm_member_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetOwnerIdByLevyId :one
SELECT realm_member_id FROM levies
WHERE levy_id = $1;

-- name: FindStationedLevies :many
SELECT * FROM levies
WHERE encampment = $1 AND stationed = true
FOR UPDATE;

-- name: UpdateLevy :exec
UPDATE levies
SET encampment = $2,
swordmen = $3,
shield_bearers = $4,
archers = $5,
lancers = $6,
supply_troop = $7,
movement_speed = $8
WHERE levy_id = $1;

-- name: UpdateLevyStatus :exec
UPDATE levies
SET stationed = $2
WHERE levy_id = $1;
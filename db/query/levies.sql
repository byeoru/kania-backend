-- name: CreateLevy :one
INSERT INTO levies (
    name,
    morale,
    encampment,
    swordmen,
    shield_bearers,
    archers,
    lancers,
    supply_troop,
    movement_speed,
    offensive_strength,
    defensive_strength,
    realm_member_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;
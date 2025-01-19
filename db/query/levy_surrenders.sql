-- name: CreateLevySurrender :exec
INSERT INTO levy_surrenders (
    levy_id,
    receiving_realm_id,
    surrender_reason,
    surrendered_at,
    surrendered_sector_location
) VALUES (
    $1, $2, $3, $4, $5
);
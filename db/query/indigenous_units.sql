-- name: InitIndigenousUnits :exec
COPY indigenous_units (
    sector_number,
    swordmen,
    archers,
    lancers,
    offensive_strength,
    defensive_strength
) FROM '/var/lib/postgresql/data/indigenous_units.csv' 
WITH DELIMITER ',' 
CSV HEADER;

-- name: FindIndigenousUnit :one
SELECT * FROM indigenous_units
WHERE sector_number = $1;
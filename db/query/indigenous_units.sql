-- name: InitIndigenousUnits :exec
COPY indigenous_units (
    sector_number,
    swordmen,
    archers,
    lancers
) FROM '/var/lib/postgresql/data/indigenous_units.csv' 
WITH DELIMITER ',' 
CSV HEADER;

-- name: FindIndigenousUnit :one
SELECT * FROM indigenous_units
WHERE sector_number = $1
LIMIT 1;

-- name: UpdateIndigenousUnits :exec
UPDATE indigenous_units
SET swordmen = $2,
archers = $3,
lancers = $4
WHERE sector_number = $1;
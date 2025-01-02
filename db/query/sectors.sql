-- name: CreateSector :exec
INSERT INTO sectors (
    cell_number,
    province_number,
    realm_id,
    population
) VALUES (
    $1, $2, $3, $4
);

-- name: UpdatePopulation :exec
UPDATE sectors 
SET population = CEIL(population * POW(1 + @rate_of_increase::float, sqlc.arg(duration_day)::float / 365.25))
WHERE realm_id = sqlc.arg(realm_id)::bigint;

-- name: GetPopulation :one
SELECT population, realm_id FROM sectors
WHERE cell_number = $1 LIMIT 1;
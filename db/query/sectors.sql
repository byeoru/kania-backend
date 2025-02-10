-- name: CreateSector :exec
INSERT INTO sectors (
    cell_number,
    province_number,
    realm_id,
    rm_id,
    population
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: UpdatePopulation :one
UPDATE sectors
SET population = population - sqlc.arg(deduction)::int
WHERE cell_number = sqlc.arg(cellNumber)::int AND population >= sqlc.arg(deduction)::int
RETURNING population;

-- name: UpdateCensusPopulation :exec
UPDATE sectors 
SET population = CEIL(population * POW(1 + @rate_of_increase::float, sqlc.arg(duration_day)::float / 365.25))
WHERE realm_id = sqlc.arg(realm_id)::bigint;

-- name: GetPopulation :one
SELECT population, realm_id FROM sectors
WHERE cell_number = $1 LIMIT 1;

-- name: GetSectorRealmId :one
SELECT realm_id FROM sectors
WHERE cell_number = $1 LIMIT 1;

-- name: GetSectorRealmIdForUpdate :one
SELECT realm_id FROM sectors
WHERE cell_number = $1 LIMIT 1
FOR UPDATE;

-- name: UpdateSectorOwnership :exec
UPDATE sectors
SET realm_id = $2, rm_id = $3
WHERE cell_number = $1;

-- name: UpdateSectorToIndigenous :exec
DELETE FROM sectors
WHERE cell_number IN (
    SELECT sectors.cell_number
    FROM sectors
    LEFT JOIN levies
    ON sectors.cell_number = levies.encampment
    WHERE sectors.realm_id = $1
    GROUP BY sectors.cell_number
    HAVING COUNT(levies.encampment) = 0
);

-- name: TransferSectorOwnershipToAttackers :exec
UPDATE sectors
SET realm_id = sqlc.arg(attacker_realm_id)::bigint
WHERE cell_number IN (
    SELECT sectors.cell_number
    FROM sectors
    LEFT JOIN levies
    ON sectors.cell_number = levies.encampment
    WHERE sectors.realm_id = sqlc.arg(defender_realm_id)::bigint
    GROUP BY sectors.cell_number
    HAVING COUNT(levies.encampment) = 0
);

-- name: GetNumberOfRealmSectors :one
SELECT COUNT(S) AS sector_count FROM sectors AS S
WHERE S.realm_id = $1
LIMIT 1;

-- name: FindSectorRealmForUpdate :one
SELECT sqlc.embed(S), sqlc.embed(R) FROM sectors AS S
INNER JOIN realms AS R
ON S.realm_id = R.realm_id
WHERE cell_number = $1
LIMIT 1 FOR UPDATE;
-- name: CreateSector :exec
INSERT INTO sectors (
    cell_number,
    province_number,
    realm_id
) VALUES (
    $1, $2, $3
);
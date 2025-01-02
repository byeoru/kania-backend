-- name: CreateRealmSectorsJsonb :exec
INSERT INTO realm_sectors_jsonb (
    realm_sectors_jsonb_id, cells_jsonb
) VALUES (
    $1, $2
);
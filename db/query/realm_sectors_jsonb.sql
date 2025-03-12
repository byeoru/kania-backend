-- name: CreateRealmSectorsJsonb :exec
INSERT INTO realm_sectors_jsonb (
    realm_sectors_jsonb_id, cells_jsonb
) VALUES (
    $1, $2
);

-- name: AddRealmSectorJsonb :exec
UPDATE realm_sectors_jsonb
SET cells_jsonb = jsonb_set(
  cells_jsonb,
  sqlc.arg(key),  -- 키의 경로
  to_jsonb(sqlc.arg(value)::int)  -- 새로운 요소 추가
) WHERE realm_sectors_jsonb_id = sqlc.arg(realm_id)::bigint;

-- name: RemoveSectorJsonb :exec
UPDATE realm_sectors_jsonb
SET cells_jsonb = cells_jsonb - sqlc.arg(key)::varchar
WHERE realm_sectors_jsonb_id = sqlc.arg(realm_id)::bigint;

-- name: FindBothJsonb :many
SELECT * FROM realm_sectors_jsonb
WHERE realm_sectors_jsonb_id 
IN (sqlc.arg(realm_id_1)::bigint, sqlc.arg(realm_id_2)::bigint);
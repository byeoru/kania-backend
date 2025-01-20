-- name: CreateWorldTimeRecord :exec
INSERT INTO world_time_records (
    stop_reason,
    world_stopped_at
) VALUES (
    $1, $2
);

-- name: FindLatestWorldTimeRecord :one
SELECT * FROM world_time_records
ORDER BY world_time_record_id DESC LIMIT 1;
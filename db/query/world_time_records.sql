-- name: CreateWorldTimeRecord :exec
INSERT INTO world_time_records (
    stop_reason,
    world_stopped_at
) VALUES (
    $1, $2
);
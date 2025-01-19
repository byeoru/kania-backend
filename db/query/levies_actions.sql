-- name: CreateLevyAction :exec
INSERT INTO levies_actions (
   levy_id,
   origin_sector,
   target_sector, 
   action_type,
   completed, 
   started_at,
   expected_completion_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: FindLevyActionCountByLevyId :one
SELECT COUNT(*) FROM levies_actions
WHERE levy_id = sqlc.arg(levy_id)::bigint AND expected_completion_at < sqlc.arg(reference_date)::timestamptz
LIMIT 1;

-- name: FindLevyAction :one
SELECT * FROM levies_actions
WHERE levy_action_id = $1 AND action_type = $2
LIMIT 1;

-- name: FindLevyActionsBeforeDate :many
SELECT sqlc.embed(L), sqlc.embed(LA) FROM levies_actions AS LA
LEFT JOIN levies AS L
ON LA.levy_id = L.levy_id
WHERE LA.expected_completion_at <= sqlc.arg(current_world_time)::timestamptz
AND LA.completed = false
ORDER BY LA.expected_completion_at ASC;

-- name: UpdateLevyActionCompleted :exec
UPDATE levies_actions
SET completed = $2
WHERE levy_action_id = $1;
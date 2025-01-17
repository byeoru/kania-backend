-- name: CreateLevyAction :exec
INSERT INTO levies_actions (
   levy_id,
   origin_sector,
   target_sector, 
   action_type,
   completed, 
   expected_completion_at
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: FindLevyActionCountByLevyId :one
SELECT COUNT(*) FROM levies_actions
WHERE levy_id = sqlc.arg(levy_id)::bigint AND expected_completion_at < sqlc.arg(reference_date)::timestamptz
LIMIT 1;

-- name: FindLevyAction :one
SELECT * FROM levies_actions
WHERE levy_action_id = $1 AND action_type = $2;

-- name: FindTargetLevyActionsSortedByDateForUpdate :many
SELECT sqlc.embed(LA), sqlc.embed(L) FROM levies_actions AS LA
LEFT JOIN levies AS L
ON LA.levy_id = L.levy_id
WHERE LA.target_sector = sqlc.arg(targetSectorId)::int 
AND LA.expected_completion_at < sqlc.arg(expectedCompletionAt)::timestamptz
AND LA.completed = false
ORDER BY LA.expected_completion_at ASC
FOR UPDATE OF levies_actions, levies;

-- name: UpdateLevyActionCompleted :exec
UPDATE levies_actions
SET completed = $2
WHERE levy_action_id = $1;
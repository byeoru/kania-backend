-- name: CreateBothPrivateCoffersLog :exec
INSERT INTO private_coffers_logs (
    rm_id,
    change_amount,
    total_coffers,
    reason,
    world_time_at
) VALUES 
(sqlc.arg(source_rm_id)::bigint, sqlc.arg(source_change_amount)::int, sqlc.arg(source_total_coffers)::int, sqlc.arg(source_reason)::varchar, sqlc.arg(world_time_at)::timestamptz),
(sqlc.arg(receiver_rm_id)::bigint, sqlc.arg(receiver_change_amount)::int, sqlc.arg(receiver_total_coffers)::int, sqlc.arg(receiver_reason)::varchar, sqlc.arg(world_time_at)::timestamptz);
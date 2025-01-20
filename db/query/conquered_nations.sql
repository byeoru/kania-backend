-- name: CreateConqueredNations :exec
INSERT INTO conquered_nations (
   rm_id,
   owner_nickname,
   country_name,
   cells_jsonb,
   conquered_at
) VALUES (
    $1, $2, $3, $4, $5
);
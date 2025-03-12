-- name: CreateBattleOutcome :exec
INSERT INTO battle_outcomes (
   levy_action_id,
   realm_id,
   attacker,
   defender
) VALUES (
    $1, $2, $3, $4
);
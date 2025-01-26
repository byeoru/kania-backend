-- name: CreateMemberAuthority :exec
INSERT INTO member_authorities (
    rm_id,
    create_unit,
    reinforce_unit,
    move_unit,
    attack_unit,
    private_troops,
    census
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
);
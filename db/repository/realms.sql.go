// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: realms.sql

package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/sqlc-dev/pqtype"
)

const checkCellOwner = `-- name: CheckCellOwner :one
SELECT EXISTS (
    SELECT 1
    FROM realms
    WHERE realm_id = $1 AND owner_id = $2
)
`

type CheckCellOwnerParams struct {
	RealmID int64 `json:"realm_id"`
	OwnerID int64 `json:"owner_id"`
}

func (q *Queries) CheckCellOwner(ctx context.Context, arg *CheckCellOwnerParams) (bool, error) {
	row := q.db.QueryRowContext(ctx, checkCellOwner, arg.RealmID, arg.OwnerID)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const createRealm = `-- name: CreateRealm :one
INSERT INTO realms (
    name,
    owner_id,
    owner_nickname,
    capital_number,
    political_entity,
    color,
    population_growth_rate,
    state_coffers,
    census_at,
    tax_collection_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING realm_id, name, owner_nickname, owner_id, capital_number, political_entity, color, population_growth_rate, state_coffers, census_at, tax_collection_at, created_at
`

type CreateRealmParams struct {
	Name                 string    `json:"name"`
	OwnerID              int64     `json:"owner_id"`
	OwnerNickname        string    `json:"owner_nickname"`
	CapitalNumber        int32     `json:"capital_number"`
	PoliticalEntity      string    `json:"political_entity"`
	Color                string    `json:"color"`
	PopulationGrowthRate float64   `json:"population_growth_rate"`
	StateCoffers         int32     `json:"state_coffers"`
	CensusAt             time.Time `json:"census_at"`
	TaxCollectionAt      time.Time `json:"tax_collection_at"`
}

func (q *Queries) CreateRealm(ctx context.Context, arg *CreateRealmParams) (*Realm, error) {
	row := q.db.QueryRowContext(ctx, createRealm,
		arg.Name,
		arg.OwnerID,
		arg.OwnerNickname,
		arg.CapitalNumber,
		arg.PoliticalEntity,
		arg.Color,
		arg.PopulationGrowthRate,
		arg.StateCoffers,
		arg.CensusAt,
		arg.TaxCollectionAt,
	)
	var i Realm
	err := row.Scan(
		&i.RealmID,
		&i.Name,
		&i.OwnerNickname,
		&i.OwnerID,
		&i.CapitalNumber,
		&i.PoliticalEntity,
		&i.Color,
		&i.PopulationGrowthRate,
		&i.StateCoffers,
		&i.CensusAt,
		&i.TaxCollectionAt,
		&i.CreatedAt,
	)
	return &i, err
}

const findAllRealmsWithJsonExcludeMe = `-- name: FindAllRealmsWithJsonExcludeMe :many
SELECT 
realm_id, 
name, 
owner_nickname, 
capital_number, 
political_entity, 
color,
J.cells_jsonb
FROM realms AS R
LEFT JOIN realm_sectors_jsonb AS J 
ON R.realm_id = J.realm_sectors_jsonb_id
WHERE R.owner_id != $1
`

type FindAllRealmsWithJsonExcludeMeRow struct {
	RealmID         int64                 `json:"realm_id"`
	Name            string                `json:"name"`
	OwnerNickname   string                `json:"owner_nickname"`
	CapitalNumber   int32                 `json:"capital_number"`
	PoliticalEntity string                `json:"political_entity"`
	Color           string                `json:"color"`
	CellsJsonb      pqtype.NullRawMessage `json:"cells_jsonb"`
}

func (q *Queries) FindAllRealmsWithJsonExcludeMe(ctx context.Context, ownerID int64) ([]*FindAllRealmsWithJsonExcludeMeRow, error) {
	rows, err := q.db.QueryContext(ctx, findAllRealmsWithJsonExcludeMe, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*FindAllRealmsWithJsonExcludeMeRow{}
	for rows.Next() {
		var i FindAllRealmsWithJsonExcludeMeRow
		if err := rows.Scan(
			&i.RealmID,
			&i.Name,
			&i.OwnerNickname,
			&i.CapitalNumber,
			&i.PoliticalEntity,
			&i.Color,
			&i.CellsJsonb,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findRealmWithJson = `-- name: FindRealmWithJson :one
SELECT 
realm_id, 
name,
owner_nickname,
political_entity, 
color, 
population_growth_rate, 
state_coffers, 
census_at, 
tax_collection_at,
capital_number,
J.cells_jsonb
FROM realms AS R
LEFT JOIN realm_sectors_jsonb AS J 
ON R.realm_id = J.realm_sectors_jsonb_id 
WHERE R.owner_id = $1 LIMIT 1
`

type FindRealmWithJsonRow struct {
	RealmID              int64                 `json:"realm_id"`
	Name                 string                `json:"name"`
	OwnerNickname        string                `json:"owner_nickname"`
	PoliticalEntity      string                `json:"political_entity"`
	Color                string                `json:"color"`
	PopulationGrowthRate float64               `json:"population_growth_rate"`
	StateCoffers         int32                 `json:"state_coffers"`
	CensusAt             time.Time             `json:"census_at"`
	TaxCollectionAt      time.Time             `json:"tax_collection_at"`
	CapitalNumber        int32                 `json:"capital_number"`
	CellsJsonb           pqtype.NullRawMessage `json:"cells_jsonb"`
}

func (q *Queries) FindRealmWithJson(ctx context.Context, ownerID int64) (*FindRealmWithJsonRow, error) {
	row := q.db.QueryRowContext(ctx, findRealmWithJson, ownerID)
	var i FindRealmWithJsonRow
	err := row.Scan(
		&i.RealmID,
		&i.Name,
		&i.OwnerNickname,
		&i.PoliticalEntity,
		&i.Color,
		&i.PopulationGrowthRate,
		&i.StateCoffers,
		&i.CensusAt,
		&i.TaxCollectionAt,
		&i.CapitalNumber,
		&i.CellsJsonb,
	)
	return &i, err
}

const getCensusAndPopulationGrowthRate = `-- name: GetCensusAndPopulationGrowthRate :one
SELECT census_at, population_growth_rate FROM realms
WHERE realm_id = $1
`

type GetCensusAndPopulationGrowthRateRow struct {
	CensusAt             time.Time `json:"census_at"`
	PopulationGrowthRate float64   `json:"population_growth_rate"`
}

func (q *Queries) GetCensusAndPopulationGrowthRate(ctx context.Context, realmID int64) (*GetCensusAndPopulationGrowthRateRow, error) {
	row := q.db.QueryRowContext(ctx, getCensusAndPopulationGrowthRate, realmID)
	var i GetCensusAndPopulationGrowthRateRow
	err := row.Scan(&i.CensusAt, &i.PopulationGrowthRate)
	return &i, err
}

const getRealmId = `-- name: GetRealmId :one
SELECT realm_id FROM realms
WHERE owner_id = $1
`

func (q *Queries) GetRealmId(ctx context.Context, ownerID int64) (int64, error) {
	row := q.db.QueryRowContext(ctx, getRealmId, ownerID)
	var realm_id int64
	err := row.Scan(&realm_id)
	return realm_id, err
}

const getRealmIdWithSector = `-- name: GetRealmIdWithSector :one
SELECT R.realm_id, name, cell_number FROM realms AS R
LEFT JOIN sectors AS S
ON R.realm_id = S.realm_id AND S.cell_number = $2
WHERE R.owner_id = $1
`

type GetRealmIdWithSectorParams struct {
	OwnerID    int64 `json:"owner_id"`
	CellNumber int32 `json:"cell_number"`
}

type GetRealmIdWithSectorRow struct {
	RealmID    int64         `json:"realm_id"`
	Name       string        `json:"name"`
	CellNumber sql.NullInt32 `json:"cell_number"`
}

func (q *Queries) GetRealmIdWithSector(ctx context.Context, arg *GetRealmIdWithSectorParams) (*GetRealmIdWithSectorRow, error) {
	row := q.db.QueryRowContext(ctx, getRealmIdWithSector, arg.OwnerID, arg.CellNumber)
	var i GetRealmIdWithSectorRow
	err := row.Scan(&i.RealmID, &i.Name, &i.CellNumber)
	return &i, err
}

const updateCensusAt = `-- name: UpdateCensusAt :exec
UPDATE realms
SET census_at = $2
WHERE realm_id = $1
`

type UpdateCensusAtParams struct {
	RealmID  int64     `json:"realm_id"`
	CensusAt time.Time `json:"census_at"`
}

func (q *Queries) UpdateCensusAt(ctx context.Context, arg *UpdateCensusAtParams) error {
	_, err := q.db.ExecContext(ctx, updateCensusAt, arg.RealmID, arg.CensusAt)
	return err
}

const updateStateCoffers = `-- name: UpdateStateCoffers :one
UPDATE realms
SET state_coffers = state_coffers - $1
WHERE realm_id = $2 AND state_coffers >= $1
RETURNING state_coffers
`

type UpdateStateCoffersParams struct {
	Deduction int32 `json:"deduction"`
	RealmID   int64 `json:"realm_id"`
}

func (q *Queries) UpdateStateCoffers(ctx context.Context, arg *UpdateStateCoffersParams) (int32, error) {
	row := q.db.QueryRowContext(ctx, updateStateCoffers, arg.Deduction, arg.RealmID)
	var state_coffers int32
	err := row.Scan(&state_coffers)
	return state_coffers, err
}
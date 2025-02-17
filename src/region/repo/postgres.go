package repo

import (
	"context"
	goerrors "errors"

	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/satori/uuid"
	"pet_adopter/src/Region"
)

const (
	getRegions    = `SELECT id, name FROM Region ORDER BY name ASC;`
	getRegionByID = `SELECT id, name FROM Region WHERE id = $1;`
	addRegion     = `INSERT INTO Region (id, name) VALUES ($1, $2);`
	removeRegion  = `DELETE FROM Region WHERE id = $1;`
)

type RegionPostgres struct {
	db pgxtype.Querier
}

func NewRegionPostgres(db pgxtype.Querier) *RegionPostgres {
	return &RegionPostgres{db: db}
}

func (repo *RegionPostgres) GetRegions(ctx context.Context) ([]region.Region, error) {
	result := make([]region.Region, 0)

	query, err := repo.db.Query(ctx, getRegions)
	if err != nil {
		return result, errors.Wrap(err, "failed to get regions from postgres")
	}
	defer query.Close()

	for query.Next() {
		var RegionRow region.Region
		if err = query.Scan(&RegionRow.ID, &RegionRow.Name); err != nil {
			return result, errors.Wrap(err, "failed to parse region")
		}
		result = append(result, RegionRow)
	}

	return result, nil
}

func (repo *RegionPostgres) GetRegionByID(ctx context.Context, id uuid.UUID) (region.Region, error) {
	result := region.Region{}
	if err := repo.db.QueryRow(ctx, getRegionByID, id).Scan(&result.ID, &result.Name); err != nil {
		if goerrors.Is(err, pgx.ErrNoRows) {
			return result, region.ErrRegionNotFound
		}
		return result, errors.Wrap(err, "failed to get region from postgres")
	}

	return result, nil
}

func (repo *RegionPostgres) AddRegion(ctx context.Context, region region.Region) error {
	if _, err := repo.db.Exec(ctx, addRegion, region.ID, region.Name); err != nil {
		return errors.Wrap(err, "failed to add region to postgres")
	}

	return nil
}

func (repo *RegionPostgres) RemoveRegionByID(ctx context.Context, id uuid.UUID) error {
	if _, err := repo.db.Exec(ctx, removeRegion, id); err != nil {
		if goerrors.Is(err, pgx.ErrNoRows) {
			return region.ErrRegionNotFound
		}
		return errors.Wrap(err, "failed to remove region from postgres")
	}

	return nil
}

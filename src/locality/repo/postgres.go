package repo

import (
	"context"
	goerrors "errors"

	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/satori/uuid"
	"pet_adopter/src/locality"
)

const (
	getLocalities           = `SELECT id, region_id, latitude, longitude, name FROM Locality ORDER BY name ASC;`
	getLocalityByID         = `SELECT id, region_id, latitude, longitude, name FROM Locality WHERE id = $1;`
	getLocalitiesByRegionID = `SELECT id, region_id, latitude, longitude, name FROM Locality WHERE region_id = $1;`
	addLocality             = `INSERT INTO Locality(id, region_id, latitude, longitude, name) VALUES($1, $2, $3, $4, $5);`
	removeLocality          = `DELETE FROM Locality WHERE id = $1;`
)

type LocalityPostgres struct {
	db pgxtype.Querier
}

func NewLocalityPostgres(db pgxtype.Querier) *LocalityPostgres {
	return &LocalityPostgres{db: db}
}

func (repo *LocalityPostgres) GetLocalities(ctx context.Context) ([]locality.Locality, error) {
	result := make([]locality.Locality, 0)

	query, err := repo.db.Query(ctx, getLocalities)
	if err != nil {
		return result, errors.Wrap(err, "failed to get localities from postgres")
	}

	for query.Next() {
		var localityRow locality.Locality
		if err = query.Scan(&localityRow.ID, &localityRow.RegionID, &localityRow.Latitude, &localityRow.Longitude, &localityRow.Name); err != nil {
			return result, errors.Wrap(err, "failed to parse locality")
		}
		result = append(result, localityRow)
	}

	return result, nil
}

func (repo *LocalityPostgres) GetLocalityByID(ctx context.Context, id uuid.UUID) (locality.Locality, error) {
	result := locality.Locality{}
	if err := repo.db.QueryRow(ctx, getLocalityByID, id).Scan(&result.ID, &result.RegionID, &result.Latitude, &result.Longitude, &result.Name); err != nil {
		if goerrors.Is(err, pgx.ErrNoRows) {
			return result, locality.ErrLocalityNotFound
		}
		return result, errors.Wrap(err, "failed to get locality from postgres")
	}

	return result, nil
}

func (repo *LocalityPostgres) GetLocalitiesByRegionID(ctx context.Context, regionID uuid.UUID) ([]locality.Locality, error) {
	result := make([]locality.Locality, 0)

	query, err := repo.db.Query(ctx, getLocalitiesByRegionID, regionID)
	if err != nil {
		return result, errors.Wrap(err, "failed to get localities by region id from postgres")
	}

	for query.Next() {
		var localityRow locality.Locality
		if err = query.Scan(&localityRow.ID, &localityRow.RegionID, &localityRow.Latitude, &localityRow.Longitude, &localityRow.Name); err != nil {
			return result, errors.Wrap(err, "failed to parse locality")
		}
		result = append(result, localityRow)
	}

	return result, nil
}

func (repo *LocalityPostgres) AddLocality(ctx context.Context, locality locality.Locality) error {
	if _, err := repo.db.Exec(ctx, addLocality, locality.ID, locality.RegionID, locality.Latitude, locality.Longitude, locality.Name); err != nil {
		return errors.Wrap(err, "failed to add locality to postgres")
	}

	return nil
}

func (repo *LocalityPostgres) RemoveLocalityByID(ctx context.Context, id uuid.UUID) error {
	if _, err := repo.db.Exec(ctx, removeLocality, id); err != nil {
		if goerrors.Is(err, pgx.ErrNoRows) {
			return locality.ErrLocalityNotFound
		}
		return errors.Wrap(err, "failed to remove locality from postgres")
	}

	return nil
}

package repo

import (
	"context"
	goerrors "errors"

	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/satori/uuid"
	"pet_adopter/src/breed"
)

const (
	getBreeds           = `SELECT id, animal_id, name FROM Breed ORDER BY name ASC;`
	getBreedByID        = `SELECT id, animal_id, name FROM Breed WHERE id = $1;`
	getBreedsByAnimalID = `SELECT id, animal_id, name FROM Breed WHERE animal_id = $1;`
	addBreed            = `INSERT INTO Breed(id, animal_id, name) VALUES($1, $2, $3);`
	removeBreed         = `DELETE FROM Breed WHERE id = $1;`
)

type BreedPostgres struct {
	db pgxtype.Querier
}

func NewBreedPostgres(db pgxtype.Querier) *BreedPostgres {
	return &BreedPostgres{db: db}
}

func (repo *BreedPostgres) GetBreeds(ctx context.Context) ([]breed.Breed, error) {
	result := make([]breed.Breed, 0)

	query, err := repo.db.Query(ctx, getBreeds)
	if err != nil {
		return result, errors.Wrap(err, "failed to get breeds from postgres")
	}
	defer query.Close()

	for query.Next() {
		var breedRow breed.Breed
		if err = query.Scan(&breedRow.ID, &breedRow.AnimalID, &breedRow.Name); err != nil {
			return result, errors.Wrap(err, "failed to parse breed")
		}
		result = append(result, breedRow)
	}

	return result, nil
}

func (repo *BreedPostgres) GetBreedByID(ctx context.Context, id uuid.UUID) (breed.Breed, error) {
	result := breed.Breed{}
	if err := repo.db.QueryRow(ctx, getBreedByID, id).Scan(&result.ID, &result.AnimalID, &result.Name); err != nil {
		if goerrors.Is(err, pgx.ErrNoRows) {
			return result, breed.ErrBreedNotFound
		}
		return result, errors.Wrap(err, "failed to get breed from postgres")
	}

	return result, nil
}

func (repo *BreedPostgres) GetBreedsByAnimalID(ctx context.Context, animalID uuid.UUID) ([]breed.Breed, error) {
	result := make([]breed.Breed, 0)

	query, err := repo.db.Query(ctx, getBreedsByAnimalID, animalID)
	if err != nil {
		return result, errors.Wrap(err, "failed to get breeds by animal id from postgres")
	}
	defer query.Close()

	for query.Next() {
		var breedRow breed.Breed
		if err = query.Scan(&breedRow.ID, &breedRow.AnimalID, &breedRow.Name); err != nil {
			return result, errors.Wrap(err, "failed to parse breed")
		}
		result = append(result, breedRow)
	}

	return result, nil
}

func (repo *BreedPostgres) AddBreed(ctx context.Context, breed breed.Breed) error {
	if _, err := repo.db.Exec(ctx, addBreed, breed.ID, breed.AnimalID, breed.Name); err != nil {
		return errors.Wrap(err, "failed to add breed to postgres")
	}

	return nil
}

func (repo *BreedPostgres) RemoveBreedByID(ctx context.Context, id uuid.UUID) error {
	if _, err := repo.db.Exec(ctx, removeBreed, id); err != nil {
		if goerrors.Is(err, pgx.ErrNoRows) {
			return breed.ErrBreedNotFound
		}
		return errors.Wrap(err, "failed to remove breed from postgres")
	}

	return nil
}

package repo

import (
	"context"
	goerrors "errors"

	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/satori/uuid"
	"pet_adopter/src/animal"
)

const (
	getAnimals    = `SELECT id, name FROM Animal ORDER BY name ASC`
	getAnimalByID = `SELECT id, name FROM Animal WHERE id = $1`
	addAnimal     = `INSERT INTO Animal (id, name) VALUES ($1, $2)`
	removeAnimal  = `DELETE FROM Animal WHERE id = $1`
)

type AnimalPostgres struct {
	db pgxtype.Querier
}

func NewAnimalsPostgres(db pgxtype.Querier) *AnimalPostgres {
	return &AnimalPostgres{db: db}
}

func (repo *AnimalPostgres) GetAnimals(ctx context.Context) ([]animal.Animal, error) {
	result := make([]animal.Animal, 0)

	query, err := repo.db.Query(ctx, getAnimals)
	if err != nil {
		return result, errors.Wrap(err, "failed to get animals from postgres")
	}

	for query.Next() {
		var animal animal.Animal
		if err = query.Scan(&animal.ID, &animal.Name); err != nil {
			return result, errors.Wrap(err, "failed to parse animal")
		}
		result = append(result, animal)
	}

	return result, nil
}

func (repo *AnimalPostgres) GetAnimalByID(ctx context.Context, id uuid.UUID) (animal.Animal, error) {
	result := animal.Animal{}
	if err := repo.db.QueryRow(ctx, getAnimalByID, id).Scan(&result.ID, &result.Name); err != nil {
		if goerrors.Is(err, pgx.ErrNoRows) {
			return result, animal.ErrAnimalNotFound
		}
		return result, errors.Wrap(err, "failed to get animal from postgres")
	}

	return result, nil
}

func (repo *AnimalPostgres) AddAnimal(ctx context.Context, animal animal.Animal) error {
	if _, err := repo.db.Exec(ctx, addAnimal, animal.ID, animal.Name); err != nil {
		return errors.Wrap(err, "failed to add animal to postgres")
	}

	return nil
}

func (repo *AnimalPostgres) RemoveAnimalByID(ctx context.Context, id uuid.UUID) error {
	if _, err := repo.db.Exec(ctx, removeAnimal, id); err != nil {
		if goerrors.Is(err, pgx.ErrNoRows) {
			return animal.ErrAnimalNotFound
		}
		return errors.Wrap(err, "failed to remove animal from postgres")
	}

	return nil
}

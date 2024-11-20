package breed

import (
	"context"

	"github.com/pkg/errors"
	"github.com/satori/uuid"
)

var (
	ErrBreedNotFound = errors.New("breed not found")
)

type Breed struct {
	ID       uuid.UUID `json:"id"`
	AnimalID uuid.UUID `json:"animal_id"`
	Name     string    `json:"name"`
}

type BreedRepo interface {
	GetBreeds(ctx context.Context) ([]Breed, error)
	GetBreedByID(ctx context.Context, id uuid.UUID) (Breed, error)
	GetBreedsByAnimalID(ctx context.Context, animalID uuid.UUID) ([]Breed, error)
	AddBreed(ctx context.Context, breed Breed) error
	RemoveBreedByID(ctx context.Context, id uuid.UUID) error
}

type BreedLogic interface {
	GetBreeds(ctx context.Context) ([]Breed, error)
	GetBreedByID(ctx context.Context, id uuid.UUID) (Breed, error)
	GetBreedsByAnimalID(ctx context.Context, animalID uuid.UUID) ([]Breed, error)
	AddBreed(ctx context.Context, name string, animalID uuid.UUID) (Breed, error)
	RemoveBreedByID(ctx context.Context, id uuid.UUID) error
}

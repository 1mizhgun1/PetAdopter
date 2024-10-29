package animal

import (
	"context"

	"github.com/pkg/errors"
	"github.com/satori/uuid"
)

var (
	ErrAnimalNotFound = errors.New("animal not found")
)

type Animal struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type AnimalRepo interface {
	GetAnimals(ctx context.Context) ([]Animal, error)
	GetAnimalByID(ctx context.Context, id uuid.UUID) (Animal, error)
	AddAnimal(ctx context.Context, animal Animal) error
	RemoveAnimalByID(ctx context.Context, id uuid.UUID) error
}

type AnimalLogic interface {
	GetAnimals(ctx context.Context) ([]Animal, error)
	GetAnimalByID(ctx context.Context, id uuid.UUID) (Animal, error)
	AddAnimal(ctx context.Context, name string) (Animal, error)
	RemoveAnimalByID(ctx context.Context, id uuid.UUID) error
}

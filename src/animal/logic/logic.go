package logic

import (
	"context"

	"github.com/satori/uuid"
	"pet_adopter/src/animal"
)

type AnimalLogic struct {
	repo animal.AnimalRepo
}

func NewAnimalLogic(repo animal.AnimalRepo) AnimalLogic {
	return AnimalLogic{repo: repo}
}

func (logic *AnimalLogic) GetAnimals(ctx context.Context) ([]animal.Animal, error) {
	return logic.repo.GetAnimals(ctx)
}

func (logic *AnimalLogic) GetAnimalByID(ctx context.Context, id uuid.UUID) (animal.Animal, error) {
	return logic.repo.GetAnimalByID(ctx, id)
}

func (logic *AnimalLogic) AddAnimal(ctx context.Context, name string) (animal.Animal, error) {
	result := animal.Animal{ID: uuid.NewV4(), Name: name}
	return result, logic.repo.AddAnimal(ctx, result)
}

func (logic *AnimalLogic) RemoveAnimalByID(ctx context.Context, id uuid.UUID) error {
	return logic.repo.RemoveAnimalByID(ctx, id)
}

package logic

import (
	"context"

	"github.com/satori/uuid"
	"pet_adopter/src/breed"
)

type BreedLogic struct {
	repo breed.BreedRepo
}

func NewBreedLogic(repo breed.BreedRepo) BreedLogic {
	return BreedLogic{repo: repo}
}

func (logic *BreedLogic) GetBreeds(ctx context.Context) ([]breed.Breed, error) {
	return logic.repo.GetBreeds(ctx)
}

func (logic *BreedLogic) GetBreedByID(ctx context.Context, id uuid.UUID) (breed.Breed, error) {
	return logic.repo.GetBreedByID(ctx, id)
}

func (logic *BreedLogic) GetBreedsByAnimalID(ctx context.Context, animalID uuid.UUID) ([]breed.Breed, error) {
	return logic.repo.GetBreedsByAnimalID(ctx, animalID)
}

func (logic *BreedLogic) AddBreed(ctx context.Context, name string, animalID uuid.UUID) (breed.Breed, error) {
	result := breed.Breed{ID: uuid.NewV4(), Name: name, AnimalID: animalID}
	return result, logic.repo.AddBreed(ctx, result)
}

func (logic *BreedLogic) RemoveBreedByID(ctx context.Context, id uuid.UUID) error {
	return logic.repo.RemoveBreedByID(ctx, id)
}

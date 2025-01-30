package logic

import (
	"context"

	"github.com/satori/uuid"
	"pet_adopter/src/locality"
)

type LocalityLogic struct {
	repo locality.LocalityRepo
}

func NewLocalityLogic(repo locality.LocalityRepo) LocalityLogic {
	return LocalityLogic{repo: repo}
}

func (logic *LocalityLogic) GetLocalities(ctx context.Context) ([]locality.Locality, error) {
	return logic.repo.GetLocalities(ctx)
}

func (logic *LocalityLogic) GetLocalityByID(ctx context.Context, id uuid.UUID) (locality.Locality, error) {
	return logic.repo.GetLocalityByID(ctx, id)
}

func (logic *LocalityLogic) GetLocalitiesByRegionID(ctx context.Context, regionID uuid.UUID) ([]locality.Locality, error) {
	return logic.repo.GetLocalitiesByRegionID(ctx, regionID)
}

func (logic *LocalityLogic) AddLocality(ctx context.Context, name string, regionID uuid.UUID, latitude float64, longitude float64) (locality.Locality, error) {
	result := locality.Locality{ID: uuid.NewV4(), Name: name, RegionID: regionID, Latitude: latitude, Longitude: longitude}
	return result, logic.repo.AddLocality(ctx, result)
}

func (logic *LocalityLogic) RemoveLocalityByID(ctx context.Context, id uuid.UUID) error {
	return logic.repo.RemoveLocalityByID(ctx, id)
}

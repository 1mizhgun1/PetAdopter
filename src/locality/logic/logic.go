package logic

import (
	"context"
	"math"

	"github.com/pkg/errors"
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

func (logic *LocalityLogic) GetLocalityByCoords(ctx context.Context, latitude float64, longitude float64) (locality.Locality, error) {
	all, err := logic.repo.GetLocalities(ctx)
	if err != nil {
		return locality.Locality{}, errors.Wrap(err, "failed to get localities")
	}

	minDist := math.MaxFloat64
	answer := locality.Locality{}
	for _, loc := range all {
		dx := latitude - loc.Latitude
		dy := longitude - loc.Longitude
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist < minDist {
			minDist = dist
			answer = loc
		}
	}

	return answer, nil
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

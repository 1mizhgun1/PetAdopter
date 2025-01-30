package logic

import (
	"context"

	"github.com/satori/uuid"
	"pet_adopter/src/region"
)

type RegionLogic struct {
	repo region.RegionRepo
}

func NewRegionLogic(repo region.RegionRepo) RegionLogic {
	return RegionLogic{repo: repo}
}

func (logic *RegionLogic) GetRegions(ctx context.Context) ([]region.Region, error) {
	return logic.repo.GetRegions(ctx)
}

func (logic *RegionLogic) GetRegionByID(ctx context.Context, id uuid.UUID) (region.Region, error) {
	return logic.repo.GetRegionByID(ctx, id)
}

func (logic *RegionLogic) AddRegion(ctx context.Context, name string) (region.Region, error) {
	result := region.Region{ID: uuid.NewV4(), Name: name}
	return result, logic.repo.AddRegion(ctx, result)
}

func (logic *RegionLogic) RemoveRegionByID(ctx context.Context, id uuid.UUID) error {
	return logic.repo.RemoveRegionByID(ctx, id)
}

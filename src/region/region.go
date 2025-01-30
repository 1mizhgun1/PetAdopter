package region

import (
	"context"

	"github.com/pkg/errors"
	"github.com/satori/uuid"
)

var (
	ErrRegionNotFound = errors.New("region not found")
)

type Region struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type RegionRepo interface {
	GetRegions(ctx context.Context) ([]Region, error)
	GetRegionByID(ctx context.Context, id uuid.UUID) (Region, error)
	AddRegion(ctx context.Context, region Region) error
	RemoveRegionByID(ctx context.Context, id uuid.UUID) error
}

type RegionLogic interface {
	GetRegions(ctx context.Context) ([]Region, error)
	GetRegionByID(ctx context.Context, id uuid.UUID) (Region, error)
	AddRegion(ctx context.Context, name string) (Region, error)
	RemoveRegionByID(ctx context.Context, id uuid.UUID) error
}

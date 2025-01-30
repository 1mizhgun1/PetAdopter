package locality

import (
	"context"

	"github.com/pkg/errors"
	"github.com/satori/uuid"
)

var (
	ErrLocalityNotFound = errors.New("locality not found")
)

type Locality struct {
	ID        uuid.UUID `json:"id"`
	RegionID  uuid.UUID `json:"region_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Name      string    `json:"name"`
}

type LocalityRepo interface {
	GetLocalities(ctx context.Context) ([]Locality, error)
	GetLocalityByID(ctx context.Context, id uuid.UUID) (Locality, error)
	GetLocalitiesByRegionID(ctx context.Context, regionID uuid.UUID) ([]Locality, error)
	AddLocality(ctx context.Context, locality Locality) error
	RemoveLocalityByID(ctx context.Context, id uuid.UUID) error
}

type LocalityLogic interface {
	GetLocalities(ctx context.Context) ([]Locality, error)
	GetLocalityByID(ctx context.Context, id uuid.UUID) (Locality, error)
	GetLocalitiesByRegionID(ctx context.Context, regionID uuid.UUID) ([]Locality, error)
	AddLocality(ctx context.Context, name string, regionID uuid.UUID, latitude float64, longitude float64) (Locality, error)
	RemoveLocalityByID(ctx context.Context, id uuid.UUID) error
}

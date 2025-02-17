package ad

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/uuid"
	"pet_adopter/src/config"
)

var (
	ErrAdNotFound = errors.New("ad not found")
	ErrNotOwner   = errors.New("not owner")
)

const (
	Actual    = 'A'
	Realised  = 'R'
	Cancelled = 'C'
)

type Ad struct {
	ID uuid.UUID `json:"id"`

	OwnerID uuid.UUID `json:"owner_id"`
	Status  AdStatus  `json:"status"`

	AdForm

	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}

type AdStatus byte

type AdForm struct {
	PhotoURL    string    `json:"photo_url"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	AnimalID    uuid.UUID `json:"animal_id"`
	BreedID     uuid.UUID `json:"breed_id"`
	Price       int       `json:"price"`
	Contacts    string    `json:"contacts"`
}

type UpdateForm struct {
	PhotoURL    *string    `json:"photo_url"`
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	AnimalID    *uuid.UUID `json:"animal_id"`
	BreedID     *uuid.UUID `json:"breed_id"`
	Price       *int       `json:"price"`
	Contacts    *string    `json:"contacts"`
	Status      *AdStatus  `json:"status"`
}

type SearchParams struct {
	OwnerID  *uuid.UUID `json:"owner_id"`
	AnimalID *uuid.UUID `json:"animal_id"`
	BreedID  *uuid.UUID `json:"breed_id"`
	MinPrice *int       `json:"min_price"`
	MaxPrice *int       `json:"max_price"`

	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func NewSearchParams(cfg config.AdConfig) SearchParams {
	return SearchParams{
		OwnerID:  nil,
		AnimalID: nil,
		BreedID:  nil,
		MinPrice: nil,
		MaxPrice: nil,
		Limit:    cfg.DefaultSearchLimit,
		Offset:   cfg.DefaultSearchOffset,
	}
}

type AdRepo interface {
	SearchAds(ctx context.Context, params SearchParams) ([]Ad, error)
	GetAd(ctx context.Context, id uuid.UUID) (Ad, error)
	CreateAd(ctx context.Context, ad Ad) error
	UpdateAd(ctx context.Context, id uuid.UUID, form UpdateForm, now time.Time) error
	DeleteAd(ctx context.Context, id uuid.UUID) error
}

type AdLogic interface {
	SearchAds(ctx context.Context, params SearchParams) ([]Ad, error)
	GetAd(ctx context.Context, id uuid.UUID) (Ad, error)
	CreateAd(ctx context.Context, form AdForm) (Ad, error)
	UpdateAd(ctx context.Context, id uuid.UUID, form UpdateForm) (Ad, error)
	UpdatePhoto(ctx context.Context, id uuid.UUID, newPhotoURL string) (Ad, error)
	Close(ctx context.Context, id uuid.UUID, status AdStatus) (Ad, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

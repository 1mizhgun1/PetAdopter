package ad

import (
	"context"
	"io"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/uuid"
	"pet_adopter/src/config"
)

var (
	ErrAdNotFound        = errors.New("ad not found")
	ErrNotOwner          = errors.New("not owner")
	ErrInvalidForeignKey = errors.New("invalid foreign key")
)

const (
	Actual    = "A"
	Realised  = "R"
	Cancelled = "C"
)

type Ad struct {
	ID uuid.UUID `json:"id"`

	OwnerID uuid.UUID `json:"owner_id"`
	Status  string    `json:"status"`

	AdForm

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AdStatus byte

type AdForm struct {
	PhotoURL    string    `json:"photo_url,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	AnimalID    uuid.UUID `json:"animal_id"`
	BreedID     uuid.UUID `json:"breed_id"`
	Price       int       `json:"price"`
	Contacts    string    `json:"contacts"`
}

type UpdateForm struct {
	PhotoURL    *string    `json:"photo_url,omitempty"`
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	AnimalID    *uuid.UUID `json:"animal_id,omitempty"`
	BreedID     *uuid.UUID `json:"breed_id,omitempty"`
	Price       *int       `json:"price,omitempty"`
	Contacts    *string    `json:"contacts,omitempty"`
	Status      *string    `json:"status,omitempty"`
}

type PhotoParams struct {
	Data      io.ReadSeeker `json:"data"`
	Extension string        `json:"extension"`
}

type AdInfo struct {
	Username     string `json:"username"`
	AnimalName   string `json:"animal_name"`
	BreedName    string `json:"breed_name"`
	LocalityName string `json:"locality_name"`
}

type RespAd struct {
	Info      Ad     `json:"info"`
	ExtraInfo AdInfo `json:"extra_info"`
}

type SearchParams struct {
	OwnerID  *uuid.UUID `json:"owner_id"`
	AnimalID *uuid.UUID `json:"animal_id"`
	BreedID  *uuid.UUID `json:"breed_id"`
	MinPrice *int       `json:"min_price"`
	MaxPrice *int       `json:"max_price"`
	Radius   *int       `json:"radius"`

	AllStatuses bool `json:"all_statuses"`

	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type SearchExtra struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func NewSearchParams(cfg config.AdConfig) SearchParams {
	return SearchParams{
		OwnerID:     nil,
		AnimalID:    nil,
		BreedID:     nil,
		MinPrice:    nil,
		MaxPrice:    nil,
		Radius:      nil,
		AllStatuses: false,
		Limit:       cfg.DefaultSearchLimit,
		Offset:      cfg.DefaultSearchOffset,
	}
}

type AdRepo interface {
	SearchAds(ctx context.Context, params SearchParams, extra SearchExtra) ([]RespAd, error)
	GetAd(ctx context.Context, id uuid.UUID) (RespAd, error)
	CreateAd(ctx context.Context, ad Ad) error
	UpdateAd(ctx context.Context, id uuid.UUID, form UpdateForm, now time.Time) error
	DeleteAd(ctx context.Context, id uuid.UUID) error
}

type AdLogic interface {
	SearchAds(ctx context.Context, params SearchParams, extra SearchExtra) ([]RespAd, error)
	GetAd(ctx context.Context, id uuid.UUID) (RespAd, error)
	CreateAd(ctx context.Context, form AdForm, photoForm PhotoParams) (RespAd, error)
	UpdateAd(ctx context.Context, id uuid.UUID, form UpdateForm) (RespAd, error)
	UpdatePhoto(ctx context.Context, id uuid.UUID, photoForm PhotoParams) (RespAd, error)
	Close(ctx context.Context, id uuid.UUID, status string) (RespAd, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

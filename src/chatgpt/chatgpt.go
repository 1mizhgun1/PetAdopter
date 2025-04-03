package chatgpt

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/uuid"
	"pet_adopter/src/ad"
	"pet_adopter/src/utils"
)

const (
	ContentTypeText  = "input_text"
	ContentTypeImage = "input_image"
)

var (
	ErrDescriptionNotFound = errors.New("description not found")

	DescribePhotoPrompt = "Определи цвет окраса животного в формате RGB и дай ответ по шаблону: {\"color\":\"243 12 123\"} формат RGB. Если на фото нет животного или сложно распознать - напиши пустой json {}. Если На фото несколько животных - выбери любого на свой выбор. В ответе напиши только результат и ничего лишнего."
)

type Content []ContentItem

type ContentItem struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

type Description struct {
	Color string `json:"color"`
}

type PostgresDescription struct {
	ID        uuid.UUID `json:"id"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChatGPTClient interface {
	SendRequest(content Content) (string, error)
}

type ChatGPTRepo interface {
	GetDescriptions(ctx context.Context) ([]PostgresDescription, error)
	GetDescription(ctx context.Context, id uuid.UUID) (PostgresDescription, error)
	CreateDescription(ctx context.Context, description PostgresDescription) error
	UpdateDescription(ctx context.Context, description PostgresDescription) error
	DeleteDescription(ctx context.Context, id uuid.UUID) error
}

type ChatGPT interface {
	GetSame(ctx context.Context, id uuid.UUID, color utils.Color) ([]ad.RespAd, error)
	GetDescriptionFromDB(ctx context.Context, id uuid.UUID) (Description, error)
	DescribePhoto(ctx context.Context, id uuid.UUID, photo ad.PhotoParams, update bool) error
	DeleteDescription(ctx context.Context, id uuid.UUID) error
}

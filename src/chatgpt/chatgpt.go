package chatgpt

import "pet_adopter/src/ad"

const (
	ContentTypeText  = "input_text"
	ContentTypeImage = "input_image"
)

var (
	DescribePhotoPrompt = "Определи цвет окраса животного в формате RGB и дай ответ по шаблону: {\"color\":\"0f56c2\"} формат RGB. Если на фото нет животного или сложно распознать - напиши пустой json {}. Если На фото несколько животных - выбери любого на свой выбор. В ответе напиши только результат и ничего лишнего."
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

type ChatGPTClient interface {
	SendRequest(content Content) (string, error)
}

type ChatGPT interface {
	DescribePhoto(photo ad.PhotoParams) (Description, error)
}

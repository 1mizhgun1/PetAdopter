package logic

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"pet_adopter/src/ad"
	"pet_adopter/src/chatgpt"
)

type ChatGPT struct {
	client chatgpt.ChatGPTClient
}

func NewChatGPT(client chatgpt.ChatGPTClient) *ChatGPT {
	return &ChatGPT{
		client: client,
	}
}

func (c *ChatGPT) DescribePhoto(photo ad.PhotoParams) (chatgpt.Description, error) {
	_, err := photo.Data.Seek(0, io.SeekStart)
	if err != nil {
		return chatgpt.Description{}, errors.Wrap(err, "failed to set zero seek offset on photo")
	}

	photoData, err := io.ReadAll(photo.Data)
	if err != nil {
		return chatgpt.Description{}, errors.Wrap(err, "failed to read bytes from photo")
	}

	content := chatgpt.Content{
		{
			Type: chatgpt.ContentTypeText,
			Text: chatgpt.DescribePhotoPrompt,
		},
		{
			Type:     chatgpt.ContentTypeImage,
			ImageURL: makeImageURL(photoData, photo.Extension),
		},
	}

	answer, err := c.client.SendRequest(content)
	if err != nil {
		return chatgpt.Description{}, errors.Wrap(err, "SendRequest failed")
	}

	answer = strings.TrimPrefix(answer, "```")
	answer = strings.TrimPrefix(answer, "json")
	answer = strings.TrimSuffix(answer, "```")

	var description chatgpt.Description
	if err = json.Unmarshal([]byte(answer), &description); err != nil {
		fmt.Printf("[DEBUG] invalid GPT answer: %s\n", answer)
		return chatgpt.Description{}, errors.Wrap(err, "failed to unmarshal answer into description")
	}

	return description, nil
}

func makeImageURL(photo []byte, extension string) string {
	return fmt.Sprintf(
		"data:image/%s;base64,%s",
		strings.TrimPrefix(extension, "."),
		base64.StdEncoding.EncodeToString(photo),
	)
}

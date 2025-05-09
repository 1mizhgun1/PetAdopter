package logic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/uuid"
	"pet_adopter/src/ad"
	"pet_adopter/src/chatgpt"
	"pet_adopter/src/config"
	"pet_adopter/src/utils"
)

type ChatGPT struct {
	client chatgpt.ChatGPTClient
	repo   chatgpt.ChatGPTRepo
	adRepo ad.AdRepo
	cfg    config.Config
}

func NewChatGPT(client chatgpt.ChatGPTClient, repo chatgpt.ChatGPTRepo, adRepo ad.AdRepo, cfg config.Config) *ChatGPT {
	return &ChatGPT{
		client: client,
		repo:   repo,
		adRepo: adRepo,
		cfg:    cfg,
	}
}

func (c *ChatGPT) GetDescriptionFromDB(ctx context.Context, id uuid.UUID) (chatgpt.Description, error) {
	desc, err := c.repo.GetDescription(ctx, id)
	if err != nil {
		return chatgpt.Description{}, errors.Wrap(err, "failed to get description")
	}
	return chatgpt.Description{Color: desc.Color}, nil
}

func (c *ChatGPT) GetSame(ctx context.Context, id uuid.UUID, color utils.Color) ([]ad.RespAd, error) {
	descriptions, err := c.repo.GetDescriptions(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get descriptions")
	}

	distances := make(map[uuid.UUID]int64)
	for _, desc := range descriptions {
		descColor, err := utils.ParseColor(desc.Color)
		if err != nil {
			fmt.Printf("invalid color: %s\n", desc.Color)
			continue
		}
		dist, near := utils.Distance(color, descColor, c.cfg.Color)
		if near {
			distances[desc.ID] = dist
		}
	}

	ads, err := c.adRepo.SearchAds(ctx, ad.NewSearchParams(c.cfg.Ad), ad.SearchExtra{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to search ads")
	}

	nearAds := make([]ad.RespAd, 0)
	for _, row := range ads {
		if _, found := distances[row.Info.ID]; found && row.Info.ID != id {
			nearAds = append(nearAds, row)
		}
	}

	sort.Slice(nearAds, func(i, j int) bool {
		return distances[nearAds[i].Info.ID] < distances[nearAds[j].Info.ID]
	})

	return nearAds, nil
}

func (c *ChatGPT) DescribePhoto(ctx context.Context, id uuid.UUID, photo ad.PhotoParams, update bool) error {
	_, err := photo.Data.Seek(0, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, "failed to set zero seek offset on photo")
	}

	photoData, err := io.ReadAll(photo.Data)
	if err != nil {
		return errors.Wrap(err, "failed to read bytes from photo")
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
		return errors.Wrap(err, "SendRequest failed")
	}

	answer = strings.TrimPrefix(answer, "```")
	answer = strings.TrimPrefix(answer, "json")
	answer = strings.TrimSuffix(answer, "```")

	var description chatgpt.Description
	if err = json.Unmarshal([]byte(answer), &description); err != nil {
		fmt.Printf("[DEBUG] invalid GPT answer: %s\n", answer)
		return errors.Wrap(err, "failed to unmarshal answer into description")
	}

	now := time.Now().Local()
	desc := chatgpt.PostgresDescription{
		ID:        id,
		Color:     description.Color,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if update {
		if err = c.repo.UpdateDescription(ctx, desc); err != nil {
			return errors.Wrap(err, "failed to update description")
		}
	} else {
		if err = c.repo.CreateDescription(ctx, desc); err != nil {
			return errors.Wrap(err, "failed to create description")
		}
	}

	fmt.Printf("ANIMAL COLOR ON PHOTO: '%s'\n", description.Color)
	return nil
}

func (c *ChatGPT) DeleteDescription(ctx context.Context, id uuid.UUID) error {
	if err := c.repo.DeleteDescription(ctx, id); err != nil {
		return errors.Wrap(err, "failed to delete description")
	}
	return nil
}

func makeImageURL(photo []byte, extension string) string {
	return fmt.Sprintf(
		"data:image/%s;base64,%s",
		strings.TrimPrefix(extension, "."),
		base64.StdEncoding.EncodeToString(photo),
	)
}

package request

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
	"pet_adopter/src/chatgpt"
	"pet_adopter/src/config"
)

type ChatGPTClient struct {
	client *http.Client
	cfg    config.ChatGPTConfig
}

func NewChatGPTClient(cfg config.ChatGPTConfig) *ChatGPTClient {
	return &ChatGPTClient{
		client: http.DefaultClient,
		cfg:    cfg,
	}
}

func (c *ChatGPTClient) SendRequest(content chatgpt.Content) (string, error) {
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal content")
	}

	url := fmt.Sprintf("%s%s", c.cfg.BaseURL, c.cfg.ResponsesURL)
	requestBody := makeRequestData(c.cfg.Model, string(contentBytes))

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(requestBody))
	if err != nil {
		return "", errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("CHATGPT_API_KEY")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to send request")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response body")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("status code: %d, respBody: %s", resp.StatusCode, string(respBody))
	}

	var result response
	if err = json.Unmarshal(respBody, &result); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal response body")
	}

	if len(result.Output) == 0 {
		return "", errors.Errorf("no answer found in response, len(output)=0")
	}
	if len(result.Output[0].Content) == 0 {
		return "", errors.Errorf("no answer found in response, len(output[0].content)=0")
	}

	return result.Output[0].Content[0].Text, nil
}

func makeRequestData(model string, content string) string {
	return fmt.Sprintf(`{"model":"%s","input":[{"role":"user","content":%s}]}`, model, content)
}

type response struct {
	Output []respOutput `json:"output"`
}

type respOutput struct {
	Content []respContent `json:"content"`
}

type respContent struct {
	Text string `json:"text"`
}

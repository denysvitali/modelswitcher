package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OpenRouterModelsResponse struct {
	Data []OpenRouterModel `json:"data"`
}

type OpenRouterModel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Fetcher struct {
	client *http.Client
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (f *Fetcher) FetchModels(apiKey string) ([]OpenRouterModel, error) {
	req, err := http.NewRequest("GET", "https://openrouter.ai/api/v1/models", nil)
	if err != nil {
		return nil, err
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result OpenRouterModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return result.Data, nil
}

func (f *Fetcher) FetchOpenAPIYAML(apiKey string) ([]OpenRouterModel, error) {
	req, err := http.NewRequest("GET", "https://openrouter.ai/openapi.yaml", nil)
	if err != nil {
		return nil, err
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseModelsFromYAML(string(body)), nil
}

func parseModelsFromYAML(content string) []OpenRouterModel {
	var models []OpenRouterModel
	lines := strings.Split(content, "\n")
	var currentModel *OpenRouterModel

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") && strings.Contains(trimmed, "/") {
			id := strings.TrimPrefix(trimmed, "- ")
			id = strings.TrimSuffix(id, ":")
			id = strings.TrimSpace(id)
			if currentModel != nil {
				models = append(models, *currentModel)
			}
			currentModel = &OpenRouterModel{ID: id}
		} else if currentModel != nil {
			if strings.HasPrefix(trimmed, "name:") {
				currentModel.Name = strings.TrimPrefix(trimmed, "name:")
				currentModel.Name = strings.TrimSpace(currentModel.Name)
			} else if strings.HasPrefix(trimmed, "description:") {
				desc := strings.TrimPrefix(trimmed, "description:")
				currentModel.Description = strings.TrimSpace(desc)
			}
		}
	}
	if currentModel != nil {
		models = append(models, *currentModel)
	}

	return models
}

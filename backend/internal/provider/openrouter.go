package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const apiBaseURL = "https://openrouter.ai/api/v1"

type Model struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Provider      string `json:"provider"`
	Description   string `json:"description,omitempty"`
	ContextLength int    `json:"contextLength,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type StreamEvent struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Message string `json:"message,omitempty"`
}

type HTTPError struct {
	Status  int
	Message string
}

func (e *HTTPError) Error() string { return e.Message }

type OpenRouter struct {
	apiKey  string
	siteURL string
	appName string
	client  *http.Client

	cacheMu      sync.RWMutex
	cachedModels []Model
	cacheUntil   time.Time
}

func NewOpenRouter(apiKey, siteURL, appName string) *OpenRouter {
	return &OpenRouter{
		apiKey:  apiKey,
		siteURL: siteURL,
		appName: appName,
		client: &http.Client{
			Transport: &http.Transport{
				Proxy:                 http.ProxyFromEnvironment,
				MaxIdleConns:          50,
				IdleConnTimeout:       90 * time.Second,
				ResponseHeaderTimeout: 60 * time.Second,
			},
		},
	}
}

func (o *OpenRouter) Configured() bool { return o.apiKey != "" }

func (o *OpenRouter) Models(ctx context.Context) ([]Model, error) {
	if !o.Configured() {
		return fallbackModels(), nil
	}

	o.cacheMu.RLock()
	if time.Now().Before(o.cacheUntil) && len(o.cachedModels) > 0 {
		models := append([]Model(nil), o.cachedModels...)
		o.cacheMu.RUnlock()
		return models, nil
	}
	o.cacheMu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBaseURL+"/models?output_modalities=text", nil)
	if err != nil {
		return nil, err
	}
	o.setHeaders(req)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch OpenRouter models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, decodeHTTPError(resp)
	}

	var payload struct {
		Data []struct {
			ID            string `json:"id"`
			Name          string `json:"name"`
			Description   string `json:"description"`
			ContextLength int    `json:"context_length"`
			Architecture  struct {
				OutputModalities []string `json:"output_modalities"`
			} `json:"architecture"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode OpenRouter models: %w", err)
	}

	models := make([]Model, 0, len(payload.Data)+1)
	models = append(models, fallbackModels()...)
	for _, item := range payload.Data {
		if !strings.HasSuffix(item.ID, ":free") || !supportsText(item.Architecture.OutputModalities) {
			continue
		}
		models = append(models, Model{
			ID:            item.ID,
			Name:          strings.TrimSuffix(item.Name, " (free)"),
			Provider:      providerName(item.ID),
			Description:   item.Description,
			ContextLength: item.ContextLength,
		})
	}

	sort.SliceStable(models[1:], func(i, j int) bool {
		return strings.ToLower(models[i+1].Name) < strings.ToLower(models[j+1].Name)
	})

	o.cacheMu.Lock()
	o.cachedModels = append([]Model(nil), models...)
	o.cacheUntil = time.Now().Add(10 * time.Minute)
	o.cacheMu.Unlock()

	return models, nil
}

func (o *OpenRouter) StreamChat(ctx context.Context, request ChatRequest, emit func(StreamEvent) error) error {
	if !o.Configured() {
		return &HTTPError{Status: http.StatusServiceUnavailable, Message: "Add OPENROUTER_API_KEY to .env and restart the server."}
	}

	payload := struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
		Stream   bool      `json:"stream"`
	}{Model: request.Model, Messages: request.Messages, Stream: true}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return err
	}
	o.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("connect to OpenRouter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeHTTPError(resp)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 2<<20)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") || !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			return emit(StreamEvent{Type: "done"})
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if chunk.Error != nil {
			return &HTTPError{Status: http.StatusBadGateway, Message: chunk.Error.Message}
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			if err := emit(StreamEvent{Type: "delta", Content: chunk.Choices[0].Delta.Content}); err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("read OpenRouter stream: %w", err)
	}
	return emit(StreamEvent{Type: "done"})
}

func (o *OpenRouter) setHeaders(req *http.Request) {
	if o.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.apiKey)
	}
	if o.siteURL != "" {
		req.Header.Set("HTTP-Referer", o.siteURL)
	}
	if o.appName != "" {
		req.Header.Set("X-Title", o.appName)
	}
}

func decodeHTTPError(resp *http.Response) error {
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	message := strings.TrimSpace(string(data))
	var payload struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(data, &payload) == nil && payload.Error.Message != "" {
		message = payload.Error.Message
	}
	if message == "" {
		message = http.StatusText(resp.StatusCode)
	}
	return &HTTPError{Status: resp.StatusCode, Message: message}
}

func fallbackModels() []Model {
	return []Model{{
		ID:          "openrouter/free",
		Name:        "Auto (free)",
		Provider:    "OpenRouter",
		Description: "Automatically picks an available free model for this message.",
	}}
}

func providerName(id string) string {
	author, _, ok := strings.Cut(id, "/")
	if !ok {
		return "OpenRouter"
	}
	words := strings.FieldsFunc(author, func(r rune) bool { return r == '-' || r == '_' })
	for i := range words {
		if len(words[i]) > 0 {
			words[i] = strings.ToUpper(words[i][:1]) + words[i][1:]
		}
	}
	return strings.Join(words, " ")
}

func supportsText(modalities []string) bool {
	if len(modalities) == 0 {
		return true
	}
	for _, modality := range modalities {
		if modality == "text" {
			return true
		}
	}
	return false
}

package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// TranslationResult is the parsed response from a translation API.
type TranslationResult struct {
	TranslatedText string `json:"translatedText"`
	DetectedLanguage struct {
		Language string  `json:"language"`
		Confidence float64 `json:"confidence"`
	} `json:"detectedLanguage"`
}

// TranslateClient wraps a free LibreTranslate-compatible endpoint.
type TranslateClient struct {
	endpoint string
	apiKey   string
	http     *http.Client
}

// NewTranslateClient returns a client for the given endpoint. If endpoint
// is empty, Translate returns ErrNoClient and the bot silently skips
// translation (so the demo still works without network).
func NewTranslateClient(endpoint, apiKey string) *TranslateClient {
	return &TranslateClient{
		endpoint: endpoint,
		apiKey:   apiKey,
		http:     &http.Client{Timeout: 5 * time.Second},
	}
}

// ErrNoClient is returned by Translate when no endpoint is configured.
var ErrNoClient = fmt.Errorf("no translation client configured")

// Translate converts text from src to dst. It uses "auto" detection if
// src is empty.
func (c *TranslateClient) Translate(ctx context.Context, text, src, dst string) (string, error) {
	if c == nil || c.endpoint == "" {
		return "", ErrNoClient
	}
	if src == dst {
		return text, nil
	}
	if src == "" {
		src = "auto"
	}

	body, _ := json.Marshal(map[string]string{
		"q":      text,
		"source": src,
		"target": dst,
		"format": "text",
	})
	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("api-key", c.apiKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("translate: %s: %s", resp.Status, string(raw))
	}

	var out TranslationResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.TranslatedText == "" {
		return "", fmt.Errorf("translate: empty response")
	}
	return out.TranslatedText, nil
}

// LogTranslateErr logs a translation error at debug verbosity. The bot
// treats translation failures as non-fatal — we just send the original.
func LogTranslateErr(err error) {
	if err == nil || err == ErrNoClient {
		return
	}
	log.Printf("[translate] %v", err)
}

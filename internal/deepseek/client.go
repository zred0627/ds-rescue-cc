package deepseek

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const DefaultEndpoint = "https://api.deepseek.com/chat/completions"

type Client struct {
	Endpoint    string
	APIKey      string
	Timeout     int // seconds
	MaxRetries  int // default 4
	BaseBackoff int // base ms; default 1000
}

func (c *Client) Call(req ChatRequest) (*ChatResponse, error) {
	if c.Endpoint == "" {
		c.Endpoint = DefaultEndpoint
	}
	if c.Timeout == 0 {
		c.Timeout = 90
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = 4
	}
	if c.BaseBackoff == 0 {
		c.BaseBackoff = 1000
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpClient := &http.Client{Timeout: time.Duration(c.Timeout) * time.Second}

	var lastErr error
	for attempt := 0; attempt < c.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(c.BaseBackoff*(1<<(attempt-1))) * time.Millisecond
			time.Sleep(backoff)
		}
		httpReq, err := http.NewRequest("POST", c.Endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

		httpResp, err := httpClient.Do(httpReq)
		if err != nil {
			lastErr = err
			continue
		}
		respBody, _ := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()

		if httpResp.StatusCode == 429 || httpResp.StatusCode >= 500 {
			lastErr = fmt.Errorf("status %d: %s", httpResp.StatusCode, respBody)
			continue
		}
		if httpResp.StatusCode != 200 {
			return nil, fmt.Errorf("non-retriable status %d: %s", httpResp.StatusCode, respBody)
		}

		var resp ChatResponse
		if err := json.Unmarshal(respBody, &resp); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		return &resp, nil
	}
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// ResolveModel maps user-friendly aliases to actual DeepSeek model IDs
func ResolveModel(alias string) string {
	switch alias {
	case "pro", "v4-pro", "":
		return "deepseek-chat"
	case "flash", "v4-flash":
		return "deepseek-chat-flash"
	case "reasoner", "r1":
		return "deepseek-reasoner"
	default:
		return alias // pass-through full model name
	}
}

package deepseek

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClient_CallSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Errorf("missing/bad Authorization header")
		}
		resp := ChatResponse{
			Choices: []Choice{{Message: Message{Role: "assistant", Content: "ok"}}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := &Client{Endpoint: srv.URL, APIKey: "sk-test", Timeout: 5}
	resp, err := c.Call(ChatRequest{Model: "deepseek-chat", Messages: []Message{{Role: "user", Content: "hi"}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Choices[0].Message.Content != "ok" {
		t.Errorf("expected content 'ok', got: %q", resp.Choices[0].Message.Content)
	}
}

func TestClient_Retry429ThenSuccess(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		json.NewEncoder(w).Encode(ChatResponse{Choices: []Choice{{Message: Message{Role: "assistant", Content: "recovered"}}}})
	}))
	defer srv.Close()

	c := &Client{Endpoint: srv.URL, APIKey: "sk-test", Timeout: 5, MaxRetries: 3, BaseBackoff: 1}
	resp, err := c.Call(ChatRequest{Model: "deepseek-chat"})
	if err != nil {
		t.Fatalf("expected recovery after 429, got error: %v", err)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got: %d", attempts)
	}
	if resp.Choices[0].Message.Content != "recovered" {
		t.Errorf("expected recovered content, got: %q", resp.Choices[0].Message.Content)
	}
}

func TestClient_Server500Backoff(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(ChatResponse{Choices: []Choice{{Message: Message{Role: "assistant", Content: "ok"}}}})
	}))
	defer srv.Close()

	c := &Client{Endpoint: srv.URL, APIKey: "sk-test", Timeout: 5, MaxRetries: 4, BaseBackoff: 1}
	_, err := c.Call(ChatRequest{Model: "deepseek-chat"})
	if err != nil {
		t.Fatalf("expected recovery after 2x 500, got: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClient_Auth401NoRetry(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := &Client{Endpoint: srv.URL, APIKey: "sk-bad", Timeout: 5, MaxRetries: 4, BaseBackoff: 1}
	_, err := c.Call(ChatRequest{Model: "deepseek-chat"})
	if err == nil {
		t.Fatal("expected auth error, got nil")
	}
	if attempts != 1 {
		t.Errorf("401 must NOT retry (got %d attempts)", attempts)
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected '401' in error, got: %v", err)
	}
}

func TestClient_TimeoutTriggersError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer srv.Close()

	c := &Client{Endpoint: srv.URL, APIKey: "sk-test", Timeout: 1, MaxRetries: 2, BaseBackoff: 1}
	_, err := c.Call(ChatRequest{Model: "deepseek-chat"})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

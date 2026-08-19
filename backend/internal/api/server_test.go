package api

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"allai/backend/internal/provider"
)

func testServer() http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewServer(provider.NewOpenRouter("", "", "Allai test"), "http://localhost:5173", logger).Handler()
}

func TestHealthShowsConfigurationState(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	recorder := httptest.NewRecorder()
	testServer().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"configured":false`) {
		t.Fatalf("expected unconfigured state, got %s", recorder.Body.String())
	}
}

func TestModelsReturnsFreeRouterWithoutKey(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/models", nil)
	recorder := httptest.NewRecorder()
	testServer().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "openrouter/free") {
		t.Fatalf("expected free router fallback, got %s", recorder.Body.String())
	}
}

func TestChatRejectsPaidModel(t *testing.T) {
	body := `{"model":"openai/gpt-4","messages":[{"role":"user","content":"hello"}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/chat/stream", strings.NewReader(body))
	recorder := httptest.NewRecorder()
	testServer().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "only allows free") {
		t.Fatalf("expected free-only validation error, got %s", recorder.Body.String())
	}
}

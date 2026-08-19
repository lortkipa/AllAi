package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"allai/backend/internal/provider"
)

type Server struct {
	openRouter    *provider.OpenRouter
	allowedOrigin string
	logger        *slog.Logger
}

func NewServer(openRouter *provider.OpenRouter, allowedOrigin string, logger *slog.Logger) *Server {
	return &Server{openRouter: openRouter, allowedOrigin: allowedOrigin, logger: logger}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("GET /api/models", s.models)
	mux.HandleFunc("POST /api/chat/stream", s.chat)
	return s.cors(s.requestLog(mux))
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":     "ok",
		"configured": s.openRouter.Configured(),
	})
}

func (s *Server) models(w http.ResponseWriter, r *http.Request) {
	models, err := s.openRouter.Models(r.Context())
	if err != nil {
		s.logger.Error("list models", "error", err)
		writeError(w, http.StatusBadGateway, "Could not load OpenRouter models. Try again shortly.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"models":     models,
		"configured": s.openRouter.Configured(),
	})
}

func (s *Server) chat(w http.ResponseWriter, r *http.Request) {
	var request provider.ChatRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "The chat request is not valid.")
		return
	}

	if message := validateChat(request); message != "" {
		writeError(w, http.StatusBadRequest, message)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "Streaming is not supported by this server.")
		return
	}

	started := false
	emit := func(event provider.StreamEvent) error {
		if !started {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache, no-transform")
			w.Header().Set("Connection", "keep-alive")
			w.Header().Set("X-Accel-Buffering", "no")
			w.WriteHeader(http.StatusOK)
			started = true
		}
		data, err := json.Marshal(event)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	err := s.openRouter.StreamChat(r.Context(), request, emit)
	if err == nil || errors.Is(err, r.Context().Err()) {
		return
	}

	s.logger.Error("stream chat", "model", request.Model, "error", err)
	if started {
		_ = emit(provider.StreamEvent{Type: "error", Message: publicError(err)})
		return
	}

	status := http.StatusBadGateway
	var httpErr *provider.HTTPError
	if errors.As(err, &httpErr) && httpErr.Status >= 400 && httpErr.Status < 600 {
		status = httpErr.Status
	}
	writeError(w, status, publicError(err))
}

func validateChat(request provider.ChatRequest) string {
	if request.Model == "" {
		return "Choose a model before sending a message."
	}
	if request.Model != "openrouter/free" && !strings.HasSuffix(request.Model, ":free") {
		return "This version of Allai only allows free OpenRouter models."
	}
	if len(request.Messages) == 0 {
		return "Write a message before sending."
	}
	if len(request.Messages) > 100 {
		return "This conversation is too long. Start a new chat and try again."
	}
	for _, message := range request.Messages {
		if message.Role != "user" && message.Role != "assistant" && message.Role != "system" {
			return "The conversation contains an unsupported message role."
		}
		if strings.TrimSpace(message.Content) == "" {
			return "The conversation contains an empty message."
		}
		if len(message.Content) > 100_000 {
			return "One of the messages is too long."
		}
	}
	return ""
}

func publicError(err error) string {
	var httpErr *provider.HTTPError
	if errors.As(err, &httpErr) && httpErr.Message != "" {
		return httpErr.Message
	}
	return "The model could not complete this response. Try again."
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin == s.allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info("request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

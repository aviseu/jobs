package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aviseu/jobs/internal/app/domain/channel"
	"github.com/go-chi/chi/v5"
)

type IntegrationHandler struct {
	s   *channel.Service
	log *slog.Logger
}

func NewIntegrationHandler(s *channel.Service, log *slog.Logger) *IntegrationHandler {
	return &IntegrationHandler{
		s:   s,
		log: log,
	}
}

func (h *IntegrationHandler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", h.ListIntegrations)

	return r
}

func (h *IntegrationHandler) ListIntegrations(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := NewListIntegrationsResponse(h.s.Integrations())
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.handleError(w, fmt.Errorf("failed to encode response: %w", err))
	}
}

func (h *IntegrationHandler) handleFail(w http.ResponseWriter, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	resp := NewErrorResponse(err)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.log.Error(err.Error(), slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *IntegrationHandler) handleError(w http.ResponseWriter, err error) {
	h.log.Error(err.Error(), slog.Any("error", err))

	h.handleFail(w, errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError)
}

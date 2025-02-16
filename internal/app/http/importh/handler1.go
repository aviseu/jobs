package importh

import (
	"encoding/json"
	"fmt"
	"github.com/aviseu/jobs-backoffice/internal/app/domain"
	"github.com/aviseu/jobs-protobuf/build/gen/commands/jobs"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	ia  *domain.ImportAction
	log *slog.Logger
}

func NewHandler(ia *domain.ImportAction, log *slog.Logger) *Handler {
	return &Handler{
		ia:  ia,
		log: log,
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Post("/", h.Import)

	return r
}

type pubSubMessage struct {
	Message struct {
		Data []byte `json:"data,omitempty"`
		ID   string `json:"id"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

func (h *Handler) Import(w http.ResponseWriter, r *http.Request) {
	h.log.Info("received message!")

	var msg pubSubMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		h.log.Error(fmt.Errorf("failed to json decode request body: %w", err).Error())
		http.Error(w, "skipped message", http.StatusOK) // 200 will ack message
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			h.log.Error(fmt.Errorf("failed to close body: %w", err).Error())
		}
	}(r.Body)

	var data jobs.ExecuteImportChannel
	if err := proto.Unmarshal(msg.Message.Data, &data); err != nil {
		h.log.Error(fmt.Errorf("failed to unmarshal pubsub message: %w", err).Error())
		http.Error(w, "skipped message", http.StatusOK) // 200 will ack message
		return
	}

	importID, err := uuid.Parse(data.ImportId)
	if err != nil {
		h.log.Error(fmt.Errorf("failed to convert import id %s to uuid: %w", data.ImportId, err).Error())
		http.Error(w, "skipped message", http.StatusOK) // 200 will ack message
		return
	}

	h.log.Info("processing import " + importID.String())
	if err := h.ia.Execute(r.Context(), importID); err != nil {
		h.log.Error(fmt.Errorf("failed to execute import %s: %w", importID, err).Error())
		http.Error(w, "skipped message", http.StatusOK) // 200 will ack message
		return
	}
	h.log.Info("completed import " + importID.String())

	w.WriteHeader(http.StatusOK)
}

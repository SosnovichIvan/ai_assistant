package handler

import (
	"encoding/json"
	"net/http"

	"ai_assistant/internal/model"
	"ai_assistant/internal/service"

	"github.com/go-playground/validator/v10"
)

// TelegramHandler handles Telegram configuration HTTP requests.
type TelegramHandler struct {
	telegramSvc *service.TelegramService
	validator   *validator.Validate
}

// NewTelegramHandler creates a new Telegram handler.
func NewTelegramHandler(telegramSvc *service.TelegramService) *TelegramHandler {
	return &TelegramHandler{
		telegramSvc: telegramSvc,
		validator:   validator.New(),
	}
}

// SetMainUser handles POST /api/v1/telegram/main-user.
func (h *TelegramHandler) SetMainUser(w http.ResponseWriter, r *http.Request) {
	var req model.SetMainUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.telegramSvc.SetMainUser(r.Context(), req.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// GetMainUser handles GET /api/v1/telegram/main-user.
func (h *TelegramHandler) GetMainUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.telegramSvc.GetMainUser(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "main user not set", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

// LinkChat handles POST /api/v1/telegram/link.
func (h *TelegramHandler) LinkChat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LinkedUserID int64 `json:"linked_user_id" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	mainUser, err := h.telegramSvc.GetMainUser(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if mainUser == nil {
		http.Error(w, "main user not set", http.StatusBadRequest)
		return
	}

	link, err := h.telegramSvc.LinkChat(r.Context(), mainUser.ID, req.LinkedUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(link)
}

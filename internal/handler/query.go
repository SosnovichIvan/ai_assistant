package handler

import (
	"encoding/json"
	"net/http"

	"ai_assistant/internal/service"

	"github.com/go-playground/validator/v10"
)

// QueryHandler handles query HTTP requests.
type QueryHandler struct {
	querySvc  *service.QueryService
	validator *validator.Validate
}

// NewQueryHandler creates a new query handler.
func NewQueryHandler(querySvc *service.QueryService) *QueryHandler {
	return &QueryHandler{
		querySvc:  querySvc,
		validator: validator.New(),
	}
}

// Query handles POST /api/v1/query.
func (h *QueryHandler) Query(w http.ResponseWriter, r *http.Request) {
	var req service.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.querySvc.Query(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

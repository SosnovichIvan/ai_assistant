package handler

import (
	"encoding/json"
	"net/http"

	"ai_assistant/internal/model"
	"ai_assistant/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// KnowledgeHandler handles knowledge base HTTP requests.
type KnowledgeHandler struct {
	knowledgeSvc *service.KnowledgeService
	validator    *validator.Validate
}

// NewKnowledgeHandler creates a new knowledge handler.
func NewKnowledgeHandler(knowledgeSvc *service.KnowledgeService) *KnowledgeHandler {
	return &KnowledgeHandler{
		knowledgeSvc: knowledgeSvc,
		validator:    validator.New(),
	}
}

// Create handles POST /api/v1/documents.
func (h *KnowledgeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.knowledgeSvc.CreateDocument(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Get handles GET /api/v1/documents/:id.
func (h *KnowledgeHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	doc, err := h.knowledgeSvc.GetDocument(r.Context(), parseUUID(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if doc == nil {
		http.Error(w, "document not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(doc)
}

// List handles GET /api/v1/documents.
func (h *KnowledgeHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)

	docs, err := h.knowledgeSvc.ListDocuments(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"documents": docs,
		"limit":     limit,
		"offset":    offset,
	})
}

// Update handles PATCH /api/v1/documents/:id.
func (h *KnowledgeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req model.UpdateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.knowledgeSvc.UpdateDocument(r.Context(), parseUUID(id), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

// Delete handles DELETE /api/v1/documents/:id.
func (h *KnowledgeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.knowledgeSvc.DeleteDocument(r.Context(), parseUUID(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

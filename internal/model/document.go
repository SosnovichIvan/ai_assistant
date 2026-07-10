package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Document represents a knowledge base document.
type Document struct {
	ID        uuid.UUID       `db:"id" json:"id"`
	Title     string          `db:"title" json:"title"`
	Metadata  json.RawMessage `db:"metadata" json:"metadata"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt time.Time       `db:"updated_at" json:"updated_at"`
}

// DocumentChunk represents a chunk of a document with embedding vector.
type DocumentChunk struct {
	ID         uuid.UUID `db:"id" json:"id"`
	DocumentID uuid.UUID `db:"document_id" json:"document_id"`
	Content    string    `db:"content" json:"content"`
	Embedding  []float32 `db:"-" json:"-"`
	Position   int       `db:"position" json:"position"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// CreateDocumentRequest is the request body for creating a document.
type CreateDocumentRequest struct {
	Title    string          `json:"title" validate:"required,min=1,max=500"`
	Content  string          `json:"content" validate:"required,min=1"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// UpdateDocumentRequest is the request body for updating a document.
type UpdateDocumentRequest struct {
	Content string `json:"content,omitempty"`
	Append  bool   `json:"append,omitempty"`
}

// DocumentResponse is the response for document operations.
type DocumentResponse struct {
	ID         uuid.UUID `json:"id"`
	Title      string    `json:"title"`
	ChunksCount int      `json:"chunks_count"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
}

// ChunkWithScore represents a chunk with its similarity score.
type ChunkWithScore struct {
	Chunk     DocumentChunk `json:"chunk"`
	Score     float64       `json:"score"`
	Text      string        `json:"text,omitempty"`
}

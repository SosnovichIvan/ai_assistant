package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"ai_assistant/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DocumentRepository handles document persistence.
type DocumentRepository struct {
	db *pgxpool.Pool
}

// NewDocumentRepository creates a new document repository.
func NewDocumentRepository(db *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// Create inserts a new document.
func (r *DocumentRepository) Create(ctx context.Context, doc *model.Document) error {
	query := `
		INSERT INTO documents (id, title, metadata)
		VALUES ($1, $2, $3)
		RETURNING created_at, updated_at`

	metadata := doc.Metadata
	if metadata == nil {
		metadata = json.RawMessage("{}")
	}

	return r.db.QueryRow(ctx, query, doc.ID, doc.Title, metadata).
		Scan(&doc.CreatedAt, &doc.UpdatedAt)
}

// GetByID retrieves a document by ID.
func (r *DocumentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Document, error) {
	var doc model.Document
	query := `SELECT id, title, metadata, created_at, updated_at FROM documents WHERE id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.Title, &doc.Metadata, &doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		return nil, nil
	}
	return &doc, nil
}

// List retrieves all documents with pagination.
func (r *DocumentRepository) List(ctx context.Context, limit, offset int) ([]model.Document, error) {
	var docs []model.Document
	query := `
		SELECT id, title, metadata, created_at, updated_at 
		FROM documents 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var doc model.Document
		if err := rows.Scan(&doc.ID, &doc.Title, &doc.Metadata, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}

	return docs, rows.Err()
}

// Delete removes a document by ID.
func (r *DocumentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM documents WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	rows := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("document not found")
	}
	return nil
}

// Update updates document metadata.
func (r *DocumentRepository) Update(ctx context.Context, doc *model.Document) error {
	query := `
		UPDATE documents 
		SET title = $2, metadata = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	return r.db.QueryRow(ctx, query, doc.ID, doc.Title, doc.Metadata).
		Scan(&doc.UpdatedAt)
}

// Count returns the total number of documents.
func (r *DocumentRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM documents").Scan(&count)
	return count, err
}

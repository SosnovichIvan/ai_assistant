package repository

import (
	"context"
	"fmt"
	"strings"

	"ai_assistant/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

// VectorRepository handles document chunks with embeddings.
type VectorRepository struct {
	db *pgxpool.Pool
}

// NewVectorRepository creates a new vector repository.
func NewVectorRepository(db *pgxpool.Pool) *VectorRepository {
	return &VectorRepository{db: db}
}

// EmbeddingData represents embedding data for database operations.
type EmbeddingData struct {
	ID         uuid.UUID
	DocumentID uuid.UUID
	Content    string
	Embedding  []float32
	Position   int
}

// Upsert inserts or updates chunks for a document.
func (r *VectorRepository) Upsert(ctx context.Context, chunks []EmbeddingData) error {
	if len(chunks) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	docID := chunks[0].DocumentID
	_, err = tx.Exec(ctx, "DELETE FROM document_chunks WHERE document_id = $1", docID)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO document_chunks (id, document_id, content, embedding, position)
		VALUES ($1, $2, $3, $4, $5)`

	for _, chunk := range chunks {
		_, err = tx.Exec(ctx, query,
			chunk.ID, chunk.DocumentID, chunk.Content, pq.Array(chunk.Embedding), chunk.Position)
		if err != nil {
			return fmt.Errorf("insert chunk: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// Search performs vector similarity search.
func (r *VectorRepository) Search(ctx context.Context, query []float32, topK int, threshold float64) ([]model.ChunkWithScore, error) {
	var results []model.ChunkWithScore

	queryStr := `
		SELECT id, document_id, content, position, created_at,
			   1 - (embedding <=> $1) as score
		FROM document_chunks
		WHERE 1 - (embedding <=> $1) > $2
		ORDER BY embedding <=> $1
		LIMIT $3`

	rows, err := r.db.Query(ctx, queryStr, pq.Array(query), threshold, topK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var chunk model.DocumentChunk
		var score float64

		err := rows.Scan(
			&chunk.ID, &chunk.DocumentID, &chunk.Content,
			&chunk.Position, &chunk.CreatedAt, &score)
		if err != nil {
			return nil, err
		}

		results = append(results, model.ChunkWithScore{
			Chunk: chunk,
			Score: score,
			Text:  truncateText(chunk.Content, 200),
		})
	}

	return results, rows.Err()
}

// GetByDocumentID retrieves all chunks for a document.
func (r *VectorRepository) GetByDocumentID(ctx context.Context, docID uuid.UUID) ([]model.DocumentChunk, error) {
	var chunks []model.DocumentChunk
	query := `
		SELECT id, document_id, content, position, created_at
		FROM document_chunks
		WHERE document_id = $1
		ORDER BY position`

	rows, err := r.db.Query(ctx, query, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var chunk model.DocumentChunk
		if err := rows.Scan(&chunk.ID, &chunk.DocumentID, &chunk.Content, &chunk.Position, &chunk.CreatedAt); err != nil {
			return nil, err
		}
		chunks = append(chunks, chunk)
	}

	return chunks, rows.Err()
}

// CountByDocumentID returns the number of chunks for a document.
func (r *VectorRepository) CountByDocumentID(ctx context.Context, docID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM document_chunks WHERE document_id = $1", docID).Scan(&count)
	return count, err
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

func StringSliceToFloat32(strs []string) ([]float32, error) {
	result := make([]float32, len(strs))
	for i, s := range strs {
		var f float64
		s = strings.TrimSpace(s)
		if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
			return nil, err
		}
		result[i] = float32(f)
	}
	return result, nil
}

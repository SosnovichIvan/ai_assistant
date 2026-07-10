package service

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"ai_assistant/internal/model"
	"ai_assistant/internal/repository"

	"github.com/google/uuid"
)

// KnowledgeService handles document processing and embeddings.
type KnowledgeService struct {
	docRepo     *repository.DocumentRepository
	vectorRepo  *repository.VectorRepository
	embeddingSvc EmbeddingService
}

// NewKnowledgeService creates a new knowledge service.
func NewKnowledgeService(
	docRepo *repository.DocumentRepository,
	vectorRepo *repository.VectorRepository,
	embeddingSvc EmbeddingService,
) *KnowledgeService {
	return &KnowledgeService{
		docRepo:    docRepo,
		vectorRepo: vectorRepo,
		embeddingSvc: embeddingSvc,
	}
}

// CreateDocument creates a document and generates embeddings for its chunks.
func (s *KnowledgeService) CreateDocument(ctx context.Context, req *model.CreateDocumentRequest) (*model.DocumentResponse, error) {
	doc := &model.Document{
		ID:       uuid.New(),
		Title:    req.Title,
		Metadata: req.Metadata,
	}

	if err := s.docRepo.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("create document: %w", err)
	}

	chunks, err := s.chunkText(req.Content)
	if err != nil {
		return nil, fmt.Errorf("chunk text: %w", err)
	}

	embeddings, err := s.embeddingSvc.EmbedTexts(ctx, chunks)
	if err != nil {
		return nil, fmt.Errorf("embed texts: %w", err)
	}

	chunkData := make([]repository.EmbeddingData, len(chunks))
	for i, chunk := range chunks {
		chunkData[i] = repository.EmbeddingData{
			ID:         uuid.New(),
			DocumentID: doc.ID,
			Content:    chunk,
			Embedding:  embeddings[i],
			Position:   i,
		}
	}

	if err := s.vectorRepo.Upsert(ctx, chunkData); err != nil {
		return nil, fmt.Errorf("save embeddings: %w", err)
	}

	return &model.DocumentResponse{
		ID:          doc.ID,
		Title:       doc.Title,
		ChunksCount: len(chunks),
		CreatedAt:   doc.CreatedAt,
	}, nil
}

// UpdateDocument appends content to an existing document.
func (s *KnowledgeService) UpdateDocument(ctx context.Context, id uuid.UUID, req *model.UpdateDocumentRequest) (*model.DocumentResponse, error) {
	doc, err := s.docRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("document not found")
	}

	existingCount, err := s.vectorRepo.CountByDocumentID(ctx, id)
	if err != nil {
		return nil, err
	}

	newChunks, err := s.chunkText(req.Content)
	if err != nil {
		return nil, fmt.Errorf("chunk text: %w", err)
	}

	embeddings, err := s.embeddingSvc.EmbedTexts(ctx, newChunks)
	if err != nil {
		return nil, fmt.Errorf("embed texts: %w", err)
	}

	chunkData := make([]repository.EmbeddingData, len(newChunks))
	for i, chunk := range newChunks {
		chunkData[i] = repository.EmbeddingData{
			ID:         uuid.New(),
			DocumentID: id,
			Content:    chunk,
			Embedding:  embeddings[i],
			Position:   existingCount + i,
		}
	}

	if err := s.vectorRepo.Upsert(ctx, chunkData); err != nil {
		return nil, fmt.Errorf("save embeddings: %w", err)
	}

	newCount, _ := s.vectorRepo.CountByDocumentID(ctx, id)
	if err := s.docRepo.Update(ctx, doc); err != nil {
		return nil, err
	}

	return &model.DocumentResponse{
		ID:          doc.ID,
		Title:       doc.Title,
		ChunksCount: newCount,
		UpdatedAt:   doc.UpdatedAt,
	}, nil
}

// DeleteDocument removes a document and its chunks.
func (s *KnowledgeService) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	return s.docRepo.Delete(ctx, id)
}

// GetDocument retrieves a document by ID.
func (s *KnowledgeService) GetDocument(ctx context.Context, id uuid.UUID) (*model.Document, error) {
	return s.docRepo.GetByID(ctx, id)
}

// ListDocuments retrieves all documents.
func (s *KnowledgeService) ListDocuments(ctx context.Context, limit, offset int) ([]model.Document, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.docRepo.List(ctx, limit, offset)
}

func (s *KnowledgeService) chunkText(text string) ([]string, error) {
	paragraphs := strings.Split(text, "\n\n")
	var chunks []string
	var currentChunk strings.Builder
	maxChunkSize := 500

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		if len(para) > maxChunkSize {
			sentences := splitSentences(para)
			for _, sentence := range sentences {
				if currentChunk.Len()+len(sentence) > maxChunkSize && currentChunk.Len() > 0 {
					chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
					currentChunk.Reset()
				}
				currentChunk.WriteString(sentence)
				currentChunk.WriteString(" ")
			}
		} else if currentChunk.Len()+len(para) > maxChunkSize && currentChunk.Len() > 0 {
			chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
			currentChunk.Reset()
			currentChunk.WriteString(para)
		} else {
			if currentChunk.Len() > 0 {
				currentChunk.WriteString("\n\n")
			}
			currentChunk.WriteString(para)
		}
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
	}

	if len(chunks) == 0 {
		chunks = []string{text}
	}

	return chunks, nil
}

func splitSentences(text string) []string {
	var sentences []string
	var current strings.Builder
	runes := []rune(text)

	for i := 0; i < len(runes); i++ {
		current.WriteRune(runes[i])
		if isSentenceEnd(runes, i) {
			sentences = append(sentences, strings.TrimSpace(current.String()))
			current.Reset()
		}
	}

	if current.Len() > 0 {
		sentences = append(sentences, strings.TrimSpace(current.String()))
	}

	return sentences
}

func isSentenceEnd(runes []rune, i int) bool {
	if i >= len(runes)-1 {
		return false
	}

	r := runes[i]
	next := runes[i+1]

	if r == '.' || r == '!' || r == '?' {
		return unicode.IsUpper(next) || unicode.IsSpace(next)
	}

	return false
}

// EmbeddingService interface for embedding generation.
type EmbeddingService interface {
	EmbedTexts(ctx context.Context, texts []string) ([][]float32, error)
}

package service

import (
	"context"
	"fmt"

	"ai_assistant/internal/model"
	"ai_assistant/internal/repository"
)

// QueryService handles semantic search queries.
type QueryService struct {
	vectorRepo  *repository.VectorRepository
	embeddingSvc EmbeddingService
}

// NewQueryService creates a new query service.
func NewQueryService(
	vectorRepo *repository.VectorRepository,
	embeddingSvc EmbeddingService,
) *QueryService {
	return &QueryService{
		vectorRepo:   vectorRepo,
		embeddingSvc: embeddingSvc,
	}
}

// QueryRequest is the request for a semantic query.
type QueryRequest struct {
	Question  string  `json:"question" validate:"required,min=1"`
	TopK      int     `json:"top_k"`
	Threshold float64 `json:"threshold"`
}

// QueryResponse is the response for a semantic query.
type QueryResponse struct {
	Answer  string                `json:"answer,omitempty"`
	Sources []model.ChunkWithScore `json:"sources"`
	Found   bool                  `json:"found"`
	Message string                `json:"message,omitempty"`
}

// Query performs semantic search and returns relevant chunks.
func (s *QueryService) Query(ctx context.Context, req *QueryRequest) (*QueryResponse, error) {
	topK := req.TopK
	if topK <= 0 {
		topK = 5
	}

	threshold := req.Threshold
	if threshold <= 0 {
		threshold = 0.7
	}

	embeddings, err := s.embeddingSvc.EmbedTexts(ctx, []string{req.Question})
	if err != nil {
		return nil, fmt.Errorf("embed question: %w", err)
	}

	results, err := s.vectorRepo.Search(ctx, embeddings[0], topK, threshold)
	if err != nil {
		return nil, fmt.Errorf("search vectors: %w", err)
	}

	if len(results) == 0 {
		return &QueryResponse{
			Answer:  "",
			Sources: []model.ChunkWithScore{},
			Found:   false,
			Message: "Не удалось найти релевантную информацию",
		}, nil
	}

	answer := s.buildAnswer(results)

	return &QueryResponse{
		Answer:  answer,
		Sources: results,
		Found:   true,
	}, nil
}

func (s *QueryService) buildAnswer(results []model.ChunkWithScore) string {
	if len(results) == 0 {
		return ""
	}

	primary := results[0].Text

	var additional []string
	for i := 1; i < len(results) && i < 3; i++ {
		additional = append(additional, results[i].Text)
	}

	if len(additional) > 0 {
		return primary + "\n\nДополнительная информация:\n" + joinTexts(additional)
	}

	return primary
}

func joinTexts(texts []string) string {
	result := ""
	for i, text := range texts {
		if i > 0 {
			result += "\n---\n"
		}
		result += text
	}
	return result
}

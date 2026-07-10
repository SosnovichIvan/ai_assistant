package service

import (
	"context"
	"fmt"

	"ai_assistant/internal/model"
	"ai_assistant/internal/repository"

	"github.com/google/uuid"
)

// TelegramService handles Telegram bot operations.
type TelegramService struct {
	telegramRepo *repository.TelegramRepository
}

// NewTelegramService creates a new Telegram service.
func NewTelegramService(telegramRepo *repository.TelegramRepository) *TelegramService {
	return &TelegramService{telegramRepo: telegramRepo}
}

// RegisterUser registers or updates a Telegram user.
func (s *TelegramService) RegisterUser(ctx context.Context, user *model.TelegramUser) error {
	return s.telegramRepo.UpsertUser(ctx, user)
}

// GetUser retrieves a user by ID.
func (s *TelegramService) GetUser(ctx context.Context, userID int64) (*model.TelegramUser, error) {
	return s.telegramRepo.GetUserByID(ctx, userID)
}

// GetMainUser retrieves the main user.
func (s *TelegramService) GetMainUser(ctx context.Context) (*model.TelegramUser, error) {
	return s.telegramRepo.GetMainUser(ctx)
}

// SetMainUser sets a user as the main user.
func (s *TelegramService) SetMainUser(ctx context.Context, userID int64) error {
	return s.telegramRepo.SetMainUser(ctx, userID)
}

// LinkChat links two users for chat relay.
func (s *TelegramService) LinkChat(ctx context.Context, mainUserID, linkedUserID int64) (*model.TelegramChatLink, error) {
	mainUser, err := s.telegramRepo.GetUserByID(ctx, mainUserID)
	if err != nil {
		return nil, err
	}
	if mainUser == nil || !mainUser.IsMainUser {
		return nil, fmt.Errorf("main user not found")
	}

	link := &model.TelegramChatLink{
		ID:           uuid.New(),
		MainUserID:   mainUserID,
		LinkedUserID: linkedUserID,
		RoomID:       fmt.Sprintf("telegram_%d_%d", mainUserID, linkedUserID),
	}

	if err := s.telegramRepo.CreateChatLink(ctx, link); err != nil {
		return nil, err
	}

	return link, nil
}

// GetChatLink retrieves a chat link between two users.
func (s *TelegramService) GetChatLink(ctx context.Context, mainUserID, linkedUserID int64) (*model.TelegramChatLink, error) {
	return s.telegramRepo.GetChatLink(ctx, mainUserID, linkedUserID)
}

// GetLinkedUsers retrieves all users linked to a main user.
func (s *TelegramService) GetLinkedUsers(ctx context.Context, mainUserID int64) ([]model.TelegramUser, error) {
	return s.telegramRepo.GetLinkedUsers(ctx, mainUserID)
}

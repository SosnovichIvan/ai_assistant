package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"ai_assistant/internal/model"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot wraps the Telegram bot API.
type Bot struct {
	api      tgbotapi.BotAPI
	logger   *slog.Logger
	handlers map[string]Handler
	relaySvc RelayService
}

// Handler is a callback for bot commands.
type Handler func(update tgbotapi.Update) error

// RelayService handles message relay between users.
type RelayService interface {
	RelayMessage(ctx context.Context, fromUserID int64, content string) error
}

// NewBot creates a new Telegram bot.
func NewBot(token string, logger *slog.Logger) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	return &Bot{
		api:      *api,
		logger:   logger,
		handlers: make(map[string]Handler),
	}, nil
}

// RegisterHandler registers a command handler.
func (b *Bot) RegisterHandler(command string, handler Handler) {
	b.handlers[command] = handler
}

// SetRelayService sets the relay service.
func (b *Bot) SetRelayService(svc RelayService) {
	b.relaySvc = svc
}

// Start starts polling for updates.
func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("starting telegram bot", slog.String("bot", b.api.Self.UserName))

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.api.StopReceivingUpdates()
			return ctx.Err()

		case update := <-updates:
			if update.Message == nil {
				continue
			}

			b.handleMessage(update)
		}
	}
}

// handleMessage handles incoming messages.
func (b *Bot) handleMessage(update tgbotapi.Update) {
	msg := update.Message

	// Handle commands
	if msg.IsCommand() {
		command := strings.TrimPrefix(msg.Command(), "/")
		if handler, ok := b.handlers[command]; ok {
			if err := handler(update); err != nil {
				b.logger.Error("handler error",
					slog.String("command", command),
					slog.Any("error", err),
				)
			}
			return
		}
	}

	// Handle regular messages - relay if service is set
	if b.relaySvc != nil && msg.Chat.Type == "private" {
		if err := b.relaySvc.RelayMessage(context.Background(), msg.From.ID, msg.Text); err != nil {
			b.logger.Error("relay error", slog.Any("error", err))
		}
	}
}

// SendMessage sends a message to a user.
func (b *Bot) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.api.Send(msg)
	return err
}

// GetMe returns bot information.
func (b *Bot) GetMe() (model.TelegramUser, error) {
	user := b.api.Self
	return model.TelegramUser{
		ID:        user.ID,
		Username:  user.UserName,
		FirstName: user.FirstName,
	}, nil
}

package model

import (
	"time"

	"github.com/google/uuid"
)

// Chat represents a chat room/thread.
type Chat struct {
	ID        uuid.UUID `db:"id" json:"id"`
	RoomID    string    `db:"room_id" json:"room_id"`
	Title     string    `db:"title,omitempty" json:"title,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// TelegramUser represents a Telegram user.
type TelegramUser struct {
	ID           int64     `db:"id" json:"id"`
	Username     string    `db:"username" json:"username,omitempty"`
	FirstName    string    `db:"first_name" json:"first_name,omitempty"`
	IsMainUser   bool      `db:"is_main_user" json:"is_main_user"`
	LinkedChatID *int64    `db:"linked_chat_id" json:"linked_chat_id,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// TelegramChatLink represents a link between two Telegram users.
type TelegramChatLink struct {
	ID           uuid.UUID `db:"id" json:"id"`
	MainUserID   int64     `db:"main_user_id" json:"main_user_id"`
	LinkedUserID int64     `db:"linked_user_id" json:"linked_user_id"`
	RoomID       string    `db:"room_id" json:"room_id"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// TelegramConfigRequest is the request for setting Telegram config.
type TelegramConfigRequest struct {
	BotToken    string `json:"bot_token" validate:"required"`
	WebhookURL  string `json:"webhook_url,omitempty"`
	WebhookPort int    `json:"webhook_port,omitempty"`
}

// TelegramConfigResponse is the response for Telegram config.
type TelegramConfigResponse struct {
	BotTokenSet  bool   `json:"bot_token_set"`
	WebhookURL   string `json:"webhook_url,omitempty"`
	WebhookPort  int    `json:"webhook_port"`
}

// SetMainUserRequest is the request for setting main user.
type SetMainUserRequest struct {
	UserID int64 `json:"user_id" validate:"required"`
}

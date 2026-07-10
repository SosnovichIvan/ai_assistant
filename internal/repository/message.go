package repository

import (
	"context"

	"ai_assistant/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MessageRepository handles message persistence.
type MessageRepository struct {
	db *pgxpool.Pool
}

// NewMessageRepository creates a new message repository.
func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create inserts a new message.
func (r *MessageRepository) Create(ctx context.Context, msg *model.Message) error {
	query := `
		INSERT INTO messages (id, room_id, sender, content, message_type)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at`

	return r.db.QueryRow(ctx, query,
		msg.ID, msg.RoomID, msg.Sender, msg.Content, msg.MessageType).
		Scan(&msg.CreatedAt)
}

// GetByRoomID retrieves messages for a room with pagination.
func (r *MessageRepository) GetByRoomID(ctx context.Context, roomID string, limit, offset int) ([]model.Message, error) {
	var messages []model.Message
	query := `
		SELECT id, room_id, sender, content, message_type, created_at
		FROM messages
		WHERE room_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, roomID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var msg model.Message
		if err := rows.Scan(&msg.ID, &msg.RoomID, &msg.Sender, &msg.Content, &msg.MessageType, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// GetHistory retrieves recent messages for a room.
func (r *MessageRepository) GetHistory(ctx context.Context, roomID string, limit int) ([]model.Message, error) {
	var messages []model.Message
	query := `
		SELECT id, room_id, sender, content, message_type, created_at
		FROM messages
		WHERE room_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := r.db.Query(ctx, query, roomID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var msg model.Message
		if err := rows.Scan(&msg.ID, &msg.RoomID, &msg.Sender, &msg.Content, &msg.MessageType, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, rows.Err()
}

// DeleteByRoomID removes all messages from a room.
func (r *MessageRepository) DeleteByRoomID(ctx context.Context, roomID string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM messages WHERE room_id = $1", roomID)
	return err
}

package repository

import (
	"context"

	"ai_assistant/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TelegramRepository handles Telegram user persistence.
type TelegramRepository struct {
	db *pgxpool.Pool
}

// NewTelegramRepository creates a new Telegram repository.
func NewTelegramRepository(db *pgxpool.Pool) *TelegramRepository {
	return &TelegramRepository{db: db}
}

// UpsertUser inserts or updates a Telegram user.
func (r *TelegramRepository) UpsertUser(ctx context.Context, user *model.TelegramUser) error {
	query := `
		INSERT INTO telegram_users (id, username, first_name, is_main_user, linked_chat_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			username = EXCLUDED.username,
			first_name = EXCLUDED.first_name,
			updated_at = NOW()
		RETURNING created_at, updated_at`

	return r.db.QueryRow(ctx, query,
		user.ID, user.Username, user.FirstName, user.IsMainUser, user.LinkedChatID).
		Scan(&user.CreatedAt, &user.UpdatedAt)
}

// GetUserByID retrieves a user by Telegram ID.
func (r *TelegramRepository) GetUserByID(ctx context.Context, id int64) (*model.TelegramUser, error) {
	var user model.TelegramUser
	query := `
		SELECT id, username, first_name, is_main_user, linked_chat_id, created_at, updated_at
		FROM telegram_users WHERE id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.FirstName, &user.IsMainUser,
		&user.LinkedChatID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, nil
	}
	return &user, nil
}

// GetMainUser retrieves the main user.
func (r *TelegramRepository) GetMainUser(ctx context.Context) (*model.TelegramUser, error) {
	var user model.TelegramUser
	query := `
		SELECT id, username, first_name, is_main_user, linked_chat_id, created_at, updated_at
		FROM telegram_users WHERE is_main_user = TRUE LIMIT 1`

	err := r.db.QueryRow(ctx, query).Scan(
		&user.ID, &user.Username, &user.FirstName, &user.IsMainUser,
		&user.LinkedChatID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, nil
	}
	return &user, nil
}

// SetMainUser sets a user as the main user.
func (r *TelegramRepository) SetMainUser(ctx context.Context, userID int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "UPDATE telegram_users SET is_main_user = FALSE WHERE is_main_user = TRUE")
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "UPDATE telegram_users SET is_main_user = TRUE WHERE id = $1", userID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// CreateChatLink creates a link between two users.
func (r *TelegramRepository) CreateChatLink(ctx context.Context, link *model.TelegramChatLink) error {
	query := `
		INSERT INTO telegram_chat_links (id, main_user_id, linked_user_id, room_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (main_user_id, linked_user_id) DO UPDATE SET
			room_id = EXCLUDED.room_id
		RETURNING created_at`

	return r.db.QueryRow(ctx, query,
		link.ID, link.MainUserID, link.LinkedUserID, link.RoomID).
		Scan(&link.CreatedAt)
}

// GetChatLink retrieves a chat link between two users.
func (r *TelegramRepository) GetChatLink(ctx context.Context, mainUserID, linkedUserID int64) (*model.TelegramChatLink, error) {
	var link model.TelegramChatLink
	query := `
		SELECT id, main_user_id, linked_user_id, room_id, created_at
		FROM telegram_chat_links
		WHERE main_user_id = $1 AND linked_user_id = $2`

	err := r.db.QueryRow(ctx, query, mainUserID, linkedUserID).Scan(
		&link.ID, &link.MainUserID, &link.LinkedUserID, &link.RoomID, &link.CreatedAt)
	if err != nil {
		return nil, nil
	}
	return &link, nil
}

// GetLinkedUsers retrieves all users linked to a main user.
func (r *TelegramRepository) GetLinkedUsers(ctx context.Context, mainUserID int64) ([]model.TelegramUser, error) {
	var users []model.TelegramUser
	query := `
		SELECT tu.id, tu.username, tu.first_name, tu.is_main_user, tu.linked_chat_id, tu.created_at, tu.updated_at
		FROM telegram_users tu
		JOIN telegram_chat_links tcl ON tu.id = tcl.linked_user_id
		WHERE tcl.main_user_id = $1`

	rows, err := r.db.Query(ctx, query, mainUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user model.TelegramUser
		if err := rows.Scan(&user.ID, &user.Username, &user.FirstName, &user.IsMainUser,
			&user.LinkedChatID, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConversationRepository struct {
	pool *pgxpool.Pool
}

func NewConversationRepository(pool *pgxpool.Pool) *ConversationRepository {
	return &ConversationRepository{pool: pool}
}

func (r *ConversationRepository) CreateConversation(ctx context.Context, userID uuid.UUID, title *string, model *string) (*Conversation, error) {
	conv := &Conversation{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     title,
		Model:     model,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO conversations (id, user_id, title, model, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, title, model, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		conv.ID,
		conv.UserID,
		conv.Title,
		conv.Model,
		conv.CreatedAt,
		conv.UpdatedAt,
	).Scan(
		&conv.ID,
		&conv.UserID,
		&conv.Title,
		&conv.Model,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return conv, nil
}

func (r *ConversationRepository) GetConversation(ctx context.Context, conversationID uuid.UUID, userID uuid.UUID) (*Conversation, error) {
	conv := &Conversation{}

	query := `
		SELECT id, user_id, title, model, created_at, updated_at
		FROM conversations
		WHERE id = $1 AND user_id = $2
	`

	err := r.pool.QueryRow(ctx, query, conversationID, userID).Scan(
		&conv.ID,
		&conv.UserID,
		&conv.Title,
		&conv.Model,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return conv, nil
}

func (r *ConversationRepository) ListConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Conversation, error) {
	query := `
		SELECT id, user_id, title, model, created_at, updated_at
		FROM conversations
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var conv Conversation
		err := rows.Scan(
			&conv.ID,
			&conv.UserID,
			&conv.Title,
			&conv.Model,
			&conv.CreatedAt,
			&conv.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

func (r *ConversationRepository) UpdateConversation(ctx context.Context, conversationID uuid.UUID, userID uuid.UUID, title *string, model *string) error {
	query := `
		UPDATE conversations
		SET title = COALESCE($3, title),
		    model = COALESCE($4, model),
		    updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`

	result, err := r.pool.Exec(ctx, query, conversationID, userID, title, model)
	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("conversation not found or access denied")
	}

	return nil
}

func (r *ConversationRepository) DeleteConversation(ctx context.Context, conversationID uuid.UUID, userID uuid.UUID) error {
	query := `
		DELETE FROM conversations
		WHERE id = $1 AND user_id = $2
	`

	result, err := r.pool.Exec(ctx, query, conversationID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("conversation not found or access denied")
	}

	return nil
}

func (r *ConversationRepository) AddMessage(ctx context.Context, conversationID uuid.UUID, role string, content interface{}, metadata map[string]interface{}) (*Message, error) {
	// Convert content to JSONB
	contentJSON, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal content: %w", err)
	}

	// Convert metadata to JSONB
	var metadataJSON []byte
	if metadata != nil {
		metadataJSON, err = json.Marshal(metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	msg := &Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		Metadata:       metadata,
		CreatedAt:      time.Now(),
	}

	query := `
		INSERT INTO messages (id, conversation_id, role, content, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, conversation_id, role, content, metadata, created_at
	`

	var contentJSONB []byte
	var metadataJSONB []byte

	err = r.pool.QueryRow(
		ctx,
		query,
		msg.ID,
		msg.ConversationID,
		msg.Role,
		contentJSON,
		metadataJSON,
		msg.CreatedAt,
	).Scan(
		&msg.ID,
		&msg.ConversationID,
		&msg.Role,
		&contentJSONB,
		&metadataJSONB,
		&msg.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to add message: %w", err)
	}

	// Unmarshal content back
	if err := json.Unmarshal(contentJSONB, &msg.Content); err != nil {
		return nil, fmt.Errorf("failed to unmarshal content: %w", err)
	}

	// Unmarshal metadata back
	if len(metadataJSONB) > 0 {
		if err := json.Unmarshal(metadataJSONB, &msg.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	// Update conversation's updated_at timestamp
	updateQuery := `
		UPDATE conversations
		SET updated_at = NOW()
		WHERE id = $1
	`
	_, _ = r.pool.Exec(ctx, updateQuery, conversationID)

	return msg, nil
}

func (r *ConversationRepository) GetMessages(ctx context.Context, conversationID uuid.UUID, userID uuid.UUID) ([]Message, error) {
	// First verify the conversation belongs to the user
	_, err := r.GetConversation(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, conversation_id, role, content, metadata, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var contentJSONB []byte
		var metadataJSONB []byte

		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.Role,
			&contentJSONB,
			&metadataJSONB,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		// Unmarshal content
		if err := json.Unmarshal(contentJSONB, &msg.Content); err != nil {
			return nil, fmt.Errorf("failed to unmarshal content: %w", err)
		}

		// Unmarshal metadata
		if len(metadataJSONB) > 0 {
			if err := json.Unmarshal(metadataJSONB, &msg.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

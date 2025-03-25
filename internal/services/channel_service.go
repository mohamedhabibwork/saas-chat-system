package services

import (
	"database/sql"
	"fmt"
	"time"

	"awesomeProject/internal/models"
)

// ChannelService handles channel-related operations
type ChannelService struct {
	db Database
}

// NewChannelService creates a new channel service
func NewChannelService(db Database) *ChannelService {
	return &ChannelService{
		db: db,
	}
}

// CreateChannel creates a new channel
func (s *ChannelService) CreateChannel(channel *models.Channel) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Insert channel
	query := `
		INSERT INTO channels (
			name, description, type, created_by, tenant_id,
			settings, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id
	`
	err = tx.QueryRow(
		query,
		channel.Name, channel.Description, channel.Type,
		channel.CreatedBy, channel.TenantID, channel.Settings,
	).Scan(&channel.ID)
	if err != nil {
		return fmt.Errorf("failed to create channel: %v", err)
	}

	// Add creator as admin
	err = s.addMember(tx, channel.ID, channel.CreatedBy, "admin")
	if err != nil {
		return fmt.Errorf("failed to add creator as admin: %v", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// GetChannel retrieves a channel by ID
func (s *ChannelService) GetChannel(channelID int) (*models.Channel, error) {
	var channel models.Channel
	query := `
		SELECT id, name, description, type, created_by,
			   tenant_id, settings, created_at, updated_at
		FROM channels
		WHERE id = $1
	`
	err := s.db.QueryRow(query, channelID).Scan(
		&channel.ID, &channel.Name, &channel.Description,
		&channel.Type, &channel.CreatedBy, &channel.TenantID,
		&channel.Settings, &channel.CreatedAt, &channel.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &channel, nil
}

// ListChannels retrieves all channels for a user
func (s *ChannelService) ListChannels(userID int) ([]models.Channel, error) {
	query := `
		SELECT c.id, c.name, c.description, c.type, c.created_by,
			   c.tenant_id, c.settings, c.created_at, c.updated_at
		FROM channels c
		JOIN channel_members cm ON c.id = cm.channel_id
		WHERE cm.user_id = $1
		ORDER BY c.created_at DESC
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []models.Channel
	for rows.Next() {
		var channel models.Channel
		err := rows.Scan(
			&channel.ID, &channel.Name, &channel.Description,
			&channel.Type, &channel.CreatedBy, &channel.TenantID,
			&channel.Settings, &channel.CreatedAt, &channel.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}

	return channels, nil
}

// UpdateChannel updates a channel's settings
func (s *ChannelService) UpdateChannel(channelID int, settings models.ChannelSettings) error {
	query := `
		UPDATE channels
		SET settings = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := s.db.Exec(query, settings, channelID)
	return err
}

// AddMember adds a user to a channel
func (s *ChannelService) AddMember(channelID int, userID int, role string) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	err = s.addMember(tx, channelID, userID, role)
	if err != nil {
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// RemoveMember removes a user from a channel
func (s *ChannelService) RemoveMember(channelID int, userID int) error {
	query := `
		DELETE FROM channel_members
		WHERE channel_id = $1 AND user_id = $2
	`
	_, err := s.db.Exec(query, channelID, userID)
	return err
}

// GetMembers retrieves all members of a channel
func (s *ChannelService) GetMembers(channelID int) ([]models.ChannelMember, error) {
	query := `
		SELECT id, channel_id, user_id, role, joined_at
		FROM channel_members
		WHERE channel_id = $1
		ORDER BY joined_at ASC
	`
	rows, err := s.db.Query(query, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.ChannelMember
	for rows.Next() {
		var member models.ChannelMember
		err := rows.Scan(
			&member.ID, &member.ChannelID, &member.UserID,
			&member.Role, &member.JoinedAt,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, nil
}

// AddMessage adds a message to a channel
func (s *ChannelService) AddMessage(channelID int, userID int, content string, msgType string, metadata map[string]interface{}) error {
	query := `
		INSERT INTO channel_messages (
			channel_id, user_id, content, type, metadata
		)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.db.Exec(query, channelID, userID, content, msgType, metadata)
	return err
}

// GetMessages retrieves messages from a channel
func (s *ChannelService) GetMessages(channelID int, limit int, offset int) ([]models.ChannelMessage, error) {
	query := `
		SELECT id, channel_id, user_id, content, type,
			   metadata, created_at
		FROM channel_messages
		WHERE channel_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := s.db.Query(query, channelID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.ChannelMessage
	for rows.Next() {
		var message models.ChannelMessage
		err := rows.Scan(
			&message.ID, &message.ChannelID, &message.UserID,
			&message.Content, &message.Type, &message.Metadata,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// Helper functions

func (s *ChannelService) addMember(tx *sql.Tx, channelID int, userID int, role string) error {
	query := `
		INSERT INTO channel_members (
			channel_id, user_id, role, joined_at
		)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (channel_id, user_id) DO UPDATE
		SET role = $3
	`
	_, err := tx.Exec(query, channelID, userID, role)
	return err
} 
package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	ID        string    `json:"id"`
	ChannelID string    `json:"channel_id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Channel struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ChatService struct {
	db       *sql.DB
	clients  map[string]map[*websocket.Conn]bool
	channels map[string]*Channel
}

func NewChatService(db *sql.DB) *ChatService {
	return &ChatService{
		db:       db,
		clients:  make(map[string]map[*websocket.Conn]bool),
		channels: make(map[string]*Channel),
	}
}

func (s *ChatService) CreateChannel(ctx context.Context, channel *Channel) error {
	query := `
		INSERT INTO channels (name, description)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at`

	err := s.db.QueryRowContext(ctx, query,
		channel.Name,
		channel.Description,
	).Scan(&channel.ID, &channel.CreatedAt, &channel.UpdatedAt)

	if err != nil {
		return err
	}

	s.channels[channel.ID] = channel
	s.clients[channel.ID] = make(map[*websocket.Conn]bool)

	return nil
}

func (s *ChatService) JoinChannel(channelID string, conn *websocket.Conn) error {
	if _, ok := s.channels[channelID]; !ok {
		return errors.New("channel not found")
	}

	if s.clients[channelID] == nil {
		s.clients[channelID] = make(map[*websocket.Conn]bool)
	}
	s.clients[channelID][conn] = true

	return nil
}

func (s *ChatService) LeaveChannel(channelID string, conn *websocket.Conn) error {
	if clients, ok := s.clients[channelID]; ok {
		delete(clients, conn)
		return nil
	}
	return errors.New("channel not found")
}

func (s *ChatService) GetMessages(ctx context.Context, channelID string, page, limit int) ([]*Message, error) {
	query := `
		SELECT id, channel_id, user_id, content, created_at
		FROM messages
		WHERE channel_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	offset := (page - 1) * limit
	rows, err := s.db.QueryContext(ctx, query, channelID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		msg := &Message{}
		err := rows.Scan(
			&msg.ID,
			&msg.ChannelID,
			&msg.UserID,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (s *ChatService) BroadcastMessage(channelID string, message *Message) error {
	clients, ok := s.clients[channelID]
	if !ok {
		return errors.New("channel not found")
	}

	for client := range clients {
		err := client.WriteJSON(message)
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}

	return nil
} 
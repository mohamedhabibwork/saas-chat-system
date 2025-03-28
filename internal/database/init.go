package database

import (
	"database/sql"
	"fmt"
	"log"
)

// InitDB initializes the database with all necessary tables and the default admin account
func InitDB(db *sql.DB) error {
	// Create tables
	if err := createTablesV2(db); err != nil {
		return fmt.Errorf("error creating tables: %v", err)
	}

	// Create indexes
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("error creating indexes: %v", err)
	}

	// Create default admin account
	if err := CreateDefaultAdmin(db); err != nil {
		return fmt.Errorf("error creating default admin: %v", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// createTablesV2 creates all necessary database tables
func createTablesV2(db *sql.DB) error {
	queries := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			first_name VARCHAR(50),
			last_name VARCHAR(50),
			role_id INTEGER NOT NULL,
			is_active BOOLEAN DEFAULT true,
			tenant_id INTEGER,
			timezone VARCHAR(50) DEFAULT 'UTC',
			last_login TIMESTAMP,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,

		// Tenants table
		`CREATE TABLE IF NOT EXISTS tenants (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			status VARCHAR(20) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,

		// Subscriptions table
		`CREATE TABLE IF NOT EXISTS subscriptions (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id),
			tenant_id INTEGER NOT NULL REFERENCES tenants(id),
			plan VARCHAR(50) NOT NULL,
			status VARCHAR(20) NOT NULL,
			start_date TIMESTAMP NOT NULL,
			end_date TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,

		// Channels table
		`CREATE TABLE IF NOT EXISTS channels (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			type VARCHAR(20) NOT NULL,
			created_by INTEGER NOT NULL REFERENCES users(id),
			tenant_id INTEGER NOT NULL REFERENCES tenants(id),
			settings JSONB NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,

		// Channel members table
		`CREATE TABLE IF NOT EXISTS channel_members (
			channel_id INTEGER NOT NULL REFERENCES channels(id),
			user_id INTEGER NOT NULL REFERENCES users(id),
			role VARCHAR(20) NOT NULL,
			joined_at TIMESTAMP NOT NULL,
			PRIMARY KEY (channel_id, user_id)
		)`,

		// Channel messages table
		`CREATE TABLE IF NOT EXISTS channel_messages (
			id SERIAL PRIMARY KEY,
			channel_id INTEGER NOT NULL REFERENCES channels(id),
			user_id INTEGER NOT NULL REFERENCES users(id),
			content TEXT NOT NULL,
			type VARCHAR(20) NOT NULL,
			metadata JSONB,
			created_at TIMESTAMP NOT NULL
		)`,

		// WebRTC connections table
		`CREATE TABLE IF NOT EXISTS webrtc_connections (
			id SERIAL PRIMARY KEY,
			channel_id INTEGER NOT NULL REFERENCES channels(id),
			user_id INTEGER NOT NULL REFERENCES users(id),
			peer_id VARCHAR(255) NOT NULL,
			stream_type VARCHAR(20) NOT NULL,
			active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,

		// Files table
		`CREATE TABLE IF NOT EXISTS files (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id),
			tenant_id INTEGER NOT NULL REFERENCES tenants(id),
			name VARCHAR(255) NOT NULL,
			path VARCHAR(255) NOT NULL,
			size BIGINT NOT NULL,
			mime_type VARCHAR(100) NOT NULL,
			metadata JSONB,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,

		// File shares table
		`CREATE TABLE IF NOT EXISTS file_shares (
			id SERIAL PRIMARY KEY,
			file_id INTEGER NOT NULL REFERENCES files(id),
			shared_by INTEGER NOT NULL REFERENCES users(id),
			shared_with INTEGER NOT NULL REFERENCES users(id),
			permissions VARCHAR(20) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			expires_at TIMESTAMP
		)`,

		// Bot configurations table
		`CREATE TABLE IF NOT EXISTS bot_configs (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id),
			tenant_id INTEGER NOT NULL REFERENCES tenants(id),
			name VARCHAR(100) NOT NULL,
			description TEXT,
			model_type VARCHAR(50) NOT NULL,
			model_config JSONB NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,

		// Bot conversations table
		`CREATE TABLE IF NOT EXISTS bot_conversations (
			id SERIAL PRIMARY KEY,
			bot_id INTEGER NOT NULL REFERENCES bot_configs(id),
			user_id INTEGER NOT NULL REFERENCES users(id),
			channel_id INTEGER NOT NULL REFERENCES channels(id),
			status VARCHAR(20) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,

		// Bot messages table
		`CREATE TABLE IF NOT EXISTS bot_messages (
			id SERIAL PRIMARY KEY,
			conversation_id INTEGER NOT NULL REFERENCES bot_conversations(id),
			content TEXT NOT NULL,
			role VARCHAR(20) NOT NULL,
			metadata JSONB,
			created_at TIMESTAMP NOT NULL
		)`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

// createIndexes creates all necessary database indexes
func createIndexes(db *sql.DB) error {
	queries := []string{
		// Users indexes
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_role ON users(role_id)`,
		`CREATE INDEX IF NOT EXISTS idx_users_status ON users(is_active)`,

		// Tenants indexes
		`CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status)`,

		// Subscriptions indexes
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_tenant_id ON subscriptions(tenant_id)`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_end_date ON subscriptions(end_date)`,

		// Channels indexes
		`CREATE INDEX IF NOT EXISTS idx_channels_tenant_id ON channels(tenant_id)`,
		`CREATE INDEX IF NOT EXISTS idx_channels_type ON channels(type)`,
		`CREATE INDEX IF NOT EXISTS idx_channels_created_by ON channels(created_by)`,

		// Channel members indexes
		`CREATE INDEX IF NOT EXISTS idx_channel_members_user_id ON channel_members(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_channel_members_role ON channel_members(role)`,

		// Channel messages indexes
		`CREATE INDEX IF NOT EXISTS idx_channel_messages_channel_id ON channel_messages(channel_id)`,
		`CREATE INDEX IF NOT EXISTS idx_channel_messages_user_id ON channel_messages(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_channel_messages_created_at ON channel_messages(created_at)`,

		// WebRTC connections indexes
		`CREATE INDEX IF NOT EXISTS idx_webrtc_connections_channel_id ON webrtc_connections(channel_id)`,
		`CREATE INDEX IF NOT EXISTS idx_webrtc_connections_user_id ON webrtc_connections(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_webrtc_connections_active ON webrtc_connections(active)`,

		// Files indexes
		`CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_files_tenant_id ON files(tenant_id)`,
		`CREATE INDEX IF NOT EXISTS idx_files_mime_type ON files(mime_type)`,
		`CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at)`,

		// File shares indexes
		`CREATE INDEX IF NOT EXISTS idx_file_shares_file_id ON file_shares(file_id)`,
		`CREATE INDEX IF NOT EXISTS idx_file_shares_shared_by ON file_shares(shared_by)`,
		`CREATE INDEX IF NOT EXISTS idx_file_shares_shared_with ON file_shares(shared_with)`,
		`CREATE INDEX IF NOT EXISTS idx_file_shares_expires_at ON file_shares(expires_at)`,

		// Bot configurations indexes
		`CREATE INDEX IF NOT EXISTS idx_bot_configs_user_id ON bot_configs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_bot_configs_tenant_id ON bot_configs(tenant_id)`,
		`CREATE INDEX IF NOT EXISTS idx_bot_configs_model_type ON bot_configs(model_type)`,

		// Bot conversations indexes
		`CREATE INDEX IF NOT EXISTS idx_bot_conversations_bot_id ON bot_conversations(bot_id)`,
		`CREATE INDEX IF NOT EXISTS idx_bot_conversations_user_id ON bot_conversations(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_bot_conversations_channel_id ON bot_conversations(channel_id)`,
		`CREATE INDEX IF NOT EXISTS idx_bot_conversations_status ON bot_conversations(status)`,

		// Bot messages indexes
		`CREATE INDEX IF NOT EXISTS idx_bot_messages_conversation_id ON bot_messages(conversation_id)`,
		`CREATE INDEX IF NOT EXISTS idx_bot_messages_role ON bot_messages(role)`,
		`CREATE INDEX IF NOT EXISTS idx_bot_messages_created_at ON bot_messages(created_at)`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
} 
package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// PostgresDB implements the Database interface for PostgreSQL
type PostgresDB struct {
	db *sql.DB
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(host, port, user, password, dbname string) (*PostgresDB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	return &PostgresDB{db: db}, nil
}

// Query implements the Database interface
func (p *PostgresDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.Query(query, args...)
}

// QueryRow implements the Database interface
func (p *PostgresDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return p.db.QueryRow(query, args...)
}

// Exec implements the Database interface
func (p *PostgresDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return p.db.Exec(query, args...)
}

// Begin implements the Database interface
func (p *PostgresDB) Begin() (*sql.Tx, error) {
	return p.db.Begin()
}

// Close implements the Database interface
func (p *PostgresDB) Close() error {
	return p.db.Close()
}

// SetupTables creates all necessary database tables
func (p *PostgresDB) SetupTables() error {
	queries := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			first_name VARCHAR(50),
			last_name VARCHAR(50),
			tenant_id INTEGER,
			role_id INTEGER,
			is_active BOOLEAN DEFAULT true,
			timezone VARCHAR(50) DEFAULT 'UTC',
			last_login TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		)`,

		// Roles table
		`CREATE TABLE IF NOT EXISTS roles (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50) UNIQUE NOT NULL,
			description TEXT,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		)`,

		// Permissions table
		`CREATE TABLE IF NOT EXISTS permissions (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50) UNIQUE NOT NULL,
			description TEXT,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		)`,

		// Role permissions table
		`CREATE TABLE IF NOT EXISTS role_permissions (
			role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
			permission_id INTEGER REFERENCES permissions(id) ON DELETE CASCADE,
			PRIMARY KEY (role_id, permission_id)
		)`,

		// Sessions table
		`CREATE TABLE IF NOT EXISTS sessions (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			token VARCHAR(255) UNIQUE NOT NULL,
			device_info TEXT,
			ip_address VARCHAR(45),
			last_activity TIMESTAMP WITH TIME ZONE NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL
		)`,

		// Login attempts table
		`CREATE TABLE IF NOT EXISTS login_attempts (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			ip_address VARCHAR(45),
			success BOOLEAN NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL
		)`,

		// Password resets table
		`CREATE TABLE IF NOT EXISTS password_resets (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			token VARCHAR(255) UNIQUE NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			used BOOLEAN DEFAULT false,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL
		)`,

		// Plans table
		`CREATE TABLE IF NOT EXISTS plans (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50) UNIQUE NOT NULL,
			description TEXT,
			price DECIMAL(10,2) NOT NULL,
			interval VARCHAR(20) NOT NULL,
			features JSONB,
			limits JSONB,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		)`,

		// Subscriptions table
		`CREATE TABLE IF NOT EXISTS subscriptions (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			plan_id INTEGER REFERENCES plans(id) ON DELETE CASCADE,
			status VARCHAR(20) NOT NULL,
			start_date TIMESTAMP WITH TIME ZONE NOT NULL,
			end_date TIMESTAMP WITH TIME ZONE NOT NULL,
			auto_renew BOOLEAN DEFAULT true,
			payment_method VARCHAR(50),
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		)`,

		// Usage table
		`CREATE TABLE IF NOT EXISTS usage (
			id SERIAL PRIMARY KEY,
			subscription_id INTEGER REFERENCES subscriptions(id) ON DELETE CASCADE,
			messages_sent INTEGER DEFAULT 0,
			tokens_used INTEGER DEFAULT 0,
			storage_used INTEGER DEFAULT 0,
			period_start TIMESTAMP WITH TIME ZONE NOT NULL,
			period_end TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		)`,

		// Payments table
		`CREATE TABLE IF NOT EXISTS payments (
			id SERIAL PRIMARY KEY,
			subscription_id INTEGER REFERENCES subscriptions(id) ON DELETE CASCADE,
			amount DECIMAL(10,2) NOT NULL,
			currency VARCHAR(3) NOT NULL,
			status VARCHAR(20) NOT NULL,
			payment_method VARCHAR(50),
			transaction_id VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		)`,

		// Bots table
		`CREATE TABLE IF NOT EXISTS bots (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50) NOT NULL,
			tenant_id INTEGER,
			token VARCHAR(255) UNIQUE NOT NULL,
			config JSONB NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		)`,

		// Bot interactions table
		`CREATE TABLE IF NOT EXISTS bot_interactions (
			id SERIAL PRIMARY KEY,
			bot_id INTEGER REFERENCES bots(id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			message TEXT NOT NULL,
			response TEXT NOT NULL,
			tokens_used INTEGER NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL
		)`,

		// Bot stats table
		`CREATE TABLE IF NOT EXISTS bot_stats (
			id SERIAL PRIMARY KEY,
			bot_id INTEGER REFERENCES bots(id) ON DELETE CASCADE,
			interactions_count INTEGER DEFAULT 0,
			total_tokens_used INTEGER DEFAULT 0,
			last_interaction TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		)`,

		`CREATE TABLE IF NOT EXISTS files (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			filename VARCHAR(255) NOT NULL,
			filepath VARCHAR(1024) NOT NULL,
			url VARCHAR(1024) NOT NULL,
			size BIGINT NOT NULL,
			content_type VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, filepath)
		)`,

		`CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at)`,

		// Channels table
		`CREATE TABLE IF NOT EXISTS channels (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			type VARCHAR(20) NOT NULL,
			created_by INTEGER REFERENCES users(id) ON DELETE CASCADE,
			tenant_id INTEGER,
			settings JSONB NOT NULL DEFAULT '{}',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,

		// Channel members table
		`CREATE TABLE IF NOT EXISTS channel_members (
			id SERIAL PRIMARY KEY,
			channel_id INTEGER REFERENCES channels(id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			role VARCHAR(20) NOT NULL DEFAULT 'member',
			joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(channel_id, user_id)
		)`,

		// Channel messages table
		`CREATE TABLE IF NOT EXISTS channel_messages (
			id SERIAL PRIMARY KEY,
			channel_id INTEGER REFERENCES channels(id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			type VARCHAR(20) NOT NULL DEFAULT 'text',
			metadata JSONB DEFAULT '{}',
			sender_timezone VARCHAR(50),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,

		// WebRTC connections table
		`CREATE TABLE IF NOT EXISTS webrtc_connections (
			id SERIAL PRIMARY KEY,
			channel_id INTEGER REFERENCES channels(id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			peer_id VARCHAR(255) NOT NULL,
			stream_type VARCHAR(20) NOT NULL,
			active BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_channels_tenant_id ON channels(tenant_id)`,
		`CREATE INDEX IF NOT EXISTS idx_channel_members_channel_id ON channel_members(channel_id)`,
		`CREATE INDEX IF NOT EXISTS idx_channel_members_user_id ON channel_members(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_channel_messages_channel_id ON channel_messages(channel_id)`,
		`CREATE INDEX IF NOT EXISTS idx_channel_messages_user_id ON channel_messages(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_webrtc_connections_channel_id ON webrtc_connections(channel_id)`,
		`CREATE INDEX IF NOT EXISTS idx_webrtc_connections_user_id ON webrtc_connections(user_id)`,
	}

	for _, query := range queries {
		_, err := p.db.Exec(query)
		if err != nil {
			log.Printf("Error executing query: %v\nQuery: %s", err, query)
			return err
		}
	}

	return nil
}

// CreateDefaultRoles creates the default roles in the system
func (p *PostgresDB) CreateDefaultRoles() error {
	roles := []struct {
		name        string
		description string
	}{
		{"admin", "System administrator with full access"},
		{"user", "Regular user with basic access"},
		{"bot", "Bot user with limited access"},
	}

	for _, role := range roles {
		query := `
			INSERT INTO roles (name, description, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())
			ON CONFLICT (name) DO NOTHING
		`
		_, err := p.db.Exec(query, role.name, role.description)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateDefaultPermissions creates the default permissions in the system
func (p *PostgresDB) CreateDefaultPermissions() error {
	permissions := []struct {
		name        string
		description string
	}{
		{"create_bot", "Create new bots"},
		{"edit_bot", "Edit existing bots"},
		{"delete_bot", "Delete bots"},
		{"view_bot", "View bot details"},
		{"manage_users", "Manage user accounts"},
		{"manage_roles", "Manage roles and permissions"},
		{"manage_subscriptions", "Manage subscriptions"},
		{"view_analytics", "View analytics and statistics"},
	}

	for _, permission := range permissions {
		query := `
			INSERT INTO permissions (name, description, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())
			ON CONFLICT (name) DO NOTHING
		`
		_, err := p.db.Exec(query, permission.name, permission.description)
		if err != nil {
			return err
		}
	}

	return nil
}

// AssignDefaultPermissions assigns default permissions to roles
func (p *PostgresDB) AssignDefaultPermissions() error {
	// Admin role gets all permissions
	query := `
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id
		FROM roles r
		CROSS JOIN permissions p
		WHERE r.name = 'admin'
		ON CONFLICT DO NOTHING
	`
	_, err := p.db.Exec(query)
	if err != nil {
		return err
	}

	// User role gets basic permissions
	query = `
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id
		FROM roles r
		CROSS JOIN permissions p
		WHERE r.name = 'user'
		AND p.name IN ('view_bot')
		ON CONFLICT DO NOTHING
	`
	_, err = p.db.Exec(query)
	if err != nil {
		return err
	}

	// Bot role gets minimal permissions
	query = `
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id
		FROM roles r
		CROSS JOIN permissions p
		WHERE r.name = 'bot'
		AND p.name IN ('view_bot')
		ON CONFLICT DO NOTHING
	`
	_, err = p.db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

// CreateDefaultAdmin creates the default admin user
func (p *PostgresDB) CreateDefaultAdmin() error {
	// Check if admin user exists
	var exists bool
	err := p.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = 'admin@example.com')").Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	// Create admin user
	query := `
		INSERT INTO users (
			username, email, password_hash, first_name, last_name,
			role_id, is_active, created_at, updated_at
		)
		VALUES (
			'admin', 'admin@example.com', $1, 'Admin', 'User',
			(SELECT id FROM roles WHERE name = 'admin'), true,
			NOW(), NOW()
		)
	`
	_, err = p.db.Exec(query, "$2a$10$your_hashed_password_here") // Replace with actual hashed password
	return err
} 
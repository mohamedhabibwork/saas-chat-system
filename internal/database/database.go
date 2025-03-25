package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// DB connection
var DB *sql.DB

// Database defines the interface for database operations
type Database interface {
	// Query executes a query that returns rows
	Query(query string, args ...interface{}) (*sql.Rows, error)

	// QueryRow executes a query that is expected to return at most one row
	QueryRow(query string, args ...interface{}) *sql.Row

	// Exec executes a query without returning any rows
	Exec(query string, args ...interface{}) (sql.Result, error)

	// Begin starts a transaction
	Begin() (*sql.Tx, error)

	// Close closes the database connection
	Close() error
}

// SetupDB initializes the database connection and creates tables
func SetupDB() (*sql.DB, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get DB connection parameters
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "messenger")

	// Connect to PostgreSQL
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Check connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	// Create tenants table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS tenants (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) UNIQUE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create users table with tenant_id
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) NOT NULL,
			password VARCHAR(100) NOT NULL,
			tenant_id INTEGER REFERENCES tenants(id),
			timezone VARCHAR(50) DEFAULT 'UTC',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(username, tenant_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create groups table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS groups (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			tenant_id INTEGER REFERENCES tenants(id),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(name, tenant_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create group_members table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS group_members (
			group_id INTEGER REFERENCES groups(id),
			user_id INTEGER REFERENCES users(id),
			joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (group_id, user_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create topics table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS topics (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			tenant_id INTEGER REFERENCES tenants(id),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(name, tenant_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create topic_subscriptions table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS topic_subscriptions (
			topic_id INTEGER REFERENCES topics(id),
			user_id INTEGER REFERENCES users(id),
			subscribed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (topic_id, user_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create messages table with enhanced fields
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			sender_id INTEGER REFERENCES users(id),
			receiver_id INTEGER REFERENCES users(id) NULL,
			group_id INTEGER REFERENCES groups(id) NULL,
			topic_id INTEGER REFERENCES topics(id) NULL,
			tenant_id INTEGER REFERENCES tenants(id),
			content TEXT NOT NULL,
			message_type VARCHAR(20) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create bots table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bots (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			tenant_id INTEGER REFERENCES tenants(id),
			token VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(name, tenant_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create channels table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS channels (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			tenant_id INTEGER REFERENCES tenants(id),
			created_by INTEGER REFERENCES users(id),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(name, tenant_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create channel_members table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS channel_members (
			channel_id INTEGER REFERENCES channels(id),
			user_id INTEGER REFERENCES users(id),
			role VARCHAR(20) NOT NULL DEFAULT 'member',
			joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (channel_id, user_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create tabs table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tabs (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id),
			name VARCHAR(100) NOT NULL,
			type VARCHAR(20) NOT NULL,
			target_id INTEGER NOT NULL,
			"order" INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create bot_messages table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bot_messages (
			id SERIAL PRIMARY KEY,
			bot_id INTEGER REFERENCES bots(id),
			content TEXT NOT NULL,
			tenant_id INTEGER REFERENCES tenants(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create channel_messages table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS channel_messages (
			id SERIAL PRIMARY KEY,
			channel_id INTEGER REFERENCES channels(id),
			sender_id INTEGER REFERENCES users(id),
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
} 
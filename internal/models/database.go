package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// Database represents the database connection and operations
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new Database instance
func NewDatabase(db *sql.DB) *Database {
	return &Database{
		db: db,
	}
}

// Begin starts a new transaction
func (db *Database) Begin() (*sql.Tx, error) {
	return db.db.Begin()
}

// Query executes a query that returns rows
func (db *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.db.Query(query, args...)
}

// QueryRow executes a query that returns a single row
func (db *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.db.QueryRow(query, args...)
}

// Exec executes a query without returning any rows
func (db *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.db.Exec(query, args...)
}

// CreateFile creates a new file record in the database
func (db *Database) CreateFile(file *File) error {
	query := `
		INSERT INTO files (
			user_id, tenant_id, filename, name, filepath,
			path, url, size, content_type, mime_type,
			metadata, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		RETURNING id
	`
	
	metadataJSON, err := json.Marshal(file.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	return db.db.QueryRow(
		query,
		file.UserID, file.TenantID, file.Filename, file.Name,
		file.Filepath, file.Path, file.URL, file.Size,
		file.ContentType, file.MimeType, metadataJSON,
	).Scan(&file.ID)
}

// GetFile retrieves a file record from the database
func (db *Database) GetFile(fileID int) (*File, error) {
	var file File
	var metadataJSON []byte

	query := `
		SELECT id, user_id, tenant_id, filename, name,
			   filepath, path, url, size, content_type,
			   mime_type, metadata, created_at, updated_at
		FROM files
		WHERE id = $1
	`
	err := db.db.QueryRow(query, fileID).Scan(
		&file.ID, &file.UserID, &file.TenantID, &file.Filename,
		&file.Name, &file.Filepath, &file.Path, &file.URL,
		&file.Size, &file.ContentType, &file.MimeType,
		&metadataJSON, &file.CreatedAt, &file.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(metadataJSON, &file.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %v", err)
	}

	return &file, nil
}

// DeleteFile removes a file record from the database
func (db *Database) DeleteFile(fileID int) error {
	query := "DELETE FROM files WHERE id = $1"
	result, err := db.db.Exec(query, fileID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("file not found")
	}

	return nil
}

// ListFiles returns all files in a channel
func (db *Database) ListFiles(channelID int) ([]*File, error) {
	query := `
		SELECT f.id, f.user_id, f.tenant_id, f.filename,
			   f.name, f.filepath, f.path, f.url, f.size,
			   f.content_type, f.mime_type, f.metadata,
			   f.created_at, f.updated_at
		FROM files f
		JOIN file_metadata fm ON f.id = fm.file_id
		WHERE fm.channel_id = $1
		ORDER BY f.created_at DESC
	`
	rows, err := db.db.Query(query, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*File
	for rows.Next() {
		var file File
		var metadataJSON []byte

		err := rows.Scan(
			&file.ID, &file.UserID, &file.TenantID, &file.Filename,
			&file.Name, &file.Filepath, &file.Path, &file.URL,
			&file.Size, &file.ContentType, &file.MimeType,
			&metadataJSON, &file.CreatedAt, &file.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadataJSON, &file.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %v", err)
		}

		files = append(files, &file)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

// GetChannel retrieves a channel from the database
func (db *Database) GetChannel(channelID int) (*Channel, error) {
	var channel Channel
	var settingsJSON []byte

	query := `
		SELECT id, name, description, type, created_by,
			   tenant_id, settings, created_at, updated_at
		FROM channels
		WHERE id = $1
	`
	err := db.db.QueryRow(query, channelID).Scan(
		&channel.ID, &channel.Name, &channel.Description,
		&channel.Type, &channel.CreatedBy, &channel.TenantID,
		&settingsJSON, &channel.CreatedAt, &channel.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(settingsJSON, &channel.Settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %v", err)
	}

	return &channel, nil
}

// GetChannelMembers returns all members of a channel
func (db *Database) GetChannelMembers(channelID int) ([]*ChannelMember, error) {
	query := `
		SELECT id, channel_id, user_id, role, joined_at
		FROM channel_members
		WHERE channel_id = $1
		ORDER BY joined_at ASC
	`
	rows, err := db.db.Query(query, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*ChannelMember
	for rows.Next() {
		var member ChannelMember
		err := rows.Scan(
			&member.ID, &member.ChannelID, &member.UserID,
			&member.Role, &member.JoinedAt,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, &member)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
} 
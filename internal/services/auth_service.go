package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"awesomeProject/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication-related operations
type AuthService struct {
	db Database
}

// NewAuthService creates a new authentication service
func NewAuthService(db Database) *AuthService {
	return &AuthService{
		db: db,
	}
}

// Register registers a new user
func (s *AuthService) Register(user *models.User) error {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedPassword)

	// Insert user
	query := `
		INSERT INTO users (username, email, password_hash, first_name, last_name,
						 tenant_id, role_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id
	`
	err = s.db.QueryRow(
		query,
		user.Username, user.Email, user.PasswordHash,
		user.FirstName, user.LastName, user.TenantID,
		user.RoleID, user.IsActive,
	).Scan(&user.ID)

	if err != nil {
		return err
	}

	return nil
}

// Login authenticates a user and creates a session
func (s *AuthService) Login(username, password, deviceInfo, ipAddress string) (*models.Session, error) {
	// Get user
	var user models.User
	query := `
		SELECT id, username, email, password_hash, first_name, last_name,
			   tenant_id, role_id, is_active, last_login
		FROM users
		WHERE username = $1
	`
	err := s.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.TenantID, &user.RoleID,
		&user.IsActive, &user.LastLogin,
	)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is inactive")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		// Log failed attempt
		s.logLoginAttempt(user.ID, ipAddress, false)
		return nil, errors.New("invalid credentials")
	}

	// Log successful attempt
	s.logLoginAttempt(user.ID, ipAddress, true)

	// Update last login
	_, err = s.db.Exec("UPDATE users SET last_login = NOW() WHERE id = $1", user.ID)
	if err != nil {
		return nil, err
	}

	// Create session
	session := &models.Session{
		UserID:       user.ID,
		Token:        s.generateToken(),
		DeviceInfo:   deviceInfo,
		IPAddress:    ipAddress,
		LastActivity: time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	query = `
		INSERT INTO sessions (user_id, token, device_info, ip_address,
							last_activity, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id
	`
	err = s.db.QueryRow(
		query,
		session.UserID, session.Token, session.DeviceInfo,
		session.IPAddress, session.LastActivity, session.ExpiresAt,
	).Scan(&session.ID)

	if err != nil {
		return nil, err
	}

	return session, nil
}

// Logout invalidates a user's session
func (s *AuthService) Logout(sessionID int) error {
	query := "DELETE FROM sessions WHERE id = $1"
	_, err := s.db.Exec(query, sessionID)
	return err
}

// ValidateSession validates a session token
func (s *AuthService) ValidateSession(token string) (*models.Session, error) {
	var session models.Session
	query := `
		SELECT id, user_id, token, device_info, ip_address,
			   last_activity, expires_at, created_at
		FROM sessions
		WHERE token = $1 AND expires_at > NOW()
	`
	err := s.db.QueryRow(query, token).Scan(
		&session.ID, &session.UserID, &session.Token,
		&session.DeviceInfo, &session.IPAddress,
		&session.LastActivity, &session.ExpiresAt,
		&session.CreatedAt,
	)
	if err != nil {
		return nil, errors.New("invalid or expired session")
	}

	// Update last activity
	_, err = s.db.Exec("UPDATE sessions SET last_activity = NOW() WHERE id = $1", session.ID)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// GetUserPermissions retrieves a user's permissions based on their role
func (s *AuthService) GetUserPermissions(userID int) ([]string, error) {
	var permissions []string
	query := `
		SELECT p.name
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN users u ON u.role_id = rp.role_id
		WHERE u.id = $1
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// HasPermission checks if a user has a specific permission
func (s *AuthService) HasPermission(userID int, permission string) (bool, error) {
	permissions, err := s.GetUserPermissions(userID)
	if err != nil {
		return false, err
	}

	for _, p := range permissions {
		if p == permission {
			return true, nil
		}
	}

	return false, nil
}

// RequestPasswordReset initiates a password reset process
func (s *AuthService) RequestPasswordReset(email string) error {
	// Get user
	var userID int
	err := s.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		// Don't reveal if email exists
		return nil
	}

	// Generate reset token
	token := s.generateToken()
	expiresAt := time.Now().Add(1 * time.Hour)

	// Save reset request
	query := `
		INSERT INTO password_resets (user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err = s.db.Exec(query, userID, token, expiresAt)
	if err != nil {
		return err
	}

	// TODO: Send reset email
	return nil
}

// ResetPassword resets a user's password using a reset token
func (s *AuthService) ResetPassword(token, newPassword string) error {
	// Get reset request
	var reset models.PasswordReset
	query := `
		SELECT id, user_id, token, expires_at, used
		FROM password_resets
		WHERE token = $1 AND expires_at > NOW() AND used = false
	`
	err := s.db.QueryRow(query, token).Scan(
		&reset.ID, &reset.UserID, &reset.Token,
		&reset.ExpiresAt, &reset.Used,
	)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password and mark reset as used
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE users SET password_hash = $1 WHERE id = $2", string(hashedPassword), reset.UserID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("UPDATE password_resets SET used = true WHERE id = $1", reset.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Helper functions

func (s *AuthService) generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (s *AuthService) logLoginAttempt(userID int, ipAddress string, success bool) error {
	query := `
		INSERT INTO login_attempts (user_id, ip_address, success, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err := s.db.Exec(query, userID, ipAddress, success)
	return err
} 
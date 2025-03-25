package database

import (
	"database/sql"
	"errors"
	"time"

	"awesomeProject/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// CreateDefaultAdmin creates the default admin account if it doesn't exist
func CreateDefaultAdmin(db *sql.DB) error {
	// Check if admin exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // Admin already exists
	}

	// Create default admin account
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &models.User{
		Username:  "admin",
		Email:     "admin@example.com",
		Password:  string(hashedPassword),
		Role:      "admin",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert admin user
	_, err = db.Exec(`
		INSERT INTO users (
			username, email, password, role, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		admin.Username,
		admin.Email,
		admin.Password,
		admin.Role,
		admin.Status,
		admin.CreatedAt,
		admin.UpdatedAt,
	)

	if err != nil {
		return err
	}

	// Create default tenant
	tenant := &models.Tenant{
		Name:      "Default Tenant",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert default tenant
	var tenantID int
	err = db.QueryRow(`
		INSERT INTO tenants (name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`,
		tenant.Name,
		tenant.Status,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	).Scan(&tenantID)

	if err != nil {
		return err
	}

	// Create default subscription for admin
	subscription := &models.Subscription{
		UserID:    1, // Admin user ID
		TenantID:  tenantID,
		Plan:      "enterprise",
		Status:    "active",
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(1, 0, 0), // 1 year subscription
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert default subscription
	_, err = db.Exec(`
		INSERT INTO subscriptions (
			user_id, tenant_id, plan, status, start_date, end_date, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		subscription.UserID,
		subscription.TenantID,
		subscription.Plan,
		subscription.Status,
		subscription.StartDate,
		subscription.EndDate,
		subscription.CreatedAt,
		subscription.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

// ResetAdminPassword resets the admin password
func ResetAdminPassword(db *sql.DB, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	result, err := db.Exec(`
		UPDATE users
		SET password = $1, updated_at = $2
		WHERE role = 'admin'
	`,
		string(hashedPassword),
		time.Now(),
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no admin user found")
	}

	return nil
}

// GetAdminUser retrieves the admin user
func GetAdminUser(db *sql.DB) (*models.User, error) {
	var user models.User
	err := db.QueryRow(`
		SELECT id, username, email, role, status, created_at, updated_at
		FROM users
		WHERE role = 'admin'
		LIMIT 1
	`).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateAdminUser updates the admin user information
func UpdateAdminUser(db *sql.DB, user *models.User) error {
	_, err := db.Exec(`
		UPDATE users
		SET username = $1, email = $2, status = $3, updated_at = $4
		WHERE role = 'admin'
	`,
		user.Username,
		user.Email,
		user.Status,
		time.Now(),
	)

	return err
} 
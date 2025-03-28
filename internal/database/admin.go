package database

import (
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"saas-chat-system/internal/models"
)

// CreateDefaultAdmin creates the default admin account if it doesn't exist
func CreateDefaultAdmin(db *sql.DB) error {
	// Check if admin exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE role_id = 1").Scan(&count)
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
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: string(hashedPassword),
		FirstName:    "Admin",
		LastName:     "User",
		RoleID:       1,  // Admin role
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Insert admin user
	_, err = db.Exec(`
		INSERT INTO users (
			username, email, password_hash, first_name, last_name, role_id, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		admin.Username,
		admin.Email,
		admin.PasswordHash,
		admin.FirstName,
		admin.LastName,
		admin.RoleID,
		admin.IsActive,
		admin.CreatedAt,
		admin.UpdatedAt,
	)

	if err != nil {
		return err
	}

	// Create default tenant
	tenant := &models.Tenant{
		Name:      "Default Tenant",
		CreatedAt: time.Now(),
	}

	// Insert default tenant
	var tenantID int
	err = db.QueryRow(`
		INSERT INTO tenants (name, created_at)
		VALUES ($1, $2)
		RETURNING id
	`,
		tenant.Name,
		tenant.CreatedAt,
	).Scan(&tenantID)

	if err != nil {
		return err
	}

	// Create default subscription for admin
	subscription := &models.Subscription{
		ID:           "1",
		TenantID:     "1", // Tenant ID as string
		Plan:         models.PlanEnterprise,
		Status:       "active",
		StartDate:    time.Now(),
		EndDate:      time.Now().AddDate(1, 0, 0), // 1 year subscription
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Insert default subscription
	_, err = db.Exec(`
		INSERT INTO subscriptions (
			id, tenant_id, plan, status, start_date, end_date, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		subscription.ID,
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
		SET password_hash = $1, updated_at = $2
		WHERE role_id = 1
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
		SELECT id, username, email, first_name, last_name, role_id, is_active, created_at, updated_at
		FROM users
		WHERE role_id = 1
		LIMIT 1
	`).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.RoleID,
		&user.IsActive,
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
		SET username = $1, email = $2, first_name = $3, last_name = $4, is_active = $5, updated_at = $6
		WHERE role_id = 1
	`,
		user.Username,
		user.Email,
		user.FirstName,
		user.LastName,
		user.IsActive,
		time.Now(),
	)

	return err
}

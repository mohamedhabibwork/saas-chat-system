package initdata

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func InitData() {
	// Connect to database
	db, err := sql.Open("postgres", "postgres://user:password@localhost:5432/dbname?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize data
	if err := initDatabase(db); err != nil {
		log.Fatal(err)
	}

	log.Println("Data initialization completed successfully")
}

func initDatabase(db *sql.DB) error {
	// Create default roles
	roles := []struct {
		name        string
		description string
	}{
		{"admin", "System administrator with full access"},
		{"user", "Regular user with basic access"},
		{"moderator", "Channel moderator with limited admin access"},
		{"bot", "Bot user with limited access"},
	}

	for _, role := range roles {
		_, err := db.Exec(`
			INSERT INTO roles (name, description, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())
			ON CONFLICT (name) DO NOTHING
		`, role.name, role.description)
		if err != nil {
			return fmt.Errorf("error creating role %s: %v", role.name, err)
		}
	}

	// Create default permissions
	permissions := []struct {
		name        string
		description string
	}{
		{"create_channel", "Create new channels"},
		{"delete_channel", "Delete channels"},
		{"manage_users", "Manage user accounts"},
		{"manage_roles", "Manage roles and permissions"},
		{"manage_subscriptions", "Manage subscriptions"},
		{"view_analytics", "View analytics and statistics"},
		{"manage_bots", "Manage bot configurations"},
		{"upload_files", "Upload files"},
		{"download_files", "Download files"},
		{"share_files", "Share files with other users"},
	}

	for _, permission := range permissions {
		_, err := db.Exec(`
			INSERT INTO permissions (name, description, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())
			ON CONFLICT (name) DO NOTHING
		`, permission.name, permission.description)
		if err != nil {
			return fmt.Errorf("error creating permission %s: %v", permission.name, err)
		}
	}

	// Create default subscription plans
	plans := []struct {
		name        string
		description string
		price       float64
		interval    string
		features    string
		limits      string
	}{
		{
			"free",
			"Free plan with basic features",
			0,
			"monthly",
			`{"video_chat": false, "screen_sharing": false, "file_sharing": true, "bot_integration": false}`,
			`{"max_storage": 1073741824, "max_files": 100, "max_daily_uploads": 10, "max_file_size": 10485760}`,
		},
		{
			"basic",
			"Basic plan with enhanced features",
			9.99,
			"monthly",
			`{"video_chat": true, "screen_sharing": false, "file_sharing": true, "bot_integration": false}`,
			`{"max_storage": 5368709120, "max_files": 1000, "max_daily_uploads": 50, "max_file_size": 52428800}`,
		},
		{
			"pro",
			"Professional plan with advanced features",
			29.99,
			"monthly",
			`{"video_chat": true, "screen_sharing": true, "file_sharing": true, "bot_integration": true}`,
			`{"max_storage": 21474836480, "max_files": 10000, "max_daily_uploads": 200, "max_file_size": 104857600}`,
		},
		{
			"enterprise",
			"Enterprise plan with all features",
			99.99,
			"monthly",
			`{"video_chat": true, "screen_sharing": true, "file_sharing": true, "bot_integration": true, "custom_bot_development": true, "priority_support": true, "dedicated_server": true}`,
			`{"max_storage": 107374182400, "max_files": 100000, "max_daily_uploads": 1000, "max_file_size": 524288000}`,
		},
	}

	for _, plan := range plans {
		_, err := db.Exec(`
			INSERT INTO plans (name, description, price, interval, features, limits, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			ON CONFLICT (name) DO NOTHING
		`, plan.name, plan.description, plan.price, plan.interval, plan.features, plan.limits)
		if err != nil {
			return fmt.Errorf("error creating plan %s: %v", plan.name, err)
		}
	}

	// Create default tenant
	var tenantID int
	err := db.QueryRow(`
		INSERT INTO tenants (name, status, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id
		ON CONFLICT (name) DO UPDATE SET id = EXCLUDED.id RETURNING id
	`, "Default Tenant", "active").Scan(&tenantID)
	if err != nil {
		return fmt.Errorf("error creating default tenant: %v", err)
	}

	// Create default admin user
	hashedPassword := "$2a$10$your_hashed_password_here" // Replace with actual hashed password
	_, err = db.Exec(`
		INSERT INTO users (
			username, email, password, role, status,
			tenant_id, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (email) DO NOTHING
	`,
		"admin",
		"admin@example.com",
		hashedPassword,
		"admin",
		"active",
		tenantID,
	)
	if err != nil {
		return fmt.Errorf("error creating admin user: %v", err)
	}

	// Create default subscription for admin
	_, err = db.Exec(`
		INSERT INTO subscriptions (
			user_id, tenant_id, plan, status,
			start_date, end_date, created_at, updated_at
		)
		SELECT u.id, $1, $2, $3, NOW(), NOW() + INTERVAL '1 year', NOW(), NOW()
		FROM users u
		WHERE u.email = $4
		ON CONFLICT (user_id) DO NOTHING
	`,
		tenantID,
		"enterprise",
		"active",
		"admin@example.com",
	)
	if err != nil {
		return fmt.Errorf("error creating admin subscription: %v", err)
	}

	// Create default channels
	channels := []struct {
		name        string
		description string
		type_       string
	}{
		{"general", "General discussion channel", "public"},
		{"announcements", "System announcements", "public"},
		{"support", "Customer support channel", "public"},
		{"admin", "Administrative channel", "private"},
	}

	for _, channel := range channels {
		var channelID int
		err := db.QueryRow(`
			INSERT INTO channels (
				name, description, type, created_by,
				tenant_id, settings, created_at, updated_at
			)
			SELECT $1, $2, $3, u.id, $4, '{}', NOW(), NOW()
			FROM users u
			WHERE u.email = 'admin@example.com'
			RETURNING id
		`, channel.name, channel.description, channel.type_, tenantID).Scan(&channelID)
		if err != nil {
			return fmt.Errorf("error creating channel %s: %v", channel.name, err)
		}

		// Add admin as member of the channel
		_, err = db.Exec(`
			INSERT INTO channel_members (
				channel_id, user_id, role, joined_at
			)
			SELECT $1, u.id, 'admin', NOW()
			FROM users u
			WHERE u.email = 'admin@example.com'
			ON CONFLICT (channel_id, user_id) DO NOTHING
		`, channelID)
		if err != nil {
			return fmt.Errorf("error adding admin to channel %s: %v", channel.name, err)
		}
	}

	return nil
}

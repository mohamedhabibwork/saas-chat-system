package services

import (
	"saas-chat-system/internal/models"
)

// RoleService handles role-related operations
type RoleService struct {
	db Database
}

// NewRoleService creates a new role service
func NewRoleService(db Database) *RoleService {
	return &RoleService{
		db: db,
	}
}

// CreateRole creates a new role
func (s *RoleService) CreateRole(role *models.Role) error {
	query := `
		INSERT INTO roles (name, description, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id
	`
	err := s.db.QueryRow(query, role.Name, role.Description).Scan(&role.ID)
	if err != nil {
		return err
	}

	return nil
}

// GetRole retrieves a role by ID
func (s *RoleService) GetRole(roleID int) (*models.Role, error) {
	var role models.Role
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM roles
		WHERE id = $1
	`
	err := s.db.QueryRow(query, roleID).Scan(
		&role.ID, &role.Name, &role.Description,
		&role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

// ListRoles retrieves all roles
func (s *RoleService) ListRoles() ([]models.Role, error) {
	var roles []models.Role
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM roles
		ORDER BY name ASC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var role models.Role
		err := rows.Scan(
			&role.ID, &role.Name, &role.Description,
			&role.CreatedAt, &role.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// UpdateRole updates an existing role
func (s *RoleService) UpdateRole(role *models.Role) error {
	query := `
		UPDATE roles
		SET name = $1,
			description = $2,
			updated_at = NOW()
		WHERE id = $3
	`
	_, err := s.db.Exec(query, role.Name, role.Description, role.ID)
	return err
}

// DeleteRole deletes a role
func (s *RoleService) DeleteRole(roleID int) error {
	query := "DELETE FROM roles WHERE id = $1"
	_, err := s.db.Exec(query, roleID)
	return err
}

// AssignPermissions assigns permissions to a role
func (s *RoleService) AssignPermissions(roleID int, permissionIDs []int) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	// Delete existing permissions
	_, err = tx.Exec("DELETE FROM role_permissions WHERE role_id = $1", roleID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert new permissions
	for _, permissionID := range permissionIDs {
		_, err = tx.Exec(
			"INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)",
			roleID, permissionID,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// GetRolePermissions retrieves all permissions assigned to a role
func (s *RoleService) GetRolePermissions(roleID int) ([]models.Permission, error) {
	var permissions []models.Permission
	query := `
		SELECT p.id, p.name, p.description, p.created_at, p.updated_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.name ASC
	`
	rows, err := s.db.Query(query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var permission models.Permission
		err := rows.Scan(
			&permission.ID, &permission.Name,
			&permission.Description, &permission.CreatedAt,
			&permission.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// CreatePermission creates a new permission
func (s *RoleService) CreatePermission(permission *models.Permission) error {
	query := `
		INSERT INTO permissions (name, description, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id
	`
	err := s.db.QueryRow(query, permission.Name, permission.Description).Scan(&permission.ID)
	if err != nil {
		return err
	}

	return nil
}

// GetPermission retrieves a permission by ID
func (s *RoleService) GetPermission(permissionID int) (*models.Permission, error) {
	var permission models.Permission
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM permissions
		WHERE id = $1
	`
	err := s.db.QueryRow(query, permissionID).Scan(
		&permission.ID, &permission.Name,
		&permission.Description, &permission.CreatedAt,
		&permission.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &permission, nil
}

// ListPermissions retrieves all permissions
func (s *RoleService) ListPermissions() ([]models.Permission, error) {
	var permissions []models.Permission
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM permissions
		ORDER BY name ASC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var permission models.Permission
		err := rows.Scan(
			&permission.ID, &permission.Name,
			&permission.Description, &permission.CreatedAt,
			&permission.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// UpdatePermission updates an existing permission
func (s *RoleService) UpdatePermission(permission *models.Permission) error {
	query := `
		UPDATE permissions
		SET name = $1,
			description = $2,
			updated_at = NOW()
		WHERE id = $3
	`
	_, err := s.db.Exec(query, permission.Name, permission.Description, permission.ID)
	return err
}

// DeletePermission deletes a permission
func (s *RoleService) DeletePermission(permissionID int) error {
	query := "DELETE FROM permissions WHERE id = $1"
	_, err := s.db.Exec(query, permissionID)
	return err
}

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"saas-chat-system/internal/models"
	"saas-chat-system/internal/services"
)

// CreateRoleRequest represents the role creation request body
type CreateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateRoleRequest represents the role update request body
type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AssignPermissionsRequest represents the permission assignment request body
type AssignPermissionsRequest struct {
	PermissionIDs []int `json:"permission_ids"`
}

// @Summary      List roles
// @Description  Get a list of all roles
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Success      200 {array} models.Role "List of roles"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/roles [get]
func HandleListRoles(roleService *services.RoleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		roles, err := roleService.ListRoles()
		if err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, roles, "", http.StatusOK)
	}
}

// @Summary      Create role
// @Description  Create a new role
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        role body CreateRoleRequest true "Role details"
// @Success      201 {object} models.Role "Role created successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/roles [post]
func HandleCreateRole(roleService *services.RoleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req CreateRoleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Name == "" {
			sendResponse(w, false, nil, "Missing role name", http.StatusBadRequest)
			return
		}

		// Create role
		role := &models.Role{
			Name:        req.Name,
			Description: req.Description,
		}

		if err := roleService.CreateRole(role); err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, role, "", http.StatusCreated)
	}
}

// @Summary      Update role
// @Description  Update an existing role
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        id query integer true "Role ID"
// @Param        role body UpdateRoleRequest true "Role details"
// @Success      200 {object} models.Role "Role updated successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/roles/{id} [put]
func HandleUpdateRole(roleService *services.RoleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get role ID from query parameters
		roleID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			sendResponse(w, false, nil, "Invalid role ID", http.StatusBadRequest)
			return
		}

		var req UpdateRoleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Name == "" {
			sendResponse(w, false, nil, "Missing role name", http.StatusBadRequest)
			return
		}

		// Update role
		role := &models.Role{
			ID:          roleID,
			Name:        req.Name,
			Description: req.Description,
		}

		if err := roleService.UpdateRole(role); err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, role, "", http.StatusOK)
	}
}

// @Summary      Delete role
// @Description  Delete a role
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        id query integer true "Role ID"
// @Success      200 {object} map[string]interface{} "Role deleted successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/roles/{id} [delete]
func HandleDeleteRole(roleService *services.RoleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get role ID from query parameters
		roleID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			sendResponse(w, false, nil, "Invalid role ID", http.StatusBadRequest)
			return
		}

		if err := roleService.DeleteRole(roleID); err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, nil, "", http.StatusOK)
	}
}

// @Summary      List permissions
// @Description  Get a list of all permissions
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Success      200 {array} models.Permission "List of permissions"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/permissions [get]
func HandleListPermissions(roleService *services.RoleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		permissions, err := roleService.ListPermissions()
		if err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, permissions, "", http.StatusOK)
	}
}

// @Summary      Assign permissions
// @Description  Assign permissions to a role
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        role_id query integer true "Role ID"
// @Param        request body AssignPermissionsRequest true "Permission IDs"
// @Success      200 {object} map[string]interface{} "Permissions assigned successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/roles/{role_id}/permissions [post]
func HandleAssignPermissions(roleService *services.RoleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get role ID from query parameters
		roleID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			sendResponse(w, false, nil, "Invalid role ID", http.StatusBadRequest)
			return
		}

		var req AssignPermissionsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if len(req.PermissionIDs) == 0 {
			sendResponse(w, false, nil, "No permissions specified", http.StatusBadRequest)
			return
		}

		if err := roleService.AssignPermissions(roleID, req.PermissionIDs); err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, nil, "", http.StatusOK)
	}
}

// HandleGetRolePermissions handles retrieving permissions for a role
func HandleGetRolePermissions(roleService *services.RoleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get role ID from query parameters
		roleID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			sendResponse(w, false, nil, "Invalid role ID", http.StatusBadRequest)
			return
		}

		permissions, err := roleService.GetRolePermissions(roleID)
		if err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, permissions, "", http.StatusOK)
	}
}

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"awesomeProject/internal/models"
	"awesomeProject/internal/services"
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

// HandleListRoles handles listing all roles
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

// HandleCreateRole handles creating a new role
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

// HandleUpdateRole handles updating an existing role
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

// HandleDeleteRole handles deleting a role
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

// HandleListPermissions handles listing all permissions
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

// HandleAssignPermissions handles assigning permissions to a role
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
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"awesomeProject/internal/database"
	"awesomeProject/internal/services"
	"awesomeProject/internal/websocket"
	"awesomeProject/internal/handlers"
	"awesomeProject/internal/middleware"
)

var (
	port = flag.String("port", "8080", "Port to listen on")
	host = flag.String("host", "localhost", "Host to listen on")
)

func main() {
	flag.Parse()

	// Initialize database
	db, err := database.NewPostgresDB(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Setup database tables
	if err := db.SetupTables(); err != nil {
		log.Fatalf("Failed to setup database tables: %v", err)
	}

	// Create default roles and permissions
	if err := db.CreateDefaultRoles(); err != nil {
		log.Fatalf("Failed to create default roles: %v", err)
	}
	if err := db.CreateDefaultPermissions(); err != nil {
		log.Fatalf("Failed to create default permissions: %v", err)
	}
	if err := db.AssignDefaultPermissions(); err != nil {
		log.Fatalf("Failed to assign default permissions: %v", err)
	}

	// Initialize services
	authService := services.NewAuthService(db)
	subscriptionService := services.NewSubscriptionService(db)
	roleService := services.NewRoleService(db)
	storageService, err := services.NewStorageService(db)
	if err != nil {
		log.Fatalf("Failed to initialize storage service: %v", err)
	}

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize bot service
	botService := services.NewBotService(hub, db)

	// Setup routes
	router := http.NewServeMux()
	setupRoutes(router, authService, subscriptionService, roleService, storageService)

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", *host, *port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s:%s", *host, *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := server.Close(); err != nil {
		log.Printf("Error closing server: %v", err)
	}
}

func setupRoutes(
	hub *websocket.Hub,
	authService *services.AuthService,
	subscriptionService *services.SubscriptionService,
	roleService *services.RoleService,
	storageService *services.StorageService,
) http.Handler {
	mux := http.NewServeMux()

	// WebSocket endpoint
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.HandleWebSocket(hub, w, r)
	})

	// Authentication endpoints
	mux.HandleFunc("/api/auth/register", handleRegister(authService))
	mux.HandleFunc("/api/auth/login", handleLogin(authService))
	mux.HandleFunc("/api/auth/logout", handleLogout(authService))
	mux.HandleFunc("/api/auth/reset-password", handleResetPassword(authService))

	// Subscription endpoints
	mux.HandleFunc("/api/subscriptions/plans", handleListPlans(subscriptionService))
	mux.HandleFunc("/api/subscriptions", handleSubscribe(subscriptionService))
	mux.HandleFunc("/api/subscriptions/cancel", handleCancelSubscription(subscriptionService))
	mux.HandleFunc("/api/subscriptions/renew", handleRenewSubscription(subscriptionService))
	mux.HandleFunc("/api/subscriptions/usage", handleGetUsage(subscriptionService))

	// Role endpoints
	mux.HandleFunc("/api/roles", handleListRoles(roleService))
	mux.HandleFunc("/api/roles/create", handleCreateRole(roleService))
	mux.HandleFunc("/api/roles/update", handleUpdateRole(roleService))
	mux.HandleFunc("/api/roles/delete", handleDeleteRole(roleService))
	mux.HandleFunc("/api/roles/permissions", handleListPermissions(roleService))
	mux.HandleFunc("/api/roles/assign-permissions", handleAssignPermissions(roleService))

	// Bot endpoints
	mux.HandleFunc("/api/bots", handleListBots(botService))
	mux.HandleFunc("/api/bots/create", handleCreateBot(botService))
	mux.HandleFunc("/api/bots/update", handleUpdateBot(botService))
	mux.HandleFunc("/api/bots/delete", handleDeleteBot(botService))
	mux.HandleFunc("/api/bots/message", handleSendMessage(botService))
	mux.HandleFunc("/api/bots/stats", handleGetBotStats(botService))

	// File handlers with middleware
	fileHandlers := handlers.NewFileHandlers(storageService)
	fileMiddleware := middleware.NewFileUploadMiddleware(subscriptionService)
	
	mux.HandleFunc("/api/files/upload", fileMiddleware.HandleFileUpload(fileHandlers.HandleUpload))
	mux.HandleFunc("/api/files/download", fileHandlers.HandleDownload)
	mux.HandleFunc("/api/files/delete", fileHandlers.HandleDelete)
	mux.HandleFunc("/api/files/list", fileHandlers.HandleList)

	return mux
}

// Handler functions will be implemented in separate files
func handleRegister(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement registration handler
	}
}

func handleLogin(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement login handler
	}
}

func handleLogout(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement logout handler
	}
}

func handleResetPassword(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement password reset handler
	}
}

func handleListPlans(subscriptionService *services.SubscriptionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement list plans handler
	}
}

func handleSubscribe(subscriptionService *services.SubscriptionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement subscription handler
	}
}

func handleCancelSubscription(subscriptionService *services.SubscriptionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement cancel subscription handler
	}
}

func handleRenewSubscription(subscriptionService *services.SubscriptionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement renew subscription handler
	}
}

func handleGetUsage(subscriptionService *services.SubscriptionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement get usage handler
	}
}

func handleListRoles(roleService *services.RoleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement list roles handler
	}
}

func handleCreateRole(roleService *services.RoleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement create role handler
	}
}

func handleUpdateRole(roleService *services.RoleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement update role handler
	}
}

func handleDeleteRole(roleService *services.RoleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement delete role handler
	}
}

func handleListPermissions(roleService *services.RoleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement list permissions handler
	}
}

func handleAssignPermissions(roleService *services.RoleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement assign permissions handler
	}
}

func handleListBots(botService *services.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement list bots handler
	}
}

func handleCreateBot(botService *services.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement create bot handler
	}
}

func handleUpdateBot(botService *services.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement update bot handler
	}
}

func handleDeleteBot(botService *services.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement delete bot handler
	}
}

func handleSendMessage(botService *services.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement send message handler
	}
}

func handleGetBotStats(botService *services.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement get bot stats handler
	}
} 
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"saas-chat-system/internal/database"
	"saas-chat-system/internal/handlers"
	"saas-chat-system/internal/middleware"
	"saas-chat-system/internal/services"
	"saas-chat-system/internal/websocket"
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

	// Initialize channel service
	channelService := services.NewChannelService(db)

	// Initialize WebRTC service
	webrtcService := services.NewWebRTCService()

	// Initialize user service
	userService := services.NewUserService(db)

	// Initialize chat service
	chatService := services.NewChatService(db)

	// Setup routes
	router := http.NewServeMux()
	setupRoutes(router, hub, authService, subscriptionService, roleService, storageService, botService, channelService, webrtcService, userService, chatService)

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
	router *http.ServeMux,
	hub *websocket.Hub,
	authService *services.AuthService,
	subscriptionService *services.SubscriptionService,
	roleService *services.RoleService,
	storageService *services.StorageService,
	botService *services.BotService,
	channelService *services.ChannelService,
	webrtcService *services.WebRTCService,
	userService *services.UserService,
	chatService *services.ChatService,
) {
	// WebSocket endpoint
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleWebSocket(hub, w, r)
	})

	// Authentication endpoints
	router.HandleFunc("/api/auth/register", handlers.HandleRegister(authService))
	router.HandleFunc("/api/auth/login", handlers.HandleLogin(authService))
	router.HandleFunc("/api/auth/logout", handlers.HandleLogout(authService))
	router.HandleFunc("/api/auth/reset-password", handlers.HandleResetPassword(authService))
	router.HandleFunc("/api/auth/new-password", handlers.HandleNewPassword(authService))

	// Subscription endpoints
	router.HandleFunc("/api/subscriptions/plans", handlers.HandleListPlans(subscriptionService))
	router.HandleFunc("/api/subscriptions", handlers.HandleSubscribe(subscriptionService))
	router.HandleFunc("/api/subscriptions/cancel", handlers.HandleCancelSubscription(subscriptionService))
	router.HandleFunc("/api/subscriptions/renew", handlers.HandleRenewSubscription(subscriptionService))
	router.HandleFunc("/api/subscriptions/usage", handlers.HandleGetUsage(subscriptionService))

	// Role endpoints
	router.HandleFunc("/api/v1/roles", handlers.HandleListRoles(roleService))
	router.HandleFunc("/api/v1/roles", handlers.HandleCreateRole(roleService))
	router.HandleFunc("/api/v1/roles/{id}", handlers.HandleUpdateRole(roleService))
	router.HandleFunc("/api/v1/roles/{id}", handlers.HandleDeleteRole(roleService))
	router.HandleFunc("/api/v1/permissions", handlers.HandleListPermissions(roleService))
	router.HandleFunc("/api/v1/roles/{role_id}/permissions", handlers.HandleAssignPermissions(roleService))

	// Bot endpoints
	router.HandleFunc("/api/v1/bots", handlers.HandleListBots(botService))
	router.HandleFunc("/api/v1/bots", handlers.HandleCreateBot(botService))
	router.HandleFunc("/api/v1/bots/{id}", handlers.HandleUpdateBot(botService))
	router.HandleFunc("/api/v1/bots/{id}", handlers.HandleDeleteBot(botService))
	router.HandleFunc("/api/v1/bots/{id}/message", handlers.HandleSendMessage(botService))
	router.HandleFunc("/api/v1/bots/{id}/stats", handlers.HandleGetBotStats(botService))

	// Channel endpoints
	channelHandlers := handlers.NewChannelHandlers(channelService, webrtcService)
	router.HandleFunc("/api/v1/channels", channelHandlers.HandleCreateChannel)
	router.HandleFunc("/api/v1/channels/{id}/members", channelHandlers.HandleGetMembers)
	router.HandleFunc("/api/v1/channels/{channel_id}/members", channelHandlers.HandleAddMember)
	router.HandleFunc("/api/v1/channels/{channel_id}/members/{user_id}", channelHandlers.HandleRemoveMember)
	router.HandleFunc("/api/v1/channels/{channel_id}/messages", channelHandlers.HandleAddMessage)
	router.HandleFunc("/api/v1/channels/{channel_id}/messages", channelHandlers.HandleGetMessages)
	router.HandleFunc("/api/v1/channels/{channel_id}/webrtc", channelHandlers.HandleWebRTCConnection)
	router.HandleFunc("/api/v1/channels/{channel_id}/webrtc/ice", channelHandlers.HandleWebRTCICECandidate)

	// User endpoints
	userHandlers := handlers.NewUserHandlers(userService)
	router.HandleFunc("/api/v1/users", userHandlers.HandleCreate)
	router.HandleFunc("/api/v1/users/{id}", userHandlers.HandleGet)
	router.HandleFunc("/api/v1/users/{id}", userHandlers.HandleUpdate)
	router.HandleFunc("/api/v1/users/{id}", userHandlers.HandleDelete)
	router.HandleFunc("/api/v1/users", userHandlers.HandleList)

	// Chat endpoints
	chatHandlers := handlers.NewChatHandlers(chatService)
	router.HandleFunc("/api/v1/channels", chatHandlers.HandleCreateChannel)
	router.HandleFunc("/api/v1/channels/{channel_id}/join", chatHandlers.HandleJoinChannel)
	router.HandleFunc("/api/v1/channels/{channel_id}/leave", chatHandlers.HandleLeaveChannel)
	router.HandleFunc("/api/v1/channels/{channel_id}/messages", chatHandlers.HandleGetMessages)
	router.HandleFunc("/api/v1/channels/{channel_id}/messages", chatHandlers.HandleSendMessage)

	// File handlers with middleware
	fileHandlers := handlers.NewFileHandlers(storageService)
	fileMiddleware := middleware.NewFileUploadMiddleware(subscriptionService)

	router.HandleFunc("/api/v1/files/upload", fileMiddleware.HandleFileUpload(fileHandlers.HandleUpload))
	router.HandleFunc("/api/v1/files/download", fileHandlers.HandleDownload)
	router.HandleFunc("/api/v1/files/delete", fileHandlers.HandleDelete)
	router.HandleFunc("/api/v1/files/list", fileHandlers.HandleList)
}

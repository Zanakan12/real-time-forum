package server

import (
	"fmt"
	"handlers"
	"middlewares"
	"net/http"
	"time"
)

func InitServer() {

	// Create a new server instance with specified timeout settings and max header bytes
	server := NewServer(":8080", "localhost.crt", "localhost.key", 10*time.Second, 10*time.Second, 30*time.Second, 2*time.Second, 1<<20) // 1 MB max header size
	// Start the broadcast messages goroutine
	middlewares.SetErrorHandlers(handlers.Err400Handler, handlers.Err500Handler)
	// Add handlers for different routes
	server.Handle("/", handlers.IndexHandler) // Root route
	//server.Handle("/about", handlers.AboutHandler) // About route
	server.Handle("/register", handlers.RegisterHandler)
	server.Handle("/login", handlers.LoginHandler)
	server.Handle("/login-validation", handlers.LoginValidationHandler)
	server.Handle("/register-validation", handlers.RegisterValidationHandler)
	server.Handle("/post-validation", handlers.PostValidationHandler)
	server.Handle("/post-delete-validation", handlers.PostDeleteValidationHandler)
	server.Handle("/post-update-validation", handlers.PostUpdateValidationHandler)
	server.Handle("/comment-validation", handlers.CommentValidationHandler)
	server.Handle("/likes-dislikes-validation", handlers.LikesDislikesValidationHandler)
	server.Handle("/profile", handlers.ProfileHandler)
	server.Handle("/update-name", handlers.UpdateNameHandler)
	server.Handle("/admin", handlers.AdminInterfaceHandler)
	server.Handle("/user-request-validation", handlers.UserValidationRequestHandler)
	server.Handle("/logout", handlers.LogoutHandler)
	server.Handle("/comment-delete-validation", handlers.CommentDeleteValidationHandler)
	server.Handle("/comment-update-validation", handlers.CommentUpdateValidationHandler)
	server.Handle("/moderator", handlers.ModeratorPowerHandler)
	server.Handle("/mod", handlers.ModeratorInterfaceHandler)

	// Ajout des routes pour l'authentification Google
	server.Handle("/google-login", handlers.HandleGoogleLogin)
	server.Handle("/google-callback", handlers.HandleGoogleCallback)
	//Reddit
	server.Handle("/reddit-login", handlers.HandleRedditLogin)
	server.Handle("/reddit-callback", handlers.HandleRedditCallback)
	//Discord
	server.Handle("/dis-login", handlers.HandleDiscordLogin)
	server.Handle("/dis-callback", handlers.HandleDiscordCallback)
	server.Handle("/notifications", handlers.NotificationsHandler)
	server.Handle("/chat", handlers.ChatPageHandler)

	// Errors
	server.Handle("/404", handlers.Err404Handler)
	server.Handle("/429", handlers.Err429Handler)
	// Add middlewares
	server.Use(middlewares.LoggingMiddleware)
	server.Use(middlewares.NotFoundMiddleware)
	server.Use(middlewares.ErrorMiddleware)
	server.Use(middlewares.RateLimitingMiddleware)
	//server.Use(middlewares.AuthMiddleware)

	// Add websocket handler
	http.HandleFunc("/ws", handlers.HandleWebSocket)
	server.Handle("/api/get-user", handlers.GetUserHandler)
	server.Handle("/api/users", handlers.GetUserListHandler)
	http.HandleFunc("/api/chat", handlers.GetChatHistory)

	// Start the server
	if err := server.Start(); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

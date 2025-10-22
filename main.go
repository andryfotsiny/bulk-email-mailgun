package main

import (
	"bulk-email-mailgun/config"
	"bulk-email-mailgun/database"
	"bulk-email-mailgun/handlers"
	"bulk-email-mailgun/middleware"
	"bulk-email-mailgun/services"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Initialiser la configuration
	config.Init()

	// Initialiser SQLite
	if err := database.Init(); err != nil {
		log.Fatal("❌ Erreur initialisation DB:", err)
	}
	defer database.Close()

	// Initialiser le nettoyage automatique des sessions
	middleware.InitCleanup()

	// Initialiser les services
	emailService := services.NewEmailService()
	wsService := services.NewWebSocketService()
	handler := handlers.NewHandler(emailService, wsService)

	// Routes publiques (sans authentification)
	http.HandleFunc("/login", handler.LoginPageHandler)
	http.HandleFunc("/api/login", handler.LoginHandler)

	// Routes protégées (avec authentification)
	http.HandleFunc("/", middleware.AuthMiddleware(handler.IndexHandler))
	http.HandleFunc("/logout", middleware.AuthMiddleware(handler.LogoutHandler))
	http.HandleFunc("/ws", middleware.AuthMiddleware(handler.WebSocketHandler))
	http.HandleFunc("/api/config", middleware.AuthMiddleware(handler.ConfigHandler))
	http.HandleFunc("/api/upload", middleware.AuthMiddleware(handler.UploadHandler))
	http.HandleFunc("/api/send", middleware.AuthMiddleware(handler.SendHandler))
	http.HandleFunc("/api/stats", middleware.AuthMiddleware(handler.StatsHandler))
	http.HandleFunc("/api/history", middleware.AuthMiddleware(handler.HistoryHandler))
	http.HandleFunc("/api/recipients", middleware.AuthMiddleware(handler.RecipientsHandler))
	http.HandleFunc("/api/reset", middleware.AuthMiddleware(handler.ResetDatabaseHandler))

	fmt.Println("Server started on http://localhost:8080")
	fmt.Printf(" Provider: %s\n", config.AppConfig.Provider)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

package main

import (
	"bulk-email-mailgun/config"
	"bulk-email-mailgun/database"
	"bulk-email-mailgun/handlers"
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

	// Initialiser les services
	emailService := services.NewEmailService()
	wsService := services.NewWebSocketService()
	handler := handlers.NewHandler(emailService, wsService)

	// Routes
	http.HandleFunc("/", handler.IndexHandler)
	http.HandleFunc("/ws", handler.WebSocketHandler)
	http.HandleFunc("/api/config", handler.ConfigHandler)
	http.HandleFunc("/api/upload", handler.UploadHandler)
	http.HandleFunc("/api/send", handler.SendHandler)
	http.HandleFunc("/api/stats", handler.StatsHandler)
	http.HandleFunc("/api/history", handler.HistoryHandler)
	http.HandleFunc("/api/recipients", handler.RecipientsHandler)

	fmt.Println("🚀 Server started on http://localhost:8080")
	fmt.Printf("📧 Provider: %s\n", config.AppConfig.Provider)
	fmt.Println("💾 SQLite Database: ./emails.db")
	fmt.Println("\n📊 Endpoints disponibles:")
	fmt.Println("   GET  /              - Interface web")
	fmt.Println("   GET  /api/stats     - Statistiques")
	fmt.Println("   GET  /api/history   - Historique des envois")
	fmt.Println("   GET  /api/recipients - Liste des destinataires")
	fmt.Println("   POST /api/upload    - Upload CSV")
	fmt.Println("   POST /api/send      - Envoyer emails")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

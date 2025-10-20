package main

import (
	"bulk-email-mailgun/config"
	"bulk-email-mailgun/handlers"
	"bulk-email-mailgun/services"
	"fmt"
	"log"
	"net/http"
)

func main() {
	config.Init()

	emailService := services.NewEmailService()
	wsService := services.NewWebSocketService()
	handler := handlers.NewHandler(emailService, wsService)

	http.HandleFunc("/", handler.IndexHandler)
	http.HandleFunc("/ws", handler.WebSocketHandler)
	http.HandleFunc("/api/config", handler.ConfigHandler)
	http.HandleFunc("/api/upload", handler.UploadHandler)
	http.HandleFunc("/api/send", handler.SendHandler)

	fmt.Println("Server started on http://localhost:8080")
	fmt.Printf("Provider: %s\n", config.AppConfig.Provider)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

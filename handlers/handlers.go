package handlers

import (
	"bulk-email-mailgun/config"
	"bulk-email-mailgun/database"
	"bulk-email-mailgun/models"
	"bulk-email-mailgun/services"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

type Handler struct {
	emailService *services.EmailService
	wsService    *services.WebSocketService
	upgrader     websocket.Upgrader
}

func NewHandler(emailService *services.EmailService, wsService *services.WebSocketService) *Handler {
	return &Handler{
		emailService: emailService,
		wsService:    wsService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

func (h *Handler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	h.wsService.AddClient(conn)
	defer h.wsService.RemoveClient(conn)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (h *Handler) ConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "POST" {
		var newConfig models.EmailConfig
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			json.NewEncoder(w).Encode(models.APIResponse{
				Success: false,
				Error:   "Invalid JSON",
			})
			return
		}

		if newConfig.SMTPServer != "" {
			config.AppConfig.SMTPServer = newConfig.SMTPServer
		}
		if newConfig.SMTPPort != 0 {
			config.AppConfig.SMTPPort = newConfig.SMTPPort
		}
		if newConfig.Email != "" {
			config.AppConfig.Email = newConfig.Email
		}
		if newConfig.Password != "" {
			config.AppConfig.Password = newConfig.Password
		}
		if newConfig.Provider != "" {
			config.AppConfig.Provider = newConfig.Provider
		}
		if newConfig.MailgunDomain != "" {
			config.AppConfig.MailgunDomain = newConfig.MailgunDomain
		}
		if newConfig.MailgunAPIKey != "" {
			config.AppConfig.MailgunAPIKey = newConfig.MailgunAPIKey
		}

		json.NewEncoder(w).Encode(models.APIResponse{
			Success: true,
			Message: "Configuration updated",
		})
		return
	}

	json.NewEncoder(w).Encode(models.ConfigResponse{
		SMTPServer:    config.AppConfig.SMTPServer,
		SMTPPort:      config.AppConfig.SMTPPort,
		Email:         config.AppConfig.Email,
		Provider:      config.AppConfig.Provider,
		MailgunDomain: config.AppConfig.MailgunDomain,
	})
}

func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	file, _, err := r.FormFile("file")
	if err != nil {
		json.NewEncoder(w).Encode(models.UploadResponse{
			Success: false,
			Error:   "Upload error",
		})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		json.NewEncoder(w).Encode(models.UploadResponse{
			Success: false,
			Error:   "CSV read error",
		})
		return
	}

	var emails []models.EmailData
	insertedCount := 0

	for i, record := range records {
		if i == 0 {
			continue
		}
		if len(record) >= 2 {
			email := models.EmailData{
				Email: strings.TrimSpace(record[0]),
				Name:  strings.TrimSpace(record[1]),
			}
			if len(record) >= 3 {
				email.Company = strings.TrimSpace(record[2])
			}
			if len(record) >= 4 {
				email.City = strings.TrimSpace(record[3])
			}

			// Insérer dans la DB
			_, err := database.InsertOrGetRecipient(email.Email, email.Name, email.Company, email.City)
			if err == nil {
				insertedCount++
			}

			emails = append(emails, email)
		}
	}

	json.NewEncoder(w).Encode(models.UploadResponse{
		Success: true,
		Count:   len(emails),
		Emails:  emails,
		Message: fmt.Sprintf("%d recipients ajoutés/mis à jour dans la base de données", insertedCount),
	})
}

func (h *Handler) SendHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if config.AppConfig.MailgunDomain == "" || config.AppConfig.MailgunAPIKey == "" {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Error:   "Mailgun not configured",
		})
		return
	}

	var req models.SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Error:   "Invalid data",
		})
		return
	}

	go h.emailService.ProcessEmails(req, h.wsService.GetBroadcastChannel())

	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Message: "Sending started",
	})
}

// StatsHandler retourne les statistiques
func (h *Handler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats, err := database.GetStats()
	if err != nil {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"stats":   stats,
	})
}

// HistoryHandler retourne l'historique des envois
func (h *Handler) HistoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	history, err := database.GetAllEmailSends()
	if err != nil {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"history": history,
	})
}

// RecipientsHandler retourne tous les recipients
func (h *Handler) RecipientsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	recipients, err := database.GetAllRecipients()
	if err != nil {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"recipients": recipients,
	})
}

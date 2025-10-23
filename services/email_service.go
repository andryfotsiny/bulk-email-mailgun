package services

import (
	"bulk-email-mailgun/config"
	"bulk-email-mailgun/database"
	"bulk-email-mailgun/models"
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"gopkg.in/gomail.v2"
)

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

// generateRandomEmail génère un email aléatoire pour Mailgun
func generateRandomEmail() string {
	rand.Seed(time.Now().UnixNano())

	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	length := 10
	result := make([]byte, length)

	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}

	romanticNames := []string{
		"secret.admirer",
		"mystery.lover",
		"anonymous.heart",
		"secret.love",
		"hidden.romance",
		"unknown.angel",
		"mystery.angel",
		"secret.angel",
	}

	randomName := romanticNames[rand.Intn(len(romanticNames))]
	randomSuffix := string(result[:6])

	return fmt.Sprintf("%s.%s@%s", randomName, randomSuffix, config.AppConfig.MailgunDomain)
}

// SendEmailWithProvider envoie un email via le provider choisi
func (s *EmailService) SendEmailWithProvider(to, subject, body, provider string) (string, error) {
	if provider == "mailgun" {
		senderEmail, err := s.sendWithMailgun(to, subject, body)
		return senderEmail, err
	}
	// Gmail par défaut
	return config.AppConfig.Email, s.sendWithSMTP(to, subject, body)
}

func (s *EmailService) sendWithMailgun(to, subject, body string) (string, error) {
	if config.AppConfig.MailgunDomain == "" || config.AppConfig.MailgunAPIKey == "" {
		return "", fmt.Errorf("mailgun not configured")
	}

	mg := mailgun.NewMailgun(config.AppConfig.MailgunDomain, config.AppConfig.MailgunAPIKey)

	// Générer un email aléatoire
	randomEmail := generateRandomEmail()
	displayName := "Admirateur Secret"
	fromAddress := fmt.Sprintf("%s <%s>", displayName, randomEmail)

	message := mg.NewMessage(
		fromAddress,
		subject,
		"",
		to,
	)
	message.SetHtml(body)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	resp, id, err := mg.Send(ctx, message)

	if err != nil {
		fmt.Printf("❌ Erreur envoi Mailgun à %s: %v\n", to, err)
		return randomEmail, err
	}

	fmt.Printf("✅ Email envoyé via Mailgun depuis %s → %s (ID: %s, Response: %s)\n", randomEmail, to, id, resp)
	return randomEmail, nil
}

func (s *EmailService) sendWithSMTP(to, subject, body string) error {
	if config.AppConfig.Email == "" || config.AppConfig.Password == "" {
		return fmt.Errorf("gmail not configured")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", config.AppConfig.Email)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(
		config.AppConfig.SMTPServer,
		config.AppConfig.SMTPPort,
		config.AppConfig.Email,
		config.AppConfig.Password,
	)

	if config.AppConfig.SMTPPort == 465 {
		d.SSL = true
	}

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	err := d.DialAndSend(m)
	if err != nil {
		fmt.Printf("❌ Erreur envoi Gmail à %s: %v\n", to, err)
		return err
	}

	fmt.Printf("✅ Email envoyé via Gmail depuis %s → %s\n", config.AppConfig.Email, to)
	return nil
}

func (s *EmailService) ProcessEmails(req models.SendRequest, broadcast chan<- models.ProgressUpdate) {
	total := len(req.Emails)
	sent := 0
	failed := 0

	// Déterminer le provider (mailgun par défaut)
	provider := req.Provider
	if provider == "" {
		provider = "mailgun"
	}

	fmt.Printf("📧 Provider sélectionné: %s\n", provider)

	// 1. Créer le contenu d'email une seule fois
	contentID, err := database.InsertEmailContent(req.Subject, req.Body)
	if err != nil {
		fmt.Printf("❌ Erreur création contenu: %v\n", err)
		return
	}
	fmt.Printf("📝 Contenu d'email créé (ID: %d)\n", contentID)

	concurrency := 10
	if provider == "mailgun" {
		concurrency = 50
	}

	semaphore := make(chan struct{}, concurrency)

	for i, emailData := range req.Emails {
		semaphore <- struct{}{}

		go func(index int, data models.EmailData) {
			defer func() { <-semaphore }()

			// 2. Insérer/récupérer le recipient
			recipientID, err := database.InsertOrGetRecipient(data.Email)
			if err != nil {
				fmt.Printf("❌ Erreur recipient: %v\n", err)
				failed++
				broadcast <- models.ProgressUpdate{
					Current:    index + 1,
					Total:      total,
					Sent:       sent,
					Failed:     failed,
					Percentage: float64(index+1) / float64(total) * 100,
				}
				return
			}

			// 3. Personnaliser le body (seulement {{email}})
			body := strings.ReplaceAll(req.Body, "{{email}}", data.Email)

			// 4. Envoyer l'email via le provider choisi
			senderEmail, sendErr := s.SendEmailWithProvider(data.Email, req.Subject, body, provider)

			// 5. Insérer/récupérer le sender
			displayName := "Expéditeur"
			if provider == "mailgun" {
				displayName = "Admirateur Secret"
			}

			senderID, err := database.InsertOrGetSender(senderEmail, displayName)
			if err != nil {
				fmt.Printf("❌ Erreur sender: %v\n", err)
				failed++
				broadcast <- models.ProgressUpdate{
					Current:    index + 1,
					Total:      total,
					Sent:       sent,
					Failed:     failed,
					Percentage: float64(index+1) / float64(total) * 100,
				}
				return
			}

			// 6. Déterminer le status
			status := "sent"
			errorMessage := ""

			if sendErr != nil {
				status = "failed"
				errorMessage = sendErr.Error()
				failed++
			} else {
				sent++
			}

			// 7. Enregistrer dans la DB
			if err := database.InsertEmailSend(contentID, senderID, recipientID, status, errorMessage); err != nil {
				fmt.Printf("❌ Erreur enregistrement DB: %v\n", err)
			}

			// 8. Broadcaster la progression
			broadcast <- models.ProgressUpdate{
				Current:    index + 1,
				Total:      total,
				Sent:       sent,
				Failed:     failed,
				Percentage: float64(index+1) / float64(total) * 100,
			}

			// 9. Délai entre les envois
			delay := 500 * time.Millisecond
			if provider == "mailgun" {
				delay = 100 * time.Millisecond
			}
			time.Sleep(delay)
		}(i, emailData)
	}

	// Attendre que tous les envois soient terminés
	for i := 0; i < cap(semaphore); i++ {
		semaphore <- struct{}{}
	}

	fmt.Printf("\n🎉 Terminé! Total: %d | Envoyés: %d | Échoués: %d\n", total, sent, failed)
}

func (s *EmailService) personalizeBody(body string, data models.EmailData) string {
	body = strings.ReplaceAll(body, "{{email}}", data.Email)
	return body
}

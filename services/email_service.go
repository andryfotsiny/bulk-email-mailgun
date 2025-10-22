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

// generateRandomEmail génère un email aléatoire
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

func (s *EmailService) SendEmail(to, subject, body string) error {
	if config.AppConfig.Provider == "mailgun" {
		return s.sendWithMailgun(to, subject, body)
	}
	return s.sendWithSMTP(to, subject, body)
}

func (s *EmailService) sendWithMailgun(to, subject, body string) error {
	if config.AppConfig.MailgunDomain == "" || config.AppConfig.MailgunAPIKey == "" {
		return fmt.Errorf("mailgun not configured")
	}

	mg := mailgun.NewMailgun(config.AppConfig.MailgunDomain, config.AppConfig.MailgunAPIKey)

	randomEmail := generateRandomEmail()
	displayName := "Admirateur Secret "
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
		fmt.Printf("❌ Erreur envoi à %s: %v\n", to, err)
		return err
	}

	fmt.Printf("Email envoyé depuis %s → %s (ID: %s, Response: %s)\n", randomEmail, to, id, resp)
	return nil
}

func (s *EmailService) sendWithSMTP(to, subject, body string) error {
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

	return d.DialAndSend(m)
}

func (s *EmailService) ProcessEmails(req models.SendRequest, broadcast chan<- models.ProgressUpdate) {
	total := len(req.Emails)
	sent := 0
	failed := 0

	// 1. Créer le contenu d'email une seule fois
	contentID, err := database.InsertEmailContent(req.Subject, req.Body)
	if err != nil {
		fmt.Printf("❌ Erreur création contenu: %v\n", err)
		return
	}
	fmt.Printf("Contenu d'email créé (ID: %d)\n", contentID)

	concurrency := 10
	if config.AppConfig.Provider == "mailgun" {
		concurrency = 50
	}

	semaphore := make(chan struct{}, concurrency)

	for i, emailData := range req.Emails {
		semaphore <- struct{}{}

		go func(index int, data models.EmailData) {
			defer func() { <-semaphore }()

			// 2. Insérer/récupérer le recipient
			recipientID, err := database.InsertOrGetRecipient(
				data.Email,
				data.Name,
				data.Company,
				data.City,
			)
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

			randomEmail := generateRandomEmail()
			senderID, err := database.InsertOrGetSender(randomEmail, "Admirateur Secret")
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

			body := s.personalizeBody(req.Body, data)

			status := "sent"
			errorMessage := ""

			if err := s.SendEmail(data.Email, req.Subject, body); err != nil {
				status = "failed"
				errorMessage = err.Error()
				failed++
			} else {
				sent++
			}

			// 6. Enregistrer dans la DB
			if err := database.InsertEmailSend(contentID, senderID, recipientID, status, errorMessage); err != nil {
				fmt.Printf("❌ Erreur enregistrement DB: %v\n", err)
			}

			// 7. Broadcaster la progression
			broadcast <- models.ProgressUpdate{
				Current:    index + 1,
				Total:      total,
				Sent:       sent,
				Failed:     failed,
				Percentage: float64(index+1) / float64(total) * 100,
			}

			delay := 500 * time.Millisecond
			if config.AppConfig.Provider == "mailgun" {
				delay = 100 * time.Millisecond
			}
			time.Sleep(delay)
		}(i, emailData)
	}

	for i := 0; i < cap(semaphore); i++ {
		semaphore <- struct{}{}
	}

	fmt.Printf("\n Terminé! Total: %d | Envoyés: %d | Échoués: %d\n", total, sent, failed)
}

func (s *EmailService) personalizeBody(body string, data models.EmailData) string {
	body = strings.ReplaceAll(body, "{{name}}", data.Name)
	body = strings.ReplaceAll(body, "{{company}}", data.Company)
	body = strings.ReplaceAll(body, "{{city}}", data.City)
	body = strings.ReplaceAll(body, "{{email}}", data.Email)
	return body
}

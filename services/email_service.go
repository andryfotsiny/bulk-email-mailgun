package services

import (
	"bulk-email-mailgun/config"
	"bulk-email-mailgun/database"
	"bulk-email-mailgun/models"
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/resend/resend-go/v2"
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

// ✅ NOUVEAU: Construire l'email Resend dynamiquement
func buildResendEmail(senderName string) string {
	// Nettoyer le nom (enlever espaces, caractères spéciaux)
	senderName = strings.TrimSpace(senderName)
	senderName = strings.ToLower(senderName)
	senderName = strings.ReplaceAll(senderName, " ", ".")

	// Si vide, utiliser "noreply" par défaut
	if senderName == "" {
		senderName = "noreply"
	}

	// Extraire le domaine de RESEND_FROM_EMAIL
	domain := config.AppConfig.ResendFromEmail
	if strings.Contains(domain, "@") {
		parts := strings.Split(domain, "@")
		domain = parts[1]
	}

	return fmt.Sprintf("%s@%s", senderName, domain)
}

// SendEmailWithProvider envoie un email via le provider choisi
func (s *EmailService) SendEmailWithProvider(to, subject, body, provider, senderName string) (string, error) {
	if provider == "mailgun" {
		senderEmail, err := s.sendWithMailgun(to, subject, body)
		return senderEmail, err
	}
	if provider == "resend" {
		// ✅ Construire l'email dynamiquement
		dynamicEmail := buildResendEmail(senderName)
		return dynamicEmail, s.sendWithResend(to, subject, body, dynamicEmail, senderName)
	}
	return "", fmt.Errorf("provider inconnu: %s", provider)
}

func (s *EmailService) sendWithMailgun(to, subject, body string) (string, error) {
	if config.AppConfig.MailgunDomain == "" || config.AppConfig.MailgunAPIKey == "" {
		return "", fmt.Errorf("mailgun not configured")
	}

	mg := mailgun.NewMailgun(config.AppConfig.MailgunDomain, config.AppConfig.MailgunAPIKey)

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

// ✅ MODIFIÉ: Accepter l'email et le nom dynamiques
func (s *EmailService) sendWithResend(to, subject, body, fromEmail, displayName string) error {
	if config.AppConfig.ResendAPIKey == "" {
		return fmt.Errorf("resend not configured")
	}

	client := resend.NewClient(config.AppConfig.ResendAPIKey)

	// Si pas de displayName, utiliser la partie avant le @
	if displayName == "" {
		if strings.Contains(fromEmail, "@") {
			displayName = strings.Split(fromEmail, "@")[0]
		}
	}

	// Formater avec le nom d'affichage
	fromAddress := fmt.Sprintf("%s <%s>", strings.Title(displayName), fromEmail)

	params := &resend.SendEmailRequest{
		From:    fromAddress,
		To:      []string{to},
		Subject: subject,
		Html:    body,
	}

	sent, err := client.Emails.Send(params)
	if err != nil {
		fmt.Printf("❌ Erreur envoi Resend à %s: %v\n", to, err)
		return err
	}

	fmt.Printf("✅ Email envoyé via Resend depuis %s → %s (ID: %s)\n", fromEmail, to, sent.Id)
	return nil
}

func (s *EmailService) ProcessEmails(req models.SendRequest, broadcast chan<- models.ProgressUpdate) {
	total := len(req.Emails)
	sent := 0
	failed := 0

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

	// Pour Resend, créer le sender UNE SEULE FOIS
	var globalSenderID int64
	if provider == "resend" {
		// ✅ Utiliser l'email dynamique
		dynamicEmail := buildResendEmail(req.SenderName)
		displayName := req.SenderName
		if displayName == "" {
			displayName = "AxSender"
		}

		globalSenderID, err = database.InsertOrGetSender(dynamicEmail, displayName)
		if err != nil {
			fmt.Printf("❌ Erreur création sender: %v\n", err)
			return
		}
		fmt.Printf("📧 Expéditeur Resend: %s <%s>\n", displayName, dynamicEmail)
	}

	concurrency := 10
	if provider == "mailgun" {
		concurrency = 50
	} else if provider == "resend" {
		concurrency = 1 // ✅ Un seul email à la fois pour éviter rate limit
	}

	semaphore := make(chan struct{}, concurrency)

	for i, emailData := range req.Emails {
		semaphore <- struct{}{}

		go func(index int, data models.EmailData) {
			defer func() { <-semaphore }()

			// Insérer/récupérer le recipient
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

			// Personnaliser le body
			body := strings.ReplaceAll(req.Body, "{{email}}", data.Email)

			// Envoyer l'email
			var senderEmail string
			var sendErr error
			var senderID int64

			if provider == "mailgun" {
				senderEmail, sendErr = s.sendWithMailgun(data.Email, req.Subject, body)

				displayName := "Admirateur Secret"
				senderID, err = database.InsertOrGetSender(senderEmail, displayName)
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
			} else if provider == "resend" {
				// ✅ Utiliser l'email dynamique
				senderEmail = buildResendEmail(req.SenderName)
				sendErr = s.sendWithResend(data.Email, req.Subject, body, senderEmail, req.SenderName)
				senderID = globalSenderID
			}

			// Déterminer le status
			status := "sent"
			errorMessage := ""

			if sendErr != nil {
				status = "failed"
				errorMessage = sendErr.Error()
				failed++
			} else {
				sent++
			}

			// Enregistrer dans la DB
			if err := database.InsertEmailSend(contentID, senderID, recipientID, status, errorMessage); err != nil {
				fmt.Printf("❌ Erreur enregistrement DB: %v\n", err)
			}

			// Broadcaster la progression
			broadcast <- models.ProgressUpdate{
				Current:    index + 1,
				Total:      total,
				Sent:       sent,
				Failed:     failed,
				Percentage: float64(index+1) / float64(total) * 100,
			}

			// Délai entre les envois
			delay := 500 * time.Millisecond
			if provider == "mailgun" {
				delay = 100 * time.Millisecond
			} else if provider == "resend" {
				delay = 1000 * time.Millisecond // ✅ 1 seconde entre chaque email
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

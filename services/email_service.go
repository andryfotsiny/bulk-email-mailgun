package services

import (
	"bulk-email-mailgun/config"
	"bulk-email-mailgun/models"
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"gopkg.in/gomail.v2"
)

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
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

	message := mg.NewMessage(
		config.AppConfig.Email,
		subject,
		"",
		to,
	)
	message.SetHtml(body)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, _, err := mg.Send(ctx, message)
	return err
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

	concurrency := 10
	if config.AppConfig.Provider == "mailgun" {
		concurrency = 50
	}

	semaphore := make(chan struct{}, concurrency)

	for i, emailData := range req.Emails {
		semaphore <- struct{}{}

		go func(index int, data models.EmailData) {
			defer func() { <-semaphore }()

			body := s.personalizeBody(req.Body, data)

			if err := s.SendEmail(data.Email, req.Subject, body); err != nil {
				failed++
			} else {
				sent++
			}

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

	fmt.Printf("Completed! Total: %d | Sent: %d | Failed: %d\n", total, sent, failed)
}

func (s *EmailService) personalizeBody(body string, data models.EmailData) string {
	body = strings.ReplaceAll(body, "{{name}}", data.Name)
	body = strings.ReplaceAll(body, "{{company}}", data.Company)
	body = strings.ReplaceAll(body, "{{city}}", data.City)
	body = strings.ReplaceAll(body, "{{email}}", data.Email)
	return body
}

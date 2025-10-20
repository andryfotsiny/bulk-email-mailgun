package config

import (
	"bulk-email-mailgun/models"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var AppConfig models.EmailConfig

func Init() {
	godotenv.Load()

	AppConfig.SMTPServer = getEnv("SMTP_SERVER", "smtp.gmail.com")
	AppConfig.SMTPPort, _ = strconv.Atoi(getEnv("SMTP_PORT", "465"))
	AppConfig.Email = getEnv("SENDER_EMAIL", "")
	AppConfig.Password = getEnv("SENDER_PASSWORD", "")
	AppConfig.Provider = getEnv("EMAIL_PROVIDER", "gmail")
	AppConfig.MailgunDomain = getEnv("MAILGUN_DOMAIN", "")
	AppConfig.MailgunAPIKey = getEnv("MAILGUN_API_KEY", "")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

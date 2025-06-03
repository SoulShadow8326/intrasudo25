package config

import (
	"os"
	"strings"
)

type EmailConfig struct {
	From     string
	Password string
	SMTPHost string
	SMTPPort string
}

type Config struct {
	Email        EmailConfig
	AdminEmails  []string
	XSecretValue string
}

func GetConfig() *Config {
	adminEmailsStr := os.Getenv("ADMIN_EMAILS")
	adminEmails := strings.Split(adminEmailsStr, ",")

	for i, email := range adminEmails {
		adminEmails[i] = strings.TrimSpace(email)
	}

	return &Config{
		Email: EmailConfig{
			From:     os.Getenv("EMAIL_FROM"),
			Password: os.Getenv("EMAIL_PASSWORD"),
			SMTPHost: os.Getenv("EMAIL_SMTP_HOST"),
			SMTPPort: os.Getenv("EMAIL_SMTP_PORT"),
		},
		AdminEmails:  adminEmails,
		XSecretValue: os.Getenv("X_SECRET_VALUE"),
	}
}

func GetAdminEmails() []string {
	return GetConfig().AdminEmails
}

func GetEmailConfig() EmailConfig {
	return GetConfig().Email
}

func GetXSecretValue() string {
	return os.Getenv("X_SECRET_VALUE")
}

func GetDiscordBotToken() string {
	return os.Getenv("DISCORD_BOT_TOKEN")
}

func GetDiscordBotURL() string {
	return os.Getenv("DISCORD_BOT_URL")
}

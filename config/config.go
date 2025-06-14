package config

import (
	"os"
	"strconv"
	"strings"
	"time"
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
	url := os.Getenv("DISCORD_BOT_URL")
	if url == "" {
		socketPath := os.Getenv("BOT_SOCKET_PATH")
		if socketPath == "" {
			socketPath = "/tmp/discord_bot.sock"
		}
		return socketPath
	}
	return url
}

func IsCountdownEnabled() bool {
	enableStr := os.Getenv("ENABLE_COUNTDOWN")
	if enableStr == "" {
		return false
	}
	enabled, err := strconv.ParseBool(enableStr)
	if err != nil {
		return false
	}
	return enabled
}

func GetCompetitionStartTime() time.Time {
	location, _ := time.LoadLocation("Asia/Kolkata")

	dateStr := os.Getenv("START_DATE")
	if dateStr == "" {
		dateStr = "2025-06-16"
	}

	timeStr := os.Getenv("START_TIME")
	if timeStr == "" {
		timeStr = "00:00:00"
	}

	datetimeStr := dateStr + "T" + timeStr
	startTime, err := time.ParseInLocation("2006-01-02T15:04:05", datetimeStr, location)
	if err != nil {
		return time.Date(2025, 6, 16, 0, 0, 0, 0, location)
	}

	return startTime
}

func GetCompetitionEndTime() time.Time {
	location, _ := time.LoadLocation("Asia/Kolkata")

	dateStr := os.Getenv("END_DATE")
	if dateStr == "" {
		dateStr = "2025-06-17"
	}

	timeStr := os.Getenv("END_TIME")
	if timeStr == "" {
		timeStr = "12:00:00"
	}

	datetimeStr := dateStr + "T" + timeStr
	endTime, err := time.ParseInLocation("2006-01-02T15:04:05", datetimeStr, location)
	if err != nil {
		return time.Date(2025, 6, 17, 12, 0, 0, 0, location)
	}

	return endTime
}
